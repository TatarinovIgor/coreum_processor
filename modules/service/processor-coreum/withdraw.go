package processor_coreum

import (
	"context"
	"coreum_processor/modules/service"
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (s CoreumProcessing) Withdraw(ctx context.Context,
	request service.CredentialWithdraw, merchantID, externalId, trxID string,
	merchantWallets service.Wallets) (*service.WithdrawResponse, error) {
	commission := 0.0
	if externalId != merchantWallets.ReceivingID || externalId != merchantWallets.SendingID {
		commission = merchantWallets.CommissionSending.Fix
		commission += merchantWallets.CommissionSending.Percent / 100. * (request.Amount - commission)
	}
	denom := s.denom
	if request.Asset == "" {
		request.Asset = s.denom
	} else {
		denom = request.Asset + "-" + request.Issuer
	}
	balance, err := s.GetAssetsBalance(ctx,
		service.BalanceRequest{Blockchain: request.Blockchain, Asset: request.Asset, Issuer: request.Issuer},
		merchantID, merchantWallets.SendingID)
	if err != nil || balance == nil {
		return nil, fmt.Errorf("can't get merchant: %v, sending wallet: %v, err: %w",
			merchantID, merchantWallets.SendingID, err)
	}
	if balance[0].Amount < request.Amount+commission {
		return nil, fmt.Errorf("merchant: %s, doesn't have enough balance to pay: %v %v, with commission: %v",
			merchantID, request.Amount, request.Asset, commission)
	}

	_, key, sendingWalletRaw, err := s.store.GetByUser(merchantID, merchantWallets.SendingID)
	if err != nil {
		return nil, err
	}

	sendingWallet := service.Wallet{}
	err = json.Unmarshal(sendingWalletRaw, &sendingWallet)
	if err != nil {
		return nil, err
	}
	//check gas
	_, err = s.updateGas(ctx, key, coreumFeeSendFT)
	if err != nil {
		return nil, err
	}

	msg := &banktypes.MsgSend{
		FromAddress: key,
		ToAddress:   s.sendingWallet.WalletAddress,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom, int64(request.Amount))),
	}
	result, err := s.broadcastTrx(ctx, merchantID, externalId, trxID, msg.FromAddress, sendingWallet, msg)
	if err != nil {
		return nil, err
	}

	return &service.WithdrawResponse{TransactionHash: result.TxHash}, nil
}
