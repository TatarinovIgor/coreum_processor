package processor_coreum

import (
	"context"
	"coreum_processor/modules/service"
	"coreum_processor/modules/storage"
	"crypto/tls"
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
	"github.com/cosmos/go-bip39"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"strconv"
	"time"
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
}

func (s CoreumProcessing) MintToken(request service.TokenRequest, merchantID, externalID string) (*service.NewTokenResponse, error) {
	_, byteAddress, err := s.store.GetByUser(merchantID, externalID)
	if err != nil {
		return nil, fmt.Errorf("can't get user: %v eth wallet from store, err: %v", externalID, err)
	}
	userWallet := service.Wallet{}
	err = json.Unmarshal(byteAddress, &userWallet)
	if err != nil {
		return nil, err
	}
	if userWallet.WalletAddress == "" {
		return nil, fmt.Errorf("empty wallet address")
	}
	amount, err := strconv.Atoi(request.Amount)
	if err != nil {
		return nil, err
	}
	token, err := s.mintCoreumToken(request.Code, request.Issuer, userWallet.WalletSeed, int64(amount))
	if err != nil {
		return nil, err
	}
	return &service.NewTokenResponse{TxHash: token}, nil
}

func (s CoreumProcessing) BurnToken(request service.TokenRequest, merchantID, externalID string) (*service.NewTokenResponse, error) {
	_, byteAddress, err := s.store.GetByUser(merchantID, externalID)
	if err != nil {
		return nil, fmt.Errorf("can't get user: %v eth wallet from store, err: %v", externalID, err)
	}
	userWallet := service.Wallet{}
	err = json.Unmarshal(byteAddress, &userWallet)
	if err != nil {
		return nil, err
	}
	if userWallet.WalletAddress == "" {
		return nil, fmt.Errorf("empty wallet address")
	}
	amount, err := strconv.Atoi(request.Amount)
	if err != nil {
		return nil, err
	}
	token, err := s.burnCoreumToken(request.Code, request.Issuer, userWallet.WalletSeed, int64(amount))
	if err != nil {
		return nil, err
	}
	return &service.NewTokenResponse{TxHash: token}, nil
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

// ToDo check token
func (s CoreumProcessing) Withdraw(request service.CredentialWithdraw, merchantID, externalId string, merchantWallets service.Wallets) (*service.WithdrawResponse, error) {
	ctx := context.Background()
	commission := 0.0
	if externalId != merchantWallets.ReceivingID || externalId != merchantWallets.SendingID {
		commission = merchantWallets.CommissionSending.Fix
		commission += merchantWallets.CommissionSending.Percent / 100. * (request.Amount - commission)
	}

	balance, err := s.GetBalance(service.BalanceRequest{}, merchantID, merchantWallets.SendingID)
	if balance.Amount < request.Amount+commission {
		return nil, fmt.Errorf("merchant: %s, doesn't have enough balance to pay: %v, with commission: %v",
			merchantID, request.Amount, commission)
	}
	if err != nil {
		return nil, fmt.Errorf("can't get merchant: %v, sending wallet: %v, err: %w",
			merchantID, merchantWallets.SendingID, err)
	}

	_, sendingWalletRaw, err := s.store.GetByUser(merchantID, merchantWallets.SendingID)
	if err != nil {
		return nil, err
	}

	sendingWallet := service.Wallet{}
	err = json.Unmarshal(sendingWalletRaw, &sendingWallet)

	msg := &banktypes.MsgSend{
		FromAddress: sendingWallet.WalletAddress,
		ToAddress:   request.WalletAddress,
		//ToDo change denom in production
		Amount: sdk.NewCoins(sdk.NewInt64Coin(constant.DenomTest, int64(request.Amount))),
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

	return &service.WithdrawResponse{TransactionHash: result.TxHash}, nil
}

// ToDo check token
func (s CoreumProcessing) TransferToReceiving(request service.TransferRequest, merchantID, externalId string) (*service.TransferResponse, error) {
	ctx := context.Background()
	_, userWallet, err := s.store.GetByUser(merchantID, externalId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	sendingWallet := service.Wallet{}
	err = json.Unmarshal(userWallet, &sendingWallet)

	msg := &banktypes.MsgSend{
		FromAddress: sendingWallet.WalletAddress,
		ToAddress:   s.receivingWallet.WalletAddress,
		//ToDo change denom in production
		Amount: sdk.NewCoins(sdk.NewInt64Coin(constant.DenomTest, int64(request.Amount))),
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
	return &service.TransferResponse{TransferHash: result.TxHash}, nil
}

// ToDo check token
func (s CoreumProcessing) TransferFromReceiving(request service.TransferRequest, merchantID, externalId string) (*service.TransferResponse, error) {
	ctx := context.Background()
	if request.Amount < s.minimumValue {
		return nil, fmt.Errorf("transaction amount is to small to be recived")
	}

	_, userWallet, err := s.store.GetByUser(merchantID, externalId)
	if err != nil {
		return nil, err
	}
	clientWallet := service.Wallet{}
	err = json.Unmarshal(userWallet, &clientWallet)

	msg := &banktypes.MsgSend{
		FromAddress: s.receivingWallet.WalletAddress,
		ToAddress:   clientWallet.WalletAddress,
		//ToDo change denom in production
		Amount: sdk.NewCoins(sdk.NewInt64Coin(constant.DenomTest, int64(request.Amount))),
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
	return &service.TransferResponse{TransferHash: result.TxHash}, nil
}

// ToDo check token
func (s CoreumProcessing) TransferFromSending(request service.TransferRequest, merchantID, receivingWallet string) (*service.TransferResponse, error) {
	ctx := context.Background()
	msg := &banktypes.MsgSend{
		FromAddress: s.receivingWallet.WalletAddress,
		ToAddress:   receivingWallet,
		//ToDo change denom in production
		Amount: sdk.NewCoins(sdk.NewInt64Coin(constant.DenomTest, int64(request.Amount))),
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
	return &service.TransferResponse{TransferHash: result.TxHash}, nil
}

func (s CoreumProcessing) IssueToken(request service.NewTokenRequest, merchantID, externalId string) (*service.NewTokenResponse, error) {
	_, byteAddress, err := s.store.GetByUser(merchantID, externalId)
	if err != nil {
		return nil, fmt.Errorf("can't get user: %v eth wallet from store, err: %v", externalId, err)
	}
	userWallet := service.Wallet{}
	err = json.Unmarshal(byteAddress, &userWallet)
	if err != nil {
		return nil, err
	}
	if userWallet.WalletAddress == "" {
		return nil, fmt.Errorf("empty wallet address")
	}
	token, err := s.createCoreumToken(request.Symbol, request.Code, request.Issuer, request.Description, userWallet.WalletSeed)
	if err != nil {
		return nil, err
	}
	return &service.NewTokenResponse{TxHash: token}, nil
}

func (s CoreumProcessing) GetBalance(request service.BalanceRequest, merchantID, externalId string) (*service.Balance, error) {
	_, byteAddress, err := s.store.GetByUser(merchantID, externalId)
	if err != nil {
		return nil, fmt.Errorf("can't get user: %v eth wallet from store, err: %v", externalId, err)
	}
	userWallet := service.Wallet{}
	err = json.Unmarshal(byteAddress, &userWallet)
	if err != nil {
		return nil, err
	}
	// Query initial balance hold by the issuer
	ctx := context.Background()
	denom := request.Asset + "-" + userWallet.WalletAddress
	bankClient := banktypes.NewQueryClient(s.clientCtx)
	resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: userWallet.WalletAddress,
		Denom:   denom,
	})
	if err != nil {
		panic(err)
	}
	log.Println(fmt.Sprintf("Issuer's balance: %s\n", resp.Balance))
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

func (s CoreumProcessing) StreamDeposit(ctx context.Context, callback service.FuncDepositCallback,
	interval time.Duration) {
	go s.streamDeposit(ctx, callback, interval)
	return
}

func (s CoreumProcessing) streamDeposit(ctx context.Context, callback service.FuncDepositCallback,
	interval time.Duration) {
	ticker := time.NewTicker(time.Second * interval)
	next := int64(0)
	for {
		select {
		case <-ctx.Done():
			log.Println("exit from coreum processor deposit stream")
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
				balance, err := s.GetBalance(service.BalanceRequest{Blockchain: s.blockchain, Asset: "coreum"}, record.MerchantID, record.ExternalID)
				if balance != nil && err == nil && balance.Amount > 0 {
					callback(balance.Blockchain, record.MerchantID, record.ExternalID, record.Key, "",
						balance.Asset, balance.Issuer, balance.Amount)
				}
				if err != nil {
					log.Println(err)
				}
				next = record.ID
			}
		}
	}
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
	}
}

func (s CoreumProcessing) createCoreumWallet() (string, string, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", "", err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", "", err
	}
	Info, err := s.clientCtx.Keyring().NewAccount(
		"key-name",
		mnemonic,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	if err != nil {
		panic(err)
	}

	return mnemonic, Info.GetAddress().String(), nil
}

func (s CoreumProcessing) createCoreumToken(symbol, subunit, issuerAddress, description, mnemonic string) (string, error) {
	msgIssue := &assetfttypes.MsgIssue{
		Issuer:        issuerAddress,
		Symbol:        symbol,
		Subunit:       subunit,
		Precision:     6,
		InitialAmount: sdk.NewInt(100_000_000),
		Description:   description,
		Features:      []assetfttypes.Feature{assetfttypes.Feature_minting, assetfttypes.Feature_burning},
	}
	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		"key-name",
		mnemonic,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	if err != nil {
		_ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress())
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
		_ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress())
		return "", err
	}
	err = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress())
	if err != nil {
		return "", err
	}
	return response.TxHash, err
}

