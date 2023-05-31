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
	"strings"
	"time"
)

const (
	coreumFeeMessage  = 50000
	coreumFeeIssueBug = 10000000
	coreumFeeIssueFT  = 70000 + coreumFeeMessage + coreumFeeIssueBug
	coerumFeeMintFT   = 11000 + coreumFeeMessage
	coreumFeeBurnFT   = 23000 + coreumFeeMessage
	coreumFeeSendFT   = 16000 + coreumFeeMessage
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

func (s CoreumProcessing) MintToken(request service.TokenRequest,
	merchantID, externalID string) (*service.NewTokenResponse, error) {

	_, byteAddress, err := s.store.GetByUser(merchantID, fmt.Sprintf("%s-%s", merchantID, request.Code))
	if err != nil {
		return nil, fmt.Errorf("can't get issuer: %v-%v coreum wallet from store, err: %v",
			request.Code, merchantID, err)
	}
	wallet := service.Wallet{}
	err = json.Unmarshal(byteAddress, &wallet)
	if err != nil {
		return nil, err
	}
	if wallet.WalletAddress == "" || wallet.WalletAddress != request.Issuer {
		return nil, fmt.Errorf("empty or incorrect issuer wallet address")
	}
	amount, err := strconv.Atoi(request.Amount)
	if err != nil {
		return nil, err
	}
	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		wallet.WalletAddress,
		wallet.WalletSeed,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()

	if err != nil {
		return nil, err
	}
	token, err := s.mintCoreumToken(request.Code, request.Issuer, int64(amount))
	if err != nil {
		return nil, err
	}
	_, byteAddress, err = s.store.GetByUser(merchantID, externalID)
	if err != nil {
		return nil, fmt.Errorf("can't get user: %v coreum wallet from store, err: %v", externalID, err)
	}
	err = json.Unmarshal(byteAddress, &wallet)
	if err != nil {
		return nil, err
	}
	token, err = s.transferCoreumTokens(request.Issuer, wallet.WalletAddress,
		fmt.Sprintf("%s-%s", request.Code, request.Issuer), int64(amount))
	if err != nil {
		return nil, err
	}

	return &service.NewTokenResponse{TxHash: token}, nil
}

func (s CoreumProcessing) BurnToken(request service.TokenRequest, merchantID, externalID string) (*service.NewTokenResponse, error) {
	_, byteAddress, err := s.store.GetByUser(merchantID, fmt.Sprintf("%s-%s", merchantID, request.Code))
	if err != nil {
		return nil, fmt.Errorf("can't get user: %v coreum wallet from store, err: %v", externalID, err)
	}
	wallet := service.Wallet{}
	err = json.Unmarshal(byteAddress, &wallet)
	if err != nil {
		return nil, err
	}
	if wallet.WalletAddress == "" || wallet.WalletAddress != request.Issuer {
		return nil, fmt.Errorf("empty or incorrect issuer wallet address")
	}
	amount, err := strconv.Atoi(request.Amount)
	if err != nil {
		return nil, err
	}
	token, err := s.burnCoreumToken(request.Code, request.Issuer, wallet.WalletSeed, int64(amount))
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

func (s CoreumProcessing) Withdraw(request service.CredentialWithdraw, merchantID, externalId string, merchantWallets service.Wallets) (*service.WithdrawResponse, error) {
	ctx := context.Background()
	commission := 0.0
	if externalId != merchantWallets.ReceivingID || externalId != merchantWallets.SendingID {
		commission = merchantWallets.CommissionSending.Fix
		commission += merchantWallets.CommissionSending.Percent / 100. * (request.Amount - commission)
	}

	balance, err := s.GetBalance(
		service.BalanceRequest{Blockchain: request.Blockchain, Asset: request.Asset, Issuer: request.Issuer},
		merchantID, merchantWallets.SendingID)
	if err != nil {
		return nil, fmt.Errorf("can't get merchant: %v, sending wallet: %v, err: %w",
			merchantID, merchantWallets.SendingID, err)
	}
	if balance[0].Amount < request.Amount+commission {
		return nil, fmt.Errorf("merchant: %s, doesn't have enough balance to pay: %v %v, with commission: %v",
			merchantID, request.Amount, request.Asset, commission)
	}

	_, sendingWalletRaw, err := s.store.GetByUser(merchantID, merchantWallets.SendingID)
	if err != nil {
		return nil, err
	}

	sendingWallet := service.Wallet{}
	err = json.Unmarshal(sendingWalletRaw, &sendingWallet)
	if err != nil {
		return nil, err
	}
	//check gas
	_, err = s.updateGas(sendingWallet.WalletAddress, coreumFeeSendFT)
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
		ToAddress:   s.sendingWallet.WalletAddress,
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
	return &service.WithdrawResponse{TransactionHash: result.TxHash}, nil
}

func (s CoreumProcessing) TransferToReceiving(request service.TransferRequest,
	merchantID, externalId string) (*service.TransferResponse, error) {
	ctx := context.Background()
	_, userWallet, err := s.store.GetByUser(merchantID, externalId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	sendingWallet := service.Wallet{}
	err = json.Unmarshal(userWallet, &sendingWallet)
	//check gas
	_, err = s.updateGas(sendingWallet.WalletAddress, coreumFeeSendFT)
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

func (s CoreumProcessing) TransferFromReceiving(request service.TransferRequest,
	merchantID, externalId string) (*service.TransferResponse, error) {
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
	_, err = s.updateGas(s.receivingWallet.WalletAddress, coreumFeeSendFT)
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

func (s CoreumProcessing) TransferFromSending(request service.TransferRequest,
	merchantID, receivingWallet string) (*service.TransferResponse, error) {
	ctx := context.Background()
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

func (s CoreumProcessing) IssueToken(request service.NewTokenRequest, merchantID,
	externalId string) (*service.NewTokenResponse, []byte, error) {
	wallet := service.Wallet{}

	issuerId := fmt.Sprintf("%s-%s", merchantID, request.Code)
	_, byteAddress, err := s.store.GetByUser(merchantID, issuerId)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return nil, nil, fmt.Errorf("can't get user: %v coreum wallet from store, err: %v", externalId, err)
	} else if errors.Is(err, storage.ErrNotFound) {
		// create issuer
		wallet.WalletSeed, wallet.WalletAddress, err = s.createCoreumWallet()
		if err != nil {
			return nil, nil, err
		}

		wallet.Blockchain = request.Blockchain
		key, err := json.Marshal(wallet)
		if err != nil {
			return nil, nil, err
		}

		_, err = s.store.Put(merchantID, issuerId, wallet.WalletAddress, key)
		if err != nil {
			return nil, nil, err
		}
	} else {
		err = json.Unmarshal(byteAddress, &wallet)
		if err != nil {
			return nil, nil, err
		}
		if wallet.WalletAddress == "" {
			return nil, nil, fmt.Errorf("empty wallet address")
		}
	}
	_, err = s.updateGas(wallet.WalletAddress, coreumFeeIssueFT)
	if err != nil {
		return nil, nil, err
	}
	token, features, err := s.createCoreumToken(request.Symbol, request.Code, request.Issuer, request.Description,
		wallet.WalletSeed, request.InitialAmount)
	if err != nil {
		return nil, nil, err
	}

	return &service.NewTokenResponse{TxHash: token, Issuer: wallet.WalletAddress}, features, nil
}

func (s CoreumProcessing) GetBalance(request service.BalanceRequest, merchantID, externalId string) ([]service.Balance, error) {
	_, byteAddress, err := s.store.GetByUser(merchantID, externalId)
	if err != nil {
		return nil, fmt.Errorf("can't get user: %v coreum wallet from store, err: %v", externalId, err)
	}
	userWallet := service.Wallet{}
	err = json.Unmarshal(byteAddress, &userWallet)
	if err != nil {
		return nil, err
	}
	// Query initial balance hold by the issuer
	ctx := context.Background()
	denom := s.denom
	bankClient := banktypes.NewQueryClient(s.clientCtx)
	var balances []service.Balance
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
			} else if strings.Contains(records[0].ExternalID, records[0].MerchantID) {
				next = records[0].ID
				continue
			} else {
				record := records[len(records)-1]
				balance, err := s.GetBalance(service.BalanceRequest{Blockchain: s.blockchain, Asset: ""}, record.MerchantID, record.ExternalID)
				if balance != nil && err == nil {
					for i := 0; i < len(balance); i++ {
						if balance[i].Amount > 0 {
							callback(balance[i].Blockchain, record.MerchantID, record.ExternalID, record.Key, "",
								balance[i].Asset, balance[i].Issuer, balance[i].Amount)
						}
					}
				}
				if err != nil {
					log.Println(err)
				}
				next = record.ID
			}
		}
	}
}

func (s CoreumProcessing) updateGas(address string, txGasPrice int64) (string, error) {

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

	trx, err := s.transferCoreumTokens(senderInfo.GetAddress().String(), address, s.denom, txGasPrice)

	return trx, err
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
	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(Info.GetAddress()) }()

	// Validate address
	_, err = sdk.AccAddressFromBech32(Info.GetAddress().String())
	if err != nil {
		panic(err)
	}

	return mnemonic, Info.GetAddress().String(), nil
}

func (s CoreumProcessing) createCoreumToken(symbol, subunit, issuerAddress, description, mnemonic string,
	initialAmount int64) (string, []byte, error) {

	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		issuerAddress,
		mnemonic,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)

	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()

	if err != nil {
		return "", nil, err
	}

	features := []assetfttypes.Feature{assetfttypes.Feature_minting, assetfttypes.Feature_burning}

	msgIssue := &assetfttypes.MsgIssue{
		Issuer:        senderInfo.GetAddress().String(),
		Symbol:        symbol,
		Subunit:       strings.ToLower(subunit),
		Precision:     6,
		InitialAmount: sdk.NewInt(initialAmount),
		Description:   description,
		Features:      features,
	}

	log.Println(senderInfo.GetAddress().String())

	ctx := context.Background()
	trx, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(senderInfo.GetAddress()),
		s.factory,
		msgIssue,
	)
	if err != nil {
		return "", nil, err
	}

	featuresJson, err := json.Marshal(features)
	if err != nil {
		return "", nil, err
	}

	return trx.TxHash, featuresJson, err
}

