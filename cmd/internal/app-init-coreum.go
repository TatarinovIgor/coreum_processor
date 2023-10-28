package internal

import (
	"coreum_processor/modules/service"
	"coreum_processor/modules/service/processor-coreum"
	"coreum_processor/modules/storage"
	"database/sql"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"log"
)

const (
	testNodeAddress = "full-node.testnet-1.coreum.dev:9090"
	signMode        = signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
)

// InitProcessorCoreum initialize Coreum crypto processing
func InitProcessorCoreum(blockchain string, db *sql.DB, callBack *service.CallBacks) service.CryptoProcessor {
	var (
		chainID                  = GetString("COREUM_CHAIN_ID", string(constant.ChainIDTest))
		nodeAddress              = GetString("COREUM_NODE_ADDRESS", testNodeAddress)
		addressPrefix            = GetString("COREUM_ADDRESS_PREFIX", constant.AddressPrefixTest)
		denom                    = GetString("COREUM_ADDRESS_PREFIX", constant.DenomTest)
		minValue                 = GetFloat("MIN_VALUE", 10.0)
		WalletReceiverAddressStr = MustString("WALLET_RECEIVER_ADDRESS")
		WalletReceiverSeedStr    = MustString("WALLET_RECEIVER_SEED")
		WalletSenderAddressStr   = MustString("WALLET_SENDER_ADDRESS")
		WalletSenderSeedStr      = MustString("WALLET_SENDER_SEED")
	)

	// Initializing store for Coreum wallets
	store, err := storage.NewKeys("coreum_wallets", db)
	if err != nil {
		log.Fatalf("could not make store for Wallets, error: %v", err)
	}

	// Initializing Coreum receiver as a structure
	WalletReceiver := service.Wallet{
		WalletAddress: WalletReceiverAddressStr,
		WalletSeed:    WalletReceiverSeedStr,
		Blockchain:    blockchain,
	}

	// Initializing Coreum sender as a structure
	WalletSender := service.Wallet{
		WalletAddress: WalletSenderAddressStr,
		WalletSeed:    WalletSenderSeedStr,
		Blockchain:    blockchain,
	}
	return processor_coreum.NewCoreumCryptoProcessor(WalletSender, WalletReceiver, blockchain, store, float64(minValue),
		constant.ChainID(chainID), nodeAddress, addressPrefix, denom, signMode, callBack)
}
