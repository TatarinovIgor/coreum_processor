package ui

import (
	"context"
	"coreum_processor/modules/asset"
	"coreum_processor/modules/internal"
	"coreum_processor/modules/service"
	"coreum_processor/modules/storage"
	"coreum_processor/modules/user"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"log"
	"net/http"
	"time"
)

func PageDashboardMerchant(ctx context.Context, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		userStore, err := internal.GetUserStore(r.Context())
		if err != nil {
			log.Println(`can't find user store`)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"message":"` + `can't find user store` + `"}`))
			return
		}
		userStore = userStore
		t, err := template.ParseFiles("./templates/lite/default/index.html")
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
		transactionRequest.Blockchain = ""
		merchantID, err := internal.GetMerchantID(r.Context())

		var emptyReq []string
		// Multi-chain
		res, err := processing.GetTransactions(transactionRequest, merchantID,
			emptyReq, emptyReq)
		_, err = processing.GetMerchantData(merchantID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `data parsing error` + `"}`))
			return
		}

		tronWallet, _ := processing.GetWalletById("tron", merchantID, merchantID+"-R")
		ethereumWallet, _ := processing.GetWalletById("ethereum", merchantID, merchantID+"-R")
		polygonWallet, _ := processing.GetWalletById("polygon", merchantID, merchantID+"-R")
		bitcoinWallet, err := processing.GetWalletById("bitcoin", merchantID, merchantID+"-R")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `data parsing error` + `"}`))
			return
		}

		templateVarMap := map[string]interface{}{
			"transactions":    res,
			"wallet_tron":     tronWallet,
			"wallet_ethereum": ethereumWallet,
			"wallet_polygon":  polygonWallet,
			"wallet_bitcoin":  bitcoinWallet,
		}
		err = t.ExecuteTemplate(w, "index.html", templateVarMap)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
	}
}

func PageMerchantTransaction(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		t, err := template.ParseFiles("./templates/lite/default/transactions.html", "./templates/lite/sidebar.html")

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

		varmap := map[string]interface{}{
			"transactions":   generateTransactionTable(res),
			"guid":           merchantID,
			"coreum_wallet":  nil,
			"coreum_balance": nil,
			"coreum_asset":   nil,
		}

		_, err = processing.GetMerchantData(merchantID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `error while trying to get merchant's data'` + `", "error":"` + err.Error() + `"} `))
			return
		}
		coreumWallet, err := processing.GetWalletById("coreum", merchantID, merchantID+"-R")
		coreumBalance, err := processing.GetBalance(service.BalanceRequest{Blockchain: "coreum", Asset: "coreum"}, merchantID, merchantID+"-R")
		if err == nil {
			varmap["coreum_wallet"] = coreumWallet
			varmap["coreum_balance"] = coreumBalance[0].Amount
			varmap["coreum_asset"] = coreumBalance[0].Asset
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
	htmlBlock := "<thead><tr><th><span>CREATED AT</span></th><th><span>CLIENT ID <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span>BLOCKCHAIN <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span>ACTION <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span>WALLET <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span>STATUS <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span>ASSET <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span>AMOUNT <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span>ACTION <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th></tr></thead>"
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

func PageMerchantUsers(userService *user.Service, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		t, err := template.ParseFiles("./templates/lite/default/users.html", "./templates/lite/sidebar.html")
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
			"users":          generateUserTable(res),
			"guid":           merchantID,
			"coreum_wallet":  nil,
			"coreum_balance": nil,
			"coreum_asset":   nil,
		}

		_, err = processing.GetMerchantData(merchantID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `data parsing error` + `"}`))
			return
		}
		coreumWallet, err := processing.GetWalletById("coreum", merchantID, merchantID+"-R")
		coreumBalance, err := processing.GetBalance(service.BalanceRequest{Blockchain: "coreum", Asset: "coreum"}, merchantID, merchantID+"-R")
		if err == nil {
			varmap["coreum_wallet"] = coreumWallet
			varmap["coreum_balance"] = coreumBalance[0].Amount
			varmap["coreum_asset"] = coreumBalance[0].Asset
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
	htmlBlock := "<thead><tr><th><span>CREATED AT</span></th><th><span>CLIENT ID <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span>FIRST NAME <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span>LAST NAME <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span>ACCESS LEVEL <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span></th></tr></thead>"
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
			"tokens": generateAssetsTable(res),
			"key":    merchantData.PublicKey,
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
	htmlBlock := "<thead><tr><th><span>Code (Symbol)</span></th><th><span>Name <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span>Issuer <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span> Issued <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span> Mint <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span>Burn <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th></tr></thead>"

	for i := 0; i < len(res); i++ {
		htmlBlock = htmlBlock + "" +
			"<tbody><tr>" +
			"<td>" + res[i].Code + "</td>" +
			"<td>" + res[i].Name + "</td>" +
			"<td>" + res[i].Issuer + "</td>" +
			"<td>" + "0" + "</td>" +
			"<td class=\"action_btn point done\" onclick=\"openFormMint('" + res[i].Code + "')\">" + "Mint" + "</td>" +
			"<td class=\"action_btn point rejected\" onclick=\"openFormBurn('" + res[i].Code + "')\">" + "Burn" + "</td>" +
			"</tr></tbody>"
	}

	return template.HTML(htmlBlock)
}

func PageMerchantAssets(assetService *asset.Service, processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		//Getting MerchantID
		merchantID, err := internal.GetMerchantID(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `data parsing error` + `"}`))
			return
		}

		//Setting template
		t, err := template.ParseFiles("./templates/lite/assets/assets-for-merchant.html", "./templates/lite/sidebar.html")
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}

		//Getting all assets for merchant merchantID
		assets, err := assetService.GetAssetList(merchantID, nil, nil, "", "active", time.Unix(0, 0), time.Now().UTC())
		//Setting RawData
		varmap := map[string]interface{}{
			"assets": assets,
			"guid":   merchantID,
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

func AssetRequestMerchant(ctx context.Context, assetService *asset.Service, processing *service.ProcessingService) httprouter.Handle {
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
		blockchain := r.Form.Get("blockchain")
		name := r.Form.Get("name")
		code := r.Form.Get("code")
		description := r.Form.Get("description")
		assetType := r.Form.Get("assetType")
		smartContractAddress := r.Form.Get("issuer")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `data parsing error` + `"}`))
			return
		}

		err = assetService.CreateAssetRequest(blockchain, code, smartContractAddress, name, description, assetType, merchantOwnerID, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `failed to add asset to database` + err.Error() + `"}`))
			return
		}

		http.Redirect(w, r, "/ui/merchant/assets", http.StatusSeeOther)
	}
}
