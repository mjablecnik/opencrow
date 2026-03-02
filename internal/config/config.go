package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Required configuration
	TelegramBotToken string
	OpenRouterAPIKey string

	// Optional configuration with defaults
	ModelName    string
	ShellTimeout time.Duration
	LogLevel     string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Required: Telegram bot token
	cfg.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	// Required: OpenRouter API key
	cfg.OpenRouterAPIKey = os.Getenv("OPENROUTER_API_KEY")
	if cfg.OpenRouterAPIKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY environment variable is required")
	}

	// Optional: Model name (default: google/gemini-2.5-flash-lite)
	cfg.ModelName = os.Getenv("MODEL_NAME")
	if cfg.ModelName == "" {
		cfg.ModelName = "google/gemini-2.5-flash-lite"
	}

	// Optional: Shell timeout (default: 30 seconds)
	shellTimeoutStr := os.Getenv("SHELL_TIMEOUT")
	if shellTimeoutStr == "" {
		cfg.ShellTimeout = 30 * time.Second
	} else {
		timeoutSec, err := strconv.Atoi(shellTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid SHELL_TIMEOUT value: %v", err)
		}
		cfg.ShellTimeout = time.Duration(timeoutSec) * time.Second
	}

	// Optional: Log level (default: info)
	cfg.LogLevel = os.Getenv("LOG_LEVEL")
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	return cfg, nil
}
