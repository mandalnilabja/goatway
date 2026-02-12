package types

// AudioSpeechRequest represents an OpenAI text-to-speech API request.
type AudioSpeechRequest struct {
	// Required: ID of the model to use (e.g., "tts-1", "tts-1-hd")
	Model string `json:"model"`

	// Required: Text to generate audio for (max 4096 characters)
	Input string `json:"input"`

	// Required: Voice to use (e.g., "alloy", "echo", "fable", "onyx", "nova", "shimmer")
	Voice string `json:"voice"`

	// Optional: Audio output format
	// Values: "mp3" (default), "opus", "aac", "flac", "wav", "pcm"
	ResponseFormat string `json:"response_format,omitempty"`

	// Optional: Speed of generated audio (0.25 to 4.0, default 1.0)
	Speed *float64 `json:"speed,omitempty"`
}

// AudioTranscriptionRequest represents an OpenAI audio transcription API request.
// This is sent as multipart/form-data with the audio file.
type AudioTranscriptionRequest struct {
	// Required: ID of the model to use (e.g., "whisper-1")
	Model string `json:"model"`

	// Required: Audio file (handled separately as multipart file)
	// Supported formats: flac, mp3, mp4, mpeg, mpga, m4a, ogg, wav, webm

	// Optional: Language of the audio in ISO-639-1 format
	Language string `json:"language,omitempty"`

	// Optional: Prompt to guide the model's style
	Prompt string `json:"prompt,omitempty"`

	// Optional: Response format
	// Values: "json" (default), "text", "srt", "verbose_json", "vtt"
	ResponseFormat string `json:"response_format,omitempty"`

	// Optional: Temperature for sampling (0 to 1)
	Temperature *float64 `json:"temperature,omitempty"`

	// Optional: Timestamp granularities (for verbose_json)
	// Values: "word", "segment"
	TimestampGranularities []string `json:"timestamp_granularities,omitempty"`
}

// AudioTranslationRequest represents an OpenAI audio translation API request.
// Translates audio into English. Sent as multipart/form-data.
type AudioTranslationRequest struct {
	// Required: ID of the model to use (e.g., "whisper-1")
	Model string `json:"model"`

	// Required: Audio file (handled separately as multipart file)

	// Optional: Prompt to guide the model's style
	Prompt string `json:"prompt,omitempty"`

	// Optional: Response format
	// Values: "json" (default), "text", "srt", "verbose_json", "vtt"
	ResponseFormat string `json:"response_format,omitempty"`

	// Optional: Temperature for sampling (0 to 1)
	Temperature *float64 `json:"temperature,omitempty"`
}

// AudioTranscriptionResponse represents a transcription/translation response.
type AudioTranscriptionResponse struct {
	Text string `json:"text"`
}

// AudioVerboseResponse represents a verbose transcription response.
type AudioVerboseResponse struct {
	Task     string         `json:"task"`
	Language string         `json:"language"`
	Duration float64        `json:"duration"`
	Text     string         `json:"text"`
	Words    []AudioWord    `json:"words,omitempty"`
	Segments []AudioSegment `json:"segments,omitempty"`
}

// AudioWord represents a word with timing information.
type AudioWord struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// AudioSegment represents a segment with timing information.
type AudioSegment struct {
	ID               int     `json:"id"`
	Seek             int     `json:"seek"`
	Start            float64 `json:"start"`
	End              float64 `json:"end"`
	Text             string  `json:"text"`
	Tokens           []int   `json:"tokens"`
	Temperature      float64 `json:"temperature"`
	AvgLogprob       float64 `json:"avg_logprob"`
	CompressionRatio float64 `json:"compression_ratio"`
	NoSpeechProb     float64 `json:"no_speech_prob"`
}
