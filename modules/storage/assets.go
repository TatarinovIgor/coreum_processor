package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type AssetStatus string
type AssetType string

const (
	AssetPending     AssetStatus = "pending"
	AssetActive      AssetStatus = "active"
	AssetDeprecated  AssetStatus = "deprecated"
	AssetDeactivated AssetStatus = "deactivated"
	AssetBlocked     AssetStatus = "blocked"
	AssetFungible    AssetType   = "fungible"
)

type AssetStore struct {
	Id            int             `json:"-"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     *time.Time      `json:"-"`
	BlockChain    string          `json:"blockchain"`
	Code          string          `json:"code"`
	Issuer        string          `json:"issuer"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	MerchantOwner string          `json:"merchant_owner"`
	Status        string          `json:"status"`
	Type          string          `json:"type"`
	Features      json.RawMessage `json:"features"`
}

type AssetPSQL struct {
	db                      *sql.DB
	assetsNamespace         string
	merchantAssetsNamespace string
	merchantListNamespace   string
	defaultAccess           UserAccess
}

// GetBlockChainAssetByCodeAndIssuer find a blockchain asset in the store by unique combination code and issuer
func (s *AssetPSQL) GetBlockChainAssetByCodeAndIssuer(blockchain, code, issuer string) (*AssetStore, error) {
	an := s.assetsNamespace
	query := fmt.Sprintf("SELECT %s.id, %s.created_at, %s.updated_at, %s.deleted_at, "+
		" %s.blockchain, %s.code, %s.issuer, %s.name, %s.description, ml.company_name, "+
		"%s.status, %s.Type, %s.Features FROM %s ",
		an, an, an, an, an, an, an, an, an, an, an, an, an)
	query += fmt.Sprintf(" join %s ma on %s.id = ma.asset_id ",
		s.merchantAssetsNamespace, an)
	query += fmt.Sprintf("join %s ml on ma.merchant_list_id = ml.id ", s.merchantListNamespace)
	query += fmt.Sprintf(" WHERE blockchain = '%s' and code = '%s' and issuer = '%s'",
		blockchain, code, issuer)
	rows, err := s.db.Query(query)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("could not execute query: %w", err)
	}

	assetsStore, err := rowsToAssets(rows)
	if err != nil {
		return nil, err
	}
	if len(assetsStore) == 1 {
		return &assetsStore[0], nil
	}
	return nil, fmt.Errorf("unexpected number of assets: %v, for blockchain: %v, code: %v, issuer: %s",
		len(assetsStore), blockchain, code, issuer)
}

// GetAssetList get a filtered list of assets
func (s *AssetPSQL) GetAssetList(merchID string, blockChain, code []string, codeLike string,
	from, to time.Time) ([]AssetStore, error) {
	an := s.assetsNamespace
	query := fmt.Sprintf("SELECT %s.id, %s.created_at, %s.updated_at, %s.deleted_at, "+
		" %s.blockchain, %s.code, %s.issuer, %s.name, %s.description, ml.company_name, "+
		"%s.status, %s.Type, %s.Features FROM %s ",
		an, an, an, an, an, an, an, an, an, an, an, an, an)
	query += fmt.Sprintf("join %s ma on %s.id = ma.asset_id ",
		s.merchantAssetsNamespace, an)
	query += fmt.Sprintf("join %s ml on ma.merchant_list_id = ml.id ", s.merchantListNamespace)

	if merchID != "" {
		query += fmt.Sprintf("where merchant_id = '%s' and ", merchID)
	} else {
		query += "where "
	}
	query += fmt.Sprintf("%s.deleted_at IS NULL and %s.created_at > '%v' and %s.created_at < '%v' ",
		an, an, from.Format(time.RFC3339), an, to.Format(time.RFC3339))

	if blockChain != nil && len(blockChain) > 0 && blockChain[0] != "" {
		query += fmt.Sprintf(" and %s.blockchain in (", an)
		for i, val := range blockChain {
			if i > 0 {
				query += ", "
			}
			query += fmt.Sprintf("%v", val)
		}
		query += ") "
	}

	if code != nil && len(code) > 0 && code[0] != "" {
		query += fmt.Sprintf(" and %s.code in (", an)
		for i, val := range blockChain {
			if i > 0 {
				query += ", "
			}
			query += fmt.Sprintf("%v", val)
		}
		query += ") "
	}
	rows, err := s.db.Query(query)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("could not execute query: %w", err)
	}

	assetStore, err := rowsToAssets(rows)
	if err != nil {
		return nil, fmt.Errorf("could not execute queryL: %w", err)
	}
	return assetStore, nil
}

