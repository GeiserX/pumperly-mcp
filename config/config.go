package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func init() {
	if strings.ToLower(os.Getenv("TRANSPORT")) != "stdio" {
		_ = godotenv.Load()
	}
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
