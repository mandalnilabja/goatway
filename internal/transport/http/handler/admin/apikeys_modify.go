package admin

import (
	"encoding/json"
	"net/http"

	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/types"
)

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

	// Invalidate cache for immediate effect
	h.InvalidateAPIKeyCache(key.KeyPrefix)

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

	// Get key first to retrieve prefix for cache invalidation
	key, err := h.Storage.GetAPIKey(id)
	if err != nil {
		if err == storage.ErrNotFound {
			types.WriteError(w, http.StatusNotFound, types.ErrNotFound("key not found"))
			return
		}
		types.WriteError(w, http.StatusInternalServerError, types.ErrServer("failed to get key"))
		return
	}

	if err := h.Storage.DeleteAPIKey(id); err != nil {
		types.WriteError(w, http.StatusInternalServerError, types.ErrServer("failed to delete key"))
		return
	}

	// Invalidate cache for immediate effect
	h.InvalidateAPIKeyCache(key.KeyPrefix)

	w.WriteHeader(http.StatusNoContent)
}
