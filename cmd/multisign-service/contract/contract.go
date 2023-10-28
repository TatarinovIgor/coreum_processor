package contract

type MultiSignAddresses struct {
	Addresses map[string]float64 `json:"addresses"`
	Threshold int                `json:"threshold"`
}

type SignTransactionRequest struct {
	ExternalID string   `json:"external_id"`
	Blockchain string   `json:"blockchain"`
	Addresses  []string `json:"addresses"`
	TrxID      string   `json:"trxID"`
	TrxData    string   `json:"trxData"`
	Threshold  int      `json:"threshold"`
}
