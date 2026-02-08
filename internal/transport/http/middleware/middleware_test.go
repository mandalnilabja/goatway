package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID(t *testing.T) {
	tests := []struct {
		name           string
		existingID     string
		wantNewID      bool
		wantPassedThru bool
	}{
		{
			name:           "generates new ID when none provided",
			existingID:     "",
			wantNewID:      true,
			wantPassedThru: false,
		},
		{
			name:           "uses existing ID from header",
			existingID:     "existing-request-id",
			wantNewID:      false,
			wantPassedThru: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedID string
			handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedID = GetRequestID(r.Context())
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.existingID != "" {
				req.Header.Set(RequestIDHeader, tt.existingID)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			// Check response header
			respID := rec.Header().Get(RequestIDHeader)
			if respID == "" {
				t.Error("expected X-Request-ID in response header")
			}

			// Check context value
			if capturedID == "" {
				t.Error("expected request ID in context")
			}

			// Check if ID was passed through or generated
			if tt.wantPassedThru && respID != tt.existingID {
				t.Errorf("expected ID %q, got %q", tt.existingID, respID)
			}

			if tt.wantNewID && respID == "" {
				t.Error("expected generated ID")
			}
		})
	}
}

func TestCORS(t *testing.T) {
	handler := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("sets CORS headers on regular request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Error("expected Access-Control-Allow-Origin header")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("handles OPTIONS preflight request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
	})
}

func TestAdminAuth(t *testing.T) {
	tests := []struct {
		name           string
		password       string
		authHeader     string
		wantStatus     int
		wantNextCalled bool
	}{
		{
			name:           "no password configured allows all",
			password:       "",
			authHeader:     "",
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
		{
			name:           "correct password passes",
			password:       "secret",
			authHeader:     "Bearer secret",
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
		{
			name:           "wrong password rejects",
			password:       "secret",
			authHeader:     "Bearer wrong",
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "missing auth header rejects",
			password:       "secret",
			authHeader:     "",
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "malformed auth header rejects",
			password:       "secret",
			authHeader:     "Basic secret",
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			handler := AdminAuth(tt.password)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
			if nextCalled != tt.wantNextCalled {
				t.Errorf("expected nextCalled=%v, got %v", tt.wantNextCalled, nextCalled)
			}
		})
	}
}

func TestRequestLogger(t *testing.T) {
	// Use a discard handler to avoid test output noise
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestResponseWriterFlush(t *testing.T) {
	handler := RequestLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Test that Flush is available
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestGetRequestID_NoContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	id := GetRequestID(req.Context())
	if id != "" {
		t.Errorf("expected empty ID, got %q", id)
	}
}
