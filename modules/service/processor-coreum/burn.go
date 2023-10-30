package processor_coreum

import (
	"context"
	"coreum_processor/modules/service"
	"encoding/json"
	"fmt"
	"github.com/CoreumFoundation/coreum/v2/pkg/client"
	assetfttypes "github.com/CoreumFoundation/coreum/v2/x/asset/ft/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"
	"strings"
)

func (s CoreumProcessing) BurnToken(ctx context.Context, request service.TokenRequest,
	merchantID, externalID string) (*service.NewTokenResponse, error) {
	_, _, byteAddress, err := s.store.GetByUser(merchantID, fmt.Sprintf("%s-%s", merchantID, request.Code))
	if err != nil {
		return nil, fmt.Errorf("can't get user: %v coreum wallet from store, err: %v", externalID, err)
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
	token, err := s.burnCoreumToken(ctx, request.Code, request.Issuer, wallet.WalletSeed, int64(amount))
	if err != nil {
		return nil, err
	}
	return &service.NewTokenResponse{TxHash: token}, nil
}

func (s CoreumProcessing) burnCoreumToken(ctx context.Context, subunit, issuerAddress, mnemonic string, amount int64) (string, error) {
	msgBurn := &assetfttypes.MsgBurn{
		Sender: issuerAddress, Coin: sdk.Coin{Denom: strings.ToLower(subunit) + "-" + issuerAddress, Amount: sdk.NewInt(amount)}}
	// update gas
	_, err := s.updateGas(ctx, issuerAddress, coreumFeeBurnFT)
	if err != nil {
		return "", err
	}

	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		issuerAddress,
		mnemonic,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)

	if err != nil {
		return "", err
	}
	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()
	response, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(senderInfo.GetAddress()),
		s.factory,
		msgBurn,
	)
	if err != nil {
		return "", err
	}

	return response.TxHash, err
}
