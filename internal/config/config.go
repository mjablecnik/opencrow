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
	StartTime    time.Time // Application start time for log file naming

	// Memory and scheduling configuration
	Memory MemoryConfig
}

// MemoryConfig holds memory and scheduling configuration
type MemoryConfig struct {
	TokenThreshold              int
	TopicSizeThreshold          int64
	NotesEnabled                bool
	NotesCleanupEnabled         bool
	NotesMaxAgeDays             int
	NotesCompletedRetentionDays int
	NotesScratchpadMaxAgeDays   int
	DailyMaintenanceTime        string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		StartTime: time.Now(), // Set start time for log file naming
	}

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

	// Memory configuration: Token threshold (default: 50000)
	memoryTokenThresholdStr := os.Getenv("MEMORY_TOKEN_THRESHOLD")
	if memoryTokenThresholdStr == "" {
		cfg.Memory.TokenThreshold = 50000
	} else {
		tokenThreshold, err := strconv.Atoi(memoryTokenThresholdStr)
		if err != nil {
			return nil, fmt.Errorf("invalid MEMORY_TOKEN_THRESHOLD value: %v", err)
		}
		cfg.Memory.TokenThreshold = tokenThreshold
	}

	// Memory configuration: Topic size threshold (default: 100KB = 102400 bytes)
	topicSizeThresholdStr := os.Getenv("TOPIC_SIZE_THRESHOLD")
	if topicSizeThresholdStr == "" {
		cfg.Memory.TopicSizeThreshold = 102400
	} else {
		topicSizeThreshold, err := strconv.ParseInt(topicSizeThresholdStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid TOPIC_SIZE_THRESHOLD value: %v", err)
		}
		cfg.Memory.TopicSizeThreshold = topicSizeThreshold
	}

	// Memory configuration: Notes enabled (default: true)
	notesEnabledStr := os.Getenv("NOTES_ENABLED")
	if notesEnabledStr == "" {
		cfg.Memory.NotesEnabled = true
	} else {
		notesEnabled, err := strconv.ParseBool(notesEnabledStr)
		if err != nil {
			return nil, fmt.Errorf("invalid NOTES_ENABLED value: %v", err)
		}
		cfg.Memory.NotesEnabled = notesEnabled
	}

	// Memory configuration: Notes cleanup enabled (default: true)
	notesCleanupEnabledStr := os.Getenv("NOTES_CLEANUP_ENABLED")
	if notesCleanupEnabledStr == "" {
		cfg.Memory.NotesCleanupEnabled = true
	} else {
		notesCleanupEnabled, err := strconv.ParseBool(notesCleanupEnabledStr)
		if err != nil {
			return nil, fmt.Errorf("invalid NOTES_CLEANUP_ENABLED value: %v", err)
		}
		cfg.Memory.NotesCleanupEnabled = notesCleanupEnabled
	}

	// Memory configuration: Notes max age days (default: 30)
	notesMaxAgeDaysStr := os.Getenv("NOTES_MAX_AGE_DAYS")
	if notesMaxAgeDaysStr == "" {
		cfg.Memory.NotesMaxAgeDays = 30
	} else {
		notesMaxAgeDays, err := strconv.Atoi(notesMaxAgeDaysStr)
		if err != nil {
			return nil, fmt.Errorf("invalid NOTES_MAX_AGE_DAYS value: %v", err)
		}
		cfg.Memory.NotesMaxAgeDays = notesMaxAgeDays
	}

	// Memory configuration: Notes completed retention days (default: 7)
	notesCompletedRetentionDaysStr := os.Getenv("NOTES_COMPLETED_RETENTION_DAYS")
	if notesCompletedRetentionDaysStr == "" {
		cfg.Memory.NotesCompletedRetentionDays = 7
	} else {
		notesCompletedRetentionDays, err := strconv.Atoi(notesCompletedRetentionDaysStr)
		if err != nil {
			return nil, fmt.Errorf("invalid NOTES_COMPLETED_RETENTION_DAYS value: %v", err)
		}
		cfg.Memory.NotesCompletedRetentionDays = notesCompletedRetentionDays
	}

	// Memory configuration: Notes scratchpad max age days (default: 7)
	notesScratchpadMaxAgeDaysStr := os.Getenv("NOTES_SCRATCHPAD_MAX_AGE_DAYS")
	if notesScratchpadMaxAgeDaysStr == "" {
		cfg.Memory.NotesScratchpadMaxAgeDays = 7
	} else {
		notesScratchpadMaxAgeDays, err := strconv.Atoi(notesScratchpadMaxAgeDaysStr)
		if err != nil {
			return nil, fmt.Errorf("invalid NOTES_SCRATCHPAD_MAX_AGE_DAYS value: %v", err)
		}
		cfg.Memory.NotesScratchpadMaxAgeDays = notesScratchpadMaxAgeDays
	}

	// Memory configuration: Daily maintenance time (default: "0 4 * * *" = 4:00 AM daily)
	cfg.Memory.DailyMaintenanceTime = os.Getenv("DAILY_MAINTENANCE_TIME")
	if cfg.Memory.DailyMaintenanceTime == "" {
		cfg.Memory.DailyMaintenanceTime = "0 4 * * *"
	}

	return cfg, nil
}
