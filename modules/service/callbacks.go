package service

import (
	"coreum_processor/modules/storage"
	"crypto/rsa"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt"
	"time"
)

type CallBacks struct {
	client          *resty.Client
	privateKey      *rsa.PrivateKey
	merchantService *Merchants
	tokenTimeToLive int
}

const (
	callBackAddresses    = "/addresses"
	callBackSign         = "/sign"
	callBackTransactions = "/transactions"
	minLengthCallBackURL = 9
)

func NewCallBackService(privateKey *rsa.PrivateKey, tokenTimeToLive, retryCount, retryWaitTime int,
	merchantService *Merchants) *CallBacks {
	return &CallBacks{
		// Create a Resty Client
		client: resty.New().SetRetryCount(retryCount).
			SetRetryWaitTime(time.Duration(retryWaitTime) * time.Second),
		privateKey: privateKey, tokenTimeToLive: tokenTimeToLive, merchantService: merchantService}
}

func (s *CallBacks) createJWTAuthorization() (string, error) {
	t := jwt.New(jwt.GetSigningMethod("RS256"))

	t.Claims = &jwt.StandardClaims{
		ExpiresAt: time.Now().UTC().Add(time.Duration(s.tokenTimeToLive) * time.Second).Unix(),
	}
	return t.SignedString(s.privateKey)
}

func (s *CallBacks) GetMultiSignAddressesFn(merchantID string) (FuncMultiSignAddrCallback, error) {
	merchant, err := s.merchantService.GetMerchantData(merchantID)
	if err != nil {
		return nil, err
	}
	if len(merchant.CallBackURL) < minLengthCallBackURL {
		return nil, nil
	}
	return func(blockChain, externalId string) (MultiSignAddress, float64, error) {
		threshold := 0.
		authorization, err := s.createJWTAuthorization()
		query := map[string]string{"blockchain": blockChain, "external_id": externalId}
		resp, err := s.client.R().SetHeader("Authorization", authorization).SetQueryParams(query).
			EnableTrace().
			Get(merchant.CallBackURL + callBackAddresses)
		if err != nil {
			return MultiSignAddress{}, threshold, err
		}
		res := struct {
			Addresses MultiSignAddress `json:"addresses"`
			Threshold float64          `json:"threshold"`
		}{}
		err = json.Unmarshal(resp.Body(), &res)
		return res.Addresses, res.Threshold, err
	}, nil
}

func (s *CallBacks) GetMultiSignFn(merchantID string) (FuncMultiSignSignature, error) {
	merchant, err := s.merchantService.GetMerchantData(merchantID)
	if err != nil {
		return nil, err
	}
	if len(merchant.CallBackURL) < minLengthCallBackURL {
		return nil, nil
	}
	return func(request MultiSignTransactionRequest) (map[string][]byte, error) {
		authorization, err := s.createJWTAuthorization()

		resp, err := s.client.R().SetHeader("Authorization", authorization).SetBody(request).
			EnableTrace().
			Post(merchant.CallBackURL + callBackSign)
		if err != nil {
			return nil, err
		}

		res := map[string][]byte{}
		err = json.Unmarshal(resp.Body(), &res)

		return res, err
	}, nil
}

func (s *CallBacks) GetTransactionFn(merchantID string) (FuncTransactionsCallback, error) {
	merchant, err := s.merchantService.GetMerchantData(merchantID)
	if err != nil {
		return nil, err
	}
	if len(merchant.CallBackURL) < minLengthCallBackURL {
		return nil, nil
	}
	return func(trx storage.TransactionStore) error {
		authorization, err := s.createJWTAuthorization()

		resp, err := s.client.R().SetHeader("Authorization", authorization).SetBody(trx).
			EnableTrace().
			Post(merchant.CallBackURL + callBackTransactions)
		if err != nil {
			return err
		}

		res := MultiSignAddress{}
		err = json.Unmarshal(resp.Body(), &res)

		return err
	}, nil
}
