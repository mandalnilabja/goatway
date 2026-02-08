package storage

// boolToInt converts a boolean to an integer (1 for true, 0 for false)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// nullString returns nil for empty strings, otherwise the string itself
// Used for nullable foreign key columns
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
