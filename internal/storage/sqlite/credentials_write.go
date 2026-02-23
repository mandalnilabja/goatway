package sqlite

import (
	"fmt"
	"time"

	"github.com/mandalnilabja/goatway/internal/storage/models"
)

// CreateCredential stores a new credential.
func (s *Storage) CreateCredential(cred *models.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStorageClosed
	}

	if cred.Provider == "" || cred.Name == "" || len(cred.Data) == 0 {
		return ErrInvalidInput
	}

	// Generate ID if not provided
	if cred.ID == "" {
		cred.ID = generateID("cred")
	}

	// Encrypt the credential data
	encryptedData, err := s.encryptor.Encrypt(string(cred.Data))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEncryptionError, err)
	}

	now := time.Now().UTC()
	cred.CreatedAt = now
	cred.UpdatedAt = now

	_, err = s.db.Exec(`
		INSERT INTO credentials (id, provider, name, data, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, cred.ID, cred.Provider, cred.Name, encryptedData, cred.CreatedAt, cred.UpdatedAt)

	return err
}

// UpdateCredential updates an existing credential.
func (s *Storage) UpdateCredential(cred *models.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStorageClosed
	}

	if cred.ID == "" {
		return ErrInvalidInput
	}

	// Encrypt the credential data
	encryptedData, err := s.encryptor.Encrypt(string(cred.Data))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEncryptionError, err)
	}

	cred.UpdatedAt = time.Now().UTC()

	result, err := s.db.Exec(`
		UPDATE credentials
		SET provider = ?, name = ?, data = ?, updated_at = ?
		WHERE id = ?
	`, cred.Provider, cred.Name, encryptedData, cred.UpdatedAt, cred.ID)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteCredential removes a credential by ID.
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
