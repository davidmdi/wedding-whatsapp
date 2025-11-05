package config

import (
	"os"
)

// Config holds the application configuration
type Config struct {
	WhatsAppDataDir string
	WeddingDate     string
	WeddingLocation string
	BrideName       string
	GroomName       string
}

// LoadConfig loads configuration from environment variables or defaults
func LoadConfig() *Config {
	return &Config{
		WhatsAppDataDir: getEnv("WHATSAPP_DATA_DIR", "data"),
		WeddingDate:     getEnv("WEDDING_DATE", "Saturday, January 1, 2025"),
		WeddingLocation: getEnv("WEDDING_LOCATION", "Venue TBD"),
		BrideName:       getEnv("BRIDE_NAME", "Bride"),
		GroomName:       getEnv("GROOM_NAME", "Groom"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
