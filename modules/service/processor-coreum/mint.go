package processor_coreum

import (
	"context"
	"coreum_processor/modules/service"
	"encoding/json"
	"fmt"
	"github.com/CoreumFoundation/coreum/v2/pkg/client"
	assetfttypes "github.com/CoreumFoundation/coreum/v2/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/v2/x/asset/nft/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"
)

func (s CoreumProcessing) MintFT(ctx context.Context,
	request service.MintTokenRequest, merchantID string) (*service.NewTokenResponse, error) {
	externalID := request.ReceivingWalletID

	_, addr, byteAddress, err := s.store.GetByUser(merchantID, fmt.Sprintf("%s-%s", merchantID, request.Code))
	if err != nil {
		return nil, fmt.Errorf("can't get issuer: %v-%v coreum wallet from store, err: %v",
			request.Code, merchantID, err)
	}

	wallet := service.Wallet{}
	err = json.Unmarshal(byteAddress, &wallet)
	if err != nil {
		return nil, err
	}
	if addr == "" || wallet.WalletAddress == "" || addr != request.Issuer {
		return nil, fmt.Errorf("empty or incorrect issuer wallet address")
	}

	amount, err := strconv.Atoi(request.Amount)
	if err != nil {
		return nil, err
	}
	denom := fmt.Sprintf("%s-%s", request.Code, request.Issuer)

	token, err := s.mintCoreumFT(ctx, merchantID, externalID, request.Code, "mint-send-"+denom,
		request.Issuer, wallet, int64(amount))
	if err != nil {
		return nil, fmt.Errorf("can't mint asset: %v, error: %w", denom, err)
	}

	_, addr, _, err = s.store.GetByUser(merchantID, externalID)
	if err != nil {
		return nil, fmt.Errorf("can't get user: %v coreum wallet from store, err: %v", externalID, err)
	}

	token, err = s.transferCoreumFT(ctx, merchantID, externalID, "mint-send-"+denom, request.Issuer, addr,
		denom, wallet, int64(amount))

	if err != nil {
		return nil, fmt.Errorf("can't transfer minted asset: %v, error: %w", denom, err)
	}

	return &service.NewTokenResponse{TxHash: token}, nil
}

func (s CoreumProcessing) MintNFT(ctx context.Context,
	request service.MintTokenRequest, merchantID string) (*service.NewTokenResponse, error) {
	_, address, byteAddress, err := s.store.GetByUser(merchantID, request.ReceivingWalletID)
	if err != nil {
		return nil, fmt.Errorf("can't get issuer: %v coreum wallet from store, err: %v",
			request.ReceivingWalletID, err)
	}

	wallet := service.Wallet{}
	err = json.Unmarshal(byteAddress, &wallet)
	if err != nil {
		return nil, err
	}
	if address == "" || wallet.WalletAddress == "" || wallet.WalletAddress != request.Issuer {
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

func (s CoreumProcessing) mintCoreumFT(ctx context.Context,
	merchantID, externalID, subunit, trxID, issuerAddress string,
	sendingWallet service.Wallet, amount int64) (string, error) {
	msgMint := &assetfttypes.MsgMint{
		Sender: issuerAddress,
		Coin:   sdk.Coin{Denom: subunit + "-" + issuerAddress, Amount: sdk.NewInt(amount)},
	}

	// update gas
	_, err := s.updateGas(ctx, issuerAddress, coerumFeeMintFT)
	if err != nil {
		return "", err
	}

	trx, err := s.broadcastTrx(ctx, merchantID, externalID, trxID, issuerAddress,
		sendingWallet, msgMint)

	if err != nil {
		return "", err
	}

	return trx.TxHash, err
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