// CreateAsset makes a new record in the asset store for a blockchain and asset description
func (s *AssetPSQL) CreateAsset(blockchain, code, name, description, assetType, merchantOwnerID string,
	features json.RawMessage) error {
	query := fmt.Sprintf("WITH merchantID AS (SELECT id FROM %s WHERE merchant_id = '%s'), "+
		"assetID AS (INSERT INTO %s "+
		"(created_at, updated_at, blockchain, code, name, description, status, type, features, merchant_owner) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, (SELECT id FROM merchantID)) RETURNING id) "+
		"INSERT INTO %s (created_at, updated_at, asset_id, merchant_list_id) VALUES "+
		"(now(), now(), (SELECT id FROM assetID), (SELECT id FROM merchantID)) RETURNING id",
		s.merchantListNamespace, merchantOwnerID, s.assetsNamespace, s.merchantAssetsNamespace)
	if features == nil {
		features = json.RawMessage{}
		_ = features.UnmarshalJSON([]byte("{}"))
	}
	_, err := s.db.Query(query,
		time.Now().UTC(), time.Now().UTC(), blockchain, code, name, description, AssetPending, assetType, features)
	if err != nil {
		return err
	}
	return nil
}

// ActivateAsset sets the assets status in store by blockchain, code and issuer from AssetStore structure
func (s *AssetPSQL) ActivateAsset(blockchain, code, issuer, merchantID string) error {
	query := fmt.Sprintf(
		"UPDATE %s SET updated_at = $4, access = $5, issuer  = $3 WHERE blockchain = $1, code  = $2 "+
			" (select id from %s where merchant_id = '%s') ",
		s.assetsNamespace, s.merchantListNamespace, merchantID)
	_, err := s.db.Query(query,
		blockchain, code, issuer, time.Now().UTC(), AssetActive)
	if err != nil {
		return err
	}
	return nil
}

// SetAssetStatus sets the assets status in store by blockchain, code and issuer from AssetStore structure
func (s *AssetPSQL) SetAssetStatus(asset AssetStore, status AssetStatus) error {
	query := fmt.Sprintf(
		"UPDATE %s SET updated_at = $4, access = $5 WHERE blockchain = $1, code  = $2, issuer  = $3",
		s.assetsNamespace)
	_, err := s.db.Query(query,
		asset.BlockChain, asset.Code, asset.Issuer, time.Now().UTC(), status)
	if err != nil {
		return err
	}
	return nil
}

// UpdateDescription updates the description in store by blockchain, code and issuer from AssetStore structure
func (s *AssetPSQL) UpdateDescription(asset AssetStore, description string) error {
	query := fmt.Sprintf(
		"UPDATE %s SET updated_at = $4, description = $5 WHERE blockchain = $1, code  = $2, issuer  = $3",
		s.assetsNamespace)
	_, err := s.db.Query(query,
		asset.BlockChain, asset.Code, asset.Issuer, time.Now().UTC(), description)
	if err != nil {
		return err
	}
	return nil
}

// LinkAssetToMerchant link the asset to a merchant store by unique user identity and merchant id
func (s *UserPSQL) LinkAssetToMerchant(identity, merchantID string, merchantAccess MerchantAccess) error {
	query := fmt.Sprintf(
		"INSERT INTO %s (created_at, updated_at, deleted_at, user_id, merchant_list_id)"+
			"values (now(), now(), null, (select id from %s where identity = '%s'), "+
			"(select id from %s where merchant_id = '%s'))",
		s.merchantUsersNamespace, s.userNamespace, identity, s.merchantListNamespace, merchantID)
	_, err := s.db.Query(query)
	if err != nil {
		return err
	}
	return nil
}

func NewAssetsStorage(defaultAccess UserAccess, assetNamespace, merchantAssetsNamespace, merchantListNamespace string,
	db *sql.DB) (*AssetPSQL, error) {
	s := AssetPSQL{
		db:                      db,
		assetsNamespace:         assetNamespace,
		merchantAssetsNamespace: merchantAssetsNamespace,
		merchantListNamespace:   merchantListNamespace,
		defaultAccess:           defaultAccess,
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not connect to DB: %v", err)
	}
	if _, err := db.Exec(fmt.Sprintf("SELECT 1 FROM %q LIMIT 1", assetNamespace)); err != nil {
		return nil, fmt.Errorf("could not connect to user storage: %v", err)
	}
	if _, err := db.Exec(fmt.Sprintf("SELECT 1 FROM %q LIMIT 1", merchantAssetsNamespace)); err != nil {
		return nil, fmt.Errorf("could not connect to merchant's users storage: %v", err)
	}
	if _, err := db.Exec(fmt.Sprintf("SELECT 1 FROM %q LIMIT 1", merchantListNamespace)); err != nil {
		return nil, fmt.Errorf("could not connect to merchant list: %v", err)
	}
	return &s, nil
}

func rowsToAssets(rows *sql.Rows) ([]AssetStore, error) {
	var assets []AssetStore
	if rows == nil {
		return assets, nil
	}
	var issuer *string
	for rows.Next() {
		asset := AssetStore{}
		if err := rows.Scan(
			&asset.Id,
			&asset.CreatedAt, &asset.UpdatedAt, &asset.DeletedAt,
			// blockchain, %s.code, %s.issuer, %s.name, %s.description, %s.company_name, %s.status
			&asset.BlockChain, &asset.Code, &issuer,
			&asset.Name, &asset.Description, &asset.MerchantOwner, &asset.Status, &asset.Type, &asset.Features,
		); err != nil {
			return nil, err
		}
		if issuer != nil {
			asset.Issuer = *issuer
		}
		assets = append(assets, asset)
	}
	return assets, nil
}
