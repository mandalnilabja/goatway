package sqlite

import "github.com/mandalnilabja/goatway/internal/storage/models"

// UpdateDailyUsage upserts daily usage data
func (s *Storage) UpdateDailyUsage(usage *models.DailyUsage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStorageClosed
	}

	// Use empty string instead of NULL for credential_id to allow ON CONFLICT to work
	credID := usage.CredentialID
	if credID == "" {
		credID = ""
	}

	_, err := s.db.Exec(`
		INSERT INTO usage_daily (date, credential_id, model, request_count,
			prompt_tokens, completion_tokens, total_tokens, error_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(date, credential_id, model) DO UPDATE SET
			request_count = request_count + excluded.request_count,
			prompt_tokens = prompt_tokens + excluded.prompt_tokens,
			completion_tokens = completion_tokens + excluded.completion_tokens,
			total_tokens = total_tokens + excluded.total_tokens,
			error_count = error_count + excluded.error_count
	`, usage.Date, credID, usage.Model, usage.RequestCount,
		usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ErrorCount)

	return err
}
