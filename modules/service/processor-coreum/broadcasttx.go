package processor_coreum

import (
	"context"
	"coreum_processor/modules/service"
	"fmt"
	"github.com/CoreumFoundation/coreum/v2/pkg/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	amomultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/dvsekhvalnov/jose2go/base64url"
)

func (s CoreumProcessing) broadcastTrx(ctx context.Context, merchantID, externalID, trxID, fromAddr string,
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
	var pubKey *amomultisig.LegacyAminoPubKey
	ok := false
	if info.GetPubKey() != nil {
		pubKey, ok = info.GetPubKey().(*amomultisig.LegacyAminoPubKey)
	}
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
			return nil, fmt.Errorf("can't add account: %v to keyring for broadcast, error: %w",
				sendingWallet.WalletAddress, err)
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
		var signAddresses []string
		for _, key := range pubKey.GetPubKeys() {
			addr, _ := bech32.ConvertAndEncode(s.addressPrefix, key.Address())
			signAddresses = append(signAddresses, addr)
		}

		sequence := info.GetSequence()
		accountNumber := info.GetAccountNumber()
		signerData := xauthsigning.SignerData{
			ChainID:       s.factory.ChainID(),
			AccountNumber: accountNumber,
			Sequence:      sequence,
		}

		signMode := s.factory.SignMode()
		gasPrice, err := client.GetGasPrice(ctx, s.clientCtx)
		if err != nil {
			return nil, fmt.Errorf("can't define gas price for multisign transaction, error: %w", err)
		}

		// TODO: gas -???
		unsignedTx, err := s.factory.WithGas(uint64(124000)).WithGasPrices(gasPrice.String()).BuildUnsignedTx(msg)
		if err != nil {
			return nil, fmt.Errorf("can't buiild multisign transaction, error: %w", err)
		}
		trxData, err := s.clientCtx.TxConfig().SignModeHandler().GetSignBytes(signMode,
			signerData, unsignedTx.GetTx())
		if err != nil {
			return nil, fmt.Errorf("can't make transaction data for signature, error: %w", err)
		}
		ms := multisig.NewMultisig(int(pubKey.Threshold))
		// set sign from internal wallet
		derivedPriv, err := hd.Secp256k1.Derive()(sendingWallet.WalletSeed, "",
			sdk.GetConfig().GetFullBIP44Path())
		if err != nil {
			return nil, fmt.Errorf("can't make multisign signature private key for wallet: %v, error: %w",
				fromAddr, err)
		}
		privKey := hd.Secp256k1.Generate()(derivedPriv)
		sign, err := privKey.Sign(trxData)
		if err != nil {
			return nil, fmt.Errorf("can't sing transaction data by processing, error: %w", err)
		}
		sigData1 := signing.SingleSignatureData{
			SignMode:  signMode,
			Signature: sign,
		}
		sigV2 := signing.SignatureV2{
			PubKey:   privKey.PubKey(),
			Data:     &sigData1,
			Sequence: sequence,
		}
		err = multisig.AddSignatureV2(ms, sigV2, pubKey.GetPubKeys())
		if err != nil {
			return nil, fmt.Errorf("can't add signature from processing signing account, error: %w", err)
		}

		mRequest := service.MultiSignTransactionRequest{
			ExternalID: externalID,
			Blockchain: s.blockchain,
			Addresses:  signAddresses,
			TrxID:      trxID,
			TrxData:    base64url.Encode(trxData),
			Threshold:  float64(pubKey.Threshold - 1),
		}
		signatures, err := callBackSignFn(mRequest)
		if err != nil {
			return nil, fmt.Errorf("can't get multisign signature, error: %w", err)
		}

		// set signs from external wallets
		for key, sign := range signatures {
			pubData, err := base64url.Decode(key)
			if err != nil {
				return nil, fmt.Errorf(
					"can't decode public key for Coreum multising signing, error: %v", err)
			}
			var acc secp256k1.PubKey
			err = acc.XXX_Unmarshal(pubData)
			if err != nil {
				return nil, fmt.Errorf("can't unmarshal public key for Coreum multising signing, error: %v", err)
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
			err = multisig.AddSignatureV2(ms, sigV2, pubKey.GetPubKeys())
			if err != nil {
				return nil, fmt.Errorf("can't add signature from multi signing account, error: %w", err)
			}
		}
		signData := signing.MultiSignatureData{Signatures: ms.Signatures, BitArray: ms.BitArray}
		sigV2.Data = &signData
		sigV2.PubKey = pubKey
		err = unsignedTx.SetSignatures(sigV2)
		if err != nil {
			return nil, fmt.Errorf("can't set signature for wallet: %s, error: %w",
				fromAddr, err)
		}
		txBytes, err := s.clientCtx.TxConfig().TxEncoder()(unsignedTx.GetTx())
		if err != nil {
			return nil, fmt.Errorf("can't get transaction bytes for broadcast, error: %w", err)
		}
		txHash, err := client.BroadcastRawTx(ctx, s.clientCtx, txBytes)
		if err != nil {
			return nil, fmt.Errorf("can't broadcast transaction, error: %w", err)
		}
		return txHash, err
	}
	return nil, fmt.Errorf("multisign callback is not defined for merhcant: %v", merchantID)
}
