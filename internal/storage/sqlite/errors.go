package sqlite

import "errors"

// Common errors returned by storage operations
var (
	ErrNotFound        = errors.New("record not found")
	ErrDuplicateKey    = errors.New("duplicate key")
	ErrInvalidInput    = errors.New("invalid input")
	ErrStorageClosed   = errors.New("storage is closed")
	ErrEncryptionError = errors.New("encryption error")
)
