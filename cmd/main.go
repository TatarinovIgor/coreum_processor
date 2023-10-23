package main

import (
	"context"
	internalApp "coreum_processor/cmd/internal"
	"coreum_processor/modules/asset"
	"coreum_processor/modules/routing"
	"coreum_processor/modules/service"
	"coreum_processor/modules/storage"
	"coreum_processor/modules/user"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/oklog/run"
	ory "github.com/ory/client-go"
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

	userStore, err := storage.NewUserStorage(storage.UserRegistered,
		"users", "merchant_users", "merchant_list",
		db)

	if err != nil {
		panic(fmt.Errorf("cant open user storage: %v", err))
	}

	assetsStore, err := storage.NewAssetsStorage(storage.UserRegistered,
		"assets", "merchant_assets", "merchant_list",
		db)
	if err != nil {
		panic(fmt.Errorf("cant open assets storage: %v", err))
	}

	// Adding processors to the unified structure
	processors := map[string]service.CryptoProcessor{
		internalApp.Coreum: internalApp.InitProcessorCoreum(internalApp.Coreum, db),
	}

	// Initializing merchant management service
	merchants := service.NewMerchantService(merchantsStore)

	// Initializing callback service
	callBack := service.NewCallBackService(cfg.PrivateKey,
		cfg.TokenTimeToLive, cfg.RetryCount, cfg.RetryWait, merchants)

	// Initializing processing services
	processingService := service.NewProcessingService(cfg.PublicKey, cfg.PrivateKey,
		cfg.TokenTimeToLive, processors, merchants, callBack, transactionStore)

	// Initializing user management service
	userService := user.NewService(userStore, merchants)
	assetService := asset.NewService(assetsStore, merchants)
	// register a new Ory client with the URL set to the Ory CLI Proxy
	// we can also read the URL from the env or a config file
	c := ory.NewConfiguration()
	c.Servers = ory.ServerConfigurations{{URL: cfg.KratosURL}}

	// Setting up API routing
	router := httprouter.New()
	urlPath := ""
	routing.InitRouter(ctx, ory.NewAPIClient(c), router, urlPath, processingService, userService, assetService)
	server := &http.Server{Addr: fmt.Sprintf(":%s", cfg.Port), Handler: router}
	log.Println("hello i am started at port:", cfg.Port)

	// Creating 2 streams one for API and another for blockchain requests
	var g run.Group
	g.Add(func() error {
		// Creating a process for API and initializing an API listener
		err := server.ListenAndServe()
		cancelFunc()
		return err
	}, func(err error) {
		cancelFunc()
	})
	g.Add(func() error {
		// Creating a process for blockchain requests and initializing a blockchain listener
		err := processingService.ListenAndServe(ctx, cfg.Interval)
		cancelFunc()
		_ = server.Shutdown(ctx)
		return err
	}, func(err error) {
		cancelFunc()
		_ = server.Shutdown(ctx)
	})
	// Shutdown
	g.Add(func() error {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, os.Kill)
		select {
		case c := <-sigChan:
			cancelFunc()
			return fmt.Errorf("interrupted with sig %q", c)
		case <-ctx.Done():
			cancelFunc()
			return nil
		}
	}, func(err error) {
		cancelFunc()
	})

	err = g.Run()
	if err != nil {
		log.Println("exit from processing, err:", err)
		return
	}

}
