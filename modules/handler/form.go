package handler

import (
	"context"
	"coreum_processor/modules/asset"
	"coreum_processor/modules/internal"
	"coreum_processor/modules/service"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"strings"
)

func UrlUpdater(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		url := r.FormValue("url")
		merchantID, err := internal.GetMerchantID(r.Context())
		data, err := processing.GetMerchantData(merchantID)
		if err != nil {
			http.Redirect(w, r, "/ui/merchant/settings", http.StatusSeeOther)
			return
		}
		merchant := service.NewMerchant{
			PublicKey:    data.PublicKey,
			MerchantName: data.MerchantName,
			Callback:     url,
		}
		_, err = processing.UpdateMerchant(merchantID, merchant)
		if err != nil {
			http.Redirect(w, r, "/ui/merchant/settings", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/ui/merchant/settings", http.StatusSeeOther)
		return
	}
}

func PublicKeySaver(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		publicKey := r.FormValue("public_key")
		merchantID, err := internal.GetMerchantID(r.Context())
		data, err := processing.GetMerchantData(merchantID)
		if err != nil {
			http.Redirect(w, r, "/ui/merchant/settings", http.StatusSeeOther)
			return
		}
		parsedKey := strings.TrimSuffix(publicKey, "\r")

		merchant := service.NewMerchant{
			PublicKey:    parsedKey,
			MerchantName: data.MerchantName,
			Callback:     data.CallBackURL,
		}
		_, err = processing.UpdateMerchant(merchantID, merchant)
		if err != nil {
			http.Redirect(w, r, "/ui/merchant/settings", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/ui/merchant/settings", http.StatusSeeOther)
		return
	}
}

func NewTokenSaver(ctx context.Context, processing *service.ProcessingService,
	assetService *asset.Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		code := r.FormValue("code")
		name := r.FormValue("name")
		description := r.FormValue("description")
		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			http.Redirect(w, r, "/ui/merchant/settings", http.StatusSeeOther)
			return
		}
		token := service.NewTokenRequest{
			Symbol:      name,
			Code:        code,
			Blockchain:  "coreum",
			Description: description,
		}
		token.Issuer, err = processing.GetWalletById(token.Blockchain, merchantID, merchantID+"-R")
		if err != nil {
			log.Println(err)
			http.Error(w, "could not perform token issuing", http.StatusBadRequest)
			return
		}
		_, features, err := processing.IssueFT(ctx, token, merchantID, merchantID+"-R")
		if err != nil {
			http.Redirect(w, r, "/ui/merchant/settings", http.StatusSeeOther)
			return
		}

		err = assetService.CreateAssetRequest(
			token.Blockchain, token.Code, "", token.Symbol, token.Description,
			"ft", merchantID, features)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not save asset", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/ui/merchant/settings", http.StatusSeeOther)
		return
	}
}
