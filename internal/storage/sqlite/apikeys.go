package sqlite

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/mandalnilabja/goatway/internal/storage/models"
)

// CreateAPIKey creates a new client API key
func (s *Storage) CreateAPIKey(key *models.ClientAPIKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStorageClosed
	}

	scopesJSON, err := json.Marshal(key.Scopes)
	if err != nil {
		return err
	}

	if key.ID == "" {
		key.ID = uuid.New().String()
	}
	key.CreatedAt = time.Now()

	_, err = s.db.Exec(`
		INSERT INTO api_keys (id, name, key_hash, key_prefix, scopes, rate_limit, is_active, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, key.ID, key.Name, key.KeyHash, key.KeyPrefix, string(scopesJSON),
		key.RateLimit, key.IsActive, key.ExpiresAt, key.CreatedAt)

	return err
}

// GetAPIKey retrieves an API key by ID
func (s *Storage) GetAPIKey(id string) (*models.ClientAPIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	var key models.ClientAPIKey
	var scopesJSON string
	var lastUsedAt, expiresAt sql.NullTime

	err := s.db.QueryRow(`
		SELECT id, name, key_hash, key_prefix, scopes, rate_limit, is_active, last_used_at, created_at, expires_at
		FROM api_keys WHERE id = ?
	`, id).Scan(
		&key.ID, &key.Name, &key.KeyHash, &key.KeyPrefix, &scopesJSON,
		&key.RateLimit, &key.IsActive, &lastUsedAt, &key.CreatedAt, &expiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(scopesJSON), &key.Scopes); err != nil {
		return nil, err
	}

	if lastUsedAt.Valid {
		key.LastUsedAt = &lastUsedAt.Time
	}
	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Time
	}

	return &key, nil
}

// GetAPIKeyByPrefix retrieves API keys matching a prefix
func (s *Storage) GetAPIKeyByPrefix(prefix string) ([]*models.ClientAPIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	rows, err := s.db.Query(`
		SELECT id, name, key_hash, key_prefix, scopes, rate_limit, is_active, last_used_at, created_at, expires_at
		FROM api_keys WHERE key_prefix = ?
	`, prefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAPIKeys(rows)
}

// ListAPIKeys returns all API keys
func (s *Storage) ListAPIKeys() ([]*models.ClientAPIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	rows, err := s.db.Query(`
		SELECT id, name, key_hash, key_prefix, scopes, rate_limit, is_active, last_used_at, created_at, expires_at
		FROM api_keys ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAPIKeys(rows)
}

// UpdateAPIKey updates an existing API key
func (s *Storage) UpdateAPIKey(key *models.ClientAPIKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStorageClosed
	}

	scopesJSON, err := json.Marshal(key.Scopes)
	if err != nil {
		return err
	}

	result, err := s.db.Exec(`
		UPDATE api_keys
		SET name = ?, key_hash = ?, key_prefix = ?, scopes = ?, rate_limit = ?, is_active = ?, expires_at = ?
		WHERE id = ?
	`, key.Name, key.KeyHash, key.KeyPrefix, string(scopesJSON),
		key.RateLimit, key.IsActive, key.ExpiresAt, key.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteAPIKey deletes an API key by ID
func (s *Storage) DeleteAPIKey(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStorageClosed
	}

	result, err := s.db.Exec("DELETE FROM api_keys WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// UpdateAPIKeyLastUsed updates the last_used_at timestamp for an API key
func (s *Storage) UpdateAPIKeyLastUsed(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStorageClosed
	}

	_, err := s.db.Exec(
		"UPDATE api_keys SET last_used_at = ? WHERE id = ?",
		time.Now(), id,
	)
	return err
}

// scanAPIKeys is a helper to scan rows into ClientAPIKey slice
func scanAPIKeys(rows *sql.Rows) ([]*models.ClientAPIKey, error) {
	var keys []*models.ClientAPIKey

	for rows.Next() {
		var key models.ClientAPIKey
		var scopesJSON string
		var lastUsedAt, expiresAt sql.NullTime

		err := rows.Scan(
			&key.ID, &key.Name, &key.KeyHash, &key.KeyPrefix, &scopesJSON,
			&key.RateLimit, &key.IsActive, &lastUsedAt, &key.CreatedAt, &expiresAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(scopesJSON), &key.Scopes); err != nil {
			return nil, err
		}

		if lastUsedAt.Valid {
			key.LastUsedAt = &lastUsedAt.Time
		}
		if expiresAt.Valid {
			key.ExpiresAt = &expiresAt.Time
		}

		keys = append(keys, &key)
	}

	return keys, rows.Err()
}
