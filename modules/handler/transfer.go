package handler

import (
	"coreum_processor/modules/internal"
	"coreum_processor/modules/service"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"strings"
)

func Deposit(processing *service.ProcessingService) httprouter.Handle {
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
		res, err = processing.Deposit(CredentialsDeposit, merchantID, externalId)
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

func NewToken(processing *service.ProcessingService) httprouter.Handle {
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

		res, err = processing.IssueToken(TokenRequest, merchantID, externalId)
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

func MintToken(processing *service.ProcessingService) httprouter.Handle {
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

		res, err = processing.MintToken(TokenRequest, merchantID, externalId)
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

func BurnToken(processing *service.ProcessingService) httprouter.Handle {
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

		res, err = processing.BurnToken(TokenRequest, merchantID, externalId)
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
