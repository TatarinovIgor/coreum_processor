package handler

import (
	"context"
	"coreum_processor/cmd/multisign-service/contract"
	"coreum_processor/cmd/multisign-service/service"
	"encoding/json"
	"fmt"
	"github.com/dvsekhvalnov/jose2go/base64url"
	"github.com/julienschmidt/httprouter"
	"io"
	"log"
	"net/http"
)

func GetAddressesHandler(multiSignService *service.MultiSignService) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
		query := request.URL.Query()

		blockchain := query.Get("blockchain")
		externalID := query.Get("external_id")
		addresses := multiSignService.GetMultiSignAddresses(blockchain, externalID)
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

func SignTransactionHandler(ctx context.Context, multiSignService *service.MultiSignService) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		signRequest := contract.SignTransactionRequest{}
		err := json.NewDecoder(request.Body).Decode(&signRequest)
		if err != nil {
			log.Println(err)
			http.Error(writer, "could not parse request data", http.StatusBadRequest)
			return
		}

		trxData, err := base64url.Decode(signRequest.TrxData)
		if err != nil {
			log.Println(err)
			http.Error(writer, "could not decode transaction data", http.StatusBadRequest)
			return
		}
		res, err := multiSignService.MultiSignTransaction(ctx, signRequest.TrxID, signRequest.Addresses,
			trxData, signRequest.Threshold)
		if err != nil {
			log.Println(err)
			http.Error(writer, "could not sign transaction data", http.StatusBadRequest)
			return
		}
		log.Println(fmt.Sprintf("On blockchain: %s \n for external id: %s \n Sign the following transaction: %s",
			signRequest.Blockchain, signRequest.ExternalID, signRequest.TrxID))

		err = json.NewEncoder(writer).Encode(res)
		if err != nil {
			log.Println(err)
			http.Error(writer, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

func InitiateTransactionHandler(multiSignService *service.MultiSignService) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		bodyBytes, err := io.ReadAll(request.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		log.Println("Request to create transaction has the following response:" + bodyString)
	}
}
