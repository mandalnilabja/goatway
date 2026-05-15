package azurefoundry

import (
	"bytes"
	"encoding/json"
	"io"
)

// rewriteModelInBody reads the request body and replaces the model field with the resolved model.
func rewriteModelInBody(optsBody io.Reader, reqBody io.Reader, resolvedModel string) (io.Reader, error) {
	var body io.Reader = reqBody
	if optsBody != nil {
		body = optsBody
	}

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	var payload map[string]any
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		return nil, err
	}

	payload["model"] = resolvedModel

	rewritten, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(rewritten), nil
}
