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

// CreateCredential stores a new credential
func (s *SQLiteStorage) CreateCredential(cred *Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStorageClosed
	}

	if cred.Provider == "" || cred.Name == "" || cred.APIKey == "" {
		return ErrInvalidInput
	}

	// Generate ID if not provided
	if cred.ID == "" {
		cred.ID = generateID("cred")
	}

	// Encrypt the API key
	encryptedKey, err := s.encryptor.Encrypt(cred.APIKey)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEncryptionError, err)
	}

	now := time.Now().UTC()
	cred.CreatedAt = now
	cred.UpdatedAt = now

	// If this credential is marked as default, unset other defaults for this provider
	if cred.IsDefault {
		_, err := s.db.Exec(
			"UPDATE credentials SET is_default = 0 WHERE provider = ?",
			cred.Provider,
		)
		if err != nil {
			return err
		}
	}

	_, err = s.db.Exec(`
		INSERT INTO credentials (id, provider, name, api_key, is_default, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, cred.ID, cred.Provider, cred.Name, encryptedKey, boolToInt(cred.IsDefault), cred.CreatedAt, cred.UpdatedAt)

	return err
}

// GetCredential retrieves a credential by ID
func (s *SQLiteStorage) GetCredential(id string) (*Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	var cred Credential
	var isDefault int
	var encryptedKey string

	err := s.db.QueryRow(`
		SELECT id, provider, name, api_key, is_default, created_at, updated_at
		FROM credentials WHERE id = ?
	`, id).Scan(&cred.ID, &cred.Provider, &cred.Name, &encryptedKey, &isDefault, &cred.CreatedAt, &cred.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Decrypt the API key
	decryptedKey, err := s.encryptor.Decrypt(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptionError, err)
	}

	cred.APIKey = decryptedKey
	cred.IsDefault = isDefault == 1

	return &cred, nil
}

// GetDefaultCredential retrieves the default credential for a provider
func (s *SQLiteStorage) GetDefaultCredential(provider string) (*Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	var cred Credential
	var isDefault int
	var encryptedKey string

	err := s.db.QueryRow(`
		SELECT id, provider, name, api_key, is_default, created_at, updated_at
		FROM credentials WHERE provider = ? AND is_default = 1
	`, provider).Scan(&cred.ID, &cred.Provider, &cred.Name, &encryptedKey, &isDefault, &cred.CreatedAt, &cred.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Decrypt the API key
	decryptedKey, err := s.encryptor.Decrypt(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptionError, err)
	}

	cred.APIKey = decryptedKey
	cred.IsDefault = true

	return &cred, nil
}

// ListCredentials retrieves all credentials
func (s *SQLiteStorage) ListCredentials() ([]*Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	rows, err := s.db.Query(`
		SELECT id, provider, name, api_key, is_default, created_at, updated_at
		FROM credentials ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials []*Credential
	for rows.Next() {
		var cred Credential
		var isDefault int
		var encryptedKey string

		err := rows.Scan(&cred.ID, &cred.Provider, &cred.Name, &encryptedKey, &isDefault, &cred.CreatedAt, &cred.UpdatedAt)
		if err != nil {
			return nil, err
		}

		decryptedKey, err := s.encryptor.Decrypt(encryptedKey)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrEncryptionError, err)
		}

		cred.APIKey = decryptedKey
		cred.IsDefault = isDefault == 1
		credentials = append(credentials, &cred)
	}

	return credentials, rows.Err()
}

// UpdateCredential updates an existing credential
func (s *SQLiteStorage) UpdateCredential(cred *Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStorageClosed
	}

	if cred.ID == "" {
		return ErrInvalidInput
	}

	// Encrypt the API key if provided
	encryptedKey, err := s.encryptor.Encrypt(cred.APIKey)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEncryptionError, err)
	}

	cred.UpdatedAt = time.Now().UTC()

	result, err := s.db.Exec(`
		UPDATE credentials
		SET provider = ?, name = ?, api_key = ?, is_default = ?, updated_at = ?
		WHERE id = ?
	`, cred.Provider, cred.Name, encryptedKey, boolToInt(cred.IsDefault), cred.UpdatedAt, cred.ID)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteCredential removes a credential by ID
