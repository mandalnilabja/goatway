package sqlite

import (
	"database/sql"
	"encoding/json"

	"github.com/mandalnilabja/goatway/internal/storage/models"
)

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
