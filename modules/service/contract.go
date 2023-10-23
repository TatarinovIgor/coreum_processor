package service

import (
	"context"
	"coreum_processor/modules/storage"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type ErrorService error

var (
	ErrNotImplemented ErrorService = fmt.Errorf("not implemented")
)

type TokenPayload struct {
	ExternalId string `json:"external_id"`
	MerchantID string `json:"merchant_id"`
}

type MerchantData struct {
	ID           uuid.UUID          `json:"id"`
	PublicKey    string             `json:"public_key"`
	MerchantName string             `json:"name"`
	CallBackURL  string             `json:"call_back_url"`
	Wallets      map[string]Wallets `json:"wallets"`
}
type Commission struct {
	Fix     float64 `json:"fix"`
	Percent float64 `json:"percent"`
}

type Wallets struct {
	CommissionReceiving Commission `json:"commission_receiving"`
	CommissionSending   Commission `json:"commission_sending"`
	ReceivingID         string     `json:"receiving_id"`
	SendingID           string     `json:"sending_id"`
	SignPublicKey       string     `json:"sign_public_key"`
}

type SmartContract struct {
	SmartContractAddress string `json:"smart_contract_address"`
	FeeLimit             int64  `json:"fee_limit"`
}

type TokenData struct {
	Payload  TokenPayload `json:"payload"`
	Subject  string       `json:"sub"`
	IssuedAt uint         `json:"iat"`
	//ExpiresIn uint         `json:"exp"`
}

type TransactionRequest struct {
	FromUnix   uint   `json:"from_unix"`
	ToUnix     uint   `json:"to_unix"`
	Blockchain string `json:"blockchain"`
}

type BalanceRequest struct {
	Blockchain string `json:"blockchain"`
	Asset      string `json:"asset"`
	Issuer     string `json:"issuer"`
}

type CredentialDeposit struct {
	Amount     float64 `json:"amount"`
	Blockchain string  `json:"blockchain"`
	Asset      string  `json:"asset"`
	Issuer     string  `json:"issuer"`
}

type CredentialWithdraw struct {
	Amount        float64 `json:"amount"`
	Blockchain    string  `json:"blockchain"`
	WalletAddress string  `json:"wallet_address"`
	Asset         string  `json:"asset"`
	Issuer        string  `json:"issuer"`
	Memo          string  `json:"memo"`
}

type WithdrawResponse struct {
	TransactionHash string `json:"result"`
}

type DepositResponse struct {
	WalletAddress string `json:"wallet_address"`
	Memo          string `json:"memo"`
	URL           string `json:"url"`
	Id            string `json:"id"`
}

type TransferResponse struct {
	PendingHash  string
	TransferHash string
}

type NewTokenResponse struct {
	Issuer string
	TxHash string
}

type Balance struct {
	Amount     float64 `json:"amount"`
	Blockchain string  `json:"blockchain"`
	Asset      string  `json:"asset"`
	Issuer     string  `json:"issuer"`
}

type TransferTokenRequest struct {
	Amount              float64
	Blockchain          string
	Subunit             string
	Issuer              string
	Type                string
	SendingExternalId   string
	ReceivingExternalId string
	NftClassId          string
	NftId               string
}

type TransferRequest struct {
	Amount     float64
	Blockchain string
	Asset      string
	Issuer     string
}

type NewTokenRequest struct {
	Symbol        string `json:"symbol"`
	Code          string `json:"code"`
	Blockchain    string `json:"blockchain"`
	Issuer        string
	Description   string `json:"description"`
	InitialAmount int64  `json:"initial_amount"`
	Type          string `json:"type"`
}

type TokenRequest struct {
	Code       string `json:"code"`
	Blockchain string `json:"blockchain"`
	Amount     string `json:"amount"`
	Issuer     string `json:"issuer"`
}

type MintTokenRequest struct {
	ClassID           string `json:"class_id"`
	NftId             string `json:"nft_id"`
	Code              string `json:"code"`
	Blockchain        string `json:"blockchain"`
	Amount            string `json:"amount"`
	Issuer            string `json:"issuer"`
	ReceivingWalletID string `json:"receiving_wallet_id"`
	Type              string `json:"type"`
}

type NewMerchant struct {
	PublicKey    string `json:"public_key"`
	MerchantName string `json:"name"`
	Callback     string `json:"callback"`
}

type NewMerchantCommission struct {
	CommissionReceiving Commission `json:"commission_receiving"`
	CommissionSending   Commission `json:"commission_sending"`
}
type Transaction struct {
	Amount     uint   `json:"amount"`
	Blockchain string `json:"blockchain"`
	Action     bool   `json:"action"`
	ExternalId string `json:"external_id"`
	Asset      string `json:"asset"`
	Issuer     string `json:"issuer"`
	Timestamp  uint   `json:"timestamp"`
}

type MerchantResponse struct {
	MerchantId string `json:"id"`
}

type DeleteWithdrawResponse struct {
	Status string `json:"status"`
}

type TransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
}

