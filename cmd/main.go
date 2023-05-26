package main

import (
	"context"
	internalApp "coreum_processor/cmd/internal"
	"coreum_processor/modules/routing"
	"coreum_processor/modules/service"
	"coreum_processor/modules/storage"
	user "coreum_processor/modules/user"
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

	// Adding processors to the unified structure
	processors := map[string]service.CryptoProcessor{
		internalApp.Coreum: internalApp.InitProcessorCoreum(internalApp.Coreum, db),
	}

	// Initializing merchant management service
	merchants := service.NewMerchantService(merchantsStore)

	// Initializing processing services
	processingService := service.NewProcessingService(cfg.PublicKey, cfg.PrivateKey,
		cfg.TokenTimeToLive, processors, merchants, transactionStore)

	// Initializing user management service
	userService := user.NewService(userStore, merchants)
	// register a new Ory client with the URL set to the Ory CLI Proxy
	// we can also read the URL from the env or a config file
	c := ory.NewConfiguration()
	c.Servers = ory.ServerConfigurations{{URL: cfg.KratosURL}}

	// Setting up API routing
	router := httprouter.New()
	urlPath := ""
	log.Println("hello i am started")
	routing.InitRouter(ctx, ory.NewAPIClient(c), router, urlPath, processingService, userService)

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
