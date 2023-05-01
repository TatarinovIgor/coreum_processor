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
	routerWrap.GET("/get_balance", middleware.AuthMiddleware(processing, handler.GetBalance(processing)))
	routerWrap.GET("/transactions", middleware.AuthMiddleware(processing, handler.GetTransactionList(processing))) //ToDo fix sorting by blockchain
	routerWrap.GET("/merchant/:id", middleware.AuthMiddleware(processing, handler.GetMerchantById(processing)))
	routerWrap.GET("/merchants", middleware.AuthMiddleware(processing, handler.GetMerchants(processing)))
	routerWrap.GET("/get_wallet_by_id", middleware.AuthMiddleware(processing, handler.GetWalletById(processing)))
	routerWrap.GET("/get_transaction_status/:id", middleware.AuthMiddleware(processing, handler.GetTransaction(processing)))

	//POST router for backend
	routerWrap.POST("/deposit", middleware.AuthMiddleware(processing, handler.Deposit(processing)))
	routerWrap.POST("/new_token", middleware.AuthMiddleware(processing, handler.NewToken(processing)))
	routerWrap.POST("/withdraw", middleware.AuthMiddleware(processing, handler.Withdraw(processing)))
	routerWrap.POST("/merchant", middleware.AuthMiddlewareAdmin(processing, handler.CreateMerchant(processing)))

	// DELETE routers for backend
	routerWrap.DELETE("/withdraw/:guid", middleware.AuthMiddleware(processing, handler.DeleteWithdraw(processing)))

	// PUT routers for backend
	routerWrap.PUT("/withdraw/:guid", middleware.AuthMiddleware(processing, handler.UpdateWithdraw(processing)))
	routerWrap.PUT("/merchant/:id", middleware.AuthMiddleware(processing, handler.UpdateMerchant(processing)))
	routerWrap.PUT("/merchant/:id/:blockchain/commission",
		middleware.AuthMiddlewareAdmin(processing, handler.UpdateMerchantCommission(processing)))

}
