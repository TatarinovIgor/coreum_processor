package service

import (
	"context"
	"fmt"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	amomultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

var (
	addressPrefix = constant.AddressPrefixTest
)

// FuncTrxIDVerification definition of a function to verify eligibility of transaction by trxID,
// if transaction is eligible the function must return true otherwise false
type FuncTrxIDVerification func(trxID string) bool

type MultiSignService struct {
	clientCtx         client.Context
	privateKey        map[string]types.PrivKey
	trxVerificationFn FuncTrxIDVerification
}

// NewMultiSignService create a new service to make a set of transaction signatures for a coreum multi sign accounts
//   - clientCtx - is a coreum client that used to extract public keys from  multi sign accounts
//   - fn - is a transaction verification function, that returns true if transaction verified and should be executed
//     otherwise return false and signature will not be created for the transaction
//   - mnemonics - a set of mnemonics to generate coreum keys for multi sign accounts
//
// the function panic in case it is not possible to create private keys from the provided mnemonics
func NewMultiSignService(clientCtx client.Context, fn FuncTrxIDVerification, mnemonics ...string) *MultiSignService {
	algo := hd.Secp256k1
	hdPath := sdk.GetConfig().GetFullBIP44Path()
	privateKey := map[string]types.PrivKey{}

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

	return &MultiSignService{clientCtx: clientCtx, privateKey: privateKey, trxVerificationFn: fn}
}

// GetMultiSignAddresses returns array of addresses that should be used to create multi sign accounts
func (s *MultiSignService) GetMultiSignAddresses() []string {
	var res []string
	for pubKey := range s.privateKey {
		res = append(res, pubKey)
	}
	return res
}

// MultiSignTransaction generate a map of signatures for each account used for multi sign account generation
//   - ctx - is a context for execution
//   - address - a multi sign account that request transaction signatures
//   - trxID - a transaction id that should be signed for execution
//   - trxData - byte serialise transaction that should be signed
//   - threshold - a minimum number of signatures required for transaction execution
//
// in case of success the result has a map of address used to generate signature and transaction signatures
func (s *MultiSignService) MultiSignTransaction(ctx context.Context, address, trxID string,
	trxData []byte, threshold int) (map[string][]byte, error) {

	// transaction verification if applicable
	if s.trxVerificationFn != nil {
		if !s.trxVerificationFn(trxID) {
			return nil, fmt.Errorf("transaction: %s, is not verified", trxID)
		}
	}

	res := map[string][]byte{}
	numSign := 0

	accAddress, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}

	info, err := client.GetAccountInfo(ctx, s.clientCtx, accAddress)
	if err != nil {
		return nil, err
	}

	pubKey, ok := info.GetPubKey().(*amomultisig.LegacyAminoPubKey)
	if !ok {
		return nil, fmt.Errorf("unacceptable key format for address: %s", address)
	}

	for _, publicKey := range pubKey.GetPubKeys() {
		addr, err := bech32.ConvertAndEncode(addressPrefix, publicKey.Address())
		if err != nil {
			continue
		}

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

		res[addr] = signature
		numSign++
		if numSign >= threshold {
			// got enough signatures
			break
		}
	}

	if numSign < threshold {
		return nil, fmt.Errorf("can't get enough signer for address: %s", address)
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
