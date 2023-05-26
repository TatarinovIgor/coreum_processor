package ui

import (
	"context"
	"coreum_processor/modules/internal"
	"coreum_processor/modules/service"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"log"
	"net/http"
)

func PageDashboardAdmin(ctx context.Context, processing *service.ProcessingService) httprouter.Handle {
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
