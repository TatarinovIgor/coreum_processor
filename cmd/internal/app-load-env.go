package internal

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"
)

func MakeConfigFromEnv() AppConfig {

	var (
		// Initializing ENV variable for listening port
		port = MustInt("PORT")
		// Initializing ENV variable for JWT time to live
		tokenTimeToLive = MustInt("TOKEN_TIME_TO_LIVE")
		// Initializing private key for internal functions
		privateKeyPath = MustString("PRIVATE_KEY")
		// Initializing public key for internal functions
		publicKeyPath = MustString("PUBLIC_KEY")
		// Initializing interval of checking transactions
		interval = GetInt("LISTEN_AND_SERVE_INTERVAL", 15)
		// Initializing number of retry for callbacks
		retryCount = GetInt("RETRY_COUNT", 15)
		// Initializing time waite interval in sec betweend callbacks
		retryWait = GetInt("RETRY_COUNT", 30)
		// Initializing ENV variable for kratos url
		kratosURL = MustString("KRATOS_URL")
	)

	if len(publicKeyPath) < 1 {
		log.Fatal("PUBLIC_KEY env variable must be set")
	}
	if len(privateKeyPath) < 1 {
		log.Fatal("PRIVATE_KEY env variable must be set")
	}

	// Parsing public key
	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Fatalf("could not read public key: %s, error: %v", publicKeyPath, err)
	}
	block, _ := pem.Decode(publicKey)
	if block == nil {
		panic("failed to parse PEM block containing the private key")
	}
	publicInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Fatalf("could not parse public key env variable: %s, error: %v", publicKeyPath, err)
	}
	public, ok := publicInterface.(*rsa.PublicKey)
	if !ok {
		log.Fatalf("public key is not of type *rsa.PublicKey")
	}

	// Parsing private key
	privateKey, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalf("could not read private key: %s, error: %v", privateKeyPath, err)
	}
	block, _ = pem.Decode(privateKey)
	if block == nil {
		panic("failed to parse PEM block containing the private key")
	}
	private, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Fatalf("could not parse private key env variable: %s, error: %v", privateKeyPath, err)
	}

	return AppConfig{
		Port:            fmt.Sprintf("%v", port),
		TokenTimeToLive: tokenTimeToLive,
		PrivateKey:      private,
		PublicKey:       public,
		Interval:        time.Duration(interval),
		RetryCount:      retryCount,
		RetryWait:       retryWait,
		KratosURL:       kratosURL,
	}
}
