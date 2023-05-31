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
	router.GET("/", handler.About)
	//GET routers for frontend auth
	routerWrap.GET("/login", ui.GetPageLogIn(ctx, ory))
	routerWrap.GET("/register", ui.GetPageSignUp(ctx, ory))
	routerWrap.GET("/logout", ui.GetPageLogOut(ctx, ory))
	routerWrap.GET("/reset", ui.PageReset)
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
		userService, ui.PageMerchantTransaction(processing)))
	routerWrap.GET("/ui/merchant/users", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.PageMerchantUsers(userService, processing)))
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
		userService, ui.PageMerchantAssets(assetService, processing)))
	routerWrap.POST("/ui/merchant/assets", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, ui.AssetRequestMerchant(ctx, assetService, processing)))

	//Urls for form submissions from frontend
	routerWrap.POST("/submit_public_key", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, handler.PublicKeySaver(processing)))
	routerWrap.POST("/submit_new_token", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, handler.NewTokenSaver(processing, assetService)))
	routerWrap.POST("/ui/merchant/mint", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, handler.MintTokenMerchant(processing, assetService)))
	routerWrap.POST("/ui/merchant/burn", middleware.AuthMiddlewareCookie(ctx, ory,
		userService, handler.BurnTokenMerchant(processing, assetService)))

	//GET routers for styles and assets
	router.ServeFiles("/assets/*filepath", http.Dir("templates/assets"))

	//GET routers for backend
	routerWrap.GET("/get_balance", middleware.AuthMiddleware(processing, handler.GetBalance(processing)))                    //Tested
	routerWrap.GET("/transactions", middleware.AuthMiddleware(processing, handler.GetTransactionList(processing)))           //Tested
	routerWrap.GET("/merchant/:id", middleware.AuthMiddleware(processing, handler.GetMerchantById(processing)))              //Tested
	routerWrap.GET("/merchants", middleware.AuthMiddleware(processing, handler.GetMerchants(processing)))                    //Tested
	routerWrap.GET("/get_wallet_by_id", middleware.AuthMiddleware(processing, handler.GetWalletById(processing)))            //Tested
	routerWrap.GET("/get_transaction_status/:id", middleware.AuthMiddleware(processing, handler.GetTransaction(processing))) //Tested

	//POST router for backend
	routerWrap.POST("/deposit", middleware.AuthMiddleware(processing, handler.Deposit(processing))) //Tested
	routerWrap.POST("/token_issue", middleware.AuthMiddleware(processing, handler.NewToken(processing, assetService)))
	routerWrap.POST("/token_mint", middleware.AuthMiddleware(processing, handler.MintToken(processing, assetService)))
	routerWrap.POST("/token_burn", middleware.AuthMiddleware(processing, handler.BurnToken(processing, assetService))) //Tested
	routerWrap.POST("/withdraw", middleware.AuthMiddleware(processing, handler.Withdraw(processing)))                  //Tested
	routerWrap.POST("/merchant", middleware.AuthMiddlewareAdmin(processing, handler.CreateMerchant(processing)))       //Tested

	// DELETE routers for backend
	routerWrap.DELETE("/withdraw/:guid", middleware.AuthMiddleware(processing, handler.DeleteWithdraw(processing))) //Tested

	// PUT routers for backend
	routerWrap.PUT("/withdraw/:guid", middleware.AuthMiddleware(processing, handler.UpdateWithdraw(processing)))
	routerWrap.PUT("/merchant/:id", middleware.AuthMiddleware(processing, handler.UpdateMerchant(processing)))
	routerWrap.PUT("/merchant/:id/:blockchain/commission",
		middleware.AuthMiddlewareAdmin(processing, handler.UpdateMerchantCommission(processing)))

}
