package sqlite

// migrateCredentials removes the is_default column from existing credentials tables.
// SQLite doesn't support DROP COLUMN, so we recreate the table.
func (s *Storage) migrateCredentials() error {
	// Check if is_default column exists
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('credentials') WHERE name = 'is_default'
	`).Scan(&count)
	if err != nil {
		return err
	}

	// No migration needed if column doesn't exist
	if count == 0 {
		return nil
	}

	// Migrate: create new table, copy data, drop old, rename new
	migration := `
	CREATE TABLE credentials_new (
		id          TEXT PRIMARY KEY,
		provider    TEXT NOT NULL,
		name        TEXT NOT NULL UNIQUE,
		data        TEXT NOT NULL,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	INSERT INTO credentials_new (id, provider, name, data, created_at, updated_at)
	SELECT id, provider, name, data, created_at, updated_at FROM credentials;

	DROP TABLE credentials;

	ALTER TABLE credentials_new RENAME TO credentials;

	CREATE INDEX IF NOT EXISTS idx_creds_provider ON credentials(provider);
	`

	_, err = s.db.Exec(migration)
	return err
}
