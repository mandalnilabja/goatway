package sqlite

import (
	"database/sql"
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
		INSERT INTO credentials (id, provider, name, data, is_default, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, cred.ID, cred.Provider, cred.Name, encryptedData, boolToInt(cred.IsDefault), cred.CreatedAt, cred.UpdatedAt)

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
		SET provider = ?, name = ?, data = ?, is_default = ?, updated_at = ?
		WHERE id = ?
	`, cred.Provider, cred.Name, encryptedData, boolToInt(cred.IsDefault), cred.UpdatedAt, cred.ID)

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

// SetDefaultCredential sets a credential as the default for its provider.
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
