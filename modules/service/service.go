package service

import (
	"bytes"
	"context"
	"coreum_processor/modules/storage"
	"crypto/rsa"
	"crypto/x509"
	"embed"
	"encoding/json"
	"encoding/pem"
	"fmt"
	encoder "github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	_ "io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	//go:embed html-templates/*.gohtml
	res          embed.FS
	pageDeposit  = "html-templates/payment_form_deposit.gohtml"
	pageWithdraw = "html-templates/payment_form_withdraw.gohtml"
)

type ProcessingService struct {
	publicKey        *rsa.PublicKey
	privateKey       *rsa.PrivateKey
	tokenTimeToLive  int64
	processorWallets []Wallet
	processors       map[string]CryptoProcessor
	merchants        Merchants
	transactionStore *storage.TransactionPSQL
}

// NewProcessingService create a service to process transaction by provided crypto processor
func NewProcessingService(publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey,
	tokenTimeToLive int64, processors map[string]CryptoProcessor,
	merchants Merchants, transactionStore *storage.TransactionPSQL) *ProcessingService {
	return &ProcessingService{
		publicKey:        publicKey,
		privateKey:       privateKey,
		tokenTimeToLive:  tokenTimeToLive,
		processors:       processors,
		merchants:        merchants,
		transactionStore: transactionStore,
	}
}

func (s ProcessingService) ListenAndServe(ctx context.Context, interval time.Duration) error {
	for _, processor := range s.processors {
		processor.StreamDeposit(ctx, s.makeDepositCallback(), interval)
	}
	ticker := time.NewTicker(time.Second * interval)
	for {
		select {
		case <-ctx.Done():
			log.Println("exit from crypto processing")
			return nil
		case <-ticker.C:
			s.processTransaction(ctx)
		}
	}
}

func (s ProcessingService) MakeCallback(store storage.TransactionStore, callBackURL string) error {
	client := &http.Client{}
	body, err := json.Marshal(store)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", callBackURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	var tokenToSend TokenPayload
	tokenToSend.MerchantID = store.MerchantId
	tokenToSend.ExternalId = store.ExternalId
	token, err := s.GenerateToken(tokenToSend)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", token)

	_, err = client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s ProcessingService) GenerateToken(tokenData TokenPayload) (string, error) {
	token := encoder.New(encoder.SigningMethodRS256)
	claims := token.Claims.(encoder.MapClaims)
	claims["exp"] = time.Now().Add(10 * time.Minute).Unix()
	claims["payload"] = tokenData
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s ProcessingService) AdminTokenDecode(token string) (TokenPayload, error) {
	tokenData := TokenData{}

	tok, err := jwt.Parse([]byte(token))
	if err != nil {
		return tokenData.Payload, fmt.Errorf("can't parse token: %s, err: %v", token, err)
	}

	tokenByte, err := json.Marshal(tok.PrivateClaims())
	if err != nil {
		return tokenData.Payload, fmt.Errorf("can't marshal token claim, err: %v", err)
	}

	err = json.Unmarshal(tokenByte, &tokenData)
	if err != nil {
		return tokenData.Payload, fmt.Errorf("can't unmarshal token data, err: %v", err)
	}

	if time.Now().Unix()-tok.IssuedAt().Unix() > s.tokenTimeToLive {
		return tokenData.Payload, fmt.Errorf("token expaired")
	}

	_, err = jwt.Parse([]byte(token), jwt.WithVerify(jwa.RS256, s.publicKey), jwt.WithValidate(true))
	if err != nil {
		return tokenData.Payload, fmt.Errorf("can't parse token: %s, err: %v", token, err)
	}

	return tokenData.Payload, nil
}
func (s ProcessingService) TokenDecode(token string) (TokenPayload, error) {
	tokenData := TokenData{}

	tok, err := jwt.Parse([]byte(token))
	if err != nil {
		return tokenData.Payload, fmt.Errorf("can't parse token: %s, err: %v", token, err)
	}
	tokenString, err := json.Marshal(tok.PrivateClaims())
	if err != nil {
		return tokenData.Payload, fmt.Errorf("can't marshal token claim, err: %v", err)
	}
	err = json.Unmarshal(tokenString, &tokenData)
	if err != nil {
		return tokenData.Payload, fmt.Errorf("can't unmarshal token data, err: %v", err)
	}
	if time.Now().Unix()-tok.IssuedAt().Unix() > s.tokenTimeToLive {
		return tokenData.Payload, fmt.Errorf("token expaired")
	}
	data, err := s.merchants.GetMerchantData(tokenData.Payload.MerchantID)
	if err != nil {
		return tokenData.Payload, fmt.Errorf("can't unmarshal token data, err: %v", err)
	}
	block, _ := pem.Decode([]byte(data.PublicKey))
	if block == nil {
		return tokenData.Payload, fmt.Errorf("can't unmarshal token data, err: %v", err)
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return tokenData.Payload, fmt.Errorf("can't unmarshal token data, err: %v", err)
	}
	_, err = jwt.Parse([]byte(token), jwt.WithVerify(jwa.RS256, pub.(*rsa.PublicKey)), jwt.WithValidate(true))
	if err != nil {
		return tokenData.Payload, fmt.Errorf("can't parse token: %s, err: %v", token, err)
	}
	return tokenData.Payload, nil
}

