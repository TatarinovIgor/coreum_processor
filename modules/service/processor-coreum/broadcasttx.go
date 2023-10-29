package processor_coreum

import (
	"context"
	"coreum_processor/modules/service"
	"fmt"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	amomultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/dvsekhvalnov/jose2go/base64url"
)

func (s CoreumProcessing) broadcastTrx(ctx context.Context, merchantID, externalID, fromAddr, trxID string,
	sendingWallet service.Wallet, msg sdk.Msg) (*sdk.TxResponse, error) {

	// define sending address type
	accAddress, err := sdk.AccAddressFromBech32(fromAddr)
	if err != nil {
		return nil, fmt.Errorf("can't get address for account: %v, error: %w", fromAddr, err)
	}
	info, err := client.GetAccountInfo(ctx, s.clientCtx, accAddress)
	if err != nil {
		return nil, fmt.Errorf("can't get info for account: %v, error: %w", fromAddr, err)
	}
	if info.GetPubKey() == nil {
		return nil, fmt.Errorf("can't get public key for account: %v", fromAddr)

	}

	pubKey, ok := info.GetPubKey().(*amomultisig.LegacyAminoPubKey)
	if !ok {
		// not multisign account
		senderInfo, err := s.clientCtx.Keyring().NewAccount(
			sendingWallet.WalletAddress,
			sendingWallet.WalletSeed,
			"",
			sdk.GetConfig().GetFullBIP44Path(),
			hd.Secp256k1,
		)
		if err != nil {
			return nil, err
		}
		defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()
		bech32, err := sdk.AccAddressFromBech32(sendingWallet.WalletAddress)
		if err != nil {
			return nil, err
		}

		return client.BroadcastTx(ctx, s.clientCtx.WithFromAddress(bech32), s.factory, msg)
	}

	// multisign account process
	callBackSignFn, err := s.callBack.GetMultiSignFn(merchantID)
	if err != nil {
		return nil, fmt.Errorf(
			"could not extract merchant: %v callback for multisign signature, error: %w", merchantID, err)
	}
	if callBackSignFn != nil {
		var addresses []string
		for _, key := range pubKey.GetPubKeys() {
			addr, _ := bech32.ConvertAndEncode(s.addressPrefix, key.Address())
			addresses = append(addresses, addr)
		}
		mRequest := service.MultiSignTransactionRequest{
			ExternalID: externalID,
			Blockchain: s.blockchain,
			Addresses:  addresses,
			TrxID:      trxID,
			Threshold:  0,
		}
		signatures, err := callBackSignFn(mRequest)
		if err != nil {
			return nil, fmt.Errorf("can't get multisign signature, error: %w", err)
		}

		unsignedTx, err := s.factory.BuildUnsignedTx(msg)
		if err != nil {
			return nil, fmt.Errorf("can't buiild multisign transaction, error: %w", err)
		}
		adr, err := sdk.AccAddressFromBech32(fromAddr)
		if err != nil {
			return nil, fmt.Errorf("can't make multisign address, error: %w", err)
		}
		info, err := client.GetAccountInfo(ctx, s.clientCtx, adr)
		if err != nil {
			return nil, fmt.Errorf("unacceptable key format to get info for address: %s, error: %w", adr, err)
		}
		pubKey, ok := info.GetPubKey().(*amomultisig.LegacyAminoPubKey)
		if !ok {
			return nil, fmt.Errorf("unacceptable key format for address: %s", adr)
		}

		signerData := xauthsigning.SignerData{
			ChainID:       s.factory.ChainID(),
			AccountNumber: info.GetAccountNumber(),
			Sequence:      info.GetSequence(),
		}
		signMode := s.factory.SignMode()
		data, err := s.clientCtx.TxConfig().SignModeHandler().GetSignBytes(signMode, signerData,
			unsignedTx.GetTx())
		if err != nil {
			return nil, fmt.Errorf("can't make multisign transaction bytes, error: %w", err)
		}
		mRequest.TrxData = base64url.Encode(data)
		ms := multisig.NewMultisig(int(sendingWallet.Threshold + 1))
		sigV2 := signing.SignatureV2{
			Sequence: info.GetSequence(),
		}
		// set sign from internal wallet
		derivedPriv, err := hd.Secp256k1.Derive()(sendingWallet.WalletSeed, "",
			sdk.GetConfig().GetFullBIP44Path())
		if err != nil {
			return nil, fmt.Errorf("can't make multisign signature private key for wallet: %v, error: %w",
				fromAddr, err)
		}
		privKey := hd.Secp256k1.Generate()(derivedPriv)
		sigV2, err = tx.SignWithPrivKey(
			signMode, signerData,
			unsignedTx, privKey, s.clientCtx.TxConfig(), signerData.Sequence)
		if err != nil {
			return nil, fmt.Errorf("can't make multisign signature for wallet: %v, error: %w",
				fromAddr, err)
		}
		sendingAdr, err := sdk.AccAddressFromBech32(sendingWallet.WalletAddress)
		if err != nil {
			return nil, fmt.Errorf("can't make multisign sending address, error: %w", err)
		}
		sendingInfo, err := client.GetAccountInfo(ctx, s.clientCtx, sendingAdr)
		if err != nil {
			return nil, fmt.Errorf("unacceptable key format to get info for sending address: %s, error: %w",
				adr, err)
		}
		sigV2.PubKey = sendingInfo.GetPubKey()
		err = multisig.AddSignatureV2(ms, sigV2, pubKey.GetPubKeys())
		if err != nil {
			return nil, fmt.Errorf("can't added signature of: %s for wallet: %s, error: %w",
				sendingWallet.WalletAddress, adr, err)
		}
		// set signs from external wallets
		for sign, data := range signatures {
			adr, err := sdk.AccAddressFromBech32(sign)
			if err != nil {
				return nil, fmt.Errorf("can't make multisign address, error: %w", err)
			}
			info, err := client.GetAccountInfo(ctx, s.clientCtx, adr)
			if err != nil {
				return nil, fmt.Errorf("unacceptable key format to get info for address: %s", adr)
			}
			sigV2.Data = &signing.SingleSignatureData{
				SignMode:  signMode,
				Signature: data,
			}
			sigV2.PubKey = info.GetPubKey()
			err = multisig.AddSignatureV2(ms, sigV2, pubKey.GetPubKeys())
			if err != nil {
				return nil, fmt.Errorf("can't added signature of: %s for wallet: %s, error: %w",
					sign, adr, err)
			}
		}
		signData := signing.MultiSignatureData{Signatures: ms.Signatures, BitArray: ms.BitArray}
		sigV2.Data = &signData
		sigV2.PubKey = pubKey
		err = unsignedTx.SetSignatures(sigV2)
		if err != nil {
			return nil, fmt.Errorf("can't set signature for wallet: %s, error: %w",
				adr, err)
		}
		txBytes, err := s.clientCtx.TxConfig().TxEncoder()(unsignedTx.GetTx())
		return client.BroadcastRawTx(ctx, s.clientCtx, txBytes)
	}
	return nil, fmt.Errorf("multisign callback is not defined for merhcant: %v", merchantID)
}
