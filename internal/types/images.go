package types

// ImageGenerationRequest represents an OpenAI image generation API request.
type ImageGenerationRequest struct {
	// Required: Text description of the image to generate
	Prompt string `json:"prompt"`

	// Optional: Model to use (e.g., "dall-e-2", "dall-e-3")
	Model string `json:"model,omitempty"`

	// Optional: Number of images to generate (1-10, default 1)
	// For dall-e-3, only n=1 is supported
	N *int `json:"n,omitempty"`

	// Optional: Quality of generated images
	// Values: "standard" (default), "hd" (dall-e-3 only)
	Quality string `json:"quality,omitempty"`

	// Optional: Response format
	// Values: "url" (default), "b64_json"
	ResponseFormat string `json:"response_format,omitempty"`

	// Optional: Size of generated images
	// dall-e-2: "256x256", "512x512", "1024x1024" (default)
	// dall-e-3: "1024x1024" (default), "1792x1024", "1024x1792"
	Size string `json:"size,omitempty"`

	// Optional: Style of generated images (dall-e-3 only)
	// Values: "vivid" (default), "natural"
	Style string `json:"style,omitempty"`

	// Optional: Unique identifier for the end-user
	User string `json:"user,omitempty"`
}

// ImageEditRequest represents an OpenAI image edit API request.
// Sent as multipart/form-data with image files.
type ImageEditRequest struct {
	// Required: Image file to edit (PNG, < 4MB, square)

	// Required: Text description of the desired edit
	Prompt string `json:"prompt"`

	// Optional: Mask image (PNG with transparency indicating edit areas)

	// Optional: Model to use (e.g., "dall-e-2")
	Model string `json:"model,omitempty"`

	// Optional: Number of images to generate (1-10, default 1)
	N *int `json:"n,omitempty"`

	// Optional: Size of generated images
	// Values: "256x256", "512x512", "1024x1024" (default)
	Size string `json:"size,omitempty"`

	// Optional: Response format
	// Values: "url" (default), "b64_json"
	ResponseFormat string `json:"response_format,omitempty"`

	// Optional: Unique identifier for the end-user
	User string `json:"user,omitempty"`
}

// ImageVariationRequest represents an OpenAI image variation API request.
// Sent as multipart/form-data with image file.
type ImageVariationRequest struct {
	// Required: Image file to create variations of (PNG, < 4MB, square)

	// Optional: Model to use (e.g., "dall-e-2")
	Model string `json:"model,omitempty"`

	// Optional: Number of images to generate (1-10, default 1)
	N *int `json:"n,omitempty"`

	// Optional: Response format
	// Values: "url" (default), "b64_json"
	ResponseFormat string `json:"response_format,omitempty"`

	// Optional: Size of generated images
	// Values: "256x256", "512x512", "1024x1024" (default)
	Size string `json:"size,omitempty"`

	// Optional: Unique identifier for the end-user
	User string `json:"user,omitempty"`
}

// ImagesResponse represents an OpenAI images API response.
type ImagesResponse struct {
	Created int64       `json:"created"` // Unix timestamp
	Data    []ImageData `json:"data"`
}

// ImageData represents a single image in the response.
type ImageData struct {
	// URL of the generated image (if response_format="url")
	URL string `json:"url,omitempty"`

	// Base64-encoded image data (if response_format="b64_json")
	B64JSON string `json:"b64_json,omitempty"`

	// Revised prompt used for generation (dall-e-3 only)
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}
