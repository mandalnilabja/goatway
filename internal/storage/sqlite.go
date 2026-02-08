package storage

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// SQLiteStorage implements the Storage interface using SQLite
type SQLiteStorage struct {
	db        *sql.DB
	encryptor *AESEncryptor
	mu        sync.RWMutex
	closed    bool
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings for better concurrency
	db.SetMaxOpenConns(1) // SQLite works best with single writer
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	encryptor, err := NewEncryptor()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	storage := &SQLiteStorage{
		db:        db,
		encryptor: encryptor,
	}

	return storage, nil
}

// Migrate creates the database schema
func (s *SQLiteStorage) Migrate() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStorageClosed
	}

	schema := `
	CREATE TABLE IF NOT EXISTS credentials (
		id          TEXT PRIMARY KEY,
		provider    TEXT NOT NULL,
		name        TEXT NOT NULL,
		api_key     TEXT NOT NULL,
		is_default  INTEGER DEFAULT 0,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS request_logs (
		id                TEXT PRIMARY KEY,
		request_id        TEXT NOT NULL,
		credential_id     TEXT,
		model             TEXT NOT NULL,
		provider          TEXT NOT NULL,
		prompt_tokens     INTEGER DEFAULT 0,
		completion_tokens INTEGER DEFAULT 0,
		total_tokens      INTEGER DEFAULT 0,
		is_streaming      INTEGER DEFAULT 0,
		status_code       INTEGER,
		error_message     TEXT,
		duration_ms       INTEGER,
		created_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (credential_id) REFERENCES credentials(id) ON DELETE SET NULL
	);

	CREATE TABLE IF NOT EXISTS usage_daily (
		date              TEXT NOT NULL,
		credential_id     TEXT,
		model             TEXT NOT NULL,
		request_count     INTEGER DEFAULT 0,
		prompt_tokens     INTEGER DEFAULT 0,
		completion_tokens INTEGER DEFAULT 0,
		total_tokens      INTEGER DEFAULT 0,
		error_count       INTEGER DEFAULT 0,
		PRIMARY KEY (date, credential_id, model),
		FOREIGN KEY (credential_id) REFERENCES credentials(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_logs_created ON request_logs(created_at);
	CREATE INDEX IF NOT EXISTS idx_logs_model ON request_logs(model);
	CREATE INDEX IF NOT EXISTS idx_logs_credential ON request_logs(credential_id);
	CREATE INDEX IF NOT EXISTS idx_usage_date ON usage_daily(date);
	CREATE INDEX IF NOT EXISTS idx_creds_provider ON credentials(provider);
	`

	_, err := s.db.Exec(schema)
	return err
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	return s.db.Close()
}

// generateID creates a new unique ID with a prefix
func generateID(prefix string) string {
	return prefix + "_" + uuid.New().String()[:8]
}
