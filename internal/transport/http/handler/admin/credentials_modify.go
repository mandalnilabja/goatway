package admin

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/transport/http/handler/shared"
)

// CreateCredential handles POST /api/admin/credentials.
func (h *Handlers) CreateCredential(w http.ResponseWriter, r *http.Request) {
	var req CreateCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Provider == "" || req.Name == "" || len(req.Data) == 0 {
		shared.WriteJSONError(w, "provider, name, and data are required", http.StatusBadRequest)
		return
	}

	cred := &storage.Credential{
		Provider:  req.Provider,
		Name:      req.Name,
		Data:      req.Data,
		IsDefault: req.IsDefault,
	}

	if err := h.Storage.CreateCredential(cred); err != nil {
		shared.WriteJSONError(w, "Failed to create credential: "+err.Error(), http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, cred.ToPreview(), http.StatusCreated)
}

// UpdateCredential handles PUT /api/admin/credentials/{id}.
func (h *Handlers) UpdateCredential(w http.ResponseWriter, r *http.Request) {
	id := extractCredentialID(r.URL.Path)
	if id == "" {
		shared.WriteJSONError(w, "Credential ID is required", http.StatusBadRequest)
		return
	}

	cred, err := h.Storage.GetCredential(id)
	if err == storage.ErrNotFound {
		shared.WriteJSONError(w, "Credential not found", http.StatusNotFound)
		return
	}
	if err != nil {
		shared.WriteJSONError(w, "Failed to get credential: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var req UpdateCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Provider != nil {
		cred.Provider = *req.Provider
	}
	if req.Name != nil {
		cred.Name = *req.Name
	}
	if req.Data != nil {
		cred.Data = *req.Data
	}
	if req.IsDefault != nil {
		cred.IsDefault = *req.IsDefault
	}
	cred.UpdatedAt = time.Now()

	if err := h.Storage.UpdateCredential(cred); err != nil {
		shared.WriteJSONError(w, "Failed to update credential: "+err.Error(), http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, cred.ToPreview(), http.StatusOK)
}

// DeleteCredential handles DELETE /api/admin/credentials/{id}.
func (h *Handlers) DeleteCredential(w http.ResponseWriter, r *http.Request) {
	id := extractCredentialID(r.URL.Path)
	if id == "" {
		shared.WriteJSONError(w, "Credential ID is required", http.StatusBadRequest)
		return
	}

	if err := h.Storage.DeleteCredential(id); err == storage.ErrNotFound {
		shared.WriteJSONError(w, "Credential not found", http.StatusNotFound)
		return
	} else if err != nil {
		shared.WriteJSONError(w, "Failed to delete credential: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetDefaultCredential handles POST /api/admin/credentials/{id}/default.
func (h *Handlers) SetDefaultCredential(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path like /api/admin/credentials/{id}/default
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/credentials/")
	path = strings.TrimSuffix(path, "/default")
	id := path

	if id == "" {
		shared.WriteJSONError(w, "Credential ID is required", http.StatusBadRequest)
		return
	}

	if err := h.Storage.SetDefaultCredential(id); err == storage.ErrNotFound {
		shared.WriteJSONError(w, "Credential not found", http.StatusNotFound)
		return
	} else if err != nil {
		shared.WriteJSONError(w, "Failed to set default credential: "+err.Error(), http.StatusInternalServerError)
		return
	}

	cred, _ := h.Storage.GetCredential(id)
	shared.WriteJSON(w, cred.ToPreview(), http.StatusOK)
}
