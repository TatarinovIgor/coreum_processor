package routing

import (
	"coreum_processor/modules/handler"
	"coreum_processor/modules/middleware"
	"coreum_processor/modules/service"
	"github.com/julienschmidt/httprouter"
)

func InitRouter(router *httprouter.Router, pathName string, processing *service.ProcessingService) {

	routerWrap := NewRouterWrap(pathName, router)

	router.GET("/about", handler.About)

	//GET routers for frontend
	routerWrap.GET("/", handler.PageLanding)

	//GET routers for backend
	routerWrap.GET("/get_balance", middleware.AuthMiddleware(processing, handler.GetBalance(processing)))                    //Tested
	routerWrap.GET("/transactions", middleware.AuthMiddleware(processing, handler.GetTransactionList(processing)))           //Tested
	routerWrap.GET("/merchant/:id", middleware.AuthMiddleware(processing, handler.GetMerchantById(processing)))              //Tested
	routerWrap.GET("/merchants", middleware.AuthMiddleware(processing, handler.GetMerchants(processing)))                    //Tested
	routerWrap.GET("/get_wallet_by_id", middleware.AuthMiddleware(processing, handler.GetWalletById(processing)))            //Tested
	routerWrap.GET("/get_transaction_status/:id", middleware.AuthMiddleware(processing, handler.GetTransaction(processing))) //Tested

	//POST router for backend
	routerWrap.POST("/deposit", middleware.AuthMiddleware(processing, handler.Deposit(processing))) //Tested
	routerWrap.POST("/token_issue", middleware.AuthMiddleware(processing, handler.NewToken(processing)))
	routerWrap.POST("/token_mint", middleware.AuthMiddleware(processing, handler.MintToken(processing)))
	routerWrap.POST("/token_burn", middleware.AuthMiddleware(processing, handler.BurnToken(processing)))         //Tested
	routerWrap.POST("/withdraw", middleware.AuthMiddleware(processing, handler.Withdraw(processing)))            //Tested
	routerWrap.POST("/merchant", middleware.AuthMiddlewareAdmin(processing, handler.CreateMerchant(processing))) //Tested

	// DELETE routers for backend
	routerWrap.DELETE("/withdraw/:guid", middleware.AuthMiddleware(processing, handler.DeleteWithdraw(processing))) //Tested

	// PUT routers for backend
	routerWrap.PUT("/withdraw/:guid", middleware.AuthMiddleware(processing, handler.UpdateWithdraw(processing)))
	routerWrap.PUT("/merchant/:id", middleware.AuthMiddleware(processing, handler.UpdateMerchant(processing)))
	routerWrap.PUT("/merchant/:id/:blockchain/commission",
		middleware.AuthMiddlewareAdmin(processing, handler.UpdateMerchantCommission(processing)))

}
