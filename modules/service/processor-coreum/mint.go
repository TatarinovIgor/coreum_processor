package processor_coreum

import (
	"context"
	"coreum_processor/modules/service"
	"encoding/json"
	"fmt"
	"github.com/CoreumFoundation/coreum/pkg/client"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/x/asset/nft/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"
)

func (s CoreumProcessing) MintFT(ctx context.Context,
	request service.MintTokenRequest, merchantID string) (*service.NewTokenResponse, error) {
	externalID := request.ReceivingWalletID

	_, byteAddress, err := s.store.GetByUser(merchantID, fmt.Sprintf("%s-%s", merchantID, request.Code))
	if err != nil {
		return nil, fmt.Errorf("can't get issuer: %v-%v coreum wallet from store, err: %v",
			request.Code, merchantID, err)
	}

	wallet := service.Wallet{}
	err = json.Unmarshal(byteAddress, &wallet)
	if err != nil {
		return nil, err
	}
	if wallet.WalletAddress == "" || wallet.WalletAddress != request.Issuer {
		return nil, fmt.Errorf("empty or incorrect issuer wallet address")
	}

	amount, err := strconv.Atoi(request.Amount)
	if err != nil {
		return nil, err
	}

	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		wallet.WalletAddress,
		wallet.WalletSeed,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()
	if err != nil {
		return nil, err
	}

	token, err := s.mintCoreumFT(ctx, request.Code, request.Issuer, int64(amount))
	if err != nil {
		return nil, err
	}

	_, byteAddress, err = s.store.GetByUser(merchantID, externalID)
	if err != nil {
		return nil, fmt.Errorf("can't get user: %v coreum wallet from store, err: %v", externalID, err)
	}

	err = json.Unmarshal(byteAddress, &wallet)
	if err != nil {
		return nil, err
	}

	token, err = s.transferCoreumFT(ctx, request.Issuer, wallet.WalletAddress,
		fmt.Sprintf("%s-%s", request.Code, request.Issuer), int64(amount))
	if err != nil {
		return nil, err
	}

	return &service.NewTokenResponse{TxHash: token}, nil
}

func (s CoreumProcessing) MintNFT(ctx context.Context,
	request service.MintTokenRequest, merchantID string) (*service.NewTokenResponse, error) {
	_, byteAddress, err := s.store.GetByUser(merchantID, request.ReceivingWalletID)
	if err != nil {
		return nil, fmt.Errorf("can't get issuer: %v coreum wallet from store, err: %v",
			request.ReceivingWalletID, err)
	}

	wallet := service.Wallet{}
	err = json.Unmarshal(byteAddress, &wallet)
	if err != nil {
		return nil, err
	}
	if wallet.WalletAddress == "" || wallet.WalletAddress != request.Issuer {
		return nil, fmt.Errorf("empty or incorrect issuer wallet address")
	}

	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		wallet.WalletAddress,
		wallet.WalletSeed,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()
	if err != nil {
		return nil, err
	}

	token, err := s.mintCoreumNFT(ctx, request.ClassID, request.Issuer, request.NftId)
	if err != nil {
		return nil, err
	}

	return &service.NewTokenResponse{TxHash: token}, nil
}

func (s CoreumProcessing) mintCoreumFT(ctx context.Context, subunit, issuerAddress string, amount int64) (string, error) {
	msgMint := &assetfttypes.MsgMint{
		Sender: issuerAddress,
		Coin:   sdk.Coin{Denom: subunit + "-" + issuerAddress, Amount: sdk.NewInt(amount)},
	}
	address, err := sdk.AccAddressFromBech32(issuerAddress)
	if err != nil {
		return "", err
	}
	senderInfo, err := s.clientCtx.Keyring().KeyByAddress(address)
	if err != nil {
		return "", err
	}
	// update gas
	_, err = s.updateGas(ctx, issuerAddress, coerumFeeMintFT)
	if err != nil {
		return "", err
	}

	response, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(senderInfo.GetAddress()),
		s.factory,
		msgMint,
	)
	if err != nil {
		return "", err
	}

	return response.TxHash, err
}

func (s CoreumProcessing) mintCoreumNFT(ctx context.Context, classId, issuerAddress, nftId string) (string, error) {
	address, err := sdk.AccAddressFromBech32(issuerAddress)
	if err != nil {
		return "", err
	}
	senderInfo, err := s.clientCtx.Keyring().KeyByAddress(address)
	if err != nil {
		return "", err
	}

	if err != nil {
		return "", err
	}

	classID := assetnfttypes.BuildClassID(classId, senderInfo.GetAddress())

	msgMint := &assetnfttypes.MsgMint{
		Sender:  senderInfo.GetAddress().String(),
		ClassID: classID,
		ID:      nftId,
	}

	trx, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(senderInfo.GetAddress()),
		s.factory,
		msgMint,
	)
	if err != nil {
		return "", err
	}

	return trx.TxHash, err
}