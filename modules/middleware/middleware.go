package middleware

import (
	"context"
	"coreum_processor/modules/internal"
	"coreum_processor/modules/service"
	user "coreum_processor/modules/user"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/ory/client-go"
	"log"
	"net/http"
)

func AuthMiddlewareAdmin(ProcessingService *service.ProcessingService, next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		authToken := r.Header.Get("Authorization")
		token, err := ProcessingService.AdminTokenDecode(authToken)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "invalid or expired jwt", http.StatusBadRequest)
			return
		}
		ctx := internal.WithExternalID(r.Context(), token.ExternalId)
		ctx = internal.WithMerchantID(ctx, token.MerchantID)
		next(w, r.WithContext(ctx), ps)
	}
}

func AuthMiddlewareCookie(ctx context.Context, ory *client.APIClient,
	userService *user.Service, next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		cookies := r.Header.Get("Cookie")
		// check if we have a session
		session, _, err := ory.FrontendApi.ToSession(ctx).Cookie(cookies).Execute()
		if (err != nil) || (err == nil && !*session.Active) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		userStore, err := userService.GetUser(session.Identity.Id)
		if err != nil {
			log.Println("can't get user, err: " + err.Error())
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		} else if user.IsBlocked(userStore.Access) {
			log.Println("user is blocked: " + session.Identity.Id)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		ctxR := internal.WithUserStore(r.Context(), userStore)
		merchantList, err := userService.GetUserMerchants(session.Identity.Id)
		if err == nil && len(merchantList) > 0 {
			ctxR = internal.WithMerchantID(ctxR, merchantList[0].MerchantID)
		} else if err != nil {
			log.Println(err)
		} else {
			log.Println("can't find merchants for user: ", session.Identity.Id)
		}
		next(w, r.WithContext(ctxR), ps)
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
