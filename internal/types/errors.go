package types

import (
	"encoding/json"
	"net/http"
)

// APIError represents an OpenAI-compatible error response.
type APIError struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains the error information.
type ErrorDetail struct {
	Message string  `json:"message"`
	Type    string  `json:"type"`
	Param   *string `json:"param,omitempty"`
	Code    *string `json:"code,omitempty"`
}

// Error type constants
const (
	ErrorTypeInvalidRequest     = "invalid_request_error"
	ErrorTypeAuthentication     = "authentication_error"
	ErrorTypePermission         = "permission_error"
	ErrorTypeNotFound           = "not_found_error"
	ErrorTypeRateLimit          = "rate_limit_error"
	ErrorTypeServer             = "server_error"
	ErrorTypeServiceUnavailable = "service_unavailable"
)

// NewAPIError creates a new API error.
func NewAPIError(message, errType string) *APIError {
	return &APIError{
		Error: ErrorDetail{
			Message: message,
			Type:    errType,
		},
	}
}

// NewAPIErrorWithCode creates a new API error with a code.
func NewAPIErrorWithCode(message, errType, code string) *APIError {
	return &APIError{
		Error: ErrorDetail{
			Message: message,
			Type:    errType,
			Code:    &code,
		},
	}
}

// NewAPIErrorWithParam creates a new API error with a parameter reference.
func NewAPIErrorWithParam(message, errType, param string) *APIError {
	return &APIError{
		Error: ErrorDetail{
			Message: message,
			Type:    errType,
			Param:   &param,
		},
	}
}

// WriteError writes an API error to the response writer.
func WriteError(w http.ResponseWriter, statusCode int, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(err)
}

// Common error constructors

// ErrInvalidRequest creates an invalid request error.
func ErrInvalidRequest(message string) *APIError {
	return NewAPIError(message, ErrorTypeInvalidRequest)
}

// ErrAuthentication creates an authentication error.
func ErrAuthentication(message string) *APIError {
	return NewAPIError(message, ErrorTypeAuthentication)
}

// ErrRateLimit creates a rate limit error.
func ErrRateLimit(message string) *APIError {
	return NewAPIError(message, ErrorTypeRateLimit)
}

// ErrServer creates a server error.
func ErrServer(message string) *APIError {
	return NewAPIError(message, ErrorTypeServer)
}

// ErrNotFound creates a not found error.
func ErrNotFound(message string) *APIError {
	return NewAPIError(message, ErrorTypeNotFound)
}
