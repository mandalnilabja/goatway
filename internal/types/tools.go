package types

import "encoding/json"

// Tool represents a tool available to the model.
type Tool struct {
	Type     string   `json:"type"` // Currently only "function"
	Function Function `json:"function"`
}

// ToolType constants
const ToolTypeFunction = "function"

// Function defines a function that can be called by the model.
type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"` // JSON Schema object
	Strict      *bool       `json:"strict,omitempty"`     // Enable strict schema
}

// ToolCall represents a tool call made by the model.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // "function"
	Function FunctionCall `json:"function"`
	Index    *int         `json:"index,omitempty"` // For streaming
}

// FunctionCall contains the function name and arguments.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// ToolChoice specifies how the model should use tools.
// Can be "none", "auto", "required", or a specific tool.
type ToolChoice struct {
	Type     string          `json:"type,omitempty"`     // "function" for specific tool
	Function *ToolChoiceFunc `json:"function,omitempty"` // Specific function to call
	Auto     bool            `json:"-"`                  // True if "auto"
	None     bool            `json:"-"`                  // True if "none"
	Required bool            `json:"-"`                  // True if "required"
}

// ToolChoiceFunc specifies a specific function to call.
type ToolChoiceFunc struct {
	Name string `json:"name"`
}

// MarshalJSON implements custom marshaling for ToolChoice.
func (tc ToolChoice) MarshalJSON() ([]byte, error) {
	if tc.None {
		return []byte(`"none"`), nil
	}
	if tc.Auto {
		return []byte(`"auto"`), nil
	}
	if tc.Required {
		return []byte(`"required"`), nil
	}
	// Specific tool choice
	type alias ToolChoice
	return json.Marshal(alias(tc))
}

// UnmarshalJSON implements custom unmarshaling for ToolChoice.
func (tc *ToolChoice) UnmarshalJSON(data []byte) error {
	// Try string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		switch str {
		case "none":
			tc.None = true
		case "auto":
			tc.Auto = true
		case "required":
			tc.Required = true
		}
		return nil
	}

	// Try object
	type alias ToolChoice
	return json.Unmarshal(data, (*alias)(tc))
}

// NewTool creates a new function tool definition.
func NewTool(name, description string, parameters interface{}) Tool {
	return Tool{
		Type: ToolTypeFunction,
		Function: Function{
			Name:        name,
			Description: description,
			Parameters:  parameters,
		},
	}
}
