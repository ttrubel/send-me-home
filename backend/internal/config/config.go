package config

import (
	"os"
)

type Config struct {
	Port            string
	GeminiAPIKey    string
	ElevenLabsAPIKey string
	FirestoreProjectID string
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		GeminiAPIKey:    getEnv("GEMINI_API_KEY", ""),
		ElevenLabsAPIKey: getEnv("ELEVENLABS_API_KEY", ""),
		FirestoreProjectID: getEnv("FIRESTORE_PROJECT_ID", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
