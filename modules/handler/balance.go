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

// GetBalance is a method for getting clients balance on given blockchain
func GetBalance(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)

		blockchain := service.BalanceRequest{}
		blockchain.Blockchain = strings.ToLower(r.URL.Query().Get("blockchain"))
		blockchain.Asset = strings.ToLower(r.URL.Query().Get("asset"))
		externalId, err := internal.GetExternalID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "incorrect request data", http.StatusBadRequest)
			return
		}
		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not find merchant", http.StatusBadRequest)
			return
		}
		res, err := processing.GetBalance(blockchain, merchantID, externalId)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not get balance for user", http.StatusBadRequest)
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