func (s *SQLiteStorage) DeleteCredential(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStorageClosed
	}

	result, err := s.db.Exec("DELETE FROM credentials WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// SetDefaultCredential sets a credential as the default for its provider
func (s *SQLiteStorage) SetDefaultCredential(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStorageClosed
	}

	// Get the credential to find its provider
	var provider string
	err := s.db.QueryRow("SELECT provider FROM credentials WHERE id = ?", id).Scan(&provider)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	// Unset all defaults for this provider
	_, err = s.db.Exec("UPDATE credentials SET is_default = 0 WHERE provider = ?", provider)
	if err != nil {
		return err
	}

	// Set this credential as default
	_, err = s.db.Exec("UPDATE credentials SET is_default = 1, updated_at = ? WHERE id = ?", time.Now().UTC(), id)
	return err
}

// LogRequest stores a request log entry
func (s *SQLiteStorage) LogRequest(log *RequestLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStorageClosed
	}

	if log.ID == "" {
		log.ID = generateID("log")
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO request_logs (id, request_id, credential_id, model, provider,
			prompt_tokens, completion_tokens, total_tokens, is_streaming,
			status_code, error_message, duration_ms, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, log.ID, log.RequestID, nullString(log.CredentialID), log.Model, log.Provider,
		log.PromptTokens, log.CompletionTokens, log.TotalTokens, boolToInt(log.IsStreaming),
		log.StatusCode, log.ErrorMessage, log.DurationMs, log.CreatedAt)

	return err
}

