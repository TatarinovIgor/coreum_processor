package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/dvsekhvalnov/jose2go/base64url"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"strings"
)

const amount = 102400

const signMode = signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON

// FuncTrxIDVerification definition of a function to verify eligibility of transaction by trxID,
// if transaction is eligible the function must return true otherwise false
type FuncTrxIDVerification func(trxID string) bool

type MultiSignService struct {
	clientCtx         client.Context
	privateKey        map[string]types.PrivKey
	trxVerificationFn FuncTrxIDVerification
	addressPrefix     string
	txFactory         client.Factory
}

// NewMultiSignService create a new service to make a set of transaction signatures for a coreum multi sign accounts
//   - clientCtx - is a coreum client that used to extract public keys from  multi sign accounts
//   - fn - is a transaction verification function, that returns true if transaction verified and should be executed
//     otherwise return false and signature will not be created for the transaction
//   - networkType - is a string that defines type of blockchain network can be ['devnet','testnet','mainnet']
//   - mnemonics - a set of mnemonics to generate coreum keys for multi sign accounts
//
// the function panic in case it is not possible to create private keys from the provided mnemonics
func NewMultiSignService(ctx context.Context, fn FuncTrxIDVerification,
	networkType string, mnemonics ...string) *MultiSignService {
	algo := hd.Secp256k1
	hdPath := sdk.GetConfig().GetFullBIP44Path()
	privateKey := map[string]types.PrivKey{}
	addressPrefix := ""
	var chainID constant.ChainID
	nodeAddress := "full-node.testnet-1.coreum.dev:9090"
	networkType = strings.ToLower(networkType)
	switch networkType {
	case "devnet":
		addressPrefix = constant.AddressPrefixDev
		chainID = constant.ChainIDDev
	case "testnet":
		addressPrefix = constant.AddressPrefixTest
		chainID = constant.ChainIDTest
	case "mainnet":
		addressPrefix = constant.AddressPrefixMain
		chainID = constant.ChainIDMain
	default:
		panic("unsupported type of blockchain network type " + networkType)
	}
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(addressPrefix, addressPrefix+"pub")
	config.SetCoinType(constant.CoinType)
	config.Seal()
	// List required modules.
	// If you need types from any other module import them and add here.
	modules := module.NewBasicManager(
		auth.AppModuleBasic{},
	)

	// Configure client context and tx factory
	// If you don't use TLS then replace `grpc.WithTransportCredentials()` with `grpc.WithInsecure()`
	grpcClient, err := grpc.Dial(nodeAddress, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		panic(err)
	}

	clientCtx := client.NewContext(client.DefaultContextConfig(), modules).
		WithChainID(string(chainID)).
		WithGRPCClient(grpcClient).
		WithKeyring(keyring.NewInMemory()).
		WithBroadcastMode(flags.BroadcastBlock)

	for _, m := range mnemonics {
		// create master key and derive first key
		derivedPrivate, err := algo.Derive()(m, "", hdPath)
		if err != nil {
			panic(err)
		}

		pKey := algo.Generate()(derivedPrivate)
		address, err := bech32.ConvertAndEncode(addressPrefix, pKey.PubKey().Address())
		if err != nil {
			panic(err)
		}
		privateKey[address] = pKey
	}

	txFactory := client.Factory{}.
		WithKeybase(clientCtx.Keyring()).
		WithChainID(clientCtx.ChainID()).
		WithTxConfig(clientCtx.TxConfig()).
		WithSignMode(signMode).
		WithGas(amount).
		WithSimulateAndExecute(true)

	return &MultiSignService{clientCtx: clientCtx, privateKey: privateKey,
		trxVerificationFn: fn, addressPrefix: addressPrefix, txFactory: txFactory}
}

// GetMultiSignAddresses returns map of addresses and their weight that should be used to create multi sign accounts
func (s *MultiSignService) GetMultiSignAddresses(blockchain, externalID string) map[string]float64 {
	res := map[string]float64{}
	msg := fmt.Sprintf("On blockchain: %s\n\tfor external id: %s\n\tGiven the following addresses:",
		blockchain, externalID)
	for adr, private := range s.privateKey {
		m, _ := private.PubKey().(*secp256k1.PubKey).Marshal()
		pubKey := base64url.Encode(m)
		res[pubKey] = 1.0
		msg = fmt.Sprintf("%s\n\t %v\n", msg, adr)
	}
	log.Println(msg)
	return res
}

// MultiSignTransaction generate a map of signatures for each account used for multi sign account generation
//   - ctx - is a context for execution
//   - trxID - a transaction id that should be signed for execution
//   - addresses - a multi sign addresses that requested for transaction signatures
//   - trxData - byte serialise transaction that should be signed
//   - threshold - a minimum number of signatures required for transaction execution
//
// in case of success the result has a map of address used to generate signature and transaction signatures
func (s *MultiSignService) MultiSignTransaction(ctx context.Context, trxID string, addresses []string,
	trxData []byte, threshold int) (map[string][]byte, error) {

	// transaction verification if applicable
	if s.trxVerificationFn != nil {
		if !s.trxVerificationFn(trxID) {
			return nil, fmt.Errorf("transaction: %s, is not verified", trxID)
		}
	}

	res := map[string][]byte{}
	numSign := 0

	for _, addr := range addresses {

		privateKey, err := s.findPrivateKeyByAddress(addr)
		if err != nil {
			continue
		} else if privateKey == nil {
			continue
		}

		signature, err := privateKey.Sign(trxData)
		if err != nil {
			continue
		}
		m, _ := privateKey.PubKey().(*secp256k1.PubKey).Marshal()
		pubKey := base64url.Encode(m)

		res[pubKey] = signature
		numSign++
		if numSign >= threshold {
			// got enough signatures
			break
		}
	}

	if numSign < threshold {
		return nil, fmt.Errorf("can't get enough signer for trxID: %s", trxID)
	}

	return res, nil
}

func (s *MultiSignService) findPrivateKeyByAddress(address string) (types.PrivKey, error) {
	privateKey, ok := s.privateKey[address]
	if ok {
		return privateKey, nil
	}
	return nil, fmt.Errorf("can't find private key for address: %s", address)
}
