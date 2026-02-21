package admin

import (
	"net/http"
	"strings"

	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/transport/http/handler/shared"
)

// ListCredentials handles GET /api/admin/credentials.
func (h *Handlers) ListCredentials(w http.ResponseWriter, r *http.Request) {
	creds, err := h.Storage.ListCredentials()
	if err != nil {
		shared.WriteJSONError(w, "Failed to list credentials: "+err.Error(), http.StatusInternalServerError)
		return
	}

	previews := make([]*storage.CredentialPreview, len(creds))
	for i, cred := range creds {
		previews[i] = cred.ToPreview()
	}

	shared.WriteJSON(w, map[string]any{"credentials": previews}, http.StatusOK)
}

// GetCredential handles GET /api/admin/credentials/{id}.
func (h *Handlers) GetCredential(w http.ResponseWriter, r *http.Request) {
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

	shared.WriteJSON(w, cred.ToPreview(), http.StatusOK)
}

// extractCredentialID extracts the credential ID from URL path.
func extractCredentialID(path string) string {
	// Path format: /api/admin/credentials/{id}
	path = strings.TrimPrefix(path, "/api/admin/credentials/")
	if idx := strings.Index(path, "/"); idx != -1 {
		return path[:idx]
	}
	return path
}
