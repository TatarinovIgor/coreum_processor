package service

import (
	"coreum_processor/modules/storage"
	"encoding/json"
	"fmt"
)

type Merchants struct {
	store storage.Storage
}

func (service *Merchants) GetMerchantData(id string) (MerchantData, error) {
	data := MerchantData{}
	_, dataRaw, err := service.store.Get(id)
	if err != nil {
		return data, err
	}
	err = json.Unmarshal(dataRaw, &data)
	return data, err
}

func (service *Merchants) GetMerchant(id string) (storage.Record, error) {
	res := storage.Record{}
	_, data, err := service.store.Get(id)
	if err != nil {
		return res, err
	}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return res, err
	}
	return res, err
}

func (service *Merchants) GetMerchants() ([]storage.Record, error) {
	data, err := service.store.GetAll()
	if err != nil {
		return data, err
	}
	return data, err
}

func (service *Merchants) CreateMerchantData(guid string, data MerchantData) (int64, error) {
	dataByte, err := json.Marshal(data)
	if err != nil {
		return 0, err
	}
	id, _, err := service.store.Set(guid, dataByte, storage.DefaultTTL)
	if err != nil {
		return 0, err
	}
	return id, err
}

func (service *Merchants) UpdateMerchantData(id string, data NewMerchant) (string, error) {
	_, dataRaw, err := service.store.Get(id)
	if err != nil {
		return "", err
	}
	dataOld := MerchantData{}
	err = json.Unmarshal(dataRaw, &dataOld)
	if err != nil {
		return "", err
	}
	dataOld.PublicKey = data.PublicKey
	dataOld.MerchantName = data.MerchantName
	dataOld.CallBackURL = data.Callback
	dataByte, err := json.Marshal(dataOld)
	if err != nil {
		return "", err
	}
	_, err = service.store.Put(id, dataByte, storage.DefaultTTL)
	if err != nil {
		return "", err
	}
	return id, err
}

func (service *Merchants) UpdateMerchantCommission(id, blockchain string,
	data NewMerchantCommission) (Wallets, error) {
	_, dataRaw, err := service.store.Get(id)
	if err != nil {
		return Wallets{}, err
	}
	dataOld := MerchantData{}
	err = json.Unmarshal(dataRaw, &dataOld)
	if err != nil {
		return Wallets{}, err
	}
	newData := Wallets{
		CommissionReceiving: data.CommissionReceiving,
		CommissionSending:   data.CommissionSending,
		ReceivingID:         fmt.Sprintf("%s-R", id),
		SendingID:           fmt.Sprintf("%s-S", id),
	}

	if dataOld.Wallets == nil {
		dataOld.Wallets = map[string]Wallets{}
	}
	if _, ok := dataOld.Wallets[blockchain]; ok {
		newData.ReceivingID = dataOld.Wallets[blockchain].ReceivingID
		newData.SendingID = dataOld.Wallets[blockchain].SendingID
	}
	dataOld.Wallets[blockchain] = newData
	dataByte, err := json.Marshal(dataOld)
	if err != nil {
		return Wallets{}, err
	}
	_, err = service.store.Put(id, dataByte, storage.DefaultTTL)
	if err != nil {
		return Wallets{}, err
	}
	return newData, err
}
func NewMerchantService(store storage.Storage) *Merchants {
	return &Merchants{
		store: store,
	}
}
