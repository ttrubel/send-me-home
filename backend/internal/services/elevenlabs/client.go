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

// EmotionType represents different emotional contexts for speech
type EmotionType string

const (
	EmotionNeutral  EmotionType = "neutral"
	EmotionHappy    EmotionType = "happy"
	EmotionAngry    EmotionType = "angry"
	EmotionFurious  EmotionType = "furious"
	EmotionSad      EmotionType = "sad"
	EmotionNervous  EmotionType = "nervous"
)

// getVoiceSettings returns optimized voice settings based on emotion
//
// Settings explanation:
// - stability: Lower = more varied/expressive, Higher = more consistent/monotone (0.0-1.0)
// - similarity_boost: How much to match the original voice (0.0-1.0)
// - style: Exaggeration of emotion/style (0.0-1.0, only works with v2 models)
// - use_speaker_boost: Enhances clarity and consistency
//
// Emotion tuning:
// - Happy: Medium-low stability for upbeat variance, high style for cheerful tone
// - Angry: Low stability for aggressive delivery, high style for intensity
// - Furious: Lowest stability for maximum rage, highest style for extreme emotion
// - Sad: Higher stability for somber tone, moderate style
// - Nervous: Low stability for jittery delivery
// - Neutral: Balanced settings
func (c *Client) getVoiceSettings(emotion EmotionType) map[string]interface{} {
	switch emotion {
	case EmotionHappy:
		// Higher stability for clear, upbeat delivery
		return map[string]interface{}{
			"stability":        0.35,
			"similarity_boost": 0.75,
			"style":            0.5,  // Exaggerate emotion
			"use_speaker_boost": true,
		}
	case EmotionAngry:
		// Lower stability for more aggressive, varied delivery
		return map[string]interface{}{
			"stability":        0.25,
			"similarity_boost": 0.65,
			"style":            0.75, // High style for emotion
			"use_speaker_boost": true,
		}
	case EmotionFurious:
		// Maximum emotion, lowest stability for intense rage
		return map[string]interface{}{
			"stability":        0.15,
			"similarity_boost": 0.60,
			"style":            0.9,  // Maximum style/emotion
			"use_speaker_boost": true,
		}
	case EmotionSad:
		// Moderate stability, lower boost for somber tone
		return map[string]interface{}{
			"stability":        0.50,
			"similarity_boost": 0.70,
			"style":            0.4,
			"use_speaker_boost": true,
		}
	case EmotionNervous:
		// Lower stability for anxious, jittery delivery
		return map[string]interface{}{
			"stability":        0.30,
			"similarity_boost": 0.75,
			"style":            0.5,
			"use_speaker_boost": true,
		}
	default: // EmotionNeutral
		return map[string]interface{}{
			"stability":        0.50,
			"similarity_boost": 0.75,
			"style":            0.0,
			"use_speaker_boost": true,
		}
	}
}

// TextToSpeech converts text to speech and returns the audio data
func (c *Client) TextToSpeech(ctx context.Context, voiceID, text string) ([]byte, error) {
	return c.TextToSpeechWithEmotion(ctx, voiceID, text, EmotionNeutral)
}

// TextToSpeechWithEmotion converts text to speech with specific emotional delivery
func (c *Client) TextToSpeechWithEmotion(ctx context.Context, voiceID, text string, emotion EmotionType) ([]byte, error) {
	// If no API key, return nil (mock mode)
	if c.apiKey == "" {
		return nil, nil
	}

	url := fmt.Sprintf("%s/text-to-speech/%s", apiBaseURL, voiceID)

	reqBody := TextToSpeechRequest{
		Text:          text,
		ModelID:       "eleven_turbo_v2_5",
		VoiceSettings: c.getVoiceSettings(emotion),
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
