package sqlite

import (
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
