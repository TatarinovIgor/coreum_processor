package asset

import (
	"coreum_processor/modules/service"
	"coreum_processor/modules/storage"
	"encoding/json"
	"strings"
	"time"
)

type Service struct {
	assetStorage *storage.AssetPSQL
	merchants    service.Merchants
}

func (s *Service) GetBlockChainAssetByCodeAndIssuer(blockchain, code, issuer string) (*storage.AssetStore, error) {
	return s.assetStorage.GetBlockChainAssetByCodeAndIssuer(blockchain, code, issuer)
}

func (s *Service) ActivateAsset(blockchain, code, issuer, merchantID string) error {
	return s.assetStorage.ActivateAsset(blockchain, code, issuer, merchantID)
}

func (s *Service) GetAssetList(merchID string, blockChain, code []string, codeLike, status string,
	from, to time.Time) ([]storage.AssetStore, error) {
	return s.assetStorage.GetAssetList(merchID, blockChain, code, codeLike, status, from, to)
}
func (s *Service) UpdateAssetRequest(asset storage.AssetStore, status string) error {
	return s.assetStorage.SetAssetStatus(asset, storage.AssetStatus(status))
}
func (s *Service) CreateAssetRequest(blockchain, code, smartContractAddress, name, description, assetType, merchantOwnerID string,
	features json.RawMessage) error {
	code = strings.ToLower(code)
	return s.assetStorage.CreateAsset(blockchain, code, smartContractAddress, name, description, assetType, merchantOwnerID, features)
}
func (s *Service) IssueAsset(blockchain, code, merchantID string) error {
	// TODO: issue asset in processing
	issuer := ""
	return s.assetStorage.ActivateAsset(blockchain, code, issuer, merchantID)
}

// NewService create a service to process operation with assets and merchant settings
func NewService(assetStorage *storage.AssetPSQL, merchants service.Merchants) *Service {
	return &Service{
		assetStorage: assetStorage,
		merchants:    merchants,
	}
}
