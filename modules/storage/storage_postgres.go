package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"time"
)

const (
	errCodePgUniqueViolation = "23505"
)

// StoragePSQL represents postgresql storage
type StoragePSQL struct {
	db        *sql.DB
	namespace string
}

// NewStorage creates new storage
func NewStorage(namespace string, db *sql.DB) (Storage, error) {
	s := StoragePSQL{
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

// Set creates a record in the storage with the key and data and returns
// numeric ID, true of the created record. If key has already exists and ttl is not expired,
// then (0, false, nil) returned.
func (s *StoragePSQL) Set(key string, data []byte, ttl time.Duration) (int64, bool, error) {
	row := s.db.QueryRow(fmt.Sprintf(`INSERT INTO %s(created_at, updated_at, key, value, ttl)
VALUES (now(), now(), $1, $2, $3)
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
WHERE now() > %s.updated_at + %s.ttl * interval '1 second'
RETURNING id`, s.namespace, s.namespace, s.namespace), key, data, ttl.Seconds())

	var id int64
	err := row.Scan(&id)

	if err == nil {
		return id, true, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}

	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		if pgErr.Code == errCodePgUniqueViolation {
			return 0, false, nil
		}
	}

	return 0, false, err
}

// Put creates or updates a record in the storage with the key and data and returns
// numberic ID of the created record
func (s *StoragePSQL) Put(key string, data []byte, ttl time.Duration) (int64, error) {
	row := s.db.QueryRow(
		fmt.Sprintf(`INSERT INTO "%s" (created_at, updated_at, key, value, ttl)
VALUES
		($1, $2, $3, $4, $5)
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, ttl=EXCLUDED.ttl
RETURNING id;`, s.namespace), time.Now(), time.Now(), key, data, ttl.Milliseconds()/1000)

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

// Get returns the data associated with the key and it's numeric ID
func (s *StoragePSQL) Get(key string) (int64, []byte, error) {
	r, err := s.GetRecord(key)
	if err != nil {
		return 0, nil, err
	}

	return r.ID, r.Data, nil
}

// GetNoTTL returns the data associated with the key ignoring the TTL value, and it's numeric ID
// only use with AdminAPI methods if needed -
// SetDepositSettled(), SetDepositInitial(), SetDepositCanceled(), SetPayoutPending(), SetPayoutSettled(), SetPayoutRejected
func (s *StoragePSQL) GetNoTTL(key string) (int64, []byte, error) {
	r, err := s.getRecordNoTTL(key)
	if err != nil {
		return 0, nil, err
	}

	return r.ID, r.Data, nil
}

// GetAll returns the data associated with the key and it's numeric ID
func (s *StoragePSQL) GetAll() ([]Record, error) {
	query := fmt.Sprintf(`SELECT id, value, updated_at
		FROM %s`,
		s.namespace,
	)
	r, err := s.getAll(query)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// GetRecord returns stored record and/or error
func (s *StoragePSQL) GetRecord(key string) (*Record, error) {
	query := fmt.Sprintf(`SELECT id, value, updated_at
		FROM %s 
		WHERE key = $1`,
		s.namespace,
	)
	return s.get(key, query)
}

// getRecordNoTTL returns stored record and/or error (TTL ignored)
func (s *StoragePSQL) getRecordNoTTL(key string) (*Record, error) {
	query := fmt.Sprintf(`SELECT id, value, updated_at FROM %s WHERE key = $1`,
		s.namespace,
	)
	return s.get(key, query)
}

// Delete deletes record from storage
func (s *StoragePSQL) Delete(key string) (int64, error) {
	row := s.db.QueryRow(fmt.Sprintf(`DELETE FROM "%s" WHERE key = $1 RETURNING id`, s.namespace), key)
	var id int64
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return id, nil
}

func (s *StoragePSQL) get(key string, query string) (*Record, error) {
	row := s.db.QueryRow(query, key)
	var (
		id        int64
		value     []byte
		updatedAt *time.Time
	)
	err := row.Scan(&id, &value, &updatedAt)
	if err == nil {
		return &Record{
				ID:        id,
				Data:      value,
				UpdatedAt: *updatedAt},
			nil
	}
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}

	return nil, fmt.Errorf("could not select row by key from storage: %w", err)
}

func (s *StoragePSQL) getAll(query string) ([]Record, error) {
	rows, err := s.db.Query(query)
	var (
		id        int64
		value     []byte
		updatedAt *time.Time
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var records []Record
	for rows.Next() {
		err := rows.Scan(&id, &value, &updatedAt)
		if err == nil {
			record := &Record{
				ID:        id,
				Data:      value,
				UpdatedAt: *updatedAt}
			records = append(records, *record)
		}
		if err != nil {
			return nil, ErrNotFound
		}
	}

	return records, nil
}
