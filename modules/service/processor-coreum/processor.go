package processor_coreum

import (
	"context"
	"coreum_processor/modules/service"
	"coreum_processor/modules/storage"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
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
	"strings"
)

const (
	coreumFeeMessage  = 50000
	coreumFeeIssueBug = 10000000
	coreumFeeIssueFT  = 70000 + coreumFeeMessage + coreumFeeIssueBug
	coerumFeeMintFT   = 11000 + coreumFeeMessage
	coreumFeeBurnFT   = 23000 + coreumFeeMessage
	coreumFeeSendFT   = 16000 + coreumFeeMessage
	coreumFeeIssueNFT = 16000
	coreumFeeMintNFT  = 39000
	coreumDecimals    = 1000000
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
	senderMnemonic  string
	denom           string
}

func NewCoreumCryptoProcessor(sendingWallet, receivingWallet service.Wallet,
	blockchain string, store *storage.KeysPSQL, minValue float64,
	chainID constant.ChainID, nodeAddress, addressPrefix, senderMnemonic string) service.CryptoProcessor {

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
		senderMnemonic:  senderMnemonic,
		denom:           constant.DenomTest, // todo change depending on env
	}
}

// GetTransactionStatus returns transaction status from the blockchain
func (s CoreumProcessing) GetTransactionStatus(_ context.Context, hash string) (service.CryptoTransactionStatus, error) {
	//todo
	return service.SuccessfulTransaction, nil
}

func (s CoreumProcessing) TransferToReceiving(ctx context.Context, request service.TransferRequest,
	merchantID, externalId string) (*service.TransferResponse, error) {
	_, _, userWallet, err := s.store.GetByUser(merchantID, externalId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	sendingWallet := service.Wallet{}
	err = json.Unmarshal(userWallet, &sendingWallet)
	//check gas
	_, err = s.updateGas(ctx, sendingWallet.WalletAddress, coreumFeeSendFT)
	if err != nil {
		return nil, err
	}
	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		sendingWallet.WalletAddress,
		string(sendingWallet.WalletSeed),
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()
	msg := &banktypes.MsgSend{
		FromAddress: sendingWallet.WalletAddress,
		ToAddress:   s.receivingWallet.WalletAddress,
		Amount: sdk.NewCoins(sdk.NewInt64Coin(fmt.Sprintf("%s-%s", request.Asset, request.Issuer),
			int64(request.Amount))),
	}
	bech32, err := sdk.AccAddressFromBech32(sendingWallet.WalletAddress)
	if err != nil {
		return nil, err
	}
	result, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(bech32),
		s.factory,
		msg,
	)
	if err != nil {
		return nil, err
	}
	return &service.TransferResponse{TransferHash: result.TxHash}, nil
}

func (s CoreumProcessing) TransferFromReceiving(ctx context.Context, request service.TransferRequest,
	merchantID, externalId string) (*service.TransferResponse, error) {
	if request.Amount < s.minimumValue {
		return nil, fmt.Errorf("transaction amount is to small to be received")
	}

	_, _, userWallet, err := s.store.GetByUser(merchantID, externalId)
	if err != nil {
		return nil, err
	}
	clientWallet := service.Wallet{}
	err = json.Unmarshal(userWallet, &clientWallet)
	if err != nil {
		return nil, err
	}
	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		s.receivingWallet.WalletAddress,
		string(s.receivingWallet.WalletSeed),
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()
	//check gas
	_, err = s.updateGas(ctx, s.receivingWallet.WalletAddress, coreumFeeSendFT)
	if err != nil {
		return nil, err
	}
	msg := &banktypes.MsgSend{
		FromAddress: s.receivingWallet.WalletAddress,
		ToAddress:   clientWallet.WalletAddress,
		Amount: sdk.NewCoins(sdk.NewInt64Coin(fmt.Sprintf("%s-%s", request.Asset, request.Issuer),
			int64(request.Amount))),
	}
	bech32, err := sdk.AccAddressFromBech32(s.receivingWallet.WalletAddress)
	if err != nil {
		return nil, err
	}
	result, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(bech32),
		s.factory,
		msg,
	)
	if err != nil {
		return nil, err
	}
	return &service.TransferResponse{TransferHash: result.TxHash}, nil
}

