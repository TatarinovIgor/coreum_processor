package handler

import (
	"context"
	"coreum_processor/cmd/multisign-service/contract"
	"coreum_processor/cmd/multisign-service/service"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func GetAddressesHandler(multiSignService *service.MultiSignService) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		addresses := multiSignService.GetMultiSignAddresses()
		res := contract.MultiSignAddresses{
			Addresses: addresses,
			Threshold: len(addresses),
		}

		err := json.NewEncoder(writer).Encode(res)
		if err != nil {
			log.Println(err)
			http.Error(writer, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

func SignTransactionHandler(multiSignService *service.MultiSignService) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		signRequest := contract.SignTransactionRequest{}
		err := json.NewDecoder(request.Body).Decode(&signRequest)
		if err != nil {
			log.Println(err)
			http.Error(writer, "could not parse request data", http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		res, err := multiSignService.MultiSignTransaction(ctx, signRequest.Address, signRequest.TrxID, []byte(signRequest.TrxData),
			signRequest.Threshold)

		err = json.NewEncoder(writer).Encode(res)
		if err != nil {
			log.Println(err)
			http.Error(writer, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}
