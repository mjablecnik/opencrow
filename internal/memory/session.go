package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"simple-telegram-chatbot/pkg/utils"
)

// SessionState represents the current state of a session
type SessionState struct {
	IsActive       bool      // Whether the session is currently active
	SessionNumber  int       // Current session number
	StartTime      time.Time // When the current session started
	LastActivity   time.Time // Last message timestamp
	MessageCount   int       // Number of messages in current session
	TokenCount     int       // Approximate token count in current session
	HasBeenReset   bool      // Whether the session has been reset at least once
	LastResetTime  time.Time // When the last reset occurred
	LastResetType  string    // Type of last reset: "scheduled", "manual", "token_based", or ""
}

// SessionManager handles session logging and tracking
type SessionManager struct {
	mu                sync.RWMutex
	currentDate       string // YYYY-MM-DD format
	currentSessionNum int
	currentLogPath    string
	memoryBasePath    string
	state             SessionState // Current session state
	logger            *utils.Logger // Logger for session operations
}

// NewSessionManager creates a new session manager
func NewSessionManager(memoryBasePath string) *SessionManager {
	return &SessionManager{
		memoryBasePath: memoryBasePath,
		logger:         utils.NewLogger("info"), // Default to info level
		state: SessionState{
			IsActive:      false,
			SessionNumber: 0,
			StartTime:     time.Time{},
			LastActivity:  time.Time{},
			MessageCount:  0,
			TokenCount:    0,
			HasBeenReset:  false,
			LastResetTime: time.Time{},
			LastResetType: "",
		},
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
		
		// Reset session state for new day
		sm.state.SessionNumber = 1
		sm.state.StartTime = now
		sm.state.MessageCount = 0
		sm.state.TokenCount = 0
	}

	// Ensure daily folder exists
	dailyFolderPath := filepath.Join(sm.memoryBasePath, "chat", currentDate)
	if err := os.MkdirAll(dailyFolderPath, 0755); err != nil {
		return fmt.Errorf("failed to create daily folder: %w", err)
	}

	// Create session log file path if not set
	if sm.currentLogPath == "" {
		sm.currentLogPath = filepath.Join(dailyFolderPath, fmt.Sprintf("session-%03d.log", sm.currentSessionNum))
		
		sm.logger.InfoWithComponent("SessionManager", "Creating new session log file",
			"session_number", sm.currentSessionNum,
			"log_path", sm.currentLogPath,
		)
		
		// Mark session as active when first message is logged
		if !sm.state.IsActive {
			sm.state.IsActive = true
			sm.state.StartTime = now
			sm.state.SessionNumber = sm.currentSessionNum
		}
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

	// Update session state
	sm.state.LastActivity = now
	sm.state.MessageCount++
	// Approximate token count: ~4 characters per token
	sm.state.TokenCount += len(content) / 4

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

// PerformSessionReset clears all in-memory conversation context and starts fresh
// This is called after scheduled summarization completes at 4:00 AM
// It increments the session number and prepares for a new session
func (sm *SessionManager) PerformSessionReset(triggerReason string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Capture session state before reset
	now := time.Now()
	currentDate := now.Format("2006-01-02")
	oldSessionNum := sm.currentSessionNum
	
	// Log the session reset operation with comprehensive details
	sm.logger.InfoWithComponent("SessionManager", "Session reset triggered",
		"trigger_reason", triggerReason,
		"session_number_before", oldSessionNum,
		"date", currentDate,
		"timestamp", now.Format("2006-01-02 15:04:05"),
	)

	// Initialize currentDate if it's empty (first time)
	if sm.currentDate == "" {
		sm.currentDate = currentDate
		sm.logger.InfoWithComponent("SessionManager", "Initializing date for first time",
			"date", currentDate,
		)
	}

	// Ensure we're on the current date
	if sm.currentDate != currentDate {
		sm.currentDate = currentDate
		sm.currentSessionNum = 1
		sm.logger.InfoWithComponent("SessionManager", "New day detected, resetting session number to 1",
			"new_date", currentDate,
		)
	} else {
		// Increment session number for same day
		// Special case: if session number is 0 (never initialized), start at 1
		if sm.currentSessionNum == 0 {
			sm.currentSessionNum = 1
		} else {
			sm.currentSessionNum++
		}
	}

	// Clear the current log path to start fresh
	// This ensures the next message will create a new session log file
	sm.currentLogPath = ""

	// Update session state
	sm.state.IsActive = false
	sm.state.SessionNumber = sm.currentSessionNum
	sm.state.StartTime = time.Time{} // Will be set on next message
	sm.state.LastActivity = now
	sm.state.MessageCount = 0
	sm.state.TokenCount = 0
	sm.state.HasBeenReset = true
	sm.state.LastResetTime = now
	sm.state.LastResetType = triggerReason

	// Log completion with new session number
	sm.logger.InfoWithComponent("SessionManager", "Session reset complete",
		"trigger_reason", triggerReason,
		"session_number_before", oldSessionNum,
		"session_number_after", sm.currentSessionNum,
		"timestamp", now.Format("2006-01-02 15:04:05"),
	)

	return nil
}

// PerformManualSessionReset handles user-initiated or agent-initiated session resets
// It triggers immediate summarization, topic extraction, and then performs session reset
// This method coordinates with SummaryManager to generate session summary and extract topics
func (sm *SessionManager) PerformManualSessionReset(summaryManager SummaryManagerInterface) error {
	sm.mu.RLock()
	sessionNum := sm.currentSessionNum
	currentDate := sm.currentDate
	logPath := sm.currentLogPath
	messageCount := sm.state.MessageCount
	isActive := sm.state.IsActive
	sm.mu.RUnlock()

	// Check if there's an active session to summarize
	// Session is considered empty if:
	// 1. Session number is 0 (never initialized)
	// 2. No log path exists (no messages written)
	// 3. Message count is 0
	// 4. Session is not marked as active
	if sessionNum == 0 || logPath == "" || messageCount == 0 || !isActive {
		sm.logger.InfoWithComponent("SessionManager", "No active session to summarize, performing simple reset",
			"session_number", sessionNum,
			"log_path", logPath,
			"message_count", messageCount,
			"is_active", isActive,
		)
		// Just perform a simple reset without summarization
		return sm.PerformSessionReset("manual_reset")
	}

	// Capture session state before reset
	now := time.Now()

	// Log the manual reset initiation
	sm.logger.InfoWithComponent("SessionManager", "Manual session reset initiated",
		"session_number", sessionNum,
		"date", currentDate,
		"timestamp", now.Format("2006-01-02 15:04:05"),
	)

	// Step 1: Generate session summary
	sm.logger.InfoWithComponent("SessionManager", "Generating session summary for manual reset",
		"session_number", sessionNum,
	)
	
	if err := summaryManager.GenerateSessionSummary(); err != nil {
		sm.logger.ErrorWithDetails("SessionManager", "Failed to generate session summary during manual reset", err,
			"session_number", sessionNum,
			"date", currentDate,
		)
		return fmt.Errorf("failed to generate session summary during manual reset: %w", err)
	}

	sm.logger.InfoWithComponent("SessionManager", "Session summary generated successfully",
		"session_number", sessionNum,
	)

	// Step 2: Perform topic extraction from the session summary
	dailyFolderPath := filepath.Join(sm.memoryBasePath, "chat", currentDate)
	summaryPath := filepath.Join(dailyFolderPath, fmt.Sprintf("session-%03d-summary.md", sessionNum))

	sm.logger.InfoWithComponent("SessionManager", "Extracting topics from session summary",
		"session_number", sessionNum,
		"summary_path", summaryPath,
	)

	// Read the generated summary
	summaryContent, err := os.ReadFile(summaryPath)
	if err != nil {
		sm.logger.ErrorWithDetails("SessionManager", "Failed to read session summary for topic extraction", err,
			"session_number", sessionNum,
			"summary_path", summaryPath,
		)
		return fmt.Errorf("failed to read session summary for topic extraction: %w", err)
	}

	// Extract topics from the summary
	if err := summaryManager.ExtractTopicsFromContent(string(summaryContent)); err != nil {
		// Log the error but don't fail the reset - topic extraction is not critical
		sm.logger.WarnWithComponent("SessionManager", "Topic extraction failed during manual reset",
			"session_number", sessionNum,
			"error", err.Error(),
		)
	} else {
		sm.logger.InfoWithComponent("SessionManager", "Topic extraction completed successfully",
			"session_number", sessionNum,
		)
	}

	// Step 3: Perform the actual session reset
	sm.logger.InfoWithComponent("SessionManager", "Performing session reset after manual reset operations",
		"session_number_before", sessionNum,
	)
	
	if err := sm.PerformSessionReset("manual_reset"); err != nil {
		sm.logger.ErrorWithDetails("SessionManager", "Failed to perform session reset", err,
			"session_number", sessionNum,
		)
		return fmt.Errorf("failed to perform session reset: %w", err)
	}

	sm.logger.InfoWithComponent("SessionManager", "Manual session reset complete",
		"session_number_before", sessionNum,
		"session_number_after", sm.GetCurrentSessionNumber(),
		"timestamp", time.Now().Format("2006-01-02 15:04:05"),
	)

	return nil
}

// SummaryManagerInterface defines the interface for summary operations needed by SessionManager
// This avoids circular dependencies between SessionManager and SummaryManager
type SummaryManagerInterface interface {
	GenerateSessionSummary() error
	ExtractTopicsFromContent(content string) error
}

// GetSessionState returns a copy of the current session state
func (sm *SessionManager) GetSessionState() SessionState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state
}

// IsSessionActive returns whether the session is currently active
func (sm *SessionManager) IsSessionActive() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.IsActive
}

// GetSessionMessageCount returns the number of messages in the current session
func (sm *SessionManager) GetSessionMessageCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.MessageCount
}

// GetSessionTokenCount returns the approximate token count for the current session
func (sm *SessionManager) GetSessionTokenCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.TokenCount
}

// GetSessionStartTime returns when the current session started
func (sm *SessionManager) GetSessionStartTime() time.Time {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.StartTime
}

// GetLastActivity returns the timestamp of the last message
func (sm *SessionManager) GetLastActivity() time.Time {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.LastActivity
}

// HasBeenReset returns whether the session has been reset at least once
func (sm *SessionManager) HasBeenReset() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.HasBeenReset
}

// GetLastResetInfo returns information about the last reset
func (sm *SessionManager) GetLastResetInfo() (time.Time, string) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.LastResetTime, sm.state.LastResetType
}
