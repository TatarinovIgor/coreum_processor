package handler

import (
	"context"
	"coreum_processor/modules/asset"
	"coreum_processor/modules/internal"
	"coreum_processor/modules/service"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"strings"
)

func Deposit(ctx context.Context, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)
		CredentialsDeposit := service.CredentialDeposit{}
		err := json.NewDecoder(r.Body).Decode(&CredentialsDeposit)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}

		res := &service.DepositResponse{}
		externalId, err := internal.GetExternalID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse client id", http.StatusBadRequest)
			return
		}

		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not find merchant", http.StatusBadRequest)
			return
		}

		CredentialsDeposit.Blockchain = strings.ToLower(CredentialsDeposit.Blockchain)
		res, err = processing.Deposit(ctx, CredentialsDeposit, merchantID, externalId)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not perform deposit", http.StatusBadRequest)
			return
		}

		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

func Withdraw(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)
		credentialsWithdraw := service.CredentialWithdraw{}
		err := json.NewDecoder(r.Body).Decode(&credentialsWithdraw)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}

		externalId, err := internal.GetExternalID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse client id", http.StatusBadRequest)
			return
		}

		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not find merchant", http.StatusBadRequest)
			return
		}

		credentialsWithdraw.Blockchain = strings.ToLower(credentialsWithdraw.Blockchain)
		res, err := processing.InitWithdraw(credentialsWithdraw, merchantID, externalId)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not perform withdraw", http.StatusBadRequest)
			return
		}

		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

func UpdateWithdraw(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w = processing.SetHeaders(w)
		transactionGuid := ps.ByName("guid")
		hash := r.URL.Query().Get("hash")
		externalId, err := internal.GetExternalID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}

		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not find merchant", http.StatusBadRequest)
			return
		}

		err = processing.UpdateWithdraw(transactionGuid, merchantID, externalId, hash)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not update withdraw", http.StatusBadRequest)
			return
		}

		deleteWithdrawReturn := service.DeleteWithdrawResponse{Status: "success"}
		err = json.NewEncoder(w).Encode(deleteWithdrawReturn)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

func DeleteWithdraw(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w = processing.SetHeaders(w)

		transactionGuid := ps.ByName("guid")
		externalId, err := internal.GetExternalID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}
		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not find merchant", http.StatusBadRequest)
			return
		}
		err = processing.DeleteWithdraw(transactionGuid, merchantID, externalId)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not delete withdraw", http.StatusBadRequest)
			return
		}
		deleteWithdrawReturn := service.DeleteWithdrawResponse{Status: "success"}
		err = json.NewEncoder(w).Encode(deleteWithdrawReturn)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

func GetTokenSupply(ctx context.Context, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)
		SupplyRequest := service.BalanceRequest{}
		SupplyRequest.Blockchain = strings.TrimSpace(strings.ToLower(r.URL.Query().Get("blockchain")))
		SupplyRequest.Asset = strings.TrimSpace(strings.ToLower(r.URL.Query().Get("asset")))
		SupplyRequest.Issuer = strings.TrimSpace(strings.ToLower(r.URL.Query().Get("issuer")))

		supply, err := processing.GetTokenSupply(ctx, SupplyRequest)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not get total supply for token", http.StatusBadRequest)
		}

		var res struct {
			Supply int64 `json:"supply"`
		}
		res.Supply = supply

		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

