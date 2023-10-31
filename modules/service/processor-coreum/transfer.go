package processor_coreum

import (
	"context"
	"coreum_processor/modules/service"
	"encoding/json"
	"fmt"
	"github.com/CoreumFoundation/coreum/v2/pkg/client"
	"github.com/CoreumFoundation/coreum/v2/x/nft"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (s CoreumProcessing) TransferFT(ctx context.Context,
	request service.TransferTokenRequest, merchantID string) (string, error) {
	SendingExternalId := request.SendingExternalId
	ReceivingExternalId := request.ReceivingExternalId

	_, key, byteAddress, err := s.store.GetByUser(merchantID, SendingExternalId)
	if err != nil {
		return "", fmt.Errorf("can't get sending wallet: %v coreum wallet from store, err: %v",
			SendingExternalId, err)
	}

	sendingWallet := service.Wallet{}
	err = json.Unmarshal(byteAddress, &sendingWallet)
	if err != nil {
		return "", err
	}

	_, _, byteAddressTo, err := s.store.GetByUser(merchantID, ReceivingExternalId)
	if err != nil {
		return "", fmt.Errorf("can't get receiving wallet: %v coreum wallet from store, err: %v",
			SendingExternalId, err)
	}

	receivingWallet := service.Wallet{}
	err = json.Unmarshal(byteAddressTo, &receivingWallet)
	if err != nil {
		return "", err
	}

	_, err = s.updateGas(ctx, sendingWallet.WalletAddress, coreumFeeSendFT)
	if err != nil {
		return "", nil
	}

	denom := fmt.Sprintf("%s-%s", request.Subunit, request.Issuer)
	//@ToDo multiply request.amount by the amount of decimals
	res, err := s.transferCoreumFT(ctx, merchantID, SendingExternalId, "transfer-"+denom, key,
		receivingWallet.WalletAddress, denom, sendingWallet, int64(request.Amount))

	if err != nil {
		return "", err
	}
	return res, nil
}

func (s CoreumProcessing) TransferNFT(ctx context.Context,
	request service.TransferTokenRequest, merchantID string) (string, error) {
	SendingExternalId := request.SendingExternalId
	ReceivingExternalId := request.ReceivingExternalId

	_, _, byteAddress, err := s.store.GetByUser(merchantID, SendingExternalId)
	if err != nil {
		return "", fmt.Errorf("can't get sending wallet: %v coreum wallet from store, err: %v",
			SendingExternalId, err)
	}

	sendingWallet := service.Wallet{}
	err = json.Unmarshal(byteAddress, &sendingWallet)
	if err != nil {
		return "", err
	}

	_, _, byteAddressTo, err := s.store.GetByUser(merchantID, ReceivingExternalId)
	if err != nil {
		return "", fmt.Errorf("can't get receiving wallet: %v coreum wallet from store, err: %v",
			SendingExternalId, err)
	}

	receivingWallet := service.Wallet{}
	err = json.Unmarshal(byteAddressTo, &receivingWallet)
	if err != nil {
		return "", err
	}

	senderInfo, err := s.clientCtx.Keyring().NewAccount(
		sendingWallet.WalletAddress,
		sendingWallet.WalletSeed,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	defer func() { _ = s.clientCtx.Keyring().DeleteByAddress(senderInfo.GetAddress()) }()
	if err != nil {
		return "", nil
	}

	_, err = s.updateGas(ctx, sendingWallet.WalletAddress, coreumFeeSendFT)
	if err != nil {
		return "", nil
	}

	//@ToDo how to use nft id and class id
	res, err := s.transferCoreumNFT(ctx,
		sendingWallet.WalletAddress, receivingWallet.WalletAddress, request.NftId, request.NftClassId)

	if err != nil {
		return "", err
	}
	return res, nil
}

func (s CoreumProcessing) transferCoreumNFT(ctx context.Context,
	senderAddress, recipientAddress, nftId, classId string) (string, error) {

	address, err := sdk.AccAddressFromBech32(senderAddress)
	if err != nil {
		return "", err
	}
	senderInfo, err := s.clientCtx.Keyring().KeyByAddress(address)
	if err != nil {
		return "", err
	}

	msgSend := &nft.MsgSend{
		Sender:   senderAddress,
		Receiver: recipientAddress,
		Id:       nftId,
		ClassId:  classId,
	}
	response, err := client.BroadcastTx(
		ctx,
		s.clientCtx.WithFromAddress(senderInfo.GetAddress()),
		s.factory,
		msgSend,
	)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return response.TxHash, nil
}

func (s CoreumProcessing) transferCoreumFT(ctx context.Context, merchantID, externalID, trxID,
	senderAddress, recipientAddress, denom string, sendingWallet service.Wallet,
	amount int64) (string, error) {

	msgSend := &banktypes.MsgSend{
		FromAddress: senderAddress,
		ToAddress:   recipientAddress,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom, amount)),
	}
	trx, err := s.broadcastTrx(ctx, merchantID, externalID, trxID, senderAddress, sendingWallet, msgSend)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return trx.TxHash, nil
}
