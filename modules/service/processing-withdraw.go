package service

import (
	"context"
	"coreum_processor/modules/storage"
	"errors"
	"fmt"
	"log"
)

const (
	limitTrxToWithdrawProcess = 1000
)

func (s ProcessingService) processWithdraw(ctx context.Context, bc string, processor CryptoProcessor,
	merch MerchantData, wallet Wallets) {

	// Process processed transactions
	s.processWithdrawProcessed(ctx, bc, processor, merch, wallet)

	// Process settled transaction
	s.processWithdrawSettled(ctx, bc, processor, merch)

}

func (s ProcessingService) processWithdrawProcessed(ctx context.Context, bc string, processor CryptoProcessor,
	merch MerchantData, wallet Wallets) {
	trx, err := s.transactionStore.GetMerchantTrxForProcessingInBlockChain(merch.ID.String(), bc,
		storage.WithdrawTransaction, storage.ProcessedTransaction, 1000)
	if err != nil && !errors.Is(storage.ErrNotFound, err) {
		log.Println(fmt.Errorf("can't get merchant transactions to settle, err: %v", err))
	} else if err == nil {

		for _, tr := range trx {

			commission := wallet.CommissionSending.Fix + tr.Amount*wallet.CommissionSending.Percent/100

			s.transactionStore.PutProcessedTransaction(tr.MerchantId, tr.ExternalId, tr.GUID.String(),
				tr.Hash1, commission)
			callBackSign, err := s.callBack.GetMultiSignFn(tr.MerchantId)
			if err != nil {
				log.Println(fmt.Errorf(
					"error in process withdraw processing for merchant: %v, due to issue with callback err: %v",
					merch.ID, err))
				continue
			}
			hash, err := processor.Withdraw(ctx, CredentialWithdraw{
				Amount:        tr.Amount,
				Blockchain:    tr.Blockchain,
				WalletAddress: tr.ExtWallet,
				Asset:         tr.Asset,
				Issuer:        tr.Issuer,
				Memo:          "",
			}, merch.ID.String(), wallet.SendingID, wallet, callBackSign)
			if err != nil {
				log.Println(fmt.Errorf("can't process transactions: %v to settle, err: %v", tr.GUID, err))
				continue
			}
			s.transactionStore.PutSettledTransaction(tr.MerchantId, tr.ExternalId, tr.GUID.String(),
				hash.TransactionHash)
			callBackTrx, err := s.callBack.GetTransactionFn(tr.MerchantId)
			if err != nil {
				log.Println(fmt.Errorf(
					"error in process withdraw processing for merchant: %v, due to issue with callback err: %v",
					merch.ID, err))
				continue
			} else if callBackTrx != nil {
				err = callBackTrx(tr)
				if err != nil {
					log.Println(fmt.Errorf(
						"error in process withdraw processing for merchant: %v, callback err: %v",
						merch.ID, err))
				}
			}
		}
	}
}
func (s ProcessingService) processWithdrawSettled(ctx context.Context, bc string, processor CryptoProcessor,
	merch MerchantData) {
	trx, err := s.transactionStore.GetMerchantTrxForProcessingInBlockChain(merch.ID.String(), bc,
		storage.WithdrawTransaction, storage.SettledTransaction, 1000)
	if err != nil && !errors.Is(storage.ErrNotFound, err) {
		log.Println(fmt.Errorf("can't get merchant transactions to done, err: %v", err))
	} else if err == nil {
		for _, tr := range trx {
			hash, err := processor.TransferFromSending(ctx, TransferRequest{
				Amount:     tr.Amount,
				Blockchain: tr.Blockchain,
				Asset:      tr.Asset,
				Issuer:     tr.Issuer,
			}, merch.ID.String(), tr.ExtWallet)
			if err != nil {
				log.Println(fmt.Errorf("can't process transactions: %v to settle, err: %v", tr.GUID, err))
				continue
			}

			err = s.transactionStore.PutDoneTransaction(tr.MerchantId, tr.ExternalId, tr.GUID.String(),
				hash.TransferHash)
			if err != nil {
				log.Println(fmt.Errorf("can't put transaction to done status, err: %v", err))
				return
			}

			callBackTrx, err := s.callBack.GetTransactionFn(tr.MerchantId)
			if err != nil {
				log.Println(fmt.Errorf(
					"error in process withdraw settlement for merchant: %v, due to issue with callback err: %v",
					merch.ID, err))
				continue
			} else if callBackTrx != nil {
				err = callBackTrx(tr)
				if err != nil {
					log.Println(fmt.Errorf(
						"error in process withdraw processing for merchant: %v, callback err: %v",
						merch.ID, err))
				}
			}
		}
	}
}
