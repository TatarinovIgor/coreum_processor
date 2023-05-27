package ui

import (
	"context"
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
		t, err := template.ParseFiles("./templates/lite/default/transactions.html")

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
			w.Write([]byte(`{"message":"` + `data parsing error` + `"}`))
			return
		}
		var emptyReq []string
		// Multi-chain
		res, err := processing.GetTransactions(transactionRequest, merchantID,
			emptyReq, emptyReq)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `data parsing error` + `"}`))
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
			w.Write([]byte(`{"message":"` + `data parsing error` + `"}`))
			return
		}
		var coreumBalance *service.Balance
		coreumWallet, err := processing.GetWalletById("coreum", merchantID, merchantID+"-R")
		coreumBalance, err = processing.GetBalance(service.BalanceRequest{Blockchain: "coreum", Asset: "coreum"}, merchantID, merchantID+"-R")
		if err == nil {
			varmap["coreum_wallet"] = coreumWallet
			varmap["coreum_balance"] = coreumBalance.Amount
			varmap["coreum_asset"] = coreumBalance.Asset
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
		t, err := template.ParseFiles("./templates/lite/default/users.html")
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
			"transactions":   generateUserTable(res),
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
		var coreumBalance *service.Balance
		coreumWallet, err := processing.GetWalletById("coreum", merchantID, merchantID+"-R")
		coreumBalance, err = processing.GetBalance(service.BalanceRequest{Blockchain: "coreum", Asset: "coreum"}, merchantID, merchantID+"-R")
		if err == nil {
			varmap["coreum_wallet"] = coreumWallet
			varmap["coreum_balance"] = coreumBalance.Amount
			varmap["coreum_asset"] = coreumBalance.Asset
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

func PageMerchantSettings(processing *service.ProcessingService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		t, err := template.ParseFiles("./templates/lite/default/settings.html")

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

		varmap := map[string]interface{}{
			"wallets": generateBlockchainsTable(processing, merchantID, merchantData.Wallets),
		}

		err = t.ExecuteTemplate(w, "settings.html", varmap)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
	}
}

func generateBlockchainsTable(processing *service.ProcessingService, merchantID string, res map[string]service.Wallets) template.HTML {
	htmlBlock := "<thead><tr><th><span>Blockchain</span></th><th><span>Commission Fix <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span>Commission Percentage <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span>Wallet ID <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th><th><span>Type <a class=\"help\" data-toggle=\"popover\" title=\"Popover title\" data-content=\"And here's some amazing content. It's very engaging. Right?\"><i class=\"feather icon-help-circle f-16\"></i></a></span></th></tr></thead>"
	blockchain := ""
	for i := 0; i < 4; i++ {
		switch i {
		case 0:
			blockchain = "tron"
		case 1:
			blockchain = "ethereum"
		case 2:
			blockchain = "polygon"
		case 3:
			blockchain = "bitcoin"
		}
		walletSending, err := processing.GetWalletById(blockchain, merchantID, merchantID+"-S")
		if err != nil {
			break
		}
		htmlBlock = htmlBlock + "" +
			"<tbody><tr>" +
			"<td>" + blockchain + "</td>" +
			"<td>" + fmt.Sprintf("%v", res[blockchain].CommissionSending.Fix) + "</td>" +
			"<td>" + fmt.Sprintf("%v", res[blockchain].CommissionSending.Percent) + "</td>" +
			"<td>" + walletSending + "</td>" +
			"<td>" + "Sending" + "</td>" +
			"</tr></tbody>"

		walletReceiving, err := processing.GetWalletById(blockchain, merchantID, merchantID+"-S")
		if err != nil {
			break
		}
		htmlBlock = htmlBlock + "" +
			"<tbody><tr>" +
			"<td>" + blockchain + "</td>" +
			"<td>" + fmt.Sprintf("%v", res[blockchain].CommissionReceiving.Fix) + "</td>" +
			"<td>" + fmt.Sprintf("%v", res[blockchain].CommissionReceiving.Percent) + "</td>" +
			"<td>" + walletReceiving + "</td>" +
			"<td>" + "Receiving" + "</td>" +
			"</tr></tbody>"
	}
	return template.HTML(htmlBlock)
}
