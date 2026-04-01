package config

import (
	"testing"
)

func TestLoadPumperlyConfig_Defaults(t *testing.T) {
	t.Setenv("PUMPERLY_URL", "")
	cfg := LoadPumperlyConfig()
	if cfg.BaseURL != "https://pumperly.com" {
		t.Errorf("BaseURL default: got %q, want %q", cfg.BaseURL, "https://pumperly.com")
	}
}

func TestLoadPumperlyConfig_EnvOverride(t *testing.T) {
	t.Setenv("PUMPERLY_URL", "http://localhost:8080")
	cfg := LoadPumperlyConfig()
	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("BaseURL: got %q, want %q", cfg.BaseURL, "http://localhost:8080")
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envVal   string
		fallback string
		want     string
	}{
		{
			name:     "returns env value when set",
			key:      "TEST_PUMPERLY_VAR",
			envVal:   "custom",
			fallback: "default",
			want:     "custom",
		},
		{
			name:     "returns default when env empty",
			key:      "TEST_PUMPERLY_EMPTY",
			envVal:   "",
			fallback: "fallback",
			want:     "fallback",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(tc.key, tc.envVal)
			got := getEnv(tc.key, tc.fallback)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