func (s CoreumProcessing) TransferBetweenMerchantWallets(ctx context.Context,
	request service.TransferRequest, merchantID string) (*service.TransferResponse, error) {

	_, _, userWallet, err := s.store.GetByUser(merchantID, merchantID+"-R")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	receivingWallet := service.Wallet{}
	err = json.Unmarshal(userWallet, &receivingWallet)

	_, _, userWallet, err = s.store.GetByUser(merchantID, merchantID+"-S")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	sendingWallet := service.Wallet{}
	err = json.Unmarshal(userWallet, &sendingWallet)
	//check gas

	_, err = s.updateGas(ctx, receivingWallet.WalletAddress, coreumFeeSendFT)
	if err != nil {
		return nil, err
	}

	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		receivingWallet.WalletAddress,
		string(receivingWallet.WalletSeed),
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	if err != nil {
		return nil, err
	}

	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()
	msg := &banktypes.MsgSend{
		FromAddress: receivingWallet.WalletAddress,
		ToAddress:   sendingWallet.WalletAddress,
		Amount: sdk.NewCoins(sdk.NewInt64Coin(fmt.Sprintf("%s-%s", request.Asset, request.Issuer),
			int64(request.Amount))),
	}

	bech32, err := sdk.AccAddressFromBech32(receivingWallet.WalletAddress)
	if err != nil {
		return nil, err
	}
	result, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(bech32),
		s.factory,
		msg,
	)
	if err != nil {
		return nil, err
	}

	return &service.TransferResponse{TransferHash: result.TxHash}, nil
}

func (s CoreumProcessing) TransferFromSending(ctx context.Context, request service.TransferRequest,
	merchantID, receivingWallet string) (*service.TransferResponse, error) {
	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		s.sendingWallet.WalletAddress,
		string(s.sendingWallet.WalletSeed),
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()

	msg := &banktypes.MsgSend{
		FromAddress: s.sendingWallet.WalletAddress,
		ToAddress:   receivingWallet,
		Amount: sdk.NewCoins(sdk.NewInt64Coin(fmt.Sprintf("%s-%s", request.Asset, request.Issuer),
			int64(request.Amount))),
	}
	bech32, err := sdk.AccAddressFromBech32(s.sendingWallet.WalletAddress)
	if err != nil {
		return nil, err
	}
	result, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(bech32),
		s.factory,
		msg,
	)
	if err != nil {
		return nil, err
	}
	return &service.TransferResponse{TransferHash: result.TxHash}, nil
}

func (s CoreumProcessing) GetTokenSupply(ctx context.Context, request service.BalanceRequest) (int64, error) {

	denom := request.Asset + "-" + request.Issuer

	bankClient := banktypes.NewQueryClient(s.clientCtx)
	// Query the balance of the recipient
	response, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{
		Denom: denom,
	})
	if err != nil {
		return 0, err
	}
	return response.Amount.Amount.Int64(), nil
}

func (s CoreumProcessing) GetBalance(ctx context.Context, merchantID, externalID string) (service.Balance, error) {
	_, _, byteAddress, err := s.store.GetByUser(merchantID, externalID)
	balance := service.Balance{
		Amount:     0,
		Blockchain: "coreum",
		Asset:      constant.DenomTest,
		Issuer:     "",
	}
	if err != nil {
		return balance, fmt.Errorf("can't get user: %v coreum wallet from store, err: %v", externalID, err)
	}

	userWallet := service.Wallet{}
	err = json.Unmarshal(byteAddress, &userWallet)
	if err != nil {
		return balance, err
	}

	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		userWallet.WalletAddress,
		userWallet.WalletSeed,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	if err != nil {
		return balance, err
	}
	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()

	amount, _, err := s.balanceCoreum(ctx, userWallet.WalletAddress, constant.DenomTest)
	balance.Amount = float64(amount) / coreumDecimals

	if err != nil {
		return balance, err
	}

	return balance, nil
}

