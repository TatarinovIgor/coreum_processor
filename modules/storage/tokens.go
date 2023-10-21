package storage

import (
	"database/sql"
	"fmt"
	"time"
)

type SmartContract struct {
	SmartContractAddress string `json:"smart_contract_address"`
	FeeLimit             int64  `json:"fee_limit"`
}

type Erc20Data struct {
	Blockchain string        `json:"blockchain"`
	Asset      string        `json:"asset"`
	Issuer     string        `json:"issuer"`
	Contract   SmartContract `json:"address"`
	Format     string        `json:"format"`
}

type TokenPSQL struct {
	db        *sql.DB
	namespace string
}

func NewTokenStorage(namespace string, db *sql.DB) (*TokenPSQL, error) {
	s := TokenPSQL{
		db:        db,
		namespace: namespace,
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not connect to DB: %v", err)
	}
	if _, err := db.Exec(fmt.Sprintf("SELECT 1 FROM %q LIMIT 1", namespace)); err != nil {
		return nil, fmt.Errorf("could not connect to transaction storage: %v", err)
	}
	return &s, nil
}

func (s *TokenPSQL) GetAllAssetsForBlockchain(blockchain string) (map[string]string, error) {
	query := `SELECT asset, smart_contract FROM $1 WHERE blockchain = $2`
	rows, err := s.db.Query(query, s.namespace, blockchain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var smartContracts map[string]string

	for rows.Next() {
		var asset, smartContract string
		if err := rows.Scan(&asset, &smartContract); err != nil {
			return nil, err
		}
		smartContracts[smartContract] = asset
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return smartContracts, nil
}

func (s *TokenPSQL) GetAll() ([]Erc20Data, error) {
	query := fmt.Sprintf(`SELECT blockchain, asset, issuer, smart_contract, fee_limit, type
		FROM %s`,
		s.namespace,
	)
	r, err := s.getAll(query)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// CreateToken adds given token to the database
func (s *TokenPSQL) CreateToken(token Erc20Data) error {
	// Define SQL statement
	query := `INSERT INTO tokens (created_at, updated_at, blockchain, asset, issuer, contractAddress, feeLimit, format) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	// Get current time
	now := time.Now()

	// Execute SQL statement
	_, err := s.db.Query(query, now, now, token.Blockchain, token.Asset, token.Issuer,
		token.Contract.SmartContractAddress, token.Contract.FeeLimit, token.Format)
	if err != nil {
		return err
	}

	return nil
}

func (s *TokenPSQL) getAll(query string) ([]Erc20Data, error) {
	rows, err := s.db.Query(query)
	var (
		blockchain, asset, issuer, contractAddress, format string
		feeLimit                                           int64
	)
	if err != nil {
		return nil, err
	}
	var records []Erc20Data
	for rows.Next() {
		err := rows.Scan(&blockchain, &asset, &issuer, &contractAddress, &format, &feeLimit)
		if err == nil {
			record := &Erc20Data{
				Blockchain: blockchain,
				Asset:      asset,
				Issuer:     issuer,
				Contract: SmartContract{
					SmartContractAddress: contractAddress,
					FeeLimit:             feeLimit,
				},
				Format: format,
			}
			records = append(records, *record)
		}
		if err != nil {
			return nil, ErrNotFound
		}
	}

	return records, nil
}
