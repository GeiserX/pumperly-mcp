package config

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env in the working directory; ignore error if the file is absent.
	_ = godotenv.Load()
}

type PumperlyConfig struct {
	BaseURL string
}

func LoadPumperlyConfig() PumperlyConfig {
	return PumperlyConfig{
		BaseURL: getEnv("PUMPERLY_URL", "https://pumperly.com"),
	}
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
