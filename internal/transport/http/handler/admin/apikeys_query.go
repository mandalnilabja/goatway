package admin

import (
	"encoding/json"
	"net/http"

	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/types"
)

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
