package handler

import (
	"coreum_processor/modules/internal"
	"coreum_processor/modules/service"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func GetTransactionList(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)

		transactionRequest := service.TransactionRequest{}
		blockchain := strings.ToLower(r.URL.Query().Get("blockchain"))
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		action := r.URL.Query().Get("action")
		status := r.URL.Query().Get("status")
		fromUnix, err := strconv.Atoi(from)
		if err != nil {
			fromUnix = 0
		}
		toUnix, err := strconv.Atoi(to)
		if err != nil {
			toUnix = int(time.Now().Unix())
		}
		transactionRequest.FromUnix = uint(fromUnix)
		transactionRequest.ToUnix = uint(toUnix)
		transactionRequest.Blockchain = blockchain
		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}
		// Multi-chain
		res, err := processing.GetTransactions(transactionRequest, merchantID,
			strings.Split(action, ","), strings.Split(status, ","))
		if err != nil {
			log.Println(err)
			http.Error(w, "could not fetch a list of transaction", http.StatusBadRequest)
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

func GetTransaction(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w = processing.SetHeaders(w)

		transactionRequest := service.TransactionRequest{}
		id := ps.ByName("id")
		fromUnix := 0
		toUnix := time.Now().Unix()
		transactionRequest.FromUnix = uint(fromUnix)
		transactionRequest.ToUnix = uint(toUnix)
		transactionRequest.Blockchain = "coreum"
		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}
		var emptyReq []string
		// Multi-chain
		res, err := processing.GetTransactions(transactionRequest, merchantID,
			emptyReq, emptyReq)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not fetch transactions", http.StatusBadRequest)
			return
		}

		for i := 0; i < len(res); i++ {
			if res[i].GUID == uuid.MustParse(id) {
				err = json.NewEncoder(w).Encode(res[i])
			}
		}
		if err != nil {
			log.Println(err)
			http.Error(w, "could not find transaction with id", http.StatusNotFound)
			return
		}
	}
}