type Wallet struct {
	WalletAddress string  `json:"wallet_address"`
	WalletSeed    string  `json:"wallet_seed"`
	Blockchain    string  `json:"blockchain"`
	Threshold     float64 `json:"threshold"`
}

type TransactionResponse struct {
	TransactionId string `json:"transaction_id"`
	TokenInfo     struct {
		Symbol   string `json:"symbol"`
		Address  string `json:"address"`
		Decimals int    `json:"decimals"`
		Name     string `json:"name"`
	} `json:"token_info"`
	BlockTimestamp int64  `json:"block_timestamp"`
	From           string `json:"from"`
	To             string `json:"to"`
	Type           string `json:"type"`
	Value          string `json:"value"`
}

type TransactionMeta struct {
	At       int `json:"at"`
	PageSize int `json:"page_size"`
}

type MultiSignAddress map[string]float64

type SignTransactionRequest struct {
	ExternalID string `json:"external_id"`
	Blockchain string `json:"blockchain"`
	Address    string `json:"address"`
	TrxID      string `json:"trxID"`
	TrxData    string `json:"trxData"`
	Threshold  int    `json:"threshold"`
}

type CryptoTransactionStatus int

const (
	NoTransaction         CryptoTransactionStatus = -1
	PendingTransaction    CryptoTransactionStatus = 1
	FailedTransaction     CryptoTransactionStatus = 2
	SuccessfulTransaction CryptoTransactionStatus = 3
)

// FuncDepositCallback defines a callback function to inform main processing service about received deposit
type FuncDepositCallback func(blockChain, merchantID, externalId, externalWallet, hash, asset, issuer string,
	amount float64)

// FuncMultiSignAddrCallback defines a callback function to get a list of address to be added to multi sig account
type FuncMultiSignAddrCallback func(blockChain, externalId string) (MultiSignAddress, float64, error)

// FuncMultiSignSignature defines a callback function to get a list of address to be added to multi sig account
type FuncMultiSignSignature func(request SignTransactionRequest) (map[string][]byte, error)

// FuncTransactionsCallback defines a callback function to post transaction for merchant
type FuncTransactionsCallback func(trx storage.TransactionStore) error

type CryptoProcessor interface {
	// CreateWallet create a wallet and put to the store under defined externalID for merchantID
	//	- merchantID - id of the merchant that request to make a new wallet
	//	- externalID - id of newly created wallet in an external system
	//	- multiSignAddresses - a func that provide a list of addresses to generate multi sign wallet,
	//						   in case of nil newly created wallet will not support multi signature
	// in case of success create new blockchain wallet and put it to the storage
	CreateWallet(ctx context.Context, merchantID, externalId string,
		multiSignAddresses FuncMultiSignAddrCallback) (*Wallet, error)

	GetWalletById(merchantID, externalId string) (string, error)

	// Deposit create a
	Deposit(ctx context.Context, request CredentialDeposit, merchantID, externalId string,
		multiSignAddresses FuncMultiSignAddrCallback) (*DepositResponse, error)
	StreamDeposit(ctx context.Context, callback FuncDepositCallback, interval time.Duration)

	// Withdraw
	Withdraw(ctx context.Context, request CredentialWithdraw,
		merchantID, externalId string, merchantWallets Wallets,
		multiSignSignature FuncMultiSignSignature) (*WithdrawResponse, error)

	IssueFT(ctx context.Context, request NewTokenRequest, merchantID, externalID string,
		multiSignAddresses FuncMultiSignAddrCallback) (*NewTokenResponse, []byte, error)
	IssueNFTClass(ctx context.Context, request NewTokenRequest, merchantID, externalId string,
		multiSignAddresses FuncMultiSignAddrCallback) (*NewTokenResponse, []byte, error)
	MintFT(ctx context.Context, request MintTokenRequest, merchantID string) (*NewTokenResponse, error)
	MintNFT(ctx context.Context, request MintTokenRequest, merchantID string) (*NewTokenResponse, error)
	BurnToken(ctx context.Context, request TokenRequest, merchantID, externalID string) (*NewTokenResponse, error)

	TransferToReceiving(ctx context.Context, request TransferRequest,
		merchantID, externalId string) (*TransferResponse, error)
	TransferFromReceiving(ctx context.Context, transfer TransferRequest,
		merchantID, externalId string) (*TransferResponse, error)
	TransferBetweenMerchantWallets(ctx context.Context, request TransferRequest,
		merchantID string) (*TransferResponse, error)
	TransferFromSending(ctx context.Context, request TransferRequest,
		merchantID, receivingWallet string) (*TransferResponse, error)
	TransferFT(ctx context.Context, request TransferTokenRequest,
		merchantID string) (string, error)
	TransferNFT(ctx context.Context, request TransferTokenRequest,
		merchantID string) (string, error)

	GetTokenSupply(ctx context.Context, request BalanceRequest) (int64, error)
	GetBalance(ctx context.Context, merchantID, externalID string) (Balance, error)
	GetAssetsBalance(ctx context.Context, request BalanceRequest, merchantID, externalId string) ([]Balance, error)
	GetTransactionStatus(ctx context.Context, hash string) (CryptoTransactionStatus, error)
}
