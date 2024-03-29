package ui

import (
	"context"
	"coreum_processor/modules/asset"
	"coreum_processor/modules/internal"
	"coreum_processor/modules/service"
	"coreum_processor/modules/storage"
	"coreum_processor/modules/user"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

func PageMerchantTransaction(ctx context.Context, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		t, err := template.ParseFiles("./templates/lite/default/transactions.html", "./templates/lite/sidebar.html", "./templates/lite/wallet_card.html")

		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}

		transactionRequest := service.TransactionRequest{}
		fromUnix := 0
		toUnix := time.Now().Unix()
		transactionRequest.FromUnix = uint(fromUnix)
		transactionRequest.ToUnix = uint(toUnix)
		transactionRequest.Blockchain = "coreum"
		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `error while trying to get merchant's id'` + `", "error":"` + err.Error() + `"} `))
			return
		}
		var emptyReq []string
		// Multi-chain
		res, err := processing.GetTransactions(transactionRequest, merchantID,
			emptyReq, emptyReq)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `error while trying to get transactions for merchant'` + `", "error":"` + err.Error() + `"} `))
			return
		}

		request := service.BalanceRequest{
			Blockchain: "coreum",
			Asset:      "",
			Issuer:     "",
		}

		merchData, err := processing.GetMerchantData(merchantID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `can't find merchant data` + `"}`))
			return
		}
		varmap := map[string]interface{}{
			"transactions":             generateTransactionTable(res),
			"balancesReceiving":        []service.Balance{},
			"balancesSending":          []service.Balance{},
			"guid":                     merchantID,
			"coreum_receiving_wallet":  "Not activated",
			"coreum_receiving_balance": 0,
			"coreum_sending_wallet":    "Not activated",
			"coreum_sending_balance":   0,
			"coreum_asset":             "testcore",
		}
		for bc, wallets := range merchData.Wallets {
			balancesReceiving, err := processing.GetAssetsBalance(ctx, request, merchantID, wallets.ReceivingID)
			if err == nil {
				varmap["balancesReceiving"] = balancesReceiving
			}

			balancesSending, err := processing.GetAssetsBalance(ctx, request, merchantID, wallets.SendingID)
			if err == nil {
				varmap["balancesSending"] = balancesSending
			}

			receivingWallet, _ := processing.GetWalletById("coreum", merchantID, wallets.ReceivingID)
			receivingBalance, err := processing.GetBalance(ctx, "coreum", merchantID, wallets.ReceivingID)
			if err == nil {
				varmap[bc+"_receiving_wallet"] = receivingWallet
				varmap[bc+"_receiving_balance"] = receivingBalance.Amount
				varmap[bc+"_asset"] = receivingBalance.Asset
			}
			sendingWallet, _ := processing.GetWalletById("coreum", merchantID, wallets.SendingID)
			sendingBalance, err := processing.GetBalance(ctx, "coreum", merchantID, wallets.SendingID)
			if err == nil {
				varmap[bc+"_sending_wallet"] = sendingWallet
				varmap[bc+"_sending_balance"] = sendingBalance.Amount
			}
		}
		err = t.ExecuteTemplate(w, "transactions.html", varmap)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
	}
}

func generateTransactionTable(res []storage.TransactionStore) template.HTML {
	htmlBlock := "<thead><tr><th><span>CREATED AT</span></th><th><span>CLIENT ID </span></th><th><span>BLOCKCHAIN </span></th><th><span>ACTION </span></th><th><span>WALLET </span></th><th><span>STATUS </span></th><th><span>ASSET </span></th><th><span>AMOUNT </span></th><th><span>ACTION </span></th></tr></thead>"
	//ToDo add actions
	for i := 0; i < len(res); i++ {
		htmlBlock = htmlBlock + "" +
			"<tbody><tr><td>" + res[i].CreatedAt.String() + "</td>" +
			"<td>" + res[i].ExternalId + "</td>" +
			"<td>" + res[i].Blockchain + "</td>" +
			"<td>" + string(res[i].Action) + "</td>" +
			"<td>" + res[i].ExtWallet + "</td>" +
			"<td class=\"action_btn" + string(res[i].Status) + "\">" + string(res[i].Status) + "</td>" +
			"<td>" + res[i].Asset + "</td>" +
			"<td>" + fmt.Sprintf("%v", res[i].Amount) + "</td>" +
			"</tr></tbody>"
	}
	return template.HTML(htmlBlock)
}

