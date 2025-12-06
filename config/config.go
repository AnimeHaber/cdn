package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	APIKey        string
	StorageAvatar string
	StorageImage  string
	CacheDuration string // Cache-Control max-age in seconds
}

func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}

	// Ensure data directories exist
	avatarPath := filepath.Join("data", "avatar")
	imagePath := filepath.Join("data", "image")

	if err := os.MkdirAll(avatarPath, 0755); err != nil {
		log.Fatalf("Error creating avatar storage: %v", err)
	}
	if err := os.MkdirAll(imagePath, 0755); err != nil {
		log.Fatalf("Error creating image storage: %v", err)
	}

	return &Config{
		Port:          getEnv("CDN_PORT", ":3000"),
		APIKey:        getEnv("API_KEY", ""),
		StorageAvatar: avatarPath,
		StorageImage:  imagePath,
		CacheDuration: getEnv("CACHE_DURATION", "31536000"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
