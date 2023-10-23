package service

import (
	"context"
	"coreum_processor/modules/storage"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

// makeDepositCallback function to create record in the transaction store to process deposit transaction
func (s ProcessingService) makeDepositCallback() FuncDepositCallback {
	return func(blockChain, merchantID, externalId, externalWallet, hash, asset, issuer string, amount float64) {

		// validate processor
		_, ok := s.processors[blockChain]
		if !ok {
			log.Println(fmt.Sprintf("error in deposit callback to define processor for a blockchain: %v",
				blockChain))
			return
		}
		// find merchant
		merch, err := s.merchants.GetMerchantData(merchantID)
		if err != nil {
			log.Println(fmt.Sprintf("error in deposit callback to get merch data: %v", err))
			return
		}

		// validate merchant wallet
		if wallet, ok := merch.Wallets[blockChain]; ok {
			if wallet.SendingID == externalId || wallet.ReceivingID == externalId {
				return
			}
		} else {
			// TODO: refund transaction
		}

		// find all initiated transaction by merchant for user and check if it covers the amount
		// if not create new transaction for reminding amount
		action := storage.DepositTransaction
		trx, err := s.transactionStore.GetInitTransactions(merchantID, externalId, blockChain, action)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			log.Println(
				fmt.Sprintf(
					"error in deposit callback to get merch: %v inititated transactions for user: %v in blockchain: %v, err: %v",
					merchantID, externalId, blockChain, err))
			return
		}
		processingAmount := 0.
		for _, tx := range trx {
			processingAmount += tx.Amount
		}
		amount -= processingAmount
		if amount <= 0.0 {
			return
		}

		// initiated transaction doesn't cover amount, create a new
		_, err = s.transactionStore.CreateTransaction(merchantID, externalId, blockChain,
			action, externalWallet, hash, asset, issuer, amount, 0)
		if err != nil {
			log.Println(fmt.Sprintf("error in storage to create transaction: %v", err))
		}
		return
	}
}

func (s ProcessingService) processTransaction(ctx context.Context) {
	// get deposit to settle
	merchants, err := s.merchants.GetMerchants()
	if err != nil {
		log.Println("processing can't get merchants to settle, err", err)
		return
	}
	for _, merchData := range merchants {
		merch := MerchantData{}
		err = json.Unmarshal(merchData.Data, &merch)
		if err != nil {
			log.Println(fmt.Errorf("can't unmarshal merchant data: %v to settle, err: %v", merchData.Data, err))
			continue
		}
		for bc, wallet := range merch.Wallets {
			bc = strings.ToLower(bc)
			processor, ok := s.processors[bc]
			if !ok {
				//log.Println(fmt.Errorf("processing can't find processor for blockchain: %v", bc))
				continue
			}
			if processor == nil {
				continue
			}

			// deposit processing
			s.processDeposit(ctx, bc, processor, merch, wallet)

			// withdraw processing
			s.processWithdraw(ctx, bc, processor, merch, wallet)
		}
	}
}