func PageMerchantUsers(ctx context.Context, userService *user.Service,
	processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		t, err := template.ParseFiles("./templates/lite/default/users.html", "./templates/lite/sidebar.html", "./templates/lite/wallet_card.html")
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `data parsing error` + `"}`))
			return
		}
		res, err := userService.GetUserList(merchantID, nil, time.Unix(0, 0), time.Now().UTC())

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `data parsing error` + `"}`))
			return
		}

		varmap := map[string]interface{}{
			"users":                    generateUserTable(res),
			"guid":                     merchantID,
			"coreum_receiving_wallet":  "Not activated",
			"coreum_receiving_balance": 0,
			"coreum_sending_wallet":    "Not activated",
			"coreum_sending_balance":   0,
			"coreum_asset":             "testcore",
		}

		_, err = processing.GetMerchantData(merchantID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `data parsing error` + `"}`))
			return
		}

		coreumReceivingWallet, _ := processing.GetWalletById("coreum", merchantID, merchantID+"-R")
		coreumReceivingBalance, err := processing.GetBalance(ctx, "coreum", merchantID, merchantID+"-R")
		if err == nil {
			varmap["coreum_receiving_wallet"] = coreumReceivingWallet
			varmap["coreum_receiving_balance"] = coreumReceivingBalance.Amount
			varmap["coreum_asset"] = coreumReceivingBalance.Asset
		}
		coreumSendingWallet, _ := processing.GetWalletById("coreum", merchantID, merchantID+"-S")
		coreumSendingBalance, err := processing.GetBalance(ctx, "coreum", merchantID, merchantID+"-S")
		if err == nil {
			varmap["coreum_sending_wallet"] = coreumSendingWallet
			varmap["coreum_sending_balance"] = coreumSendingBalance.Amount
		}

		err = t.ExecuteTemplate(w, "users.html", varmap)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
	}
}

func generateUserTable(res []storage.UserStore) template.HTML {
	htmlBlock := "<thead><tr><th><span>CREATED AT</span></th><th><span>CLIENT ID</span></th><th><span>FIRST NAME</span></th><th><span>LAST NAME</span></th><th><span>ACCESS LEVEL</span></th><th><span></th></tr></thead>"
	//ToDo add actions
	for i := 0; i < len(res); i++ {
		htmlBlock = htmlBlock + "" +
			"<tbody><tr><td>" + res[i].CreatedAt.String() + "</td>" +
			"<td>" + res[i].Identity + "</td>" +
			"<td>" + res[i].FirstName + "</td>" +
			"<td>" + res[i].LastName + "</td>" +
			"<td>" + fmt.Sprintf("%v", res[i].Access) + "</td>" +
			"</tr></tbody>"
	}
	return template.HTML(htmlBlock)
}

func PageMerchantSettings(processing *service.ProcessingService, assetService *asset.Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		t, err := template.ParseFiles("./templates/lite/default/settings.html", "./templates/lite/sidebar.html")

		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}

		merchantID, err := internal.GetMerchantID(r.Context())
		merchantData, err := processing.GetMerchantData(merchantID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `data parsing error` + `"}`))
			return
		}
		var blockchains, code []string
		blockchains = append(blockchains, "")
		code = append(code, "")
		res, err := assetService.GetAssetList(merchantID, blockchains, code, "", "",
			time.Unix(0, 0), time.Now().UTC())

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `data parsing error` + `"}`))
			return
		}
		varmap := map[string]interface{}{
			"tokens":       generateAssetsTable(res),
			"key":          merchantData.PublicKey,
			"callback_url": merchantData.CallBackURL,
		}

		err = t.ExecuteTemplate(w, "settings.html", varmap)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
	}
}

