package middleware

import (
	"coreum_processor/modules/internal"
	"coreum_processor/modules/service"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func AuthMiddlewareAdmin(ProcessingService *service.ProcessingService, next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		authToken := r.Header.Get("Authorization")
		token, err := ProcessingService.AdminTokenDecode(authToken)
		if err != nil {
			log.Error(err.Error())
			http.Error(w, "invalid or expired jwt", http.StatusBadRequest)
			return
		}
		ctx := internal.WithExternalID(r.Context(), token.ExternalId)
		ctx = internal.WithMerchantID(ctx, token.MerchantID)
		next(w, r.WithContext(ctx), ps)
	}
}

func AuthMiddleware(ProcessingService *service.ProcessingService, next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		authToken := r.Header.Get("Authorization")
		token, err := ProcessingService.TokenDecode(authToken)
		if err != nil {

			http.Error(w, fmt.Sprintf("invalid or expired jwt, err: %s", err.Error()), http.StatusBadRequest)
			return
		}
		ctx := internal.WithExternalID(r.Context(), token.ExternalId)
		ctx = internal.WithMerchantID(ctx, token.MerchantID)
		next(w, r.WithContext(ctx), ps)
	}
}

func AuthMiddlewareForm(ProcessingService *service.ProcessingService, next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		authToken := r.URL.Query().Get("auth_key")
		token, err := ProcessingService.TokenDecode(authToken)
		if err != nil {
			http.Error(w, "invalid or expired jwt", http.StatusBadRequest)
			return
		}
		ctx := internal.WithExternalID(r.Context(), token.ExternalId)
		ctx = internal.WithMerchantID(ctx, token.MerchantID)
		next(w, r.WithContext(ctx), ps)
	}
}

func CorsResponse(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")     // normal header
	w.Header().Set("Cache-Control", "public, max-age=600") // normal header

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, OPTIONS_CORS, PUT, DELETE, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Authorization, Authorization")
}
