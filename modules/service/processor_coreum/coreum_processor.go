package processor_coreum

import (
	"context"
	"crypto/tls"
	"crypto_processor/modules/service"
	"crypto_processor/modules/storage"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"time"
)

const (
	senderMnemonic   = "unit resource ramp note attitude allow pipe hollow above kingdom siren social bless crystal student appear today orchard drive prosper during report burden film" // put mnemonic here
	chainID          = constant.ChainIDTest
	addressPrefix    = constant.AddressPrefixTest
	nodeAddress      = "full-node.testnet-1.coreum.dev:9090"
	denom            = constant.DenomTest
	recipientAddress = "testcore1534s8rz2e36lwycr6gkm9vpfe5yf67wkuca7zs"
)

type CoreumProcessing struct {
	blockchain      string
	client          *grpc.ClientConn
	config          *sdk.Config
	factory         tx.Factory
	clientCtx       client.Context
	smartContract   service.SmartContract
	sendingWallet   service.Wallet
	receivingWallet service.Wallet
	store           *storage.KeysPSQL
	apiURL          string
	minimumValue    float64
}

// GetTransactionStatus returns transaction status from the blockchain
func (s CoreumProcessing) GetTransactionStatus(hash string) (service.CryptoTransactionStatus, error) {
	//todo
	return service.SuccessfulTransaction, nil
}

func (s CoreumProcessing) Deposit(request service.CredentialDeposit, merchantID, externalId string) (*service.DepositResponse, error) {
	depositData := service.DepositResponse{}
	wallet := service.Wallet{Blockchain: s.receivingWallet.Blockchain}
	_, walletByte, err := s.store.GetByUser(merchantID, externalId)
	if err != nil && errors.Is(err, storage.ErrNotFound) {
		wallet.WalletSeed, wallet.WalletAddress, err = s.createCoreumWallet()
		if err != nil {
			return nil, err
		}
		wallet.Blockchain = request.Blockchain
		key, err := json.Marshal(wallet)

		if err != nil {
			return nil, err
		}
		_, err = s.store.Put(merchantID, externalId, wallet.WalletAddress, key)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		err = json.Unmarshal(walletByte, &wallet)
		if err != nil {
			return nil, err
		}
	}
	depositData.WalletAddress = wallet.WalletAddress
	depositData.Memo = ""

	return &depositData, nil

}

func (s CoreumProcessing) Withdraw(request service.CredentialWithdraw, merchantID, externalId string, merchantWallets service.Wallets) (*service.WithdrawResponse, error) {
	return &service.WithdrawResponse{}, nil
}

func (s CoreumProcessing) TransferToReceiving(request service.TransferRequest, merchantID, externalId string) (*service.TransferResponse, error) {
	return &service.TransferResponse{}, nil
}

func (s CoreumProcessing) TransferFromReceiving(request service.TransferRequest, merchantID, externalId string) (*service.TransferResponse, error) {
	return &service.TransferResponse{}, nil
}

func (s CoreumProcessing) TransferFromSending(request service.TransferRequest, merchantID, receivingWallet string) (*service.TransferResponse, error) {
	return &service.TransferResponse{}, nil
}

func (s CoreumProcessing) GetBalance(request service.BalanceRequest, merchantID, externalId string) (*service.Balance, error) {
	// Query initial balance hold by the issuer
	ctx := context.Background()
	hexAddress := ""
	address, err := sdk.AccAddressFromHex(hexAddress)
	if err != nil {
		return nil, err
	}
	senderInfo, _ := s.clientCtx.Keyring().KeyByAddress(address)
	denom := request.Blockchain + "-" + senderInfo.GetAddress().String()
	bankClient := banktypes.NewQueryClient(s.clientCtx)
	resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: senderInfo.GetAddress().String(),
		Denom:   denom,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Issuer's balance: %s\n", resp.Balance)
	return &service.Balance{Blockchain: request.Blockchain, Amount: float64(resp.Balance.Amount.Uint64()), Asset: resp.Balance.Denom, Issuer: ""}, nil
}

func (s CoreumProcessing) GetWalletById(merchantID, externalId string) (string, error) {
	_, walletByte, err := s.store.GetByUser(merchantID, externalId)
	wallet := service.Wallet{Blockchain: s.receivingWallet.Blockchain}
	err = json.Unmarshal(walletByte, &wallet)
	if err != nil {
		return "", err
	}
	return wallet.WalletAddress, nil
}

func (s CoreumProcessing) StreamDeposit(ctx context.Context, callback service.FuncDepositCallback) {
	go s.streamDeposit(ctx, callback)
	return
}

