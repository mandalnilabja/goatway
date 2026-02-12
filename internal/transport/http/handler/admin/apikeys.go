package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/types"
)

// CreateAPIKeyRequest is the request body for creating an API key.
type CreateAPIKeyRequest struct {
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`     // ["proxy", "admin"]
	RateLimit int      `json:"rate_limit"` // Requests per minute (0 = unlimited)
	ExpiresIn *int     `json:"expires_in"` // Seconds until expiry (optional)
}

// CreateAPIKeyResponse includes the plaintext key (shown only once).
type CreateAPIKeyResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Key       string     `json:"key"` // Plaintext - shown only once!
	KeyPrefix string     `json:"key_prefix"`
	Scopes    []string   `json:"scopes"`
	RateLimit int        `json:"rate_limit"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

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

// ListAPIKeys returns all API keys (GET /api/admin/apikeys).
func (h *Handlers) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := h.Storage.ListAPIKeys()
	if err != nil {
		types.WriteError(w, http.StatusInternalServerError, types.ErrServer("failed to list keys"))
		return
	}

	// Convert to previews (no hashes)
	previews := make([]*storage.ClientAPIKeyPreview, len(keys))
	for i, k := range keys {
		previews[i] = k.ToPreview()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": previews,
	})
}

// GetAPIKeyByID returns a specific API key (GET /api/admin/apikeys/{id}).
func (h *Handlers) GetAPIKeyByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("id required"))
		return
	}

	key, err := h.Storage.GetAPIKey(id)
	if err != nil {
		if err == storage.ErrNotFound {
			types.WriteError(w, http.StatusNotFound, types.ErrNotFound("key not found"))
			return
		}
		types.WriteError(w, http.StatusInternalServerError, types.ErrServer("failed to get key"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(key.ToPreview())
}

// UpdateAPIKeyRequest is the request body for updating an API key.
type UpdateAPIKeyRequest struct {
	Name      *string  `json:"name"`
	Scopes    []string `json:"scopes"`
	RateLimit *int     `json:"rate_limit"`
	IsActive  *bool    `json:"is_active"`
}

// UpdateAPIKey updates an API key (PUT /api/admin/apikeys/{id}).
func (h *Handlers) UpdateAPIKey(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("id required"))
		return
	}

	var updates UpdateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("invalid request body"))
		return
	}

	key, err := h.Storage.GetAPIKey(id)
	if err != nil {
		if err == storage.ErrNotFound {
			types.WriteError(w, http.StatusNotFound, types.ErrNotFound("key not found"))
			return
		}
		types.WriteError(w, http.StatusInternalServerError, types.ErrServer("failed to get key"))
		return
	}

	if updates.Name != nil {
		key.Name = *updates.Name
	}
	if updates.Scopes != nil {
		// Validate scopes
		validScopes := map[string]bool{"proxy": true, "admin": true}
		for _, scope := range updates.Scopes {
			if !validScopes[scope] {
				types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("invalid scope: "+scope))
				return
			}
		}
		key.Scopes = updates.Scopes
	}
	if updates.RateLimit != nil {
		key.RateLimit = *updates.RateLimit
	}
	if updates.IsActive != nil {
		key.IsActive = *updates.IsActive
	}

	if err := h.Storage.UpdateAPIKey(key); err != nil {
		types.WriteError(w, http.StatusInternalServerError, types.ErrServer("failed to update key"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(key.ToPreview())
}

// DeleteAPIKey deletes an API key (DELETE /api/admin/apikeys/{id}).
func (h *Handlers) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("id required"))
		return
	}

	if err := h.Storage.DeleteAPIKey(id); err != nil {
		if err == storage.ErrNotFound {
			types.WriteError(w, http.StatusNotFound, types.ErrNotFound("key not found"))
			return
		}
		types.WriteError(w, http.StatusInternalServerError, types.ErrServer("failed to delete key"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RotateAPIKey generates a new key (POST /api/admin/apikeys/{id}/rotate).
func (h *Handlers) RotateAPIKey(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("id required"))
		return
	}

	key, err := h.Storage.GetAPIKey(id)
	if err != nil {
		if err == storage.ErrNotFound {
			types.WriteError(w, http.StatusNotFound, types.ErrNotFound("key not found"))
			return
		}
		types.WriteError(w, http.StatusInternalServerError, types.ErrServer("failed to get key"))
		return
	}

	// Generate new key
	plainKey, err := storage.GenerateAPIKey()
	if err != nil {
		types.WriteError(w, http.StatusInternalServerError, types.ErrServer("failed to generate key"))
		return
	}

	// Hash the new key
	hash, err := storage.HashPassword(plainKey, storage.DefaultArgon2Params())
	if err != nil {
		types.WriteError(w, http.StatusInternalServerError, types.ErrServer("failed to hash key"))
		return
	}

	// Update key with new hash and prefix
	key.KeyHash = hash
	key.KeyPrefix = storage.ExtractKeyPrefix(plainKey)

	if err := h.Storage.UpdateAPIKey(key); err != nil {
		types.WriteError(w, http.StatusInternalServerError, types.ErrServer("failed to update key"))
		return
	}

	// Return new key
	resp := CreateAPIKeyResponse{
		ID:        key.ID,
		Name:      key.Name,
		Key:       plainKey,
		KeyPrefix: key.KeyPrefix,
		Scopes:    key.Scopes,
		RateLimit: key.RateLimit,
		IsActive:  key.IsActive,
		CreatedAt: key.CreatedAt,
		ExpiresAt: key.ExpiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
