package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mandalnilabja/goatway/internal/config"
	"github.com/mandalnilabja/goatway/internal/storage/models"
	"github.com/mandalnilabja/goatway/internal/types"
)

// mockProvider implements types.Provider for testing.
type mockProvider struct {
	name      string
	lastModel string
}

func (m *mockProvider) Name() string                                               { return m.name }
func (m *mockProvider) BaseURL() string                                            { return "https://mock.test" }
func (m *mockProvider) PrepareRequest(ctx context.Context, req *http.Request) error { return nil }
func (m *mockProvider) ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request, opts *types.ProxyOptions) (*types.ProxyResult, error) {
	m.lastModel = opts.Model
	w.WriteHeader(http.StatusOK)
	return &types.ProxyResult{Model: opts.Model, StatusCode: http.StatusOK}, nil
}

// mockStorage implements storage.Storage for tests.
type mockStorage struct{}

func (m *mockStorage) GetDefaultCredential(provider string) (*models.Credential, error) {
	return &models.Credential{ID: "test-cred", Provider: provider}, nil
}

// Stub implementations for storage.Storage interface
func (m *mockStorage) CreateCredential(cred *models.Credential) error            { return nil }
func (m *mockStorage) GetCredential(id string) (*models.Credential, error)       { return nil, nil }
func (m *mockStorage) ListCredentials() ([]*models.Credential, error)            { return nil, nil }
func (m *mockStorage) UpdateCredential(cred *models.Credential) error            { return nil }
func (m *mockStorage) DeleteCredential(id string) error                          { return nil }
func (m *mockStorage) SetDefaultCredential(id string) error                      { return nil }
func (m *mockStorage) LogRequest(log *models.RequestLog) error                   { return nil }
func (m *mockStorage) GetRequestLogs(f models.LogFilter) ([]*models.RequestLog, error) {
	return nil, nil
}
func (m *mockStorage) DeleteRequestLogs(olderThan string) (int64, error)         { return 0, nil }
func (m *mockStorage) GetUsageStats(f models.StatsFilter) (*models.UsageStats, error) {
	return nil, nil
}
func (m *mockStorage) GetDailyUsage(start, end string) ([]*models.DailyUsage, error) { return nil, nil }
func (m *mockStorage) UpdateDailyUsage(usage *models.DailyUsage) error           { return nil }
func (m *mockStorage) CreateAPIKey(key *models.ClientAPIKey) error               { return nil }
func (m *mockStorage) GetAPIKey(id string) (*models.ClientAPIKey, error)         { return nil, nil }
func (m *mockStorage) GetAPIKeyByPrefix(prefix string) ([]*models.ClientAPIKey, error) {
	return nil, nil
}
func (m *mockStorage) ListAPIKeys() ([]*models.ClientAPIKey, error)              { return nil, nil }
func (m *mockStorage) UpdateAPIKey(key *models.ClientAPIKey) error               { return nil }
func (m *mockStorage) DeleteAPIKey(id string) error                              { return nil }
func (m *mockStorage) UpdateAPIKeyLastUsed(id string) error                      { return nil }
func (m *mockStorage) GetAdminPasswordHash() (string, error)                     { return "", nil }
func (m *mockStorage) SetAdminPasswordHash(hash string) error                    { return nil }
func (m *mockStorage) HasAdminPassword() (bool, error)                           { return false, nil }
func (m *mockStorage) Close() error                                              { return nil }

func TestRouter_ResolveKnownAlias(t *testing.T) {
	mock := &mockProvider{name: "openrouter"}
	providers := map[string]types.Provider{"openrouter": mock}

	cfg := &config.Config{
		Models: []config.ModelAlias{
			{Slug: "gpt4", Provider: "openrouter", Model: "openai/gpt-4o"},
		},
	}

	router := NewRouter(providers, cfg, &mockStorage{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	opts := &types.ProxyOptions{Model: "gpt4"}

	result, err := router.ProxyRequest(context.Background(), w, req, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Model != "openai/gpt-4o" {
		t.Errorf("expected model 'openai/gpt-4o', got '%s'", result.Model)
	}
	if mock.lastModel != "openai/gpt-4o" {
		t.Errorf("expected provider to receive 'openai/gpt-4o', got '%s'", mock.lastModel)
	}
}

func TestRouter_ResolveWithDefault(t *testing.T) {
	mock := &mockProvider{name: "openrouter"}
	providers := map[string]types.Provider{"openrouter": mock}

	cfg := &config.Config{
		Default: &config.DefaultRoute{Provider: "openrouter", Model: "openai/gpt-4o"},
		Models:  []config.ModelAlias{},
	}

	router := NewRouter(providers, cfg, &mockStorage{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	opts := &types.ProxyOptions{Model: "unknown-model"}

	result, err := router.ProxyRequest(context.Background(), w, req, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With default, the original slug is passed through as the model name
	if result.Model != "unknown-model" {
		t.Errorf("expected model 'unknown-model', got '%s'", result.Model)
	}
}

func TestRouter_ResolveWithoutDefault(t *testing.T) {
	mock := &mockProvider{name: "openrouter"}
	providers := map[string]types.Provider{"openrouter": mock}

	cfg := &config.Config{
		Default: nil, // No default
		Models:  []config.ModelAlias{},
	}

	router := NewRouter(providers, cfg, &mockStorage{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	opts := &types.ProxyOptions{Model: "unknown-model"}

	_, err := router.ProxyRequest(context.Background(), w, req, opts)
	if err != ErrModelNotFound {
		t.Errorf("expected ErrModelNotFound, got %v", err)
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
