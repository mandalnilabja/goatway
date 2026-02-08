package storage

import (
	"database/sql"
	"fmt"
)

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
