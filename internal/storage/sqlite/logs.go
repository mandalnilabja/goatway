package sqlite

import (
	"fmt"
	"time"

	"github.com/mandalnilabja/goatway/internal/storage/models"
)

// LogRequest stores a request log entry
func (s *Storage) LogRequest(log *models.RequestLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStorageClosed
	}

	if log.ID == "" {
		log.ID = generateID("log")
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO request_logs (id, request_id, credential_id, model, provider,
			prompt_tokens, completion_tokens, total_tokens, is_streaming,
			status_code, error_message, duration_ms, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, log.ID, log.RequestID, nullString(log.CredentialID), log.Model, log.Provider,
		log.PromptTokens, log.CompletionTokens, log.TotalTokens, boolToInt(log.IsStreaming),
		log.StatusCode, log.ErrorMessage, log.DurationMs, log.CreatedAt)

	return err
}

// GetRequestLogs retrieves request logs with filtering
func (s *Storage) GetRequestLogs(filter models.LogFilter) ([]*models.RequestLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	query := `SELECT id, request_id, COALESCE(credential_id, ''), model, provider,
		prompt_tokens, completion_tokens, total_tokens, is_streaming,
		status_code, COALESCE(error_message, ''), duration_ms, created_at
		FROM request_logs WHERE 1=1`

	var args []interface{}

	if filter.CredentialID != "" {
		query += " AND credential_id = ?"
		args = append(args, filter.CredentialID)
	}
	if filter.Model != "" {
		query += " AND model = ?"
		args = append(args, filter.Model)
	}
	if filter.Provider != "" {
		query += " AND provider = ?"
		args = append(args, filter.Provider)
	}
	if filter.StatusCode != nil {
		query += " AND status_code = ?"
		args = append(args, *filter.StatusCode)
	}
	if filter.StartDate != nil {
		query += " AND created_at >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		query += " AND created_at <= ?"
		args = append(args, *filter.EndDate)
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.RequestLog
	for rows.Next() {
		var log models.RequestLog
		var isStreaming int

		err := rows.Scan(&log.ID, &log.RequestID, &log.CredentialID, &log.Model, &log.Provider,
			&log.PromptTokens, &log.CompletionTokens, &log.TotalTokens, &isStreaming,
			&log.StatusCode, &log.ErrorMessage, &log.DurationMs, &log.CreatedAt)
		if err != nil {
			return nil, err
		}

		log.IsStreaming = isStreaming == 1
		logs = append(logs, &log)
	}

	return logs, rows.Err()
}

// DeleteRequestLogs removes logs older than the specified date
func (s *Storage) DeleteRequestLogs(olderThan string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return 0, ErrStorageClosed
	}

	result, err := s.db.Exec("DELETE FROM request_logs WHERE DATE(created_at) < ?", olderThan)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
