package storage

import "errors"

// Common errors returned by storage operations
var (
	ErrNotFound        = errors.New("record not found")
	ErrDuplicateKey    = errors.New("duplicate key")
	ErrInvalidInput    = errors.New("invalid input")
	ErrStorageClosed   = errors.New("storage is closed")
	ErrEncryptionError = errors.New("encryption error")
)

// Storage defines the interface for persistent data storage
type Storage interface {
	// Credential operations
	CreateCredential(cred *Credential) error
	GetCredential(id string) (*Credential, error)
	GetDefaultCredential(provider string) (*Credential, error)
	ListCredentials() ([]*Credential, error)
	UpdateCredential(cred *Credential) error
	DeleteCredential(id string) error
	SetDefaultCredential(id string) error

	// Request logging operations
	LogRequest(log *RequestLog) error
	GetRequestLogs(filter LogFilter) ([]*RequestLog, error)
	DeleteRequestLogs(olderThan string) (int64, error)

	// Usage statistics operations
	GetUsageStats(filter StatsFilter) (*UsageStats, error)
	GetDailyUsage(startDate, endDate string) ([]*DailyUsage, error)
	UpdateDailyUsage(usage *DailyUsage) error

	// Maintenance operations
	Close() error
	Migrate() error
}
