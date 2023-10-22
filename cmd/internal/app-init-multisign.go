package internal

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

func LoadMultiSignEnv() MultiSignConfig {

	var (
		// Initializing ENV variable for listening port
		port = MustInt("PORT")
		// Initializing mnemonics for signature
		mnemonics = MustString("MNEMONICS")
		// Initializing public key for internal functions
		publicKeyPath = MustString("PUBLIC_KEY")
		networkType   = MustString("NETWORK_TYPE")
		threshold     = GetInt("THRESHOLD", 1)
	)

	if len(publicKeyPath) < 1 {
		log.Fatal("PUBLIC_KEY env variable must be set")
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

	return MultiSignConfig{
		Port:        fmt.Sprintf("%v", port),
		Threshold:   threshold,
		Mnemonics:   mnemonics,
		PublicKey:   public,
		NetworkType: networkType,
	}
}
