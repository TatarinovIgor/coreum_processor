package internal

import (
	"coreum_processor/modules/service"
	"coreum_processor/modules/service/processor_coreum"
	"coreum_processor/modules/storage"
	"database/sql"
	"log"
)

// ToDo inti in the system
func InitProcessorCoreum(blockchain string, db *sql.DB) service.CryptoProcessor {
	var (
		minValue                 = GetFloat("MIN_VALUE", 10.0)
		WalletReceiverAddressStr = MustString("WALLET_RECEIVER_ADDRESS")
		WalletReceiverSeedStr    = MustString("WALLET_RECEIVER_SEED")
		WalletSenderAddressStr   = MustString("WALLET_SENDER_ADDRESS")
		WalletSenderSeedStr      = MustString("WALLET_SENDER_SEED")
	)

	store, err := storage.NewKeys("coreum_wallets", db)
	if err != nil {
		log.Fatalf("could not make store for Wallets, error: %v", err)
	}

	// Initializing Ethereum receiver as a structure
	WalletReceiver := service.Wallet{
		WalletAddress: WalletReceiverAddressStr,
		WalletSeed:    WalletReceiverSeedStr,
		Blockchain:    blockchain,
	}

	// Initializing Ethereum sender as a structure
	WalletSender := service.Wallet{
		WalletAddress: WalletSenderAddressStr,
		WalletSeed:    WalletSenderSeedStr,
		Blockchain:    blockchain,
	}
	return processor_coreum.NewCoreumCryptoProcessor(WalletSender, WalletReceiver, blockchain, store, float64(minValue))
}
