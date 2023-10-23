package processor_coreum

import (
	"context"
	"coreum_processor/modules/service"
	"coreum_processor/modules/storage"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"
)

func (s CoreumProcessing) Deposit(ctx context.Context, request service.CredentialDeposit, merchantID, externalId string,
	multiSignAddresses service.FuncMultiSignAddrCallback) (*service.DepositResponse, error) {
	depositData := service.DepositResponse{}

	wallet := service.Wallet{Blockchain: s.receivingWallet.Blockchain}
	_, walletByte, err := s.store.GetByUser(merchantID, externalId)

	if err != nil && errors.Is(err, storage.ErrNotFound) {
		wallet.WalletSeed, wallet.WalletAddress, err = s.createCoreumWallet(ctx,
			externalId, multiSignAddresses)
		if err != nil {
			return nil, err
		}

		wallet.Blockchain = request.Blockchain
		key, err := json.Marshal(wallet)
		if err != nil {
			return nil, err
		}

		_, err = s.store.Put(merchantID, externalId, wallet.WalletAddress, key)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		err = json.Unmarshal(walletByte, &wallet)
		if err != nil {
			return nil, err
		}
	}
	depositData.WalletAddress = wallet.WalletAddress
	depositData.Memo = ""

	return &depositData, nil

}

func (s CoreumProcessing) StreamDeposit(ctx context.Context, callback service.FuncDepositCallback,
	interval time.Duration) {
	go s.streamDeposit(ctx, callback, interval)
	return
}

func (s CoreumProcessing) streamDeposit(ctx context.Context, callback service.FuncDepositCallback,
	interval time.Duration) {
	ticker := time.NewTicker(time.Second * interval)
	next := int64(0)
	for {
		select {
		case <-ctx.Done():
			log.Println("exit from coreum processor deposit stream")
			return
		case <-ticker.C:
			records, err := s.store.GetNext(next, 1)
			if err != nil || len(records) == 0 {
				next = 0
				if err != nil {
					log.Println("Error while getting DB records:", err)
				}
			} else if strings.Contains(records[0].ExternalID, records[0].MerchantID) {
				next = records[0].ID
				continue
			} else {
				record := records[len(records)-1]
				balance, err := s.GetAssetsBalance(ctx,
					service.BalanceRequest{Blockchain: s.blockchain, Asset: ""}, record.MerchantID, record.ExternalID)
				if balance != nil && err == nil {
					for i := 0; i < len(balance); i++ {
						if balance[i].Amount > 0 {
							callback(balance[i].Blockchain, record.MerchantID, record.ExternalID, record.Key, "",
								balance[i].Asset, balance[i].Issuer, balance[i].Amount)
						}
					}
				}
				if err != nil {
					log.Println(err)
				}
				next = record.ID
			}
		}
	}
}
