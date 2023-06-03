package ui

import (
	"context"
	"coreum_processor/modules/asset"
	"coreum_processor/modules/internal"
	"coreum_processor/modules/service"
	"coreum_processor/modules/user"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

func PageMerchantsAdmin(ctx context.Context, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		userStore, err := internal.GetUserStore(r.Context())
		if err != nil {
			log.Println(`can't find user store`)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"message":"` + `can't find user store` + `"}`))
			return
		}
		t, err := template.ParseFiles("./templates/lite/dashboard/dashboard.html")
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
		userStore = userStore

		merchants, err := processing.GetMerchants()
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}

		err = t.Execute(w, merchants)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
	}
}

func PageRequestsAdmin(ctx context.Context, userService *user.Service, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		userStore, err := internal.GetUserStore(r.Context())
		if err != nil {
			log.Println(`can't find user store`)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"message":"` + `can't find user store` + `"}`))
			return
		}
		t, err := template.ParseFiles("./templates/lite/merchants/merchants.html", "./templates/lite/sidebar.html")
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
		userStore = userStore

		requests, err := userService.GetUserList("", nil, time.Unix(0, 0), time.Now().UTC())

		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}

		err = t.Execute(w, requests)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
	}
}

type RawUser struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Identity  string `json:"identity"`
}

func PageRequestsAdminUpdate(ctx context.Context, userService *user.Service, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		//Getting data from the request
		w = processing.SetHeaders(w)
		raw := struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Identity  string `json:"identity"`
		}{}
		err := json.NewDecoder(r.Body).Decode(&raw)
		if err != nil {
			http.Error(w, "Failed to parse request body", http.StatusBadRequest)
			return
		}

		//Creating new merchant
		merchant := service.MerchantData{
			PublicKey:    "",
			MerchantName: raw.FirstName,
			ID:           uuid.New(),
			CallBackURL:  "",
		}

		_, err = processing.CreateMerchants(merchant.ID.String(), merchant)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not create new merchant", http.StatusBadRequest)
			return
		}

		//Creating userStore
		userStore, err := internal.GetUserStore(r.Context())
		if err != nil {
			log.Println(`can't find user store`)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"message":"` + `can't find user store` + `"}`))
			return
		}

		userStore.Identity = strings.Trim(raw.Identity, " ")
		userStore.FirstName = raw.FirstName
		userStore.LastName = raw.LastName

		//Updating user
		err = userService.UpdateUser(*userStore)
		if err != nil {
			log.Println("Failed to update user: ", err)
			http.Error(w, "Failed to update user", http.StatusBadRequest)
		}

		//Adding merchant ID to merchant_list
		err = userService.ApproveUserMerchant(userStore.Identity, merchant.ID.String())
		if err != nil {
			log.Println("Failed to add merchant ID to merchant_list: ", err)
			http.Error(w, "Failed to add merchant ID to merchant_list", http.StatusBadRequest)
		}

		//Linking user to merchant
		err = userService.LinkUserToMerchant(userStore.Identity, merchant.ID.String())
		if err != nil {
			log.Println("Failed to link user to merchant: ", err)
			http.Error(w, "Failed to link user to merchant", http.StatusBadRequest)
		}

		_, err = userService.SetUserAccess(userStore.Identity, user.SetOnboarded(userStore.Access))
		if err != nil {
			log.Println(err)
		}

		// Send a response
		response := map[string]string{"message": "Updated successfully"}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, "Failed to send response", http.StatusInternalServerError)
			return
		}
	}
}

func PageAssetRequestsAdmin(ctx context.Context, assetService *asset.Service, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		t, err := template.ParseFiles("./templates/lite/assets/asset-requests.html", "./templates/lite/sidebar.html")
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}

		requests, err := assetService.GetAssetList("", nil, nil, "", "pending",
			time.Unix(0, 0), time.Now().UTC())

		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}

		err = t.Execute(w, requests)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
	}
}

func PageAssetRequestsAdminUpdate(ctx context.Context, assetService *asset.Service, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		//Getting data from the request
		w = processing.SetHeaders(w)
		raw := struct {
			Merchant   string `json:"merchant"`
			Blockchain string `json:"blockchain"`
			Code       string `json:"code"`
			Issuer     string `json:"issuer"`
			Action     string `json:"action"`
		}{}
		err := json.NewDecoder(r.Body).Decode(&raw)
		if err != nil {
			http.Error(w, "Failed to parse request body", http.StatusBadRequest)
			return
		}

		token, err := assetService.GetBlockChainAssetByCodeAndIssuer(raw.Blockchain, raw.Code, raw.Issuer)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not get asset", http.StatusBadRequest)
			return
		}

		if raw.Action == "Reject" {
			err := assetService.DeleteAssetRequest(*token, raw.Merchant)
			if err != nil {
				log.Println(err)
				http.Error(w, "could not delete asset", http.StatusBadRequest)
				return
			}
		} else if raw.Issuer != "" {
			err = assetService.ActivateAsset(raw.Blockchain, raw.Code, raw.Issuer, raw.Merchant)
			if err != nil {
				log.Println(err)
				http.Error(w, "could not activate asset", http.StatusBadRequest)
				return
			}
		} else {

			requestAsset := service.NewTokenRequest{
				Symbol:        token.Name,
				Code:          token.Code,
				Blockchain:    token.BlockChain,
				Issuer:        "",
				Description:   token.Description,
				InitialAmount: 1000000,
			}

			resp, _, err := processing.IssueToken(requestAsset, raw.Merchant, raw.Merchant)
			if err != nil {
				log.Println(err)
				http.Error(w, "could not issue asset", http.StatusInternalServerError)
				return
			}

			err = assetService.ActivateAsset(raw.Blockchain, strings.ToLower(raw.Code), resp.Issuer, raw.Merchant)
			if err != nil {
				log.Println(err)
				http.Error(w, "could not save asset", http.StatusInternalServerError)
				return
			}
		}

		// Send a response
		response := map[string]string{"message": "Updated successfully"}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, "Failed to send response", http.StatusInternalServerError)
			return
		}
	}
}
