package tokenizer

import (
	"strings"

	"github.com/mandalnilabja/goatway/internal/types"
)

// countContent counts tokens for message content (text or multimodal).
func (t *TiktokenTokenizer) countContent(content types.Content, model string) (int, error) {
	// Simple text content
	if content.Text != "" {
		return t.CountTokens(content.Text, model)
	}

	// Multimodal content
	total := 0
	for _, part := range content.Parts {
		switch part.Type {
		case types.ContentTypeText:
			tokens, err := t.CountTokens(part.Text, model)
			if err != nil {
				return 0, err
			}
			total += tokens

		case types.ContentTypeImageURL:
			total += t.countImageTokens(part.ImageURL)
		}
	}

	return total, nil
}

// countImageTokens calculates token cost for an image based on OpenAI's rules.
func (t *TiktokenTokenizer) countImageTokens(img *types.ImageURL) int {
	if img == nil {
		return 0
	}

	detail := strings.ToLower(img.Detail)
	switch detail {
	case "low":
		// Low detail: fixed cost
		return imageBaseTokens + (imageLowDetailTiles * imageTileTokens)
	case "high":
		// High detail: base + tiles (simplified without actual image dimensions)
		// In practice, you'd need image dimensions to calculate exactly.
		// Using a reasonable estimate of 4 tiles for high detail.
		return imageBaseTokens + (imageHighDetailMax * imageTileTokens)
	default:
		// "auto" or unspecified: assume high detail
		return imageBaseTokens + (imageHighDetailMax * imageTileTokens)
	}
}
