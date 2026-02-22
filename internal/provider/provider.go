package provider

import "github.com/mandalnilabja/goatway/internal/types"

// Re-export types from internal/types for backward compatibility
type (
	Provider     = types.Provider
	ProxyOptions = types.ProxyOptions
	ProxyResult  = types.ProxyResult
)

// ErrNoAPIKey is re-exported for backward compatibility
var ErrNoAPIKey = types.ErrNoAPIKey
