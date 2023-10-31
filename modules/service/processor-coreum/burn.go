package processor_coreum

import (
	"context"
	"coreum_processor/modules/service"
	"encoding/json"
	"fmt"
	assetfttypes "github.com/CoreumFoundation/coreum/v2/x/asset/ft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"
	"strings"
)

func (s CoreumProcessing) BurnFT(ctx context.Context, request service.TokenRequest,
	merchantID, externalID string) (*service.NewTokenResponse, error) {
	_, key, byteAddress, err := s.store.GetByUser(merchantID, fmt.Sprintf("%s-%s", merchantID, request.Code))
	if err != nil {
		return nil, fmt.Errorf("can't get user: %v coreum wallet from store to burn: %v, err: %w",
			externalID, request.Code, err)
	}
	wallet := service.Wallet{}
	err = json.Unmarshal(byteAddress, &wallet)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal wallet: %v to burn: %v, error: %w", key, request.Code, err)
	}
	if wallet.WalletAddress == "" || key != request.Issuer {
		return nil, fmt.Errorf("empty or incorrect issuer wallet address")
	}
	amount, err := strconv.Atoi(request.Amount)
	if err != nil {
		return nil, fmt.Errorf("can't define amount: %v  to burn: %v for: %v, err: %w",
			request.Amount, request.Code, request.Issuer, err)
	}
	denom := fmt.Sprintf("%s-%s", request.Code, request.Issuer)

	txHash, err := s.burnCoreumFT(ctx, merchantID, externalID, request.Code, "burn-"+denom,
		request.Issuer, wallet, int64(amount))
	if err != nil {
		return nil, fmt.Errorf("can't burn coreum FT, error: %w", err)
	}
	return &service.NewTokenResponse{TxHash: txHash}, nil
}

func (s CoreumProcessing) burnCoreumFT(ctx context.Context,
	merchantID, externalID, subunit, trxID, issuerAddress string,
	sendingWallet service.Wallet, amount int64) (string, error) {
	msgBurn := &assetfttypes.MsgBurn{
		Sender: issuerAddress,
		Coin:   sdk.Coin{Denom: strings.ToLower(subunit) + "-" + issuerAddress, Amount: sdk.NewInt(amount)}}

	// update gas
	_, err := s.updateGas(ctx, issuerAddress, coreumFeeBurnFT)
	if err != nil {
		return "", fmt.Errorf("can't update gas to burn: %v, error: %w", msgBurn.Coin.Denom, err)
	}

	trx, err := s.broadcastTrx(ctx, merchantID, externalID, trxID, issuerAddress,
		sendingWallet, msgBurn)
	if err != nil {
		return "", fmt.Errorf("can't broadcast transaction to burn: %v, error: %w", msgBurn.Coin.Denom, err)
	}

	return trx.TxHash, err
}
