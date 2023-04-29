package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type StatusTx string
type ActionTx string

const (
	InitTransaction      StatusTx = "init"
	ProcessedTransaction StatusTx = "processed"
	SettledTransaction   StatusTx = "settle"
	DoneTransaction      StatusTx = "done"
	RejectedTransaction  StatusTx = "rejected"
)

const (
	DepositTransaction  ActionTx = "deposit"
	WithdrawTransaction ActionTx = "withdraw"
)

type TransactionStore struct {
	Id         int        `json:"-"`
	GUID       uuid.UUID  `json:"GUID"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"-"`
	MerchantId string     `json:"merchant_id"`
	ExternalId string     `json:"external_id"`
	Blockchain string     `json:"blockchain"`
	Action     ActionTx   `json:"action"`
	ExtWallet  string     `json:"ext_wallet"`
	Status     StatusTx   `json:"status"`
	Asset      string     `json:"asset"`
	Issuer     string     `json:"issuer"`
	Amount     float64    `json:"amount"`
	Commission float64    `json:"-"`
	Hash1      string     `json:"-"`
	Hash2      string     `json:"-"`
	Hash3      string     `json:"-"`
	Hash4      string     `json:"-"`
	Hash5      string     `json:"-"`
	Callback   string     `json:"-"`
}

type TransactionPSQL struct {
	db        *sql.DB
	namespace string
}

func (s *TransactionPSQL) GetTransactionByGuid(merchID, guid string) (*TransactionStore, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE guid = '%s' AND merchant_id = '%s'", s.namespace, guid, merchID)
	rows, err := s.db.Query(query)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("could not execute query: %w", err)
	}

	transactionStore, err := rowsToTransaction(rows)
	if err != nil {
		return nil, err
	}
	if len(transactionStore) == 1 {
		return &transactionStore[0], nil
	}
	return nil, fmt.Errorf("unexpected number of transactions: %v, for guid: %v", len(transactionStore), guid)
}

func (s *TransactionPSQL) GetTransactionsByMerchant(merchID, blockchain string,
	actionFilter []string, statusFilter []string,
	from, to time.Time) ([]TransactionStore, error) {
	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE deleted_at IS NULL and merchant_id = '%s' and created_at > '%v' and created_at < '%v' and blockchain = '%s'",
		s.namespace, merchID, from.Format(time.RFC3339), to.Format(time.RFC3339), blockchain)
	if actionFilter != nil && len(actionFilter) > 0 && actionFilter[0] != "" {
		query += " and action in ("
		for i, val := range actionFilter {
			if i > 0 {
				query += ", "
			}
			query += fmt.Sprintf(string("'" + val + "'"))
		}
		query += ") "
	}
	if statusFilter != nil && len(statusFilter) > 0 && statusFilter[0] != "" {
		query += "and status in ("
		for i, val := range statusFilter {
			if i > 0 {
				query += ", "
			}
			query += fmt.Sprintf(string("'" + val + "'"))
		}
		query += ") "
	}
	query += "order by created_at"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("could not execute query: %w", err)
	}

	var transactions []TransactionStore

	for rows.Next() {
		transaction := TransactionStore{}
		if err := rows.Scan(
			&transaction.Id, &transaction.GUID,
			&transaction.CreatedAt, &transaction.UpdatedAt, &transaction.DeletedAt,
			&transaction.MerchantId, &transaction.ExternalId, &transaction.Blockchain,
			&transaction.Action, &transaction.ExtWallet, &transaction.Status,
			&transaction.Asset, &transaction.Issuer,
			&transaction.Amount, &transaction.Commission,
			&transaction.Hash1, &transaction.Hash2,
			&transaction.Hash3, &transaction.Hash4,
			&transaction.Hash5, &transaction.Callback,
		); err != nil {
			return []TransactionStore{}, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func (s *TransactionPSQL) GetMerchantTrxForProcessingInBlockChain(merchantID, blockchain string,
	action ActionTx, status StatusTx, limit uint) ([]TransactionStore, error) {

	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE deleted_at IS NULL and merchant_id = '%s' and blockchain = '%s' and action = '%s' and status = '%s' order by created_at limit %v",
		s.namespace, merchantID, blockchain, action, status, limit)
	rows, err := s.db.Query(query)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("could not execute query: %w", err)
	}
	transactions, err := rowsToTransaction(rows)
	if err != nil {
		return nil, fmt.Errorf("could not get tensaction from query: %w", err)
	}
	if len(transactions) == 0 {
		return nil, ErrNotFound
	}
	return transactions, nil
}