func (s CoreumProcessing) GetAssetsBalance(ctx context.Context,
	request service.BalanceRequest, merchantID, externalId string) ([]service.Balance, error) {
	_, _, byteAddress, err := s.store.GetByUser(merchantID, externalId)
	if err != nil {
		return nil, fmt.Errorf("can't get user: %v coreum wallet from store, err: %v", externalId, err)
	}
	userWallet := service.Wallet{}
	err = json.Unmarshal(byteAddress, &userWallet)
	if err != nil {
		return nil, err
	}
	// Query initial balance hold by the issuer
	denom := s.denom
	bankClient := banktypes.NewQueryClient(s.clientCtx)
	var balances []service.Balance
	//Check whether request wants specific token or all of them
	if request.Asset == "" {
		resp, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
			Address: userWallet.WalletAddress,
		})
		if err != nil {
			return []service.Balance{}, err
		}
		for i := 0; i < resp.Balances.Len(); i++ {
			balanceDenom := resp.Balances[i].Denom
			if balanceDenom == s.denom {
				continue
			}
			asset := strings.Split(balanceDenom, "-")
			balances = append(balances, service.Balance{Blockchain: request.Blockchain,
				Amount: float64(resp.Balances[i].Amount.Int64()),
				Asset:  asset[0], Issuer: asset[1]})
		}
		return balances, nil
	} else if request.Asset != denom {
		denom = request.Asset + "-" + request.Issuer
	}
	resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: userWallet.WalletAddress,
		Denom:   denom,
	})
	if err != nil {
		return nil, err
	}
	balances = append(balances, service.Balance{Blockchain: request.Blockchain,
		Amount: float64(resp.Balance.Amount.Int64()),
		Asset:  request.Asset, Issuer: request.Issuer})

	return balances, nil

}

func (s CoreumProcessing) GetWalletById(merchantID, externalId string) (string, error) {
	_, _, walletByte, err := s.store.GetByUser(merchantID, externalId)
	wallet := service.Wallet{Blockchain: s.receivingWallet.Blockchain}
	err = json.Unmarshal(walletByte, &wallet)
	if err != nil {
		return "", err
	}
	return wallet.WalletAddress, nil
}

func (s CoreumProcessing) updateGas(ctx context.Context, address string, txGasPrice int64) (string, error) {

	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		s.sendingWallet.WalletAddress,
		string(s.sendingWallet.WalletSeed),
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)

	if err != nil {
		return "", err
	}
	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()

	trx, err := s.transferCoreumFT(ctx, senderInfo.GetAddress().String(), address, s.denom, txGasPrice)

	return trx, err
}

func (s CoreumProcessing) balanceCoreum(ctx context.Context, userAddress, denom string) (int, string, error) {

	address, err := sdk.AccAddressFromBech32(userAddress)
	if err != nil {
		return 0, "", err
	}
	Info, err := s.clientCtx.Keyring().KeyByAddress(address)
	if err != nil {
		return 0, "", err
	}

	bankClient := banktypes.NewQueryClient(s.clientCtx)
	// Query the balance of the recipient
	response, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: Info.GetAddress().String(),
		Denom:   denom,
	})
	if err != nil {
		return 0, denom, err
	}
	return int(float64(response.Balance.Amount.Uint64())), response.Balance.Denom, nil
}

func (s CoreumProcessing) balanceCoreumTokens(ctx context.Context, userAddress, subunit string) (int, string, error) {

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
		return 0, "", err
	}
	return int(float64(response.Balance.Amount.Uint64())), response.Balance.Denom, nil
}
