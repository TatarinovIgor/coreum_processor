package service

import (
	"context"
	"coreum_processor/modules/storage"
	"errors"
	"fmt"
	"log"
)

const (
	limitTrxToDepositProcess = 1000
)

func (s ProcessingService) processDeposit(ctx context.Context, bc string, processor CryptoProcessor,
	merch MerchantData, wallet Wallets) {

	// Process initiated trx
	s.processDepositInitiated(ctx, bc, processor, merch)

	// Process processed trx
	s.processDepositProcessed(ctx, bc, processor, merch, wallet)

	// Process settled trx
	s.processDepositSettled(ctx, bc, processor, merch)

}

func (s ProcessingService) processDepositInitiated(ctx context.Context, bc string, processor CryptoProcessor,
	merch MerchantData) {
	trx, err := s.transactionStore.GetMerchantTrxForProcessingInBlockChain(merch.ID.String(), bc,
		storage.DepositTransaction, storage.InitTransaction, limitTrxToDepositProcess)
	if err != nil && errors.Is(storage.ErrNotFound, err) {
		// no deposit transaction to process
		return
	} else if err != nil {
		log.Println(fmt.Errorf(
			"can't get initiated transaction: %v, err: %v, blockchain: %v", merch.ID, err, bc))
		return
	}
	for _, tr := range trx {
		if tr.Hash2 == "" {
			// blockchain transaction is not created for processing
			res, err := processor.TransferToReceiving(ctx, TransferRequest{
				Amount:     tr.Amount,
				Blockchain: tr.Blockchain,
				Asset:      tr.Asset,
				Issuer:     tr.Issuer,
			}, tr.MerchantId, tr.ExternalId)
			if err != nil {
				log.Println(fmt.Sprintf("error in process deposit in transfer to receiving: %v", err))
				continue
			}
			// TODO: update store if err?
			err = s.transactionStore.PutInitiatedPendingTransaction(tr.MerchantId, tr.ExternalId,
				tr.GUID.String(), res.TransferHash)
			if err != nil {
				log.Println(
					fmt.Sprintf("error in process deposit to put inititated pending transaction for: %v, err: %v",
						tr.GUID.String(), err))
				continue
			}
		} else {
			res, err := processor.GetTransactionStatus(ctx, tr.Hash2)
			if res == SuccessfulTransaction {
				err = s.transactionStore.PutProcessedTransaction(tr.MerchantId, tr.ExternalId,
					tr.GUID.String(), tr.Hash2, 0)
			} else if res == FailedTransaction {
				// reset transaction hash to create new transaction for processing
				err = s.transactionStore.PutInitiatedPendingTransaction(tr.MerchantId, tr.ExternalId,
					tr.GUID.String(), "")
			}
			if err != nil {
				log.Println(
					fmt.Sprintf("error in process deposit to put processed status for transaction: %v, err: %v",
						tr.GUID.String(), err))
			}
			continue
		}
	}
}

func (s ProcessingService) processDepositProcessed(ctx context.Context, bc string, processor CryptoProcessor,
	merch MerchantData, wallet Wallets) {

	trx, err := s.transactionStore.GetMerchantTrxForProcessingInBlockChain(merch.ID.String(), bc,
		storage.DepositTransaction, storage.ProcessedTransaction, limitTrxToDepositProcess)
	if err != nil && errors.Is(storage.ErrNotFound, err) {
		// no deposit transaction to settle
		return
	} else if err != nil {
		log.Println(fmt.Errorf(
			"can't get processed transactions for merchant: %v, err: %v, blockchain: %v", merch.ID, err, bc))
		return
	}
	if len(trx) == 0 {
		return
	}
	amount := 0.
	asset := ""
	issuer := ""
	for _, tr := range trx {
		commission := wallet.CommissionReceiving.Fix + tr.Amount*wallet.CommissionReceiving.Percent/100
		amount += tr.Amount - commission
		s.transactionStore.PutProcessedTransaction(tr.MerchantId, tr.ExternalId, tr.GUID.String(),
			tr.Hash2, commission)
		asset = tr.Asset
		issuer = tr.Issuer
		if amount > 0 {
			//ToDo group transactions by assets
			hash, err := processor.TransferFromReceiving(ctx, TransferRequest{
				Amount:     amount,
				Blockchain: bc,
				Asset:      asset,
				Issuer:     issuer,
			}, merch.ID.String(), wallet.ReceivingID)
			if err != nil {
				log.Println(fmt.Errorf("can't settle transactions to merchant: %v, err: %v", merch.ID, err))
				return
			}
			callBack, err := s.callBack.GetTransactionFn(tr.MerchantId)
			if err != nil {
				log.Println(fmt.Errorf(
					"can't settle transactions to merchant: %v, due to issue with callback err: %v",
					merch.ID, err))
				return
			}
			for _, tr := range trx {
				s.transactionStore.PutSettledTransaction(tr.MerchantId, tr.ExternalId,
					tr.GUID.String(), hash.TransferHash)
				if callBack != nil {
					callBack(tr)
				}
			}
		}
	}

}

func (s ProcessingService) processDepositSettled(ctx context.Context, bc string, processor CryptoProcessor,
	merch MerchantData) {
	trx, err := s.transactionStore.GetMerchantTrxForProcessingInBlockChain(merch.ID.String(), bc,
		storage.DepositTransaction, storage.SettledTransaction, limitTrxToDepositProcess)
	if err != nil && errors.Is(storage.ErrNotFound, err) {
		// no deposit transaction to settle
		return
	} else if err != nil {
		log.Println(fmt.Errorf(
			"can't get settled transaction for merchant: %v, err: %v, blockchain: %v", merch.ID, err, bc))
		return
	}
	for _, tr := range trx {
		if tr.Hash3 != "" {
			res, err := processor.GetTransactionStatus(ctx, tr.Hash3)
			if res == SuccessfulTransaction {
				// TODO: gas should be returned to sending wallet
				s.transactionStore.PutDoneTransaction(tr.MerchantId, tr.ExternalId,
					tr.GUID.String(), "")
				callBack, err := s.callBack.GetTransactionFn(tr.MerchantId)
				if err != nil {
					log.Println(fmt.Errorf(
						"error in process deposit for merchant: %v, due to issue with callback err: %v",
						merch.ID, err))
					continue
				} else if callBack != nil {
					callBack(tr)
				}
			} else if res == FailedTransaction {
				// reset transaction for settlement
				s.transactionStore.PutSettledTransaction(tr.MerchantId, tr.ExternalId,
					tr.GUID.String(), "")
			}
			if err != nil {
				log.Println(
					fmt.Sprintf("error in process deposit to put processed status for transaction: %v, err: %v",
						tr.GUID.String(), err))
			}
			continue
		} else {
			log.Println(fmt.Errorf(
				"can't get merchant: %v, transactions to settle, err: %v, blockchain: %v", merch.ID, err, bc))
		}
	}
}
