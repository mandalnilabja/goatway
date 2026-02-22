package proxy

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/mandalnilabja/goatway/internal/types"
)

const openRouterModelsURL = "https://openrouter.ai/api/v1/models"

// getDefaultAPIKey gets the API key from the default openrouter credential.
func (h *Handlers) getDefaultAPIKey() string {
	if h.Storage == nil {
		return ""
	}
	cred, err := h.Storage.GetDefaultCredential("openrouter")
	if err != nil || cred == nil {
		return ""
	}
	return cred.GetAPIKey()
}

// ListModels proxies GET /v1/models to OpenRouter.
// Returns the list of available models in OpenAI-compatible format.
func (h *Handlers) ListModels(w http.ResponseWriter, r *http.Request) {
	apiKey := h.getDefaultAPIKey()
	if apiKey == "" {
		types.WriteError(w, http.StatusUnauthorized, types.ErrAuthentication("No credential configured for openrouter"))
		return
	}

	resp, err := h.fetchModels(r, apiKey, openRouterModelsURL)
	if err != nil {
		types.WriteError(w, http.StatusBadGateway, types.ErrServer("upstream error: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	h.forwardModelsResponse(w, resp)
}

// GetModel proxies GET /v1/models/{model} to OpenRouter.
// Returns details for a specific model.
func (h *Handlers) GetModel(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("model")
	if modelID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("model ID required"))
		return
	}

	apiKey := h.getDefaultAPIKey()
	if apiKey == "" {
		types.WriteError(w, http.StatusUnauthorized, types.ErrAuthentication("No credential configured for openrouter"))
		return
	}

	// OpenRouter doesn't have a single model endpoint, so fetch all and filter
	resp, err := h.fetchModels(r, apiKey, openRouterModelsURL)
	if err != nil {
		types.WriteError(w, http.StatusBadGateway, types.ErrServer("upstream error: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.forwardModelsResponse(w, resp)
		return
	}

	// Parse response to find the specific model
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		types.WriteError(w, http.StatusBadGateway, types.ErrServer("failed to read upstream response"))
		return
	}

	var modelsList modelsListResponse
	if err := json.Unmarshal(body, &modelsList); err != nil {
		types.WriteError(w, http.StatusBadGateway, types.ErrServer("failed to parse models response"))
		return
	}

	// Find the requested model
	for _, m := range modelsList.Data {
		if m.ID == modelID {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(m)
			return
		}
	}

	types.WriteError(w, http.StatusNotFound, types.ErrInvalidRequest("model '"+modelID+"' not found"))
}

// fetchModels makes a request to the upstream models endpoint.
func (h *Handlers) fetchModels(r *http.Request, apiKey, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	return client.Do(req)
}

// forwardModelsResponse forwards the upstream response to the client.
func (h *Handlers) forwardModelsResponse(w http.ResponseWriter, resp *http.Response) {
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

// modelsListResponse represents the OpenAI models list response.
type modelsListResponse struct {
	Object string  `json:"object"`
	Data   []model `json:"data"`
}

// model represents a single model in the list.
type model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}