func (s CoreumProcessing) mintCoreumToken(subunit, issuerAddress, mnemonic string, amount int64) (string, error) {
	msgMint := &assetfttypes.MsgMint{Sender: issuerAddress, Coin: sdk.Coin{Denom: subunit + "-" + issuerAddress, Amount: sdk.NewInt(amount)}}

	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		"key-name",
		mnemonic,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	if err != nil {
		_ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress())
		return "", err
	}

	ctx := context.Background()
	response, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(senderInfo.GetAddress()),
		s.factory,
		msgMint,
	)
	if err != nil {
		_ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress())
		return "", err
	}
	err = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress())
	if err != nil {
		return "", err
	}
	return response.TxHash, err
}

func (s CoreumProcessing) burnCoreumToken(subunit, issuerAddress, mnemonic string, amount int64) (string, error) {
	msgBurn := &assetfttypes.MsgBurn{Sender: issuerAddress, Coin: sdk.Coin{Denom: subunit + "-" + issuerAddress, Amount: sdk.NewInt(amount)}}

	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		"key-name",
		mnemonic,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	if err != nil {
		_ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress())
		return "", err
	}

	ctx := context.Background()
	response, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(senderInfo.GetAddress()),
		s.factory,
		msgBurn,
	)
	if err != nil {
		_ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress())
		return "", err
	}
	err = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress())
	if err != nil {
		return "", err
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
