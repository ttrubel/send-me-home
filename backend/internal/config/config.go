package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	ElevenLabsAPIKey string
	GCPProjectID     string
}

func Load() *Config {
	// Load .env file if it exists (ignore error if not found)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using environment variables")
	}

	return &Config{
		Port:             getEnv("PORT", "8080"),
		ElevenLabsAPIKey: getEnv("ELEVENLABS_API_KEY", ""),
		GCPProjectID:     getEnv("GOOGLE_CLOUD_PROJECT", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
