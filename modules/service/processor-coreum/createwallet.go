package processor_coreum

import (
	"context"
	"coreum_processor/modules/service"
	"encoding/json"
	"fmt"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	amomultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
)

func (s CoreumProcessing) CreateWallet(ctx context.Context, merchantID, externalId string,
	multiSignAddresses service.FuncMultiSignAddrCallback) (*service.Wallet, error) {

	wallet := service.Wallet{Blockchain: s.blockchain}

	WalletSeed, WalletAddress, key, threshold, err := s.createCoreumWallet(ctx, externalId, multiSignAddresses)
	if err != nil {
		return nil, err
	}

	wallet.WalletAddress = WalletAddress
	wallet.WalletSeed = WalletSeed
	wallet.Threshold = threshold

	value, err := json.Marshal(wallet)
	if err != nil {
		return nil, err
	}

	_, err = s.store.Put(merchantID, externalId, key, value)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

func (s CoreumProcessing) createCoreumWallet(ctx context.Context, externalId string,
	multiSignAddresses service.FuncMultiSignAddrCallback) (string, string, string, float64, error) {
	threshold := 0.

	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", "", "", threshold, err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", "", "", threshold, err
	}

	Info, err := s.clientCtx.Keyring().NewAccount(
		"key-name",
		mnemonic,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)

	if err != nil {
		return "", "", "", threshold, err
	}
	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(Info.GetAddress()) }()
	walletAddress := Info.GetAddress().String()
	if multiSignAddresses != nil {
		addresses, threshold, err := multiSignAddresses(s.blockchain, externalId)
		if err != nil {
			return "", "", "", threshold, fmt.Errorf("can't create Coreum multising Wallet, error: %v", err)
		} else if addresses != nil && len(addresses) > 0 && threshold > 0 {
			signKeys := []types.PubKey{Info.GetPubKey()}
			// create multi sig wallet
			for key := range addresses {
				accAddress, err := sdk.AccAddressFromBech32(key)
				if err != nil {
					return "", "", "", threshold, fmt.Errorf(
						"can't create Coreum multising Wallet from provided key: %s, error: %w", key, err)
				}
				info, err := client.GetAccountInfo(ctx, s.clientCtx, accAddress)
				if err != nil {
					return "", "", "", threshold, fmt.Errorf(
						"can't create Coreum multising Wallet from info key: %s, error: %w", key, err)
				}
				signKeys = append(signKeys, info.GetPubKey())
			}
			pubKey := amomultisig.NewLegacyAminoPubKey(2, signKeys)
			Info, err := s.clientCtx.Keyring().SaveMultisig("multi-sign", pubKey)
			if err != nil {
				return "", "", "", threshold, fmt.Errorf("can't save Coreum multising Wallet, error: %v", err)
			}
			defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(Info.GetAddress()) }()
			// Validate address
			_, err = sdk.AccAddressFromBech32(Info.GetAddress().String())
			if err != nil {
				return "", "", "", threshold, fmt.Errorf("can't validate Coreum multising Wallet, error: %v", err)
			}
			return "", walletAddress, Info.GetAddress().String(), threshold, nil
		}
	}

	// Validate address
	_, err = sdk.AccAddressFromBech32(Info.GetAddress().String())
	if err != nil {
		return "", "", "", threshold, err
	}

	return mnemonic, walletAddress, Info.GetAddress().String(), threshold, nil
}