// GetInitTransactions returns an array of transaction that was initiated by merchant for user
// in specified blockchain and action ["deposit"/"withdrawal"]
func (s *TransactionPSQL) GetInitTransactions(merchantID, externalID,
	blockchain string, action ActionTx) ([]TransactionStore, error) {

	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE deleted_at IS NULL and merchant_id = '%s' and external_id = '%s' and blockchain = '%s' and action = '%s' and status = '%s' order by created_at",
		s.namespace, merchantID, externalID, blockchain, action, InitTransaction)
	rows, err := s.db.Query(query)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("could not execute query: %w", err)
	}

	transactions, err := rowsToTransaction(rows)
	if err != nil {
		return nil, fmt.Errorf("could not get transaction from query: %w", err)
	}

	if len(transactions) == 0 {
		return nil, ErrNotFound
	}

	return transactions, nil
}

// CreateTransaction makes a new record in the transaction store with a unique transaction guid
// and return guid new created transaction
func (s *TransactionPSQL) CreateTransaction(merchantID, externalID, blockchain string, action ActionTx,
	externalWallet, hash, asset, issuer string,
	amount, commission float64) (string, error) {
	guid, err := uuid.NewUUID()
	query := fmt.Sprintf("INSERT INTO %s (guid, created_at, updated_at, merchant_id, external_id, blockchain, action, ext_wallet, status, asset, issuer, amount, commission, hash1)",
		s.namespace)
	query += "VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING guid"
	_, err = s.db.Query(
		query,
		guid, time.Now().UTC(), time.Now().UTC(), merchantID, externalID, blockchain, action, externalWallet,
		InitTransaction, asset, issuer, amount, commission, hash)
	if err != nil {
		return "", err
	}
	return guid.String(), nil
}

// RejectTransaction marks a specified transaction created by merchant for the user as rejected
func (s *TransactionPSQL) RejectTransaction(merchantID, externalID, transaction string) error {
	// TODO: check status transaction only "init" can be rejected
	query := fmt.Sprintf("UPDATE %s set status = '%s' where guid = $1 and merchant_id = $2 and external_id = $3",
		s.namespace, RejectedTransaction)
	_, err := s.db.Query(query, transaction, merchantID, externalID)

	if err != nil {
		return err
	}
	return nil
}

// PutInitiatedPendingTransaction add a hash to initiated transaction to trace status of a blockchain transaction
// status is not changed for the transaction
func (s *TransactionPSQL) PutInitiatedPendingTransaction(merchantID, externalID, transaction, hash string) error {
	// TODO: check status transaction only "init" can be pending
	query := fmt.Sprintf("UPDATE %s set status = '%s', hash2 = $1 where guid = $2 and merchant_id = $3 and external_id = $4",
		s.namespace, InitTransaction)
	_, err := s.db.Query(query, hash, transaction, merchantID, externalID)
	if err != nil {
		return err
	}
	return nil
}

func (s *TransactionPSQL) PutProcessedTransaction(merchantID, externalID, transaction, hash string, commission float64) error {
	// TODO: check status transaction only "init" can be processed
	query := fmt.Sprintf("UPDATE %s set status = '%s', hash2 = $1, commission =$5 where guid = $2 and merchant_id = $3 and external_id = $4",
		s.namespace, ProcessedTransaction)
	_, err := s.db.Query(query, hash, transaction, merchantID, externalID, commission)
	if err != nil {
		return err
	}
	return nil
}

func (s *TransactionPSQL) PutSettledTransaction(merchantID, externalID, transaction, hash string) error {
	// TODO: check status transaction only "processed" can be settled
	query := fmt.Sprintf("UPDATE %s set status = '%s', hash3 = $1 where guid = $2 and merchant_id = $3 and external_id = $4",
		s.namespace, SettledTransaction)
	_, err := s.db.Query(query, hash, transaction, merchantID, externalID)
	if err != nil {
		return err
	}
	return nil
}

func (s *TransactionPSQL) PutDoneTransaction(merchantID, externalID, transaction, hash string) error {
	// TODO: check status transaction only "settled" can be done
	query := fmt.Sprintf("UPDATE %s set status = '%s', hash5 = $1 where guid = $2 and merchant_id = $3 and external_id = $4",
		s.namespace, DoneTransaction)
	_, err := s.db.Query(query, hash, transaction, merchantID, externalID)
	if err != nil {
		return err
	}
	return nil
}

func NewTransactionStorage(namespace string, db *sql.DB) (*TransactionPSQL, error) {
	s := TransactionPSQL{
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

func rowsToTransaction(rows *sql.Rows) ([]TransactionStore, error) {
	var transactions []TransactionStore
	if rows == nil {
		return transactions, nil
	}
	for rows.Next() {
		transaction := TransactionStore{}
		callBack := ""
		if err := rows.Scan(
			&transaction.Id, &transaction.GUID,
			&transaction.CreatedAt, &transaction.UpdatedAt, &transaction.DeletedAt,
			&transaction.MerchantId, &transaction.ExternalId, &transaction.Blockchain,
			&transaction.Action, &transaction.ExtWallet, &transaction.Status,
			&transaction.Asset, &transaction.Issuer,
			&transaction.Amount, &transaction.Commission,
			&transaction.Hash1, &transaction.Hash2,
			&transaction.Hash3, &transaction.Hash4,
			&transaction.Hash5, &callBack,
		); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}
