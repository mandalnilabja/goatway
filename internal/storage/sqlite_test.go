package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) (*SQLiteStorage, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "goatway-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := NewSQLiteStorage(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create storage: %v", err)
	}

	if err := storage.Migrate(); err != nil {
		storage.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to migrate: %v", err)
	}

	cleanup := func() {
		storage.Close()
		os.RemoveAll(tmpDir)
	}

	return storage, cleanup
}

func TestCredentialCRUD(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	// Create credential
	cred := &Credential{
		Provider:  "openrouter",
		Name:      "Test Key",
		APIKey:    "sk-test-key-12345",
		IsDefault: true,
	}

	err := storage.CreateCredential(cred)
	if err != nil {
		t.Fatalf("CreateCredential failed: %v", err)
	}

	if cred.ID == "" {
		t.Error("expected ID to be generated")
	}

	// Get credential
	retrieved, err := storage.GetCredential(cred.ID)
	if err != nil {
		t.Fatalf("GetCredential failed: %v", err)
	}

	if retrieved.Name != cred.Name {
		t.Errorf("expected name %q, got %q", cred.Name, retrieved.Name)
	}
	if retrieved.APIKey != cred.APIKey {
		t.Errorf("expected API key %q, got %q", cred.APIKey, retrieved.APIKey)
	}
	if !retrieved.IsDefault {
		t.Error("expected credential to be default")
	}

	// Update credential
	retrieved.Name = "Updated Key"
	err = storage.UpdateCredential(retrieved)
	if err != nil {
		t.Fatalf("UpdateCredential failed: %v", err)
	}

	updated, err := storage.GetCredential(cred.ID)
	if err != nil {
		t.Fatalf("GetCredential after update failed: %v", err)
	}
	if updated.Name != "Updated Key" {
		t.Errorf("expected name %q, got %q", "Updated Key", updated.Name)
	}

	// List credentials
	list, err := storage.ListCredentials()
	if err != nil {
		t.Fatalf("ListCredentials failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 credential, got %d", len(list))
	}

	// Delete credential
	err = storage.DeleteCredential(cred.ID)
	if err != nil {
		t.Fatalf("DeleteCredential failed: %v", err)
	}

	_, err = storage.GetCredential(cred.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDefaultCredential(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	// Create first credential as default
	cred1 := &Credential{
		Provider:  "openrouter",
		Name:      "First Key",
		APIKey:    "sk-first-key",
		IsDefault: true,
	}
	if err := storage.CreateCredential(cred1); err != nil {
		t.Fatalf("CreateCredential failed: %v", err)
	}

	// Create second credential as default (should unset first)
	cred2 := &Credential{
		Provider:  "openrouter",
		Name:      "Second Key",
		APIKey:    "sk-second-key",
		IsDefault: true,
	}
	if err := storage.CreateCredential(cred2); err != nil {
		t.Fatalf("CreateCredential failed: %v", err)
	}

	// Get default should return second
	defaultCred, err := storage.GetDefaultCredential("openrouter")
	if err != nil {
		t.Fatalf("GetDefaultCredential failed: %v", err)
	}
	if defaultCred.ID != cred2.ID {
		t.Errorf("expected default to be %q, got %q", cred2.ID, defaultCred.ID)
	}

	// Set first as default
	err = storage.SetDefaultCredential(cred1.ID)
	if err != nil {
		t.Fatalf("SetDefaultCredential failed: %v", err)
	}

	defaultCred, err = storage.GetDefaultCredential("openrouter")
	if err != nil {
		t.Fatalf("GetDefaultCredential failed: %v", err)
	}
	if defaultCred.ID != cred1.ID {
		t.Errorf("expected default to be %q, got %q", cred1.ID, defaultCred.ID)
	}
}

func TestRequestLogging(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a credential first
	cred := &Credential{
		Provider:  "openrouter",
		Name:      "Test Key",
		APIKey:    "sk-test",
		IsDefault: true,
	}
	if err := storage.CreateCredential(cred); err != nil {
		t.Fatalf("CreateCredential failed: %v", err)
	}

	// Log a request
	log := &RequestLog{
		RequestID:        "req-123",
		CredentialID:     cred.ID,
		Model:            "gpt-4",
		Provider:         "openrouter",
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
		IsStreaming:      true,
		StatusCode:       200,
		DurationMs:       1500,
	}

	err := storage.LogRequest(log)
	if err != nil {
		t.Fatalf("LogRequest failed: %v", err)
	}

	// Retrieve logs
	logs, err := storage.GetRequestLogs(LogFilter{Limit: 10})
	if err != nil {
		t.Fatalf("GetRequestLogs failed: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}

	if logs[0].Model != "gpt-4" {
		t.Errorf("expected model %q, got %q", "gpt-4", logs[0].Model)
	}
	if logs[0].TotalTokens != 150 {
		t.Errorf("expected total tokens %d, got %d", 150, logs[0].TotalTokens)
	}

	// Filter by model
	logs, err = storage.GetRequestLogs(LogFilter{Model: "gpt-3.5"})
	if err != nil {
		t.Fatalf("GetRequestLogs with filter failed: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs for gpt-3.5, got %d", len(logs))
	}
}

func TestDailyUsage(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	today := time.Now().Format("2006-01-02")

	// Create usage entry
	usage := &DailyUsage{
		Date:             today,
		Model:            "gpt-4",
		RequestCount:     10,
		PromptTokens:     1000,
		CompletionTokens: 500,
		TotalTokens:      1500,
		ErrorCount:       1,
	}

	err := storage.UpdateDailyUsage(usage)
	if err != nil {
		t.Fatalf("UpdateDailyUsage failed: %v", err)
	}

	// Update again (should add)
	usage2 := &DailyUsage{
		Date:             today,
		Model:            "gpt-4",
		RequestCount:     5,
		PromptTokens:     500,
		CompletionTokens: 250,
		TotalTokens:      750,
		ErrorCount:       0,
	}
	err = storage.UpdateDailyUsage(usage2)
	if err != nil {
		t.Fatalf("UpdateDailyUsage second time failed: %v", err)
	}

	// Get daily usage
	dailyUsage, err := storage.GetDailyUsage(today, today)
	if err != nil {
		t.Fatalf("GetDailyUsage failed: %v", err)
	}

	if len(dailyUsage) != 1 {
		t.Fatalf("expected 1 daily usage entry, got %d", len(dailyUsage))
	}

	if dailyUsage[0].RequestCount != 15 {
		t.Errorf("expected request count %d, got %d", 15, dailyUsage[0].RequestCount)
	}
	if dailyUsage[0].TotalTokens != 2250 {
		t.Errorf("expected total tokens %d, got %d", 2250, dailyUsage[0].TotalTokens)
	}
}

func TestUsageStats(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	today := time.Now().Format("2006-01-02")

	// Create usage entries for different models
	if err := storage.UpdateDailyUsage(&DailyUsage{
		Date:         today,
		Model:        "gpt-4",
		RequestCount: 10,
		TotalTokens:  1500,
	}); err != nil {
		t.Fatalf("UpdateDailyUsage failed: %v", err)
	}
	if err := storage.UpdateDailyUsage(&DailyUsage{
		Date:         today,
		Model:        "claude-3",
		RequestCount: 5,
		TotalTokens:  1000,
	}); err != nil {
		t.Fatalf("UpdateDailyUsage failed: %v", err)
	}

	stats, err := storage.GetUsageStats(StatsFilter{})
	if err != nil {
		t.Fatalf("GetUsageStats failed: %v", err)
	}

	if stats.TotalRequests != 15 {
		t.Errorf("expected total requests %d, got %d", 15, stats.TotalRequests)
	}
	if stats.TotalTokens != 2500 {
		t.Errorf("expected total tokens %d, got %d", 2500, stats.TotalTokens)
	}
	if len(stats.ModelBreakdown) != 2 {
		t.Errorf("expected 2 models in breakdown, got %d", len(stats.ModelBreakdown))
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sk-or-v1-abcdefghijklmnop", "sk-or-...mnop"},
		{"short", "***"},
		{"1234567890", "***"},
		{"12345678901", "123456...8901"},
	}

	for _, tc := range tests {
		result := MaskAPIKey(tc.input)
		if result != tc.expected {
			t.Errorf("MaskAPIKey(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestCredentialToPreview(t *testing.T) {
	cred := &Credential{
		ID:        "cred_123",
		Provider:  "openrouter",
		Name:      "Test Key",
		APIKey:    "sk-or-v1-abcdefghijklmnop",
		IsDefault: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	preview := cred.ToPreview()

	if preview.ID != cred.ID {
		t.Errorf("expected ID %q, got %q", cred.ID, preview.ID)
	}
	if preview.APIKeyPreview == cred.APIKey {
		t.Error("preview should not contain full API key")
	}
	if preview.APIKeyPreview != "sk-or-...mnop" {
		t.Errorf("expected masked key %q, got %q", "sk-or-...mnop", preview.APIKeyPreview)
	}
}

func TestStorageClosedError(t *testing.T) {
	storage, cleanup := setupTestDB(t)

	// Close the storage
	storage.Close()
	defer cleanup()

	// All operations should return ErrStorageClosed
	_, err := storage.GetCredential("test")
	if err != ErrStorageClosed {
		t.Errorf("expected ErrStorageClosed, got %v", err)
	}

	err = storage.CreateCredential(&Credential{
		Provider: "test",
		Name:     "test",
		APIKey:   "test",
	})
	if err != ErrStorageClosed {
		t.Errorf("expected ErrStorageClosed, got %v", err)
	}
}
