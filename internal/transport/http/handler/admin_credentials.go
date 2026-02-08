package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/mandalnilabja/goatway/internal/storage"
)

// CreateCredentialRequest is the request body for creating a credential
type CreateCredentialRequest struct {
	Provider  string `json:"provider"`
	Name      string `json:"name"`
	APIKey    string `json:"api_key"`
	IsDefault bool   `json:"is_default"`
}

// UpdateCredentialRequest is the request body for updating a credential
type UpdateCredentialRequest struct {
	Provider  *string `json:"provider,omitempty"`
	Name      *string `json:"name,omitempty"`
	APIKey    *string `json:"api_key,omitempty"`
	IsDefault *bool   `json:"is_default,omitempty"`
}

// CreateCredential handles POST /api/admin/credentials
func (h *Repo) CreateCredential(w http.ResponseWriter, r *http.Request) {
	var req CreateCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Provider == "" || req.Name == "" || req.APIKey == "" {
		writeJSONError(w, "provider, name, and api_key are required", http.StatusBadRequest)
		return
	}

	cred := &storage.Credential{
		Provider:  req.Provider,
		Name:      req.Name,
		APIKey:    req.APIKey,
		IsDefault: req.IsDefault,
	}

	if err := h.Storage.CreateCredential(cred); err != nil {
		writeJSONError(w, "Failed to create credential: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, cred.ToPreview(), http.StatusCreated)
}

// ListCredentials handles GET /api/admin/credentials
func (h *Repo) ListCredentials(w http.ResponseWriter, r *http.Request) {
	creds, err := h.Storage.ListCredentials()
	if err != nil {
		writeJSONError(w, "Failed to list credentials: "+err.Error(), http.StatusInternalServerError)
		return
	}

	previews := make([]*storage.CredentialPreview, len(creds))
	for i, cred := range creds {
		previews[i] = cred.ToPreview()
	}

	writeJSON(w, map[string]any{"credentials": previews}, http.StatusOK)
}

// GetCredential handles GET /api/admin/credentials/{id}
func (h *Repo) GetCredential(w http.ResponseWriter, r *http.Request) {
	id := extractCredentialID(r.URL.Path)
	if id == "" {
		writeJSONError(w, "Credential ID is required", http.StatusBadRequest)
		return
	}

	cred, err := h.Storage.GetCredential(id)
	if err == storage.ErrNotFound {
		writeJSONError(w, "Credential not found", http.StatusNotFound)
		return
	}
	if err != nil {
		writeJSONError(w, "Failed to get credential: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, cred.ToPreview(), http.StatusOK)
}

// UpdateCredential handles PUT /api/admin/credentials/{id}
func (h *Repo) UpdateCredential(w http.ResponseWriter, r *http.Request) {
	id := extractCredentialID(r.URL.Path)
	if id == "" {
		writeJSONError(w, "Credential ID is required", http.StatusBadRequest)
		return
	}

	cred, err := h.Storage.GetCredential(id)
	if err == storage.ErrNotFound {
		writeJSONError(w, "Credential not found", http.StatusNotFound)
		return
	}
	if err != nil {
		writeJSONError(w, "Failed to get credential: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var req UpdateCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Provider != nil {
		cred.Provider = *req.Provider
	}
	if req.Name != nil {
		cred.Name = *req.Name
	}
	if req.APIKey != nil {
		cred.APIKey = *req.APIKey
	}
	if req.IsDefault != nil {
		cred.IsDefault = *req.IsDefault
	}
	cred.UpdatedAt = time.Now()

	if err := h.Storage.UpdateCredential(cred); err != nil {
		writeJSONError(w, "Failed to update credential: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, cred.ToPreview(), http.StatusOK)
}

// DeleteCredential handles DELETE /api/admin/credentials/{id}
func (h *Repo) DeleteCredential(w http.ResponseWriter, r *http.Request) {
	id := extractCredentialID(r.URL.Path)
	if id == "" {
		writeJSONError(w, "Credential ID is required", http.StatusBadRequest)
		return
	}

	if err := h.Storage.DeleteCredential(id); err == storage.ErrNotFound {
		writeJSONError(w, "Credential not found", http.StatusNotFound)
		return
	} else if err != nil {
		writeJSONError(w, "Failed to delete credential: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetDefaultCredential handles POST /api/admin/credentials/{id}/default
func (h *Repo) SetDefaultCredential(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path like /api/admin/credentials/{id}/default
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/credentials/")
	path = strings.TrimSuffix(path, "/default")
	id := path

	if id == "" {
		writeJSONError(w, "Credential ID is required", http.StatusBadRequest)
		return
	}

	if err := h.Storage.SetDefaultCredential(id); err == storage.ErrNotFound {
		writeJSONError(w, "Credential not found", http.StatusNotFound)
		return
	} else if err != nil {
		writeJSONError(w, "Failed to set default credential: "+err.Error(), http.StatusInternalServerError)
		return
	}

	cred, _ := h.Storage.GetCredential(id)
	writeJSON(w, cred.ToPreview(), http.StatusOK)
}

// extractCredentialID extracts the credential ID from URL path
func extractCredentialID(path string) string {
	// Path format: /api/admin/credentials/{id}
	path = strings.TrimPrefix(path, "/api/admin/credentials/")
	if idx := strings.Index(path, "/"); idx != -1 {
		return path[:idx]
	}
	return path
}
