package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SessionManager handles session logging and tracking
type SessionManager struct {
	mu                sync.RWMutex
	currentDate       string // YYYY-MM-DD format
	currentSessionNum int
	currentLogPath    string
	memoryBasePath    string
}

// NewSessionManager creates a new session manager
func NewSessionManager(memoryBasePath string) *SessionManager {
	return &SessionManager{
		memoryBasePath: memoryBasePath,
	}
}

// AppendToSessionLog appends a message to the current session log
// Creates daily folder and session log file if they don't exist
func (sm *SessionManager) AppendToSessionLog(role, content string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Get current date in YYYY-MM-DD format
	now := time.Now()
	currentDate := now.Format("2006-01-02")

	// Check if we need to create a new session (new day)
	if sm.currentDate != currentDate {
		sm.currentDate = currentDate
		sm.currentSessionNum = 1
		sm.currentLogPath = ""
	}

	// Ensure daily folder exists
	dailyFolderPath := filepath.Join(sm.memoryBasePath, "chat", currentDate)
	if err := os.MkdirAll(dailyFolderPath, 0755); err != nil {
		return fmt.Errorf("failed to create daily folder: %w", err)
	}

	// Create session log file path if not set
	if sm.currentLogPath == "" {
		sm.currentLogPath = filepath.Join(dailyFolderPath, fmt.Sprintf("session-%03d.log", sm.currentSessionNum))
	}

	// Format the log entry with timestamp
	timestamp := now.Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s: %s\n\n", timestamp, role, content)

	// Append to session log file
	f, err := os.OpenFile(sm.currentLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open session log file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(logEntry); err != nil {
		return fmt.Errorf("failed to write to session log: %w", err)
	}

	return nil
}

// GetCurrentSessionNumber returns the current session number
func (sm *SessionManager) GetCurrentSessionNumber() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentSessionNum
}

// GetCurrentSessionPath returns the path to the current session log file
func (sm *SessionManager) GetCurrentSessionPath() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentLogPath
}

// GetCurrentDate returns the current date in YYYY-MM-DD format
func (sm *SessionManager) GetCurrentDate() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentDate
}

// IncrementSession increments the session number and resets the log path
// This should be called when a session reset occurs
func (sm *SessionManager) IncrementSession() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Ensure we're on the current date
	now := time.Now()
	currentDate := now.Format("2006-01-02")

	if sm.currentDate != currentDate {
		sm.currentDate = currentDate
		sm.currentSessionNum = 1
	} else {
		sm.currentSessionNum++
	}

	// Reset the log path so it will be created on next append
	sm.currentLogPath = ""

	return nil
}
