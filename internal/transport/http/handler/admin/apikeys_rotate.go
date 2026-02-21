package admin

import (
	"encoding/json"
	"net/http"

	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/types"
)

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
