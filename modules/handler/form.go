package handler

// Depreciated

/*
func FormDepositPage(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		blockchain := strings.ToLower(ps.ByName("blockchain"))
		externalId, err := internal.GetExternalID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
		}
		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
		}
		processing.MakeFormDeposit(w, r, blockchain, merchantID, externalId)

	}
}

func FormWithdrawPage(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		blockchain := strings.ToLower(ps.ByName("blockchain"))
		externalId, err := internal.GetExternalID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
		}
		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
		}
		processing.MakeFormWithdraw(w, r, blockchain, merchantID, externalId)

	}
}
*/
