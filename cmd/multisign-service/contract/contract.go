package contract

type MultiSignAddresses struct {
	Addresses []string `json:"addresses"`
	Threshold int      `json:"threshold"`
}

type SignTransactionRequest struct {
	Blockchain string `json:"blockchain"`
	Address    string `json:"address"`
	TrxID      string `json:"trxID"`
	TrxData    string `json:"trxData"`
	Threshold  int    `json:"threshold"`
}
