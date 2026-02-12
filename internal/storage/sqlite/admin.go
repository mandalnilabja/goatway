package sqlite

import "database/sql"

const adminPasswordKey = "admin_password_hash"

// GetAdminPasswordHash retrieves the stored admin password hash
func (s *Storage) GetAdminPasswordHash() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return "", ErrStorageClosed
	}

	var hash string
	err := s.db.QueryRow(
		"SELECT value FROM admin_settings WHERE key = ?",
		adminPasswordKey,
	).Scan(&hash)

	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return hash, nil
}

// SetAdminPasswordHash stores the admin password hash
func (s *Storage) SetAdminPasswordHash(hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStorageClosed
	}

	_, err := s.db.Exec(`
		INSERT INTO admin_settings (key, value, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = CURRENT_TIMESTAMP
	`, adminPasswordKey, hash, hash)

	return err
}

// HasAdminPassword checks if an admin password has been configured
func (s *Storage) HasAdminPassword() (bool, error) {
	hash, err := s.GetAdminPasswordHash()
	if err != nil {
		return false, err
	}
	return hash != "", nil
}