func (s ProcessingService) GetMerchantData(merchantID string) (MerchantData, error) {
	data, err := s.merchants.GetMerchantData(merchantID)
	if err != nil {
		return MerchantData{}, err
	}
	return data, nil
}

func (s ProcessingService) GetMerchants() ([]MerchantData, error) {
	data, err := s.merchants.GetMerchants()
	if err != nil {
		return nil, err
	}
	response := make([]MerchantData, len(data))
	for i, res := range data {
		newData := MerchantData{}
		_ = json.Unmarshal(res.Data, &newData)
		response[i].ID = newData.ID
		response[i].MerchantName = newData.MerchantName
		response[i].PublicKey = newData.PublicKey
		response[i].Wallets = newData.Wallets
	}
	return response, nil
}

func (s ProcessingService) CreateMerchants(guid string, merchant MerchantData) (int64, error) {
	return s.merchants.CreateMerchantData(guid, merchant)
}

func (s ProcessingService) UpdateMerchant(guid string, merchant NewMerchant) (string, error) {
	return s.merchants.UpdateMerchantData(guid, merchant)
}

func (s ProcessingService) GetWalletById(blockchain, merchantID, externalId string) (string, error) {
	processor, ok := s.processors[blockchain]
	if !ok {
		return "", fmt.Errorf("%s blockchain not found", blockchain)
	}
	return processor.GetWalletById(merchantID, externalId)
}

func (s ProcessingService) UpdateMerchantCommission(guid, blockchain string, merchant NewMerchantCommission) (Wallets, error) {
	wallets, err := s.merchants.UpdateMerchantCommission(guid, blockchain, merchant)
	if err != nil {
		return wallets, err
	}
	processor, ok := s.processors[blockchain]
	if !ok {
		return Wallets{}, fmt.Errorf("%s blockchain not found", blockchain)
	}
	_, err = processor.Deposit(CredentialDeposit{Blockchain: blockchain, Amount: 0}, guid, wallets.SendingID)
	if err != nil {
		return Wallets{}, err
	}
	//wallets.SendingID = res.WalletAddress
	_, err = processor.Deposit(CredentialDeposit{Blockchain: blockchain, Amount: 0}, guid, wallets.ReceivingID)
	if err != nil {
		return Wallets{}, err
	}
	//wallets.ReceivingID = res.WalletAddress
	return wallets, nil
}

func (s ProcessingService) Deposit(deposit CredentialDeposit, merchantID, externalId string) (*DepositResponse, error) {
	processor, ok := s.processors[deposit.Blockchain]
	if !ok {
		return nil, fmt.Errorf("%s blockchain not found", deposit.Blockchain)
	}

	response, err := processor.Deposit(deposit, merchantID, externalId)
	if err != nil {
		return nil, fmt.Errorf("could not perform deposit: %s", err)
	}
	/*
		id, err := s.transactionStore.CreateTransaction(merchantID, externalId, deposit.Blockchain, "deposit", "", "", deposit.Blockchain, deposit.Issuer, deposit.Amount, 0)
		if err != nil {
			return nil, fmt.Errorf("could not create transaction: %s", err)
		}
		response.Id = id
		err = s.transactionStore.UpdateWallet(response.WalletAddress, merchantID, externalId, id)
		if err != nil {
			return nil, fmt.Errorf("could not save external wallet: %s", err)
		}
	*/
	return response, nil
}

