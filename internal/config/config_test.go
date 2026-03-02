package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_RequiredVariables(t *testing.T) {
	// Save original environment
	origToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	origAPIKey := os.Getenv("OPENROUTER_API_KEY")
	defer func() {
		os.Setenv("TELEGRAM_BOT_TOKEN", origToken)
		os.Setenv("OPENROUTER_API_KEY", origAPIKey)
	}()

	tests := []struct {
		name          string
		telegramToken string
		apiKey        string
		wantErr       bool
		errContains   string
	}{
		{
			name:          "missing telegram token",
			telegramToken: "",
			apiKey:        "test-api-key",
			wantErr:       true,
			errContains:   "TELEGRAM_BOT_TOKEN",
		},
		{
			name:          "missing openrouter api key",
			telegramToken: "test-token",
			apiKey:        "",
			wantErr:       true,
			errContains:   "OPENROUTER_API_KEY",
		},
		{
			name:          "all required variables present",
			telegramToken: "test-token",
			apiKey:        "test-api-key",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TELEGRAM_BOT_TOKEN", tt.telegramToken)
			os.Setenv("OPENROUTER_API_KEY", tt.apiKey)

			cfg, err := Load()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() expected error containing %q, got nil", tt.errContains)
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Load() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Load() unexpected error = %v", err)
				}
				if cfg == nil {
					t.Error("Load() returned nil config")
				}
			}
		})
	}
}

func TestLoad_OptionalVariables(t *testing.T) {
	// Save and set required environment variables
	origToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	origAPIKey := os.Getenv("OPENROUTER_API_KEY")
	origModel := os.Getenv("MODEL_NAME")
	origTimeout := os.Getenv("SHELL_TIMEOUT")
	origLogLevel := os.Getenv("LOG_LEVEL")

	defer func() {
		os.Setenv("TELEGRAM_BOT_TOKEN", origToken)
		os.Setenv("OPENROUTER_API_KEY", origAPIKey)
		os.Setenv("MODEL_NAME", origModel)
		os.Setenv("SHELL_TIMEOUT", origTimeout)
		os.Setenv("LOG_LEVEL", origLogLevel)
	}()

	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	os.Setenv("OPENROUTER_API_KEY", "test-api-key")

	tests := []struct {
		name            string
		modelName       string
		shellTimeout    string
		logLevel        string
		wantModelName   string
		wantTimeout     time.Duration
		wantLogLevel    string
		wantErr         bool
	}{
		{
			name:          "default values",
			modelName:     "",
			shellTimeout:  "",
			logLevel:      "",
			wantModelName: "google/gemini-2.5-flash-lite",
			wantTimeout:   30 * time.Second,
			wantLogLevel:  "info",
			wantErr:       false,
		},
		{
			name:          "custom values",
			modelName:     "custom/model",
			shellTimeout:  "60",
			logLevel:      "debug",
			wantModelName: "custom/model",
			wantTimeout:   60 * time.Second,
			wantLogLevel:  "debug",
			wantErr:       false,
		},
		{
			name:         "invalid timeout",
			modelName:    "",
			shellTimeout: "invalid",
			logLevel:     "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("MODEL_NAME", tt.modelName)
			os.Setenv("SHELL_TIMEOUT", tt.shellTimeout)
			os.Setenv("LOG_LEVEL", tt.logLevel)

			cfg, err := Load()

			if tt.wantErr {
				if err == nil {
					t.Error("Load() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Load() unexpected error = %v", err)
				return
			}

			if cfg.ModelName != tt.wantModelName {
				t.Errorf("ModelName = %v, want %v", cfg.ModelName, tt.wantModelName)
			}
			if cfg.ShellTimeout != tt.wantTimeout {
				t.Errorf("ShellTimeout = %v, want %v", cfg.ShellTimeout, tt.wantTimeout)
			}
			if cfg.LogLevel != tt.wantLogLevel {
				t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, tt.wantLogLevel)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
