package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/types"
)

// CreateAPIKey creates a new client API key (POST /api/admin/apikeys).
func (h *Handlers) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("invalid request body"))
		return
	}

	if req.Name == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("name is required"))
		return
	}

	// Default scopes to ["proxy"] if not specified
	if len(req.Scopes) == 0 {
		req.Scopes = []string{"proxy"}
	}

	// Validate scopes
	validScopes := map[string]bool{"proxy": true, "admin": true}
	for _, scope := range req.Scopes {
		if !validScopes[scope] {
			types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("invalid scope: "+scope))
			return
		}
	}

	// Generate API key
	plainKey, err := storage.GenerateAPIKey()
	if err != nil {
		types.WriteError(w, http.StatusInternalServerError, types.ErrServer("failed to generate key"))
		return
	}

	// Hash the key
	hash, err := storage.HashPassword(plainKey, storage.DefaultArgon2Params())
	if err != nil {
		types.WriteError(w, http.StatusInternalServerError, types.ErrServer("failed to hash key"))
		return
	}

	// Calculate expiry
	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(*req.ExpiresIn) * time.Second)
		expiresAt = &t
	}

	// Create key record
	apiKey := &storage.ClientAPIKey{
		ID:        uuid.New().String(),
		Name:      req.Name,
		KeyHash:   hash,
		KeyPrefix: storage.ExtractKeyPrefix(plainKey),
		Scopes:    req.Scopes,
		RateLimit: req.RateLimit,
		IsActive:  true,
		ExpiresAt: expiresAt,
	}

	if err := h.Storage.CreateAPIKey(apiKey); err != nil {
		types.WriteError(w, http.StatusInternalServerError, types.ErrServer("failed to create key"))
		return
	}

	// Return response with plaintext key (shown only once)
	resp := CreateAPIKeyResponse{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		Key:       plainKey,
		KeyPrefix: apiKey.KeyPrefix,
		Scopes:    apiKey.Scopes,
		RateLimit: apiKey.RateLimit,
		IsActive:  apiKey.IsActive,
		CreatedAt: apiKey.CreatedAt,
		ExpiresAt: apiKey.ExpiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}
