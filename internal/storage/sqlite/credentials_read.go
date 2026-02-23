package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mandalnilabja/goatway/internal/storage/models"
)

// GetCredential retrieves a credential by ID.
func (s *Storage) GetCredential(id string) (*models.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	var cred models.Credential
	var encryptedData string

	err := s.db.QueryRow(`
		SELECT id, provider, name, data, created_at, updated_at
		FROM credentials WHERE id = ?
	`, id).Scan(&cred.ID, &cred.Provider, &cred.Name, &encryptedData, &cred.CreatedAt, &cred.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Decrypt the credential data
	decryptedData, err := s.encryptor.Decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptionError, err)
	}

	cred.Data = json.RawMessage(decryptedData)

	return &cred, nil
}

// GetCredentialByName retrieves a credential by its unique name.
func (s *Storage) GetCredentialByName(name string) (*models.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	var cred models.Credential
	var encryptedData string

	err := s.db.QueryRow(`
		SELECT id, provider, name, data, created_at, updated_at
		FROM credentials WHERE name = ?
	`, name).Scan(&cred.ID, &cred.Provider, &cred.Name, &encryptedData, &cred.CreatedAt, &cred.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Decrypt the credential data
	decryptedData, err := s.encryptor.Decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptionError, err)
	}

	cred.Data = json.RawMessage(decryptedData)

	return &cred, nil
}

// ListCredentials retrieves all credentials.
func (s *Storage) ListCredentials() ([]*models.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	rows, err := s.db.Query(`
		SELECT id, provider, name, data, created_at, updated_at
		FROM credentials ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials []*models.Credential
	for rows.Next() {
		var cred models.Credential
		var encryptedData string

		err := rows.Scan(&cred.ID, &cred.Provider, &cred.Name, &encryptedData, &cred.CreatedAt, &cred.UpdatedAt)
		if err != nil {
			return nil, err
		}

		decryptedData, err := s.encryptor.Decrypt(encryptedData)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrEncryptionError, err)
		}

		cred.Data = json.RawMessage(decryptedData)
		credentials = append(credentials, &cred)
	}

	return credentials, rows.Err()
}
