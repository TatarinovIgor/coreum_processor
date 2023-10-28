package routing

import (
	"context"
	"coreum_processor/cmd/multisign-service/handler"
	multiSignService "coreum_processor/cmd/multisign-service/service"
	"coreum_processor/modules/middleware"
	"coreum_processor/modules/service"
	"github.com/julienschmidt/httprouter"
)

func InitRouter(ctx context.Context, router *httprouter.Router, pathName string,
	processing *service.ProcessingService, multiSign *multiSignService.MultiSignService) {
	routerWrap := NewRouterWrap(pathName, router)

	routerWrap.GET("/addresses", handler.GetAddressesHandler(multiSign))
	routerWrap.POST("/sign", middleware.AuthMiddlewareAdmin(processing, handler.SignTransactionHandler(ctx, multiSign)))
	routerWrap.POST("/transaction", handler.InitiateTransactionHandler(multiSign))
}
