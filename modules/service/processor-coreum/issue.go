package processor_coreum

import (
	"context"
	"coreum_processor/modules/service"
	"coreum_processor/modules/storage"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CoreumFoundation/coreum/pkg/client"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/x/asset/nft/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"
)

func (s CoreumProcessing) IssueNFTClass(ctx context.Context, request service.NewTokenRequest,
	merchantID, externalId string, multiSignAddresses service.FuncMultiSignAddrCallback,
	multiSignature service.FuncMultiSignSignature) (*service.NewTokenResponse, []byte, error) {
	wallet := service.Wallet{}

	_, _, byteAddress, err := s.store.GetByUser(merchantID, externalId)
	key := ""
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return nil, nil, fmt.Errorf("can't get user: %v coreum wallet from store, err: %v", externalId, err)
	} else if errors.Is(err, storage.ErrNotFound) {
		// create issuer
		wallet.WalletSeed, wallet.WalletAddress, key, wallet.Threshold, err = s.createCoreumWallet(ctx,
			externalId, multiSignAddresses, multiSignature)
		if err != nil {
			return nil, nil, err
		}

		wallet.Blockchain = request.Blockchain
		value, err := json.Marshal(wallet)
		if err != nil {
			return nil, nil, err
		}

		_, err = s.store.Put(merchantID, externalId, key, value)
		if err != nil {
			return nil, nil, err
		}
	} else {
		err = json.Unmarshal(byteAddress, &wallet)
		if err != nil {
			return nil, nil, err
		}
		if wallet.WalletAddress == "" {
			return nil, nil, fmt.Errorf("empty wallet address")
		}
	}
	_, err = s.updateGas(ctx, wallet.WalletAddress, coreumFeeMintNFT+coreumFeeIssueNFT)
	if err != nil {
		return nil, nil, err
	}
	token, features, err := s.createCoreumNFTClass(request.Symbol, request.Code, request.Issuer, request.Description,
		wallet.WalletSeed)
	if err != nil {
		return nil, nil, err
	}

	return &service.NewTokenResponse{TxHash: token, Issuer: wallet.WalletAddress}, features, nil
}

func (s CoreumProcessing) IssueFT(ctx context.Context, request service.NewTokenRequest,
	merchantID, externalId string, multiSignAddresses service.FuncMultiSignAddrCallback,
	multiSignature service.FuncMultiSignSignature) (*service.NewTokenResponse, []byte, error) {
	wallet := service.Wallet{}

	issuerId := fmt.Sprintf("%s-%s", merchantID, request.Code)
	_, _, byteAddress, err := s.store.GetByUser(merchantID, issuerId)
	key := ""
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return nil, nil, fmt.Errorf("can't get user: %v coreum wallet from store, err: %v", externalId, err)
	} else if errors.Is(err, storage.ErrNotFound) {
		// create issuer
		wallet.WalletSeed, wallet.WalletAddress, key, wallet.Threshold, err = s.createCoreumWallet(ctx,
			externalId, multiSignAddresses, multiSignature)
		if err != nil {
			return nil, nil, err
		}

		wallet.Blockchain = request.Blockchain
		value, err := json.Marshal(wallet)
		if err != nil {
			return nil, nil, err
		}

		_, err = s.store.Put(merchantID, issuerId, key, value)
		if err != nil {
			return nil, nil, err
		}
	} else {
		err = json.Unmarshal(byteAddress, &wallet)
		if err != nil {
			return nil, nil, err
		}
		if wallet.WalletAddress == "" {
			return nil, nil, fmt.Errorf("empty wallet address")
		}
	}
	_, err = s.updateGas(ctx, wallet.WalletAddress, coreumFeeIssueFT)
	if err != nil {
		return nil, nil, err
	}
	token, features, err := s.createCoreumFT(ctx, request.Symbol, request.Code, request.Issuer, request.Description,
		wallet.WalletSeed, request.InitialAmount)
	if err != nil {
		return nil, nil, err
	}

	return &service.NewTokenResponse{TxHash: token, Issuer: wallet.WalletAddress}, features, nil
}

func (s CoreumProcessing) createCoreumFT(ctx context.Context,
	symbol, subunit, issuerAddress, description, mnemonic string,
	initialAmount int64) (string, []byte, error) {

	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		issuerAddress,
		mnemonic,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)

	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()

	if err != nil {
		return "", nil, err
	}

	features := []assetfttypes.Feature{assetfttypes.Feature_minting, assetfttypes.Feature_burning}

	msgIssue := &assetfttypes.MsgIssue{
		Issuer:        senderInfo.GetAddress().String(),
		Symbol:        symbol,
		Subunit:       strings.ToLower(subunit),
		Precision:     6,
		InitialAmount: sdk.NewInt(initialAmount),
		Description:   description,
		Features:      features,
	}

	trx, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(senderInfo.GetAddress()),
		s.factory,
		msgIssue,
	)
	if err != nil {
		return "", nil, err
	}

	featuresJson, err := json.Marshal(features)
	if err != nil {
		return "", nil, err
	}

	return trx.TxHash, featuresJson, err
}

func (s CoreumProcessing) createCoreumNFTClass(symbol, name, issuerAddress, description, mnemonic string) (string, []byte, error) {

	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		issuerAddress,
		mnemonic,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)

	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()

	if err != nil {
		return "", nil, err
	}

	features := []assetnfttypes.ClassFeature{assetnfttypes.ClassFeature_burning}

	msgIssue := &assetnfttypes.MsgIssueClass{
		Issuer:      senderInfo.GetAddress().String(),
		Symbol:      symbol,
		Name:        name,
		Description: description,
		Features:    features,
	}

	ctx := context.Background()
	trx, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(senderInfo.GetAddress()),
		s.factory,
		msgIssue,
	)
	if err != nil {
		return "", nil, err
	}

	featuresJson, err := json.Marshal(features)
	if err != nil {
		return "", nil, err
	}

	return trx.TxHash, featuresJson, err
}
