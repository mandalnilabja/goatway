// Package types provides OpenAI-compatible type definitions for chat completions.
package types

import "encoding/json"

// Role constants for message roles
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// Message represents a chat message with polymorphic content support.
// Content can be a string or an array of ContentPart for multimodal input.
type Message struct {
	Role       string     `json:"role"`
	Content    Content    `json:"content,omitempty"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`   // For assistant messages
	ToolCallID string     `json:"tool_call_id,omitempty"` // For tool messages
}

// Content represents message content that can be a string or array of parts.
type Content struct {
	Text  string        // Simple string content
	Parts []ContentPart // Multimodal content parts
}

// MarshalJSON implements custom JSON marshaling for Content.
// Outputs string if Text is set, array if Parts is set.
func (c Content) MarshalJSON() ([]byte, error) {
	if len(c.Parts) > 0 {
		return json.Marshal(c.Parts)
	}
	return json.Marshal(c.Text)
}

// UnmarshalJSON implements custom JSON unmarshaling for Content.
// Accepts both string and array formats.
func (c *Content) UnmarshalJSON(data []byte) error {
	// Try string first
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		c.Text = text
		c.Parts = nil
		return nil
	}

	// Try array of content parts
	var parts []ContentPart
	if err := json.Unmarshal(data, &parts); err == nil {
		c.Parts = parts
		c.Text = ""
		return nil
	}

	return nil // Allow null/empty content
}

// String returns the text content, concatenating parts if multimodal.
func (c Content) String() string {
	if c.Text != "" {
		return c.Text
	}
	var result string
	for _, part := range c.Parts {
		if part.Type == ContentTypeText {
			result += part.Text
		}
	}
	return result
}

// ContentPart represents a single part of multimodal content.
type ContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// Content type constants
const (
	ContentTypeText     = "text"
	ContentTypeImageURL = "image_url"
)

// ImageURL represents an image reference in multimodal content.
type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"` // "auto", "low", "high"
}

// NewTextMessage creates a simple text message.
func NewTextMessage(role, content string) Message {
	return Message{
		Role:    role,
		Content: Content{Text: content},
	}
}

// NewImageMessage creates a message with text and image content.
func NewImageMessage(role, text, imageURL string) Message {
	return Message{
		Role: role,
		Content: Content{
			Parts: []ContentPart{
				{Type: ContentTypeText, Text: text},
				{Type: ContentTypeImageURL, ImageURL: &ImageURL{URL: imageURL}},
			},
		},
	}
}
