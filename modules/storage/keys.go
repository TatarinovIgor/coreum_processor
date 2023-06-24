package storage

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"time"
)

// KeyRecord represents a record in the storage
type KeyRecord struct {
	ID         int64
	MerchantID string
	ExternalID string
	Key        string
	Data       []byte
	UpdatedAt  time.Time
}
type KeysPSQL struct {
	db        *sql.DB
	namespace string
}

// NewKeys creates new storage for blockchain wallets
func NewKeys(namespace string, db *sql.DB) (*KeysPSQL, error) {
	s := KeysPSQL{
		db:        db,
		namespace: namespace,
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not connect to DB: %v", err)
	}
	if _, err := db.Exec(fmt.Sprintf("SELECT 1 FROM %q LIMIT 1", namespace)); err != nil {
		return nil, fmt.Errorf("could not connect to storage: %v", err)
	}
	return &s, nil
}

// Put creates or updates a record in the storage with the key and data and returns
// numberic ID of the created record
func (s *KeysPSQL) Put(merchantID, externalID, key string, data []byte) (int64, error) {
	row := s.db.QueryRow(
		fmt.Sprintf(`INSERT INTO "%s" (created_at, updated_at, merchant_id, external_id, key, value)
VALUES
		($1, $2, $3, $4, $5, $6)
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
RETURNING id;`, s.namespace), time.Now().UTC(), time.Now().UTC(), merchantID, externalID, key, data)

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

// Set creates a record in the storage with the key and data and returns
// numberic ID, true of the created record. If key has already exists,
// then (0, false, nil) returned.
func (s *KeysPSQL) Set(merchantID, externalID, key string, data []byte) (int64, bool, error) {
	row := s.db.QueryRow(fmt.Sprintf(`INSERT INTO %s(merchant_id, external_id, key, value)
VALUES ($1, $2, $3, $4)
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
RETURNING id`, s.namespace, s.namespace, s.namespace), merchantID, externalID, key, data)

	var id int64
	err := row.Scan(&id)

	if err == nil {
		return id, true, nil
	}

	if err == sql.ErrNoRows {
		return 0, false, nil
	}

	if pgErr, ok := err.(*pq.Error); ok {
		if pgErr.Code == errCodePgUniqueViolation {
			return 0, false, nil
		}
	}

	return 0, false, err
}

// GetByKey returns the data associated with the key and it's numeric ID
func (s *KeysPSQL) GetByKey(key string) (int64, []byte, error) {
	r, err := s.GetRecordByKey(key)
	if err != nil {
		return 0, nil, err
	}

	return r.ID, r.Data, nil
}

// GetRecordByKey returns a record associated with the key
func (s *KeysPSQL) GetRecordByKey(key string) (*KeyRecord, error) {
	query := fmt.Sprintf(`SELECT id, merchant_id, external_id, value, updated_at
		FROM %s 
		WHERE key = $1`,
		s.namespace,
	)
	return s.getByKey(key, query)
}

// GetByUser returns the data associated with the key and it's numeric ID
func (s *KeysPSQL) GetByUser(merchantID, externalID string) (int64, []byte, error) {
	r, err := s.GetRecordByUser(merchantID, externalID)
	if err != nil {
		return 0, nil, err
	}

	return r.ID, r.Data, nil
}

// GetRecordByUser returns a record associated with the key
func (s *KeysPSQL) GetRecordByUser(merchantID, externalID string) (*KeyRecord, error) {
	query := fmt.Sprintf(`SELECT id, key, value, updated_at
		FROM %s 
		WHERE merchant_id = $1 and external_id = $2`,
		s.namespace,
	)
	return s.getByUser(merchantID, externalID, query)
}

// GetRecordsByMerchant returns a record associated with the key
func (s *KeysPSQL) GetRecordsByMerchant(merchantID, externalID string) ([]KeyRecord, error) {
	panic("implement me")
}

// DeleteByKey deletes a record associated with the given key and returns the numeric ID of the deleted record.
// numeric ID of the deleted record. If key doesn't exist
// then (0, nil) returned
func (s *KeysPSQL) DeleteByKey(key string) (int64, error) {
	panic("implement me")
}

func (s *KeysPSQL) getByKey(key string, query string) (*KeyRecord, error) {
	row := s.db.QueryRow(query, key)
	var (
		id         int64
		merchantID string
		externalID string
		value      []byte
		updatedAt  *time.Time
	)
	err := row.Scan(&id, &merchantID, &externalID, &value, &updatedAt)
	if err == nil {
		return &KeyRecord{
				ID:         id,
				MerchantID: merchantID,
				ExternalID: externalID,
				Key:        key,
				Data:       value,
				UpdatedAt:  *updatedAt},
			nil
	}
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}

	return nil, fmt.Errorf("could not select row by key from storage: %w", err)
}

func (s *KeysPSQL) getByUser(merchantID, externalID string, query string) (*KeyRecord, error) {
	row := s.db.QueryRow(query, merchantID, externalID)
	var (
		id        int64
		key       string
		value     []byte
		updatedAt *time.Time
	)
	err := row.Scan(&id, &key, &value, &updatedAt)
	if err == nil {
		return &KeyRecord{
				ID:         id,
				MerchantID: merchantID,
				ExternalID: externalID,
				Key:        key,
				Data:       value,
				UpdatedAt:  *updatedAt},
			nil
	}
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}

	return nil, fmt.Errorf("could not select row by user from storage: %w", err)
}
func (s *KeysPSQL) GetNext(id, limit int64) ([]KeyRecord, error) {
	query := fmt.Sprintf(`SELECT id, merchant_id, external_id, key, value, updated_at
		FROM %s 
		WHERE id > $1 and deleted_at is null order by id LIMIT $2`,
		s.namespace,
	)
	rows, err := s.db.Query(query, id, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var records []KeyRecord
	for rows.Next() {
		var (
			key        string
			merchantID string
			externalID string
			value      []byte
			updatedAt  *time.Time
		)
		err := rows.Scan(&id, &merchantID, &externalID, &key, &value, &updatedAt)
		if err == nil {
			record := &KeyRecord{
				ID:         id,
				MerchantID: merchantID,
				ExternalID: externalID,
				Key:        key,
				Data:       value,
				UpdatedAt:  *updatedAt}
			records = append(records, *record)
		}
		if err != nil {
			return nil, ErrNotFound
		}
	}
	return records, nil
}
