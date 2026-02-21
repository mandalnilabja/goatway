package tokenizer

import (
	"testing"

	"github.com/mandalnilabja/goatway/internal/types"
)

func TestCountImageTokens(t *testing.T) {
	tok := New()

	tests := []struct {
		name     string
		image    *types.ImageURL
		expected int
	}{
		{
			name:     "nil image",
			image:    nil,
			expected: 0,
		},
		{
			name:     "low detail",
			image:    &types.ImageURL{URL: "http://example.com/img.jpg", Detail: "low"},
			expected: imageBaseTokens + imageLowDetailTiles*imageTileTokens,
		},
		{
			name:     "high detail",
			image:    &types.ImageURL{URL: "http://example.com/img.jpg", Detail: "high"},
			expected: imageBaseTokens + imageHighDetailMax*imageTileTokens,
		},
		{
			name:     "auto detail",
			image:    &types.ImageURL{URL: "http://example.com/img.jpg", Detail: "auto"},
			expected: imageBaseTokens + imageHighDetailMax*imageTileTokens,
		},
		{
			name:     "no detail specified",
			image:    &types.ImageURL{URL: "http://example.com/img.jpg"},
			expected: imageBaseTokens + imageHighDetailMax*imageTileTokens,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tok.countImageTokens(tc.image)
			if result != tc.expected {
				t.Errorf("countImageTokens() = %d, want %d", result, tc.expected)
			}
		})
	}
}
