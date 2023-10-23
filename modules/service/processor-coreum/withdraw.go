package processor_coreum

import (
	"context"
	"coreum_processor/modules/service"
	"encoding/json"
	"fmt"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (s CoreumProcessing) Withdraw(ctx context.Context,
	request service.CredentialWithdraw, merchantID, externalId string, merchantWallets service.Wallets,
	multiSignSignature service.FuncMultiSignSignature) (*service.WithdrawResponse, error) {
	commission := 0.0
	if externalId != merchantWallets.ReceivingID || externalId != merchantWallets.SendingID {
		commission = merchantWallets.CommissionSending.Fix
		commission += merchantWallets.CommissionSending.Percent / 100. * (request.Amount - commission)
	}

	balance, err := s.GetAssetsBalance(ctx,
		service.BalanceRequest{Blockchain: request.Blockchain, Asset: request.Asset, Issuer: request.Issuer},
		merchantID, merchantWallets.SendingID)
	if err != nil {
		return nil, fmt.Errorf("can't get merchant: %v, sending wallet: %v, err: %w",
			merchantID, merchantWallets.SendingID, err)
	}
	if balance[0].Amount < request.Amount+commission {
		return nil, fmt.Errorf("merchant: %s, doesn't have enough balance to pay: %v %v, with commission: %v",
			merchantID, request.Amount, request.Asset, commission)
	}

	_, sendingWalletRaw, err := s.store.GetByUser(merchantID, merchantWallets.SendingID)
	if err != nil {
		return nil, err
	}

	sendingWallet := service.Wallet{}
	err = json.Unmarshal(sendingWalletRaw, &sendingWallet)
	if err != nil {
		return nil, err
	}
	//check gas
	_, err = s.updateGas(ctx, sendingWallet.WalletAddress, coreumFeeSendFT)
	if err != nil {
		return nil, err
	}
	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		sendingWallet.WalletAddress,
		string(sendingWallet.WalletSeed),
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()

	msg := &banktypes.MsgSend{
		FromAddress: sendingWallet.WalletAddress,
		ToAddress:   s.sendingWallet.WalletAddress,
		Amount: sdk.NewCoins(sdk.NewInt64Coin(fmt.Sprintf("%s-%s", request.Asset, request.Issuer),
			int64(request.Amount))),
	}
	bech32, err := sdk.AccAddressFromBech32(sendingWallet.WalletAddress)
	if err != nil {
		return nil, err
	}
	result, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(bech32),
		s.factory,
		msg,
	)
	if err != nil {
		return nil, err
	}
	return &service.WithdrawResponse{TransactionHash: result.TxHash}, nil
}
