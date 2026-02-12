// Package storage provides the storage interface and implementations.
package storage

import (
	"github.com/mandalnilabja/goatway/internal/storage/models"
	"github.com/mandalnilabja/goatway/internal/storage/sqlite"
)

// Re-export types from models package for convenience
type (
	Credential          = models.Credential
	CredentialPreview   = models.CredentialPreview
	ClientAPIKey        = models.ClientAPIKey
	ClientAPIKeyPreview = models.ClientAPIKeyPreview
	RequestLog          = models.RequestLog
	LogFilter           = models.LogFilter
	DailyUsage          = models.DailyUsage
	ModelStats          = models.ModelStats
	UsageStats          = models.UsageStats
	StatsFilter         = models.StatsFilter
)

// Re-export functions from models package
var MaskAPIKey = models.MaskAPIKey

// Re-export errors from sqlite package
var (
	ErrNotFound        = sqlite.ErrNotFound
	ErrDuplicateKey    = sqlite.ErrDuplicateKey
	ErrInvalidInput    = sqlite.ErrInvalidInput
	ErrStorageClosed   = sqlite.ErrStorageClosed
	ErrEncryptionError = sqlite.ErrEncryptionError
)

// Storage defines the interface for persistent data storage
type Storage interface {
	// Credential operations
	CreateCredential(cred *models.Credential) error
	GetCredential(id string) (*models.Credential, error)
	GetDefaultCredential(provider string) (*models.Credential, error)
	ListCredentials() ([]*models.Credential, error)
	UpdateCredential(cred *models.Credential) error
	DeleteCredential(id string) error
	SetDefaultCredential(id string) error

	// Request logging operations
	LogRequest(log *models.RequestLog) error
	GetRequestLogs(filter models.LogFilter) ([]*models.RequestLog, error)
	DeleteRequestLogs(olderThan string) (int64, error)

	// Usage statistics operations
	GetUsageStats(filter models.StatsFilter) (*models.UsageStats, error)
	GetDailyUsage(startDate, endDate string) ([]*models.DailyUsage, error)
	UpdateDailyUsage(usage *models.DailyUsage) error

	// Client API key operations
	CreateAPIKey(key *models.ClientAPIKey) error
	GetAPIKey(id string) (*models.ClientAPIKey, error)
	GetAPIKeyByPrefix(prefix string) ([]*models.ClientAPIKey, error)
	ListAPIKeys() ([]*models.ClientAPIKey, error)
	UpdateAPIKey(key *models.ClientAPIKey) error
	DeleteAPIKey(id string) error
	UpdateAPIKeyLastUsed(id string) error

	// Admin password operations
	GetAdminPasswordHash() (string, error)
	SetAdminPasswordHash(hash string) error
	HasAdminPassword() (bool, error)

	// Maintenance operations
	Close() error
}

// NewSQLiteStorage creates a new SQLite storage instance
// This is the main factory function for creating storage
func NewSQLiteStorage(dbPath string) (Storage, error) {
	return sqlite.New(dbPath)
}
