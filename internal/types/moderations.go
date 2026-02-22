package types

// ModerationRequest represents an OpenAI moderations API request.
type ModerationRequest struct {
	// Required: Input text to classify
	// Can be string or array of strings
	Input ModerationInput `json:"input"`

	// Optional: Model to use (e.g., "omni-moderation-latest", "text-moderation-latest")
	Model string `json:"model,omitempty"`
}

// ModerationInput handles both string and array inputs for moderation.
type ModerationInput struct {
	Values []string
}

// MarshalJSON implements custom marshaling for ModerationInput.
func (m ModerationInput) MarshalJSON() ([]byte, error) {
	if len(m.Values) == 0 {
		return []byte(`""`), nil
	}
	if len(m.Values) == 1 {
		return []byte(`"` + m.Values[0] + `"`), nil
	}
	return marshalStringArray(m.Values)
}

// UnmarshalJSON implements custom unmarshaling for ModerationInput.
func (m *ModerationInput) UnmarshalJSON(data []byte) error {
	m.Values = nil
	// Try string first
	var single string
	if err := unmarshalString(data, &single); err == nil {
		m.Values = []string{single}
		return nil
	}
	// Try array of strings
	return unmarshalStringArray(data, &m.Values)
}

// ModerationResponse represents an OpenAI moderations API response.
type ModerationResponse struct {
	ID      string             `json:"id"`
	Model   string             `json:"model"`
	Results []ModerationResult `json:"results"`
}

// ModerationResult represents a single moderation result.
type ModerationResult struct {
	Flagged                   bool                     `json:"flagged"`
	Categories                ModerationCategories     `json:"categories"`
	CategoryScores            ModerationCategoryScores `json:"category_scores"`
	CategoryAppliedInputTypes map[string][]string      `json:"category_applied_input_types,omitempty"`
}

// ModerationCategories contains boolean flags for each category.
type ModerationCategories struct {
	Sexual                bool `json:"sexual"`
	Hate                  bool `json:"hate"`
	Harassment            bool `json:"harassment"`
	SelfHarm              bool `json:"self-harm"`
	SexualMinors          bool `json:"sexual/minors"`
	HateThreatening       bool `json:"hate/threatening"`
	ViolenceGraphic       bool `json:"violence/graphic"`
	SelfHarmIntent        bool `json:"self-harm/intent"`
	SelfHarmInstructions  bool `json:"self-harm/instructions"`
	HarassmentThreatening bool `json:"harassment/threatening"`
	Violence              bool `json:"violence"`
	Illicit               bool `json:"illicit,omitempty"`
	IllicitViolent        bool `json:"illicit/violent,omitempty"`
}

// ModerationCategoryScores contains confidence scores for each category.
type ModerationCategoryScores struct {
	Sexual                float64 `json:"sexual"`
	Hate                  float64 `json:"hate"`
	Harassment            float64 `json:"harassment"`
	SelfHarm              float64 `json:"self-harm"`
	SexualMinors          float64 `json:"sexual/minors"`
	HateThreatening       float64 `json:"hate/threatening"`
	ViolenceGraphic       float64 `json:"violence/graphic"`
	SelfHarmIntent        float64 `json:"self-harm/intent"`
	SelfHarmInstructions  float64 `json:"self-harm/instructions"`
	HarassmentThreatening float64 `json:"harassment/threatening"`
	Violence              float64 `json:"violence"`
	Illicit               float64 `json:"illicit,omitempty"`
	IllicitViolent        float64 `json:"illicit/violent,omitempty"`
}
