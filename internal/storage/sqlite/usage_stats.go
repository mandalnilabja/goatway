package sqlite

import "github.com/mandalnilabja/goatway/internal/storage/models"

// GetUsageStats retrieves aggregated usage statistics
func (s *Storage) GetUsageStats(filter models.StatsFilter) (*models.UsageStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	query := `SELECT
		COALESCE(SUM(request_count), 0),
		COALESCE(SUM(prompt_tokens), 0),
		COALESCE(SUM(completion_tokens), 0),
		COALESCE(SUM(total_tokens), 0),
		COALESCE(SUM(error_count), 0)
		FROM usage_daily WHERE 1=1`

	var args []interface{}

	if filter.CredentialID != "" {
		query += " AND credential_id = ?"
		args = append(args, filter.CredentialID)
	}
	if filter.StartDate != nil {
		query += " AND date >= ?"
		args = append(args, filter.StartDate.Format("2006-01-02"))
	}
	if filter.EndDate != nil {
		query += " AND date <= ?"
		args = append(args, filter.EndDate.Format("2006-01-02"))
	}

	stats := &models.UsageStats{
		ModelBreakdown: make(map[string]*models.ModelStats),
	}

	err := s.db.QueryRow(query, args...).Scan(
		&stats.TotalRequests,
		&stats.TotalPromptTokens,
		&stats.TotalCompletionTokens,
		&stats.TotalTokens,
		&stats.ErrorCount,
	)
	if err != nil {
		return nil, err
	}

	// Get model breakdown
	modelQuery := `SELECT model,
		COALESCE(SUM(request_count), 0),
		COALESCE(SUM(prompt_tokens), 0),
		COALESCE(SUM(completion_tokens), 0),
		COALESCE(SUM(total_tokens), 0),
		COALESCE(SUM(error_count), 0)
		FROM usage_daily WHERE 1=1`

	if filter.CredentialID != "" {
		modelQuery += " AND credential_id = ?"
	}
	if filter.StartDate != nil {
		modelQuery += " AND date >= ?"
	}
	if filter.EndDate != nil {
		modelQuery += " AND date <= ?"
	}
	modelQuery += " GROUP BY model"

	rows, err := s.db.Query(modelQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ms models.ModelStats
		err := rows.Scan(&ms.Model, &ms.RequestCount, &ms.PromptTokens,
			&ms.CompletionTokens, &ms.TotalTokens, &ms.ErrorCount)
		if err != nil {
			return nil, err
		}
		stats.ModelBreakdown[ms.Model] = &ms
	}

	return stats, rows.Err()
}

// GetDailyUsage retrieves daily usage data for a date range
func (s *Storage) GetDailyUsage(startDate, endDate string) ([]*models.DailyUsage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStorageClosed
	}

	rows, err := s.db.Query(`
		SELECT date, COALESCE(credential_id, ''), model, request_count,
			prompt_tokens, completion_tokens, total_tokens, error_count
		FROM usage_daily
		WHERE date >= ? AND date <= ?
		ORDER BY date ASC, model ASC
	`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usage []*models.DailyUsage
	for rows.Next() {
		var u models.DailyUsage
		err := rows.Scan(&u.Date, &u.CredentialID, &u.Model, &u.RequestCount,
			&u.PromptTokens, &u.CompletionTokens, &u.TotalTokens, &u.ErrorCount)
		if err != nil {
			return nil, err
		}
		usage = append(usage, &u)
	}

	return usage, rows.Err()
}
