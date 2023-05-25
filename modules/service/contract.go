package service

import (
	"context"
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
	TxHash string
}

type Balance struct {
	Amount     float64 `json:"amount"`
	Blockchain string  `json:"blockchain"`
	Asset      string  `json:"asset"`
	Issuer     string  `json:"issuer"`
}

type TransferRequest struct {
	Amount     float64
	Blockchain string
	Asset      string
	Issuer     string
}

type NewTokenRequest struct {
	Symbol      string `json:"symbol"`
	Subunit     string `json:"subunit"`
	Blockchain  string `json:"blockchain"`
	Description string `json:"description"`
}

type TokenRequest struct {
	Subunit    string `json:"subunit"`
	Blockchain string `json:"blockchain"`
	Amount     string `json:"amount"`
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
	WalletAddress string `json:"wallet_address"`
	WalletSeed    string `json:"wallet_seed"`
	Blockchain    string `json:"blockchain"`
}

type UserWalletDB struct {
	ExternalId    string
	WalletAddress string
	WalletSeed    string
	Blockchain    string
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
type TronTransactions struct {
	Data    []TransactionResponse `json:"data"`
	Success bool                  `json:"success"`
	Meta    TransactionMeta       `json:"meta"`
}
type CryptoTransactionStatus int

const (
	NoTransaction         CryptoTransactionStatus = -1
	PendingTransaction    CryptoTransactionStatus = 1
	FailedTransaction     CryptoTransactionStatus = 2
	SuccessfulTransaction CryptoTransactionStatus = 3
)

type FuncDepositCallback func(blockChain, merchantID, externalId, externalWallet, hash, asset, issuer string, amount float64)
type CryptoProcessor interface {
	Deposit(request CredentialDeposit, merchantID, externalId string) (*DepositResponse, error)
	Withdraw(request CredentialWithdraw, merchantID, externalId string, merchantWallets Wallets) (*WithdrawResponse, error)
	TransferToReceiving(request TransferRequest, merchantID, externalId string) (*TransferResponse, error)
	TransferFromReceiving(transfer TransferRequest, merchantID, externalId string) (*TransferResponse, error)
	GetBalance(request BalanceRequest, merchantID, externalId string) (*Balance, error)
	GetWalletById(merchantID, externalId string) (string, error)
	GetTransactionStatus(hash string) (CryptoTransactionStatus, error)
	StreamDeposit(ctx context.Context, callback FuncDepositCallback, interval time.Duration)
	TransferFromSending(request TransferRequest, merchantID, receivingWallet string) (*TransferResponse, error)
	IssueToken(request NewTokenRequest, merchantID, externalID string) (*NewTokenResponse, error)
	MintToken(request TokenRequest, merchantID, externalID string) (*NewTokenResponse, error)
	BurnToken(request TokenRequest, merchantID, externalID string) (*NewTokenResponse, error)
}