func NewToken(ctx context.Context, processing *service.ProcessingService,
	assetService *asset.Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)
		TokenRequest := service.NewTokenRequest{}
		err := json.NewDecoder(r.Body).Decode(&TokenRequest)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}

		res := &service.NewTokenResponse{}
		var features []byte
		externalId, err := internal.GetExternalID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse client id", http.StatusBadRequest)
			return
		}

		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not find merchant", http.StatusBadRequest)
			return
		}

		TokenRequest.Issuer, err = processing.GetWalletById(TokenRequest.Blockchain, merchantID, externalId)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not perform token issuing", http.StatusBadRequest)
			return
		}

		if TokenRequest.Type == "FT" {
			res, features, err = processing.IssueFT(ctx, TokenRequest, merchantID, externalId)

			log.Println(res)
			if err != nil {
				log.Println(err)
				http.Error(w, "could not perform token issuing", http.StatusBadRequest)
				return
			}
		} else {
			res, features, err = processing.IssueNFT(ctx, TokenRequest, merchantID, externalId)

			log.Println(res)
			if err != nil {
				log.Println(err)
				http.Error(w, "could not perform token issuing", http.StatusBadRequest)
				return
			}
		}

		err = assetService.CreateAssetRequest(TokenRequest.Blockchain, TokenRequest.Code, TokenRequest.Issuer,
			TokenRequest.Symbol, TokenRequest.Description, TokenRequest.Type, merchantID, features)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not save asset", http.StatusInternalServerError)
			return
		}
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

func MintToken(ctx context.Context, processing *service.ProcessingService,
	assetService *asset.Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)
		TokenRequest := service.MintTokenRequest{}
		err := json.NewDecoder(r.Body).Decode(&TokenRequest)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}

		res := &service.NewTokenResponse{}

		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not find merchant", http.StatusBadRequest)
			return
		}

		if TokenRequest.Type == "FT" {
			res, err = processing.MintFT(ctx, TokenRequest, merchantID)
			if err != nil {
				log.Println(err)
				http.Error(w, "could not perform token issuing", http.StatusBadRequest)
				return
			}
		} else if TokenRequest.Type == "NFT" {
			res, err = processing.MintNFT(ctx, TokenRequest, merchantID)
			if err != nil {
				log.Println(err)
				http.Error(w, "could not perform token issuing", http.StatusBadRequest)
				return
			}
		}
		err = assetService.IssueAsset(TokenRequest.Blockchain, TokenRequest.Code, merchantID)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not save asset", http.StatusInternalServerError)
			return
		}
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

func TransferToken(ctx context.Context, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)
		TransferRequest := service.TransferTokenRequest{}
		var res string
		err := json.NewDecoder(r.Body).Decode(&TransferRequest)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}

		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not find merchant", http.StatusBadRequest)
			return
		}
		if TransferRequest.Type == "FT" {
			res, err = processing.TransferFungibleToken(ctx, TransferRequest, merchantID)
			if err != nil {
				log.Println(err)
				http.Error(w, "could not perform token transferring", http.StatusBadRequest)
				return
			}
		} else if TransferRequest.Type == "NFT" {
			res, err = processing.TransferNonFungibleToken(ctx, TransferRequest, merchantID)
			if err != nil {
				log.Println(err)
				http.Error(w, "could not perform token transferring", http.StatusBadRequest)
				return
			}
		}

		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

func MintTokenMerchant(ctx context.Context, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)
		TokenRequest := service.MintTokenRequest{}
		err := json.NewDecoder(r.Body).Decode(&TokenRequest)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}

		res := &service.NewTokenResponse{}

		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not find merchant", http.StatusBadRequest)
			return
		}

		TokenRequest.ReceivingWalletID = merchantID + "-S"

		res, err = processing.MintFT(ctx, TokenRequest, merchantID)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not perform token issuing", http.StatusBadRequest)
			return
		}
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

func BurnTokenMerchant(ctx context.Context, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)
		TokenRequest := service.TokenRequest{}
		err := json.NewDecoder(r.Body).Decode(&TokenRequest)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}

		res := &service.NewTokenResponse{}

		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not find merchant", http.StatusBadRequest)
			return
		}
		res, err = processing.BurnToken(ctx, TokenRequest, merchantID, merchantID+"-S")
		if err != nil {
			log.Println(err)
			http.Error(w, "could not perform token issuing", http.StatusBadRequest)
			return
		}
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}
