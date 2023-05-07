package main

import (
	"context"
	internalApp "coreum_processor/cmd/internal"
	"coreum_processor/modules/routing"
	"coreum_processor/modules/service"
	"coreum_processor/modules/storage"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/oklog/run"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
	ctx, cancelFunc := context.WithCancel(context.Background())

	// Initialize application configuration from ENV
	cfg := internalApp.MakeConfigFromEnv()

	// Connecting to database
	db := internalApp.DBConnect()
	merchantsStore, err := storage.NewStorage("merchants", db)
	if err != nil {
		panic(fmt.Errorf("cant open merchant storage: %v", err))
	}
	transactionStore, err := storage.NewTransactionStorage("transactions", db)
	if err != nil {
		panic(fmt.Errorf("cant open transactions storage: %v", err))
	}

	// Adding processors to the unified structure
	processors := map[string]service.CryptoProcessor{
		internalApp.Coreum: internalApp.InitProcessorCoreum(internalApp.Coreum, db),
	}

	// Initializing merchant management service
	merchants := service.NewMerchantService(merchantsStore)

	// Initializing processing services
	processingService := service.NewProcessingService(cfg.PublicKey, cfg.PrivateKey,
		cfg.TokenTimeToLive, processors, merchants, transactionStore)

	// Setting up API routing
	router := httprouter.New()
	urlPath := ""
	log.Println("hello i am started")
	routing.InitRouter(router, urlPath, processingService)

	// Creating 2 streams one for API and another for blockchain requests
	var g run.Group
	g.Add(func() error {
		// Creating a process for API and initializing an API listener
		err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), router)
		cancelFunc()
		return err
	}, func(err error) {
		cancelFunc()
	})
	g.Add(func() error {
		// Creating a process for blockchain requests and initializing a blockchain listener
		err := processingService.ListenAndServe(ctx, cfg.Interval)
		cancelFunc()
		return err
	}, func(err error) {
		cancelFunc()
	})
	// Shutdown
	g.Add(func() error {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, os.Kill)

		select {
		case c := <-sigChan:
			return fmt.Errorf("interrupted with sig %q", c)
		case <-ctx.Done():
			return nil
		}
	}, func(err error) {
		cancelFunc()
	})

	err = g.Run()
	if err != nil {
		return
	}

}