// GetRequestLogs retrieves request logs with filtering
func (s *SQLiteStorage) GetRequestLogs(filter LogFilter) ([]*RequestLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	query := `SELECT id, request_id, COALESCE(credential_id, ''), model, provider,
		prompt_tokens, completion_tokens, total_tokens, is_streaming,
		status_code, COALESCE(error_message, ''), duration_ms, created_at
		FROM request_logs WHERE 1=1`

	var args []interface{}

	if filter.CredentialID != "" {
		query += " AND credential_id = ?"
		args = append(args, filter.CredentialID)
	}
	if filter.Model != "" {
		query += " AND model = ?"
		args = append(args, filter.Model)
	}
	if filter.Provider != "" {
		query += " AND provider = ?"
		args = append(args, filter.Provider)
	}
	if filter.StatusCode != nil {
		query += " AND status_code = ?"
		args = append(args, *filter.StatusCode)
	}
	if filter.StartDate != nil {
		query += " AND created_at >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		query += " AND created_at <= ?"
		args = append(args, *filter.EndDate)
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*RequestLog
	for rows.Next() {
		var log RequestLog
		var isStreaming int

		err := rows.Scan(&log.ID, &log.RequestID, &log.CredentialID, &log.Model, &log.Provider,
			&log.PromptTokens, &log.CompletionTokens, &log.TotalTokens, &isStreaming,
			&log.StatusCode, &log.ErrorMessage, &log.DurationMs, &log.CreatedAt)
		if err != nil {
			return nil, err
		}

		log.IsStreaming = isStreaming == 1
		logs = append(logs, &log)
	}

	return logs, rows.Err()
}

// DeleteRequestLogs removes logs older than the specified date
func (s *SQLiteStorage) DeleteRequestLogs(olderThan string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return 0, ErrStorageClosed
	}

	result, err := s.db.Exec("DELETE FROM request_logs WHERE DATE(created_at) < ?", olderThan)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// GetUsageStats retrieves aggregated usage statistics
func (s *SQLiteStorage) GetUsageStats(filter StatsFilter) (*UsageStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	query := `SELECT
		COALESCE(SUM(request_count), 0),
		COALESCE(SUM(prompt_tokens), 0),
		COALESCE(SUM(completion_tokens), 0),
		COALESCE(SUM(total_tokens), 0),
		COALESCE(SUM(error_count), 0)
		FROM usage_daily WHERE 1=1`

	var args []interface{}

	if filter.CredentialID != "" {
		query += " AND credential_id = ?"
		args = append(args, filter.CredentialID)
	}
	if filter.StartDate != nil {
		query += " AND date >= ?"
		args = append(args, filter.StartDate.Format("2006-01-02"))
	}
	if filter.EndDate != nil {
		query += " AND date <= ?"
		args = append(args, filter.EndDate.Format("2006-01-02"))
	}

	stats := &UsageStats{
		ModelBreakdown: make(map[string]*ModelStats),
	}

	err := s.db.QueryRow(query, args...).Scan(
		&stats.TotalRequests,
		&stats.TotalPromptTokens,
		&stats.TotalCompletionTokens,
		&stats.TotalTokens,
		&stats.ErrorCount,
	)
	if err != nil {
		return nil, err
	}

	// Get model breakdown
	modelQuery := `SELECT model,
		COALESCE(SUM(request_count), 0),
		COALESCE(SUM(prompt_tokens), 0),
		COALESCE(SUM(completion_tokens), 0),
		COALESCE(SUM(total_tokens), 0),
		COALESCE(SUM(error_count), 0)
		FROM usage_daily WHERE 1=1`

	if filter.CredentialID != "" {
		modelQuery += " AND credential_id = ?"
	}
	if filter.StartDate != nil {
		modelQuery += " AND date >= ?"
	}
	if filter.EndDate != nil {
		modelQuery += " AND date <= ?"
	}
	modelQuery += " GROUP BY model"

	rows, err := s.db.Query(modelQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ms ModelStats
		err := rows.Scan(&ms.Model, &ms.RequestCount, &ms.PromptTokens,
			&ms.CompletionTokens, &ms.TotalTokens, &ms.ErrorCount)
		if err != nil {
			return nil, err
		}
		stats.ModelBreakdown[ms.Model] = &ms
	}

	return stats, rows.Err()
}

// GetDailyUsage retrieves daily usage data for a date range
func (s *SQLiteStorage) GetDailyUsage(startDate, endDate string) ([]*DailyUsage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	rows, err := s.db.Query(`
		SELECT date, COALESCE(credential_id, ''), model, request_count,
			prompt_tokens, completion_tokens, total_tokens, error_count
		FROM usage_daily
		WHERE date >= ? AND date <= ?
		ORDER BY date ASC, model ASC
	`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usage []*DailyUsage
	for rows.Next() {
		var u DailyUsage
		err := rows.Scan(&u.Date, &u.CredentialID, &u.Model, &u.RequestCount,
			&u.PromptTokens, &u.CompletionTokens, &u.TotalTokens, &u.ErrorCount)
		if err != nil {
			return nil, err
		}
		usage = append(usage, &u)
	}

	return usage, rows.Err()
}

// UpdateDailyUsage upserts daily usage data
func (s *SQLiteStorage) UpdateDailyUsage(usage *DailyUsage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStorageClosed
	}

	// Use empty string instead of NULL for credential_id to allow ON CONFLICT to work
	credID := usage.CredentialID
	if credID == "" {
		credID = ""
	}

	_, err := s.db.Exec(`
		INSERT INTO usage_daily (date, credential_id, model, request_count,
			prompt_tokens, completion_tokens, total_tokens, error_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(date, credential_id, model) DO UPDATE SET
			request_count = request_count + excluded.request_count,
			prompt_tokens = prompt_tokens + excluded.prompt_tokens,
			completion_tokens = completion_tokens + excluded.completion_tokens,
			total_tokens = total_tokens + excluded.total_tokens,
			error_count = error_count + excluded.error_count
	`, usage.Date, credID, usage.Model, usage.RequestCount,
		usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ErrorCount)

	return err
}

// Helper functions

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