func generateAssetsTable(res []storage.AssetStore) template.HTML {
	htmlBlock := "<thead><tr><th><span>Code (Symbol of token)</span></th><th><span>Name</span></th><th><span>Issuer</span></th><th><span> Issued </span></th><th><span> Mint </span></th><th><span>Burn </span></th></tr></thead>"

	for i := 0; i < len(res); i++ {
		htmlBlock = htmlBlock + "" +
			"<tbody><tr>" +
			"<td>" + res[i].Code + "</td>" +
			"<td>" + res[i].Name + "</td>" +
			"<td>" + res[i].Issuer + "</td>" +
			"<td>" + "0" + "</td>" +
			"<td class=\"action_btn point done\" onclick=\"openFormMint('" + res[i].Code + "','" + res[i].Issuer + "')\">" + "Mint" + "</td>" +
			"<td class=\"action_btn point rejected\" onclick=\"openFormBurn('" + res[i].Code + "','" + res[i].Issuer + "')\">" + "Burn" + "</td>" +
			"</tr></tbody>"
	}

	return template.HTML(htmlBlock)
}

func PageMerchantAssets(ctx context.Context, assetService *asset.Service,
	processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		//Getting MerchantID
		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `data parsing error` + `"}`))
			return
		}

		//Setting template
		t, err := template.ParseFiles("./templates/lite/assets/assets-for-merchant.html", "./templates/lite/sidebar.html", "./templates/lite/wallet_card.html")
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}

		//Getting all assets for merchant merchantID
		assets, err := assetService.GetAssetList(merchantID, nil, nil, "", "", time.Unix(0, 0), time.Now().UTC())

		//Setting RawData
		varmap := map[string]interface{}{
			"assets":                   assets,
			"guid":                     merchantID,
			"coreum_receiving_wallet":  "Not activated",
			"coreum_receiving_balance": 0,
			"coreum_sending_wallet":    "Not activated",
			"coreum_sending_balance":   0,
			"coreum_asset":             "utestcore",
		}
		_, err = processing.GetMerchantData(merchantID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `data parsing error` + `"}`))
			return
		}

		coreumReceivingWallet, _ := processing.GetWalletById("coreum", merchantID, merchantID+"-R")
		coreumReceivingBalance, err := processing.GetBalance(ctx, "coreum", merchantID, merchantID+"-R")
		if err == nil {
			varmap["coreum_receiving_wallet"] = coreumReceivingWallet
			varmap["coreum_receiving_balance"] = coreumReceivingBalance.Amount
			varmap["coreum_asset"] = coreumReceivingBalance.Asset
		}
		coreumSendingWallet, _ := processing.GetWalletById("coreum", merchantID, merchantID+"-S")
		coreumSendingBalance, err := processing.GetBalance(ctx, "coreum", merchantID, merchantID+"-S")
		if err == nil {
			varmap["coreum_sending_wallet"] = coreumSendingWallet
			varmap["coreum_sending_balance"] = coreumSendingBalance.Amount
		}

		err = t.Execute(w, varmap)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
	}
}

func AssetRequestMerchant(assetService *asset.Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		//Getting merchant ID
		merchantOwnerID, err := internal.GetMerchantID(r.Context())
		// Parse the form data from the request
		err = r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form data", http.StatusBadRequest)
			return
		}

		// Access the form data values by their name
		blockchain := strings.ToLower(r.Form.Get("blockchain"))
		name := r.Form.Get("name")
		code := strings.ToLower(r.Form.Get("code"))
		description := r.Form.Get("description")
		assetType := r.Form.Get("assetType")
		issuer := r.Form.Get("issuer")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `data parsing error` + `"}`))
			return
		}
		//ToDo what is the reason for this???
		/*if issuer != "" {
			merchantOwnerID = ""
		}*/

		err = assetService.CreateAssetRequest(blockchain, code, issuer, name, description, assetType, merchantOwnerID, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `failed to add asset to database` + err.Error() + `"}`))
			return
		}

		http.Redirect(w, r, "/ui/merchant/assets", http.StatusSeeOther)
	}
}

