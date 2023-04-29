package internal

import "fmt"

// Deprecated???

var (
	// ErrNotFound Standardizing error "not found"
	ErrNotFound = fmt.Errorf("key not found")
	// ErrTypeMismatch Standardizing error "mismatch"
	ErrTypeMismatch = fmt.Errorf("key type mismatch")
)
