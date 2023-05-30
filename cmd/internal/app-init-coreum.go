package internal

import (
	"coreum_processor/modules/service"
	"coreum_processor/modules/service/processor-coreum"
	"coreum_processor/modules/storage"
	"database/sql"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"log"
)

const (
	senderMnemonic    = "unit resource ramp note attitude allow pipe hollow above kingdom siren social bless crystal student appear today orchard drive prosper during report burden film" // put mnemonic here
	testChainID       = constant.ChainIDTest
	testAddressPrefix = constant.AddressPrefixTest
	testNodeAddress   = "full-node.testnet-1.coreum.dev:9090"
)

// ToDo inti in the system
func InitProcessorCoreum(blockchain string, db *sql.DB) service.CryptoProcessor {
	var (
		chainID                  = GetString("COREUM_CHAIN_ID", string(testChainID))
		nodeAddress              = GetString("COREUM_NODE_ADDRESS", testNodeAddress)
		addressPrefix            = GetString("COREUM_ADDRESS_PREFIX", testAddressPrefix)
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
	return processor_coreum.NewCoreumCryptoProcessor(WalletSender, WalletReceiver, blockchain, store, float64(minValue),
		constant.ChainID(chainID), nodeAddress, addressPrefix, senderMnemonic)
}