func (s CoreumProcessing) streamDeposit(ctx context.Context, callback service.FuncDepositCallback) {
	ticker := time.NewTicker(time.Second)
	next := int64(0)
	for {
		select {
		case <-ctx.Done():
			log.Println("exit from ethereum processor deposit stream")
			return
		case <-ticker.C:
			records, err := s.store.GetNext(next, 1)
			if err != nil || len(records) == 0 {
				next = 0
				if err != nil {
					log.Println("Error while getting DB records:", err)
				}
			} else {
				record := records[len(records)-1]
				balance, err := s.GetBalance(service.BalanceRequest{Blockchain: s.blockchain}, record.MerchantID, record.ExternalID)
				if balance != nil && err == nil && balance.Amount > 0 {
					callback(balance.Blockchain, record.MerchantID, record.ExternalID, record.Key, "",
						balance.Asset, balance.Issuer, balance.Amount)
				}
				if err != nil {
					fmt.Println(err)
				}
				next = record.ID
			}
		}
	}
}

func NewCoreumCryptoProcessor(sendingWallet, receivingWallet service.Wallet,
	blockchain string, store *storage.KeysPSQL, minValue float64) service.CryptoProcessor {

	// Configure Cosmos SDK
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

	txFactory := client.Factory{}.
		WithKeybase(clientCtx.Keyring()).
		WithChainID(clientCtx.ChainID()).
		WithTxConfig(clientCtx.TxConfig()).
		WithSimulateAndExecute(true)

	return &CoreumProcessing{
		blockchain:      blockchain,
		client:          grpcClient,
		clientCtx:       clientCtx,
		factory:         txFactory,
		config:          config,
		sendingWallet:   sendingWallet,
		receivingWallet: receivingWallet,
		store:           store,
		minimumValue:    minValue,
	}
}

func (s CoreumProcessing) createCoreumWallet() (string, string, error) {

	Info, err := s.clientCtx.Keyring().NewAccount(
		"key-name",
		senderMnemonic,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	if err != nil {
		panic(err)
	}

	return senderMnemonic, Info.GetAddress().String(), nil
}

func (s CoreumProcessing) createCoreumToken(symbol, subunit, issuerAddress string) (string, error) {
	msgIssue := &assetfttypes.MsgIssue{
		Issuer:        issuerAddress,
		Symbol:        symbol,
		Subunit:       subunit,
		Precision:     6,
		InitialAmount: sdk.NewInt(100_000_000),
		Description:   "Make me Unique!",
		Features:      []assetfttypes.Feature{assetfttypes.Feature_freezing},
	}

	address, err := sdk.AccAddressFromBech32(issuerAddress)
	if err != nil {
		return "", err
	}
	senderInfo, err := s.clientCtx.Keyring().KeyByAddress(address)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	response, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(senderInfo.GetAddress()),
		s.factory,
		msgIssue,
	)
	if err != nil {
		panic(err)
	}
	return response.TxHash, err
}

func (s CoreumProcessing) transferCoreumTokens(senderAddress, recipientAddress, subunit string) (string, error) {
	ctx := context.Background()

	address, err := sdk.AccAddressFromBech32(senderAddress)
	if err != nil {
		return "", err
	}
	senderInfo, err := s.clientCtx.Keyring().KeyByAddress(address)
	if err != nil {
		return "", err
	}

	denom := subunit + "-" + senderInfo.GetAddress().String()

	msgSend := &banktypes.MsgSend{
		FromAddress: senderAddress,
		ToAddress:   recipientAddress,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom, 1_000_000)),
	}
	response, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(senderInfo.GetAddress()),
		s.factory,
		msgSend,
	)
	if err != nil {
		panic(err)
	}
	return response.TxHash, nil
}

func (s CoreumProcessing) balanceCoreumTokens(userAddress, subunit string) (int, string, error) {
	ctx := context.Background()

	address, err := sdk.AccAddressFromBech32(userAddress)
	if err != nil {
		return 0, "", err
	}
	Info, err := s.clientCtx.Keyring().KeyByAddress(address)
	if err != nil {
		return 0, "", err
	}

	denom := subunit + "-" + Info.GetAddress().String()

	bankClient := banktypes.NewQueryClient(s.clientCtx)
	// Query the balance of the recipient
	response, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: Info.GetAddress().String(),
		Denom:   denom,
	})
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	return int(float64(response.Balance.Amount.Uint64())), response.Balance.Denom, nil
}
