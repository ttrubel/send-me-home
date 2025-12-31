package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	apiBaseURL = "https://api.elevenlabs.io/v1"
)

// Client handles ElevenLabs text-to-speech API
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new ElevenLabs client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// TextToSpeechRequest represents the API request payload
type TextToSpeechRequest struct {
	Text          string                 `json:"text"`
	ModelID       string                 `json:"model_id"`
	VoiceSettings map[string]interface{} `json:"voice_settings,omitempty"`
}

// TextToSpeech converts text to speech and returns the audio data
func (c *Client) TextToSpeech(ctx context.Context, voiceID, text string) ([]byte, error) {
	// If no API key, return nil (mock mode)
	if c.apiKey == "" {
		return nil, nil
	}

	url := fmt.Sprintf("%s/text-to-speech/%s", apiBaseURL, voiceID)

	reqBody := TextToSpeechRequest{
		Text:    text,
		ModelID: "eleven_turbo_v2_5",
		VoiceSettings: map[string]interface{}{
			"stability":        0.5,
			"similarity_boost": 0.75,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("elevenlabs API error (status %d): %s", resp.StatusCode, string(body))
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return audioData, nil
}

// TextToSpeechStream converts text to speech and returns a streaming reader
// This is useful for streaming audio chunks in real-time
func (c *Client) TextToSpeechStream(ctx context.Context, voiceID, text string) (io.ReadCloser, error) {
	// If no API key, return nil (mock mode)
	if c.apiKey == "" {
		return nil, nil
	}

	url := fmt.Sprintf("%s/text-to-speech/%s/stream", apiBaseURL, voiceID)

	reqBody := TextToSpeechRequest{
		Text:    text,
		ModelID: "eleven_turbo_v2_5",
		VoiceSettings: map[string]interface{}{
			"stability":        0.5,
			"similarity_boost": 0.75,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("elevenlabs API error (status %d): %s", resp.StatusCode, string(body))
	}

	return resp.Body, nil
}