func (s ProcessingService) InitWithdraw(withdraw CredentialWithdraw, merchantID, externalId string) (*WithdrawResponse, error) {
	_, ok := s.processors[withdraw.Blockchain]
	if !ok {
		return nil, fmt.Errorf("%s blockchain not found", withdraw.Blockchain)
	}
	merchData, err := s.merchants.GetMerchantData(merchantID)
	if err != nil {
		return nil, err
	}
	_, ok = merchData.Wallets[withdraw.Blockchain]
	if !ok {
		return nil, fmt.Errorf("%s blockchain not found for mercchant: %s", withdraw.Blockchain, merchantID)
	}
	guid, err := s.transactionStore.CreateTransaction(merchantID, externalId, withdraw.Blockchain,
		storage.WithdrawTransaction, withdraw.WalletAddress, "", withdraw.Asset, withdraw.Issuer, withdraw.Amount, 0)
	return &WithdrawResponse{TransactionHash: guid}, err
}

func (s ProcessingService) UpdateWithdraw(transactionID, merchantID, externalId, hash string) error {
	merchant, err := s.merchants.GetMerchantData(merchantID)
	if err != nil {
		return err
	}

	transaction, err := s.transactionStore.GetTransactionByGuid(merchantID, transactionID)
	if err != nil {
		return err
	}

	wallet, ok := merchant.Wallets[transaction.Blockchain]
	if !ok {
		return fmt.Errorf("cannot find blockchain: '%v' for merchant", transaction.Blockchain)
	}

	commission := transaction.Amount*wallet.CommissionSending.Percent/100.0 + wallet.CommissionSending.Fix

	err = s.transactionStore.PutProcessedTransaction(merchantID, externalId, transactionID, hash, commission)
	if err != nil {
		return err
	}
	return nil
}

func (s ProcessingService) DeleteWithdraw(transaction, merchantID, externalId string) error {
	err := s.transactionStore.RejectTransaction(merchantID, externalId, transaction)
	if err != nil {
		return err
	}
	return nil
}

func (s ProcessingService) IssueToken(request NewTokenRequest, merchantID, externalId string) (*NewTokenResponse, error) {
	processor, ok := s.processors[request.Blockchain]
	if !ok {
		return nil, fmt.Errorf("%s blockchain not found", request.Blockchain)
	}

	response, err := processor.IssueToken(request, merchantID, externalId)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s ProcessingService) GetBalance(request BalanceRequest, merchantID, externalId string) (*Balance, error) {
	processor, ok := s.processors[request.Blockchain]
	if !ok {
		return nil, fmt.Errorf("%s blockchain not found", request.Blockchain)
	}
	return processor.GetBalance(request, merchantID, externalId)
}

func (s ProcessingService) GetTransactions(request TransactionRequest, merchantID string,
	actionFilter []string, statusFilter []string) ([]storage.TransactionStore, error) {
	transactions, err := s.transactionStore.GetTransactionsByMerchant(merchantID, request.Blockchain,
		actionFilter, statusFilter,
		time.Unix(int64(request.FromUnix), 0).UTC(), time.Unix(int64(request.ToUnix), 0).UTC())
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

/*
	func (service ProcessingService) MakeFormDeposit(w http.ResponseWriter, r *http.Request, blockchain string, merchantID, externalId string) {
		processor, ok := service.processors[blockchain]
		if !ok {
			w.WriteHeader(400)
			return
		}
		processor.MakeFormDeposit(w, r, merchantID, externalId)
	}

	func (service ProcessingService) MakeFormWithdraw(w http.ResponseWriter, r *http.Request, blockchain string, merchantID, externalId string) {
		processor, ok := service.processors[blockchain]
		if !ok {
			w.WriteHeader(400)
			return
		}
		processor.MakeFormWithdraw(w, r, merchantID, externalId)
	}
*/
func (s ProcessingService) NameToTickerConvert(name string) (ticker string) {
	if name == "tron" {
		return "trx"
	}
	return ""
}

func (s ProcessingService) SetHeaders(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Content-Type", "application/json") // normal header
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Authorization, origin, x-requested-with, content-type")
	return w
}
