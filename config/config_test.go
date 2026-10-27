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

func TestPumperlyConfig_Validate(t *testing.T) {
	for url, ok := range map[string]bool{
		"https://pumperly.com":  true,
		"http://localhost:3000": true,
		"pumperly.com":          false,
		"ftp://pumperly.com":    false,
		"https://":              false,
	} {
		err := PumperlyConfig{BaseURL: url}.Validate()
		if (err == nil) != ok {
			t.Errorf("Validate(%q) = %v, want ok=%v", url, err, ok)
		}
	}
}

func TestLoadServerConfig(t *testing.T) {
	t.Setenv("TRANSPORT", "STDIO")
	t.Setenv("LISTEN_ADDR", "")
	t.Setenv("MCP_AUTH_TOKEN", " s3cret ")
	t.Setenv("MCP_ALLOWED_ORIGINS", "https://a.example, ,https://b.example")
	cfg := LoadServerConfig()
	if cfg.Transport != "stdio" {
		t.Errorf("Transport = %q", cfg.Transport)
	}
	if cfg.ListenAddr != "127.0.0.1:8080" {
		t.Errorf("ListenAddr = %q", cfg.ListenAddr)
	}
	if cfg.AuthToken != "s3cret" {
		t.Errorf("AuthToken = %q", cfg.AuthToken)
	}
	if len(cfg.AllowedOrigins) != 2 || cfg.AllowedOrigins[1] != "https://b.example" {
		t.Errorf("AllowedOrigins = %q", cfg.AllowedOrigins)
	}

	t.Setenv("TRANSPORT", "")
	if got := LoadServerConfig().Transport; got != "http" {
		t.Errorf("default Transport = %q, want http", got)
	}
}
