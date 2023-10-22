package routing

import (
	"context"
	"coreum_processor/modules/asset"
	"coreum_processor/modules/handler"
	"coreum_processor/modules/handler/ui"
	"coreum_processor/modules/middleware"
	"coreum_processor/modules/service"
	user "coreum_processor/modules/user"
	"github.com/julienschmidt/httprouter"
	"github.com/ory/client-go"
	"net/http"
)

func InitRouter(ctx context.Context, ory *client.APIClient,
	router *httprouter.Router, pathName string,
	processing *service.ProcessingService, userService *user.Service, assetService *asset.Service) {

	routerWrap := NewRouterWrap(pathName, router)

	router.GET("/about", handler.About)
	router.GET("/", handler.PageLanding)
	//GET routers for frontend auth
	routerWrap.GET("/login", ui.GetPageLogIn(ctx, ory))
	routerWrap.GET("/register", ui.GetPageSignUp(ctx, ory))
	routerWrap.GET("/logout", ui.GetPageLogOut(ctx, ory))
	routerWrap.GET("/reset", ui.PageReset)
	routerWrap.GET("/recovery", ui.PageRecovery)
	routerWrap.GET("/error", handler.About)

	// routers for ory
	routerWrap.POST("/kratos/create-user", handler.CreateUser(userService))
	routerWrap.POST("/kratos/login-user", handler.LoginUser(userService))

	// routers for UI
	routerWrap.GET("/ui/dashboard", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.PageDashboard(ctx, userService, processing)))

	routerWrap.GET("/ui/merchant/onboarding-wizard", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.PageWizardMerchant(ctx, userService)))
	routerWrap.POST("/ui/merchant/onboarding-wizard", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.PageWizardMerchantUpdate(ctx, userService)))

	routerWrap.GET("/ui/merchant/transactions", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.PageMerchantTransaction(ctx, processing)))
	routerWrap.GET("/ui/merchant/users", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.PageMerchantUsers(ctx, userService, processing)))
	routerWrap.GET("/ui/merchant/settings", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.PageMerchantSettings(processing, assetService)))
	routerWrap.GET("/ui/admin/merchant-requests", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.PageRequestsAdmin(ctx, userService, processing)))
	routerWrap.POST("/ui/admin/merchant-requests", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.PageRequestsAdminUpdate(ctx, userService, processing)))
	routerWrap.GET("/ui/admin/asset-requests", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.PageAssetRequestsAdmin(ctx, assetService, processing)))
	routerWrap.POST("/ui/admin/asset-requests", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.PageAssetRequestsAdminUpdate(ctx, assetService, processing)))
	routerWrap.GET("/ui/merchant/assets", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.PageMerchantAssets(ctx, assetService, processing)))
	routerWrap.POST("/ui/merchant/assets", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.AssetRequestMerchant(assetService)))

	//Urls for form submissions from frontend
	routerWrap.POST("/submit_public_key", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, handler.PublicKeySaver(processing)))
	routerWrap.POST("/submit_callback_url", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, handler.UpdateCallbackUrl(processing)))
	routerWrap.POST("/submit_new_token", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, handler.NewTokenSaver(ctx, processing, assetService)))
	routerWrap.POST("/reset_password", ui.PasswordReset(ctx, ory))
	routerWrap.POST("/set_password", ui.PasswordSet(ctx, ory))

	routerWrap.POST("/ui/merchant/mint", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, handler.MintTokenMerchant(ctx, processing)))
	routerWrap.POST("/ui/merchant/burn", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, handler.BurnTokenMerchant(ctx, processing)))
	routerWrap.POST("/ui/merchant/create_wallet", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, handler.CreateWallet(ctx, processing)))
	routerWrap.POST("/ui/merchant/deposit", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.Deposit(ctx, processing)))
	routerWrap.POST("/ui/merchant/withdraw", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.Withdraw(processing)))
	routerWrap.POST("/ui/merchant/update_withdraw", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.UpdateWithdraw(processing)))
	routerWrap.POST("/ui/merchant/transfer", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.TransferMerchantWallets(ctx, processing)))

	//GET routers for styles and assets
	router.ServeFiles("/assets/*filepath", http.Dir("templates/assets"))

	//GET routers for backend
	routerWrap.GET("/get_balance", middleware.AuthMiddleware(processing, handler.GetBalance(ctx, processing)))               //Tested
	routerWrap.GET("/transactions", middleware.AuthMiddleware(processing, handler.GetTransactionList(processing)))           //Tested
	routerWrap.GET("/merchant/:id", middleware.AuthMiddleware(processing, handler.GetMerchantById(processing)))              //Tested
	routerWrap.GET("/merchants", middleware.AuthMiddleware(processing, handler.GetMerchants(processing)))                    //Tested
	routerWrap.GET("/get_wallet_by_id", middleware.AuthMiddleware(processing, handler.GetWalletById(processing)))            //Tested
	routerWrap.GET("/get_transaction_status/:id", middleware.AuthMiddleware(processing, handler.GetTransaction(processing))) //Tested
	routerWrap.GET("/get_supply", middleware.AuthMiddlewareCookie(ctx, ory, userService, handler.GetTokenSupply(ctx, processing)))

	//POST router for backend
	routerWrap.POST("/deposit", middleware.AuthMiddleware(processing, handler.Deposit(ctx, processing))) //Tested
	routerWrap.POST("/token_issue", middleware.AuthMiddleware(processing, handler.NewToken(ctx, processing, assetService)))
	routerWrap.POST("/token_mint", middleware.AuthMiddleware(processing, handler.MintToken(ctx, processing, assetService)))
	routerWrap.POST("/token_burn", middleware.AuthMiddleware(processing, handler.BurnTokenMerchant(ctx, processing))) //Tested
	routerWrap.POST("/withdraw", middleware.AuthMiddleware(processing, handler.Withdraw(processing)))                 //Tested
	routerWrap.POST("/merchant", middleware.AuthMiddlewareAdmin(processing, handler.CreateMerchant(processing)))      //Tested

	// DELETE routers for backend
	routerWrap.DELETE("/withdraw/:guid", middleware.AuthMiddleware(processing, handler.DeleteWithdraw(processing))) //Tested

	// PUT routers for backend
	routerWrap.PUT("/withdraw/:guid", middleware.AuthMiddleware(processing, handler.UpdateWithdraw(processing)))
	routerWrap.PUT("/merchant/:id", middleware.AuthMiddleware(processing, handler.UpdateMerchant(processing)))
	routerWrap.PUT("/merchant/:id/:blockchain/commission",
		middleware.AuthMiddlewareAdmin(processing, handler.UpdateMerchantCommission(ctx, processing)))

}
