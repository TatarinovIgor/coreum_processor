package main

import (
	"context"
	"coreum_processor/cmd/internal"
	"coreum_processor/cmd/multisign-service/routing"
	MultiSignService "coreum_processor/cmd/multisign-service/service"
	"coreum_processor/modules/service"
	"fmt"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
	ctx := context.Background()

	cfg := internal.LoadMultiSignEnv()

	processingService := service.NewProcessingService(cfg.PublicKey, nil,
		cfg.TokenTimeToLive, nil, service.Merchants{}, nil)

	// @ToDo write transaction check callback function and find client context
	multiSignService := MultiSignService.NewMultiSignService(client.Context{}, nil, cfg.Mnemonics)

	router := httprouter.New()
	urlPath := ""

	log.Println("Multisignature service has been started")
	routing.InitRouter(ctx, router, urlPath, processingService, multiSignService)

	server := &http.Server{Addr: fmt.Sprintf(":%s", cfg.Port), Handler: router}
	err := server.ListenAndServe()
	log.Println(err)
}
