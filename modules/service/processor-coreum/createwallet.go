package processor_coreum

import (
	"context"
	"coreum_processor/modules/service"
	"encoding/json"
	"fmt"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	amomultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"
	"github.com/dvsekhvalnov/jose2go/base64url"
)

func (s CoreumProcessing) CreateWallet(ctx context.Context, merchantID, externalId string) (*service.Wallet, error) {

	wallet := service.Wallet{Blockchain: s.blockchain}

	walletSeed, walletAddress, key, threshold, err := s.createCoreumWallet(ctx, merchantID, externalId)
	if err != nil {
		return nil, err
	}

	wallet.WalletAddress = walletAddress
	wallet.WalletSeed = walletSeed
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

func (s CoreumProcessing) createCoreumWallet(ctx context.Context,
	merchantID, externalId string) (string, string, string, float64, error) {
	threshold := 0.
	algo := hd.Secp256k1
	hdPath := sdk.GetConfig().GetFullBIP44Path()

	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", "", "", threshold, fmt.Errorf(
			"could not create new entropy for externalid: %v, error: %w", externalId, err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", "", "", threshold, fmt.Errorf(
			"could not create new mnemonic for externalid: %v, error: %w", externalId, err)
	}

	// create master key and derive first key
	derivedPrivate, err := algo.Derive()(mnemonic, "", hdPath)
	if err != nil {
		return "", "", "", threshold, fmt.Errorf(
			"could not create master key and derive first key for externalid: %v, error: %w", externalId, err)
	}

	pKey := algo.Generate()(derivedPrivate)

	walletAddress := sdk.AccAddress(pKey.PubKey().Address()).String()
	signAddress := walletAddress
	callBackAdrFn, err := s.callBack.GetMultiSignAddressesFn(merchantID)
	if err != nil {
		return "", "", "", threshold, fmt.Errorf(
			"could not extract merchant: %v callback for multisign adresses, error: %w", merchantID, err)
	}
	callBackSignFn, err := s.callBack.GetMultiSignFn(merchantID)
	if err != nil {
		return "", "", "", threshold, fmt.Errorf(
			"could not extract merchant: %v callback for multisign signature, error: %w", merchantID, err)
	}
	callBackTrxFn, err := s.callBack.GetTransactionFn(merchantID)
	if err != nil {
		return "", "", "", threshold, fmt.Errorf(
			"could not extract merchant: %v callback for multisign signature, error: %w", merchantID, err)
	}
	if callBackAdrFn != nil && callBackSignFn != nil && callBackTrxFn != nil {
		addresses, threshold, err := callBackAdrFn(s.blockchain, externalId)
		if err != nil {
			return "", "", "", threshold, fmt.Errorf("can't create Coreum multising Wallet, error: %v", err)
		} else if addresses != nil && len(addresses) > 0 && threshold > 0 {
			signKeys := []types.PubKey{pKey.PubKey()}
			var signAddresses []string
			// create multi sig wallet
			for key := range addresses {
				pubData, err := base64url.Decode(key)
				if err != nil {
					return "", "", "", threshold, fmt.Errorf(
						"can't decode public key for Coreum multising Wallet, error: %w", err)
				}
				var acc secp256k1.PubKey
				err = acc.XXX_Unmarshal(pubData)
				if err != nil {
					return "", "", "", threshold, fmt.Errorf(
						"can't unmarshal public key for Coreum multising Wallet, error: %w", err)
				}
				signKeys = append(signKeys, &acc)
				signAddresses = append(signAddresses, sdk.AccAddress(acc.Address()).String())
			}
			pubKey := amomultisig.NewLegacyAminoPubKey(int(threshold+1), signKeys)
			info, err := s.clientCtx.Keyring().SaveMultisig(externalId, pubKey)
			if err != nil {
				return "", "", "", threshold, fmt.Errorf("can't save Coreum multising Wallet, error: %v", err)
			}
			defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(info.GetAddress()) }()

			// Validate address
			_, err = sdk.AccAddressFromBech32(info.GetAddress().String())
			if err != nil {
				return "", "", "", threshold, fmt.Errorf("can't validate Coreum multising Wallet, error: %v", err)
			}

			// top up newly created account for some amount
			// TODO: should be from config?
			amount := int64(102400)
			_, err = s.updateGas(ctx, info.GetAddress().String(), amount*5)
			if err != nil {
				return "", "", "", 0, fmt.Errorf("can't put gas for multisin account actiovation, err: %w", err)
			}

			// withdraw to activate
			msg := &banktypes.MsgSend{
				FromAddress: info.GetAddress().String(),
				ToAddress:   s.sendingWallet.WalletAddress,
				Amount:      sdk.NewCoins(sdk.NewInt64Coin(s.denom, amount)),
			}

			gasPrice, err := client.GetGasPrice(ctx, s.clientCtx)
			signMode := signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON

			txFactory := client.Factory{}.
				WithKeybase(s.clientCtx.Keyring()).
				WithChainID(s.clientCtx.ChainID()).
				WithTxConfig(s.clientCtx.TxConfig()).
				WithSignMode(signMode).
				WithGas(uint64(amount)).WithGasPrices(gasPrice.String()).
				WithSimulateAndExecute(true)
			unsignedTx, err := txFactory.BuildUnsignedTx(msg)
			if err != nil {
				return "", "", "", 0, fmt.Errorf("can't build multisign unsigned transaction, error: %w", err)
			}

			infoAcc, err := client.GetAccountInfo(ctx, s.clientCtx, info.GetAddress())
			if err != nil {
				return "", "", "", 0, fmt.Errorf(
					"can't get multisign account info to make signer data, error: %w", err)
			}

			sequence := infoAcc.GetSequence()
			accNumber := infoAcc.GetAccountNumber()
			signerData := xauthsigning.SignerData{
				ChainID:       txFactory.ChainID(),
				AccountNumber: accNumber,
				Sequence:      sequence,
			}
			trxData, err := s.clientCtx.TxConfig().SignModeHandler().GetSignBytes(signMode,
				signerData, unsignedTx.GetTx())
			if err != nil {
				return "", "", "", 0, fmt.Errorf("can't make transaction data for signature, error: %w", err)
			}

			ms := multisig.NewMultisig(int(pubKey.Threshold))
			// sign by processing
			sign, err := pKey.Sign(trxData)
			if err != nil {
				return "", "", "", 0, fmt.Errorf("can't sing transaction data by processing, error: %w", err)
			}
			sigData1 := signing.SingleSignatureData{
				SignMode:  signMode,
				Signature: sign,
			}
			sigV2 := signing.SignatureV2{
				PubKey:   pKey.PubKey(),
				Data:     &sigData1,
				Sequence: sequence,
			}
			err = multisig.AddSignatureV2(ms, sigV2, signKeys)
			if err != nil {
				return "", "", "", 0, fmt.Errorf(
					"can't add signature from processing signing account, error: %w", err)
			}
			request := service.MultiSignTransactionRequest{
				ExternalID: externalId,
				Blockchain: s.blockchain,
				Addresses:  signAddresses,
				TrxID:      "",
				TrxData:    base64url.Encode(trxData),
				Threshold:  threshold,
			}

			// get signature from callback
			signatures, err := callBackSignFn(request)
			if err != nil {
				return "", "", "", 0, fmt.Errorf(
					"can't get signatures from signing account, error: %w", err)
			}

			for key, sign := range signatures {
				pubData, err := base64url.Decode(key)
				if err != nil {
					return "", "", "", threshold, fmt.Errorf(
						"can't decode public key for Coreum multising signing, error: %v", err)
				}
				var acc secp256k1.PubKey
				err = acc.XXX_Unmarshal(pubData)
				if err != nil {
					return "", "", "", threshold, fmt.Errorf(
						"can't unmarshal public key for Coreum multising signing, error: %v", err)
				}
				sigData1 := signing.SingleSignatureData{
					SignMode:  signMode,
					Signature: sign,
				}
				sigV2 := signing.SignatureV2{
					PubKey:   &acc,
					Data:     &sigData1,
					Sequence: sequence,
				}
				err = multisig.AddSignatureV2(ms, sigV2, signKeys)
				if err != nil {
					return "", "", "", 0, fmt.Errorf(
						"can't add signature from multi signing account, error: %w", err)
				}
			}

			err = unsignedTx.SetSignatures([]signing.SignatureV2{{
				PubKey:   pubKey,
				Data:     &signing.MultiSignatureData{Signatures: ms.Signatures, BitArray: ms.BitArray},
				Sequence: sequence,
			}}...)
			if err != nil {
				fmt.Println("can't set signatures for transaction, error:", err)
				return "", "", "", 0, fmt.Errorf("can't set signatures for transaction, error: %w", err)
			}

			txBytes, err := s.clientCtx.TxConfig().TxEncoder()(unsignedTx.GetTx())
			if err != nil {
				return "", "", "", 0, fmt.Errorf("can't get transaction bytes for broadcast, error: %w", err)
			}

			// Broadcast signed transaction
			_, err = client.BroadcastRawTx(ctx, s.clientCtx, txBytes)
			if err != nil {
				return "", "", "", 0, fmt.Errorf("can't broadcast transaction, error: %w", err)
			}
			return mnemonic, walletAddress, info.GetAddress().String(), threshold, nil
		}
	}

	// Validate address
	_, err = sdk.AccAddressFromBech32(signAddress)
	if err != nil {
		return "", "", "", threshold, err
	}

	return mnemonic, walletAddress, signAddress, threshold, nil
}
