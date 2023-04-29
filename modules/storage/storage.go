package storage

import (
	"errors"
	"time"
)

var (
	// ErrNotFound returned in case of key does not exist
	ErrNotFound     = errors.New("not found")
	ErrNotSupported = errors.New("unsupported")
)

const DefaultTTL = time.Hour * 24 * 365 * 111 // more than 100 years

// Record represents a record in the storage
type Record struct {
	ID        int64         `json:"ID"`
	Data      []byte        `json:"Data"`
	TTL       time.Duration `json:"TTL"`
	UpdatedAt time.Time     `json:"UpdatedAt"`
}

// Storage can store and retrieve transactions by Key
type Storage interface {
	// Put creates or updates a record in the storage with the key and data and returns
	// numberic ID of the created record
	Put(key string, data []byte, ttl time.Duration) (int64, error)

	// Set creates a record in the storage with the key and data and returns
	// numberic ID, true of the created record. If key has already exists,
	// then (0, false, nil) returned.
	Set(key string, data []byte, ttl time.Duration) (int64, bool, error)

	// Get returns the data associated with the key and it's numeric ID
	Get(key string) (int64, []byte, error)

	// GetNoTTL returns the data associated with the key ignoring the TTL value, and it's numeric ID
	// only use with AdminAPI methods if needed -
	// SetDepositSettled(), SetDepositInitial(), SetDepositCanceled(), SetPayoutPending(), SetPayoutSettled(), SetPayoutRejected
	GetNoTTL(key string) (int64, []byte, error)

	// GetRecord returns a record associated with the key
	GetRecord(key string) (*Record, error)

	// Delete deletes a record associated with the given key and returns the numeric ID of the deleted record.
	// numeric ID of the deleted record. If key doesn't exists
	// then (0, nil) returned
	Delete(key string) (int64, error)

	// GetAll returns every record from database
	GetAll() ([]Record, error)
}
