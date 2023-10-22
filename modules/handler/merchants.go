package handler

import (
	"context"
	"coreum_processor/modules/internal"
	"coreum_processor/modules/service"
	"coreum_processor/modules/storage"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"strings"
	"time"
)

type RecordParsed struct {
	ID        int64                `json:"id"`
	Data      service.MerchantData `json:"data"`
	TTL       time.Duration        `json:"ttl"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// GetMerchantById returns a data about merchant by given id, this data includes: Public Key, Name, Callback URL, and Merchant's wallets
func GetMerchantById(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w = processing.SetHeaders(w)

		merchantID := ps.ByName("id")
		merchantData, err := processing.GetMerchantData(merchantID)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}
		err = json.NewEncoder(w).Encode(merchantData)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

// GetMerchants returns all the records of merchants from database
func GetMerchants(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)

		records, err := processing.GetMerchants()
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}
		err = json.NewEncoder(w).Encode(records)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

// CreateMerchant method for creating a new record in database with merchants data
func CreateMerchant(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)

		newMerchantData := service.NewMerchant{}
		_, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}
		err = json.NewDecoder(r.Body).Decode(&newMerchantData)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse new merchant data", http.StatusBadRequest)
			return
		}
		merchant := parseMerchantData(newMerchantData)
		_, err = processing.CreateMerchants(merchant.ID.String(), merchant)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not create new merchant", http.StatusBadRequest)
			return
		}
		err = json.NewEncoder(w).Encode(merchant)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

// CreateWallet method for creating a new wallets
func CreateWallet(ctx context.Context, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)
		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}

		var data struct {
			Blockchain string `json:"blockchain"`
		}
		err = json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}
		merchData, err := processing.GetMerchantData(merchantID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `can't find merchant data` + `"}`))
			return
		}
		_, ok := merchData.Wallets[data.Blockchain]
		if ok {
			err := fmt.Errorf("wallet for merchant already exists")
			log.Println(err)
			http.Error(w, "wallet all ready exists", http.StatusBadRequest)
			return
		}
		// TODO define default commission
		commission := service.Commission{
			Fix:     1,
			Percent: 1,
		}
		wallets := service.Wallets{
			CommissionReceiving: commission,
			CommissionSending:   commission,
			ReceivingID:         merchantID + "-R",
			SendingID:           merchantID + "-S",
		}
		_, err = processing.CreateWallet(ctx, data.Blockchain, merchantID, wallets.ReceivingID)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not create receiving wallet", http.StatusBadRequest)
			return
		}

		_, err = processing.CreateWallet(ctx, data.Blockchain, merchantID, wallets.SendingID)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not create sending wallet", http.StatusBadRequest)
			return
		}
		if merchData.Wallets == nil {
			merchData.Wallets = map[string]service.Wallets{}
		}
		merchData.Wallets[data.Blockchain] = wallets
		_, err = processing.SaveMerchantData(merchantID, merchData)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not create sending wallet", http.StatusBadRequest)
			return
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

func CreateClientWallet(ctx context.Context, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)
		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}

		var data struct {
			Identity   string `json:"identity"`
			Blockchain string `json:"blockchain"`
		}

		err = json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}

		_, err = processing.CreateWallet(ctx, data.Blockchain, merchantID, data.Identity)

		//@ToDo figure out what to do with this
		//merchData, err := processing.GetMerchantData(merchantID)
		//if err != nil {
		//	w.WriteHeader(http.StatusInternalServerError)
		//	w.Write([]byte(`{"message":"` + `can't find merchant data` + `"}`))
		//	return
		//}
		//if merchData.Wallets == nil {
		//	merchData.Wallets = map[string]service.Wallets{}
		//}
		//merchData.Wallets[data.Blockchain] = wallets
		//_, err = processing.SaveMerchantData(merchantID, merchData)
		//if err != nil {
		//	log.Println(err)
		//	http.Error(w, "could not create sending wallet", http.StatusBadRequest)
		//	return
		//}

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

// UpdateMerchant method for updating a new record with merchants data
func UpdateMerchant(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w = processing.SetHeaders(w)

		newMerchantData := service.NewMerchant{}
		merchantID := ps.ByName("id")
		err := json.NewDecoder(r.Body).Decode(&newMerchantData)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}
		_, err = processing.UpdateMerchant(merchantID, newMerchantData)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not update merchant data", http.StatusBadRequest)
			return
		}
		merchantReturn := service.MerchantResponse{MerchantId: merchantID}
		err = json.NewEncoder(w).Encode(merchantReturn)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

// UpdateMerchantCommission method for setting a individual commission for a merchant
func UpdateMerchantCommission(ctx context.Context, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w = processing.SetHeaders(w)

		newMerchantCommission := service.NewMerchantCommission{}
		merchantID := ps.ByName("id")
		blockchain := strings.ToLower(ps.ByName("blockchain"))

		err := json.NewDecoder(r.Body).Decode(&newMerchantCommission)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}
		_, err = processing.UpdateMerchantCommission(ctx, merchantID, blockchain, newMerchantCommission)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not update merchants commission", http.StatusBadRequest)
			return
		}
		merchantReturn := service.MerchantResponse{MerchantId: merchantID}
		err = json.NewEncoder(w).Encode(merchantReturn)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

// GetWalletById method for getting a wallet data on given blockchain by its id
func GetWalletById(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w = processing.SetHeaders(w)

		blockchain := strings.ToLower(r.URL.Query().Get("blockchain"))
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
		wallet, err := processing.GetWalletById(blockchain, merchantID, externalId)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not get wallet", http.StatusBadRequest)
			return
		}
		merchantReturn := service.MerchantResponse{MerchantId: wallet}
		err = json.NewEncoder(w).Encode(merchantReturn)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}

// parseMerchantData parsing a merchant's data for storing it in database
func parseMerchantData(newMerchant service.NewMerchant) service.MerchantData {
	merchant := service.MerchantData{
		PublicKey:       newMerchant.PublicKey,
		MerchantName:    newMerchant.MerchantName,
		ID:              uuid.New(),
		CallBackURL:     newMerchant.Callback,
		SignCallBackURL: newMerchant.SignCallBack,
	}
	return merchant
}

// parseRecord parsing a merchant's data for exporting it from database
func parseRecord(record storage.Record) RecordParsed {
	recordParsed := RecordParsed{}
	merchantData := service.MerchantData{}
	err := json.Unmarshal(record.Data, &merchantData)
	if err != nil {
		return RecordParsed{}
	}
	recordParsed.ID = record.ID
	recordParsed.TTL = record.TTL
	recordParsed.UpdatedAt = record.UpdatedAt
	recordParsed.Data = merchantData
	return recordParsed
}

// parseRecords parse an array of given records form database
func parseRecords(records []storage.Record) []service.MerchantData {
	var recordsParsed []service.MerchantData
	for i := 0; i < len(records); i++ {
		recordParsed := parseRecord(records[i])
		recordsParsed = append(recordsParsed, recordParsed.Data)
	}
	return recordsParsed
}
