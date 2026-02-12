package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/mandalnilabja/goatway/internal/storage/models"
)

// CreateCredential stores a new credential
func (s *Storage) CreateCredential(cred *models.Credential) error {
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
func (s *Storage) GetCredential(id string) (*models.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	var cred models.Credential
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
func (s *Storage) GetDefaultCredential(provider string) (*models.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	var cred models.Credential
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
func (s *Storage) ListCredentials() ([]*models.Credential, error) {
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

	var credentials []*models.Credential
	for rows.Next() {
		var cred models.Credential
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
func (s *Storage) UpdateCredential(cred *models.Credential) error {
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
func (s *Storage) DeleteCredential(id string) error {
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
func (s *Storage) SetDefaultCredential(id string) error {
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