func Deposit(ctx context.Context, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)
		var raw struct {
			Blockchain string `json:"blockchain"`
			ExternalID string `json:"externalID"`
		}
		err := json.NewDecoder(r.Body).Decode(&raw)
		CredentialsDeposit := service.CredentialDeposit{
			Amount:     0,
			Blockchain: raw.Blockchain,
			Asset:      "",
			Issuer:     "",
		}

		res := &service.DepositResponse{}
		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not find merchant", http.StatusBadRequest)
			return
		}

		CredentialsDeposit.Blockchain = strings.ToLower(CredentialsDeposit.Blockchain)
		res, err = processing.Deposit(ctx, CredentialsDeposit, merchantID, raw.ExternalID)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not perform deposit", http.StatusBadRequest)
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

func TransferMerchantWallets(ctx context.Context, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)

		var raw struct {
			Amount     float64 `json:"amount"`
			Blockchain string  `json:"blockchain"`
			Asset      string  `json:"asset"`
			Issuer     string  `json:"issuer"`
		}
		err := json.NewDecoder(r.Body).Decode(&raw)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}

		credentialsWithdraw := service.TransferRequest{
			Amount:     raw.Amount,
			Blockchain: raw.Blockchain,
			Asset:      raw.Asset,
			Issuer:     raw.Issuer,
		}

		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not find merchant", http.StatusBadRequest)
			return
		}

		credentialsWithdraw.Blockchain = strings.ToLower(credentialsWithdraw.Blockchain)
		res, err := processing.TransferMerchantWallets(ctx, credentialsWithdraw, merchantID)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not perform withdraw", http.StatusBadRequest)
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

func Withdraw(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w = processing.SetHeaders(w)

		var raw struct {
			Amount        float64 `json:"amount"`
			Blockchain    string  `json:"blockchain"`
			WalletAddress string  `json:"wallet_address"`
			Asset         string  `json:"asset"`
			Issuer        string  `json:"issuer"`
			ExternalID    string  `json:"externalID"`
		}
		err := json.NewDecoder(r.Body).Decode(&raw)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse request data", http.StatusBadRequest)
			return
		}

		credentialsWithdraw := service.CredentialWithdraw{
			Amount:        raw.Amount,
			Blockchain:    raw.Blockchain,
			WalletAddress: raw.WalletAddress,
			Asset:         raw.Asset,
			Issuer:        raw.Issuer,
			Memo:          "",
		}

		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "could not find merchant", http.StatusBadRequest)
			return
		}

		credentialsWithdraw.Blockchain = strings.ToLower(credentialsWithdraw.Blockchain)
		res, err := processing.InitWithdraw(credentialsWithdraw, merchantID, raw.ExternalID)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not perform withdraw", http.StatusBadRequest)
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

func UpdateWithdraw(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w = processing.SetHeaders(w)

		var raw struct {
			Guid       string `json:"guid"`
			ExternalID string `json:"externalID"`
		}

		err := json.NewDecoder(r.Body).Decode(&raw)
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

		err = processing.UpdateWithdraw(raw.Guid, merchantID, raw.ExternalID, "")
		if err != nil {
			log.Println(err)
			http.Error(w, "could not update withdraw", http.StatusBadRequest)
			return
		}

		deleteWithdrawReturn := service.DeleteWithdrawResponse{Status: "success"}
		err = json.NewEncoder(w).Encode(deleteWithdrawReturn)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not parse response from server", http.StatusInternalServerError)
			return
		}
	}
}