func (s CoreumProcessing) mintCoreumToken(subunit, issuerAddress string, amount int64) (string, error) {
	msgMint := &assetfttypes.MsgMint{
		Sender: issuerAddress,
		Coin:   sdk.Coin{Denom: subunit + "-" + issuerAddress, Amount: sdk.NewInt(amount)},
	}
	address, err := sdk.AccAddressFromBech32(issuerAddress)
	if err != nil {
		return "", err
	}
	senderInfo, err := s.clientCtx.Keyring().KeyByAddress(address)
	if err != nil {
		return "", err
	}
	// update gas
	_, err = s.updateGas(issuerAddress, coerumFeeMintFT)
	if err != nil {
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
		return "", err
	}

	return response.TxHash, err
}

func (s CoreumProcessing) burnCoreumToken(subunit, issuerAddress, mnemonic string, amount int64) (string, error) {
	msgBurn := &assetfttypes.MsgBurn{
		Sender: issuerAddress, Coin: sdk.Coin{Denom: strings.ToLower(subunit) + "-" + issuerAddress, Amount: sdk.NewInt(amount)}}
	// update gas
	_, err := s.updateGas(issuerAddress, coreumFeeBurnFT)
	if err != nil {
		return "", err
	}

	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		issuerAddress,
		mnemonic,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)

	if err != nil {
		return "", err
	}
	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()

	ctx := context.Background()
	response, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(senderInfo.GetAddress()),
		s.factory,
		msgBurn,
	)
	if err != nil {
		return "", err
	}

	return response.TxHash, err
}

func (s CoreumProcessing) transferCoreumTokens(senderAddress, recipientAddress, denom string,
	amount int64) (string, error) {
	ctx := context.Background()

	address, err := sdk.AccAddressFromBech32(senderAddress)
	if err != nil {
		return "", err
	}
	senderInfo, err := s.clientCtx.Keyring().KeyByAddress(address)
	if err != nil {
		return "", err
	}

	msgSend := &banktypes.MsgSend{
		FromAddress: senderAddress,
		ToAddress:   recipientAddress,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom, amount)),
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
