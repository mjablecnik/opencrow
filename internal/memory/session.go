package memory

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"simple-telegram-chatbot/pkg/utils"
)

// SessionState represents the current state of a session
type SessionState struct {
	IsActive       bool      // Whether the session is currently active
	SessionNumber  int       // Current session number (for archived sessions in daily folders)
	StartTime      time.Time // When the current session started (first message timestamp)
	LastActivity   time.Time // Last message timestamp
	MessageCount   int       // Number of messages in current session
	TokenCount     int       // Approximate token count in current session
	HasBeenReset   bool      // Whether the session has been reset at least once
	LastResetTime  time.Time // When the last reset occurred
	LastResetType  string    // Type of last reset: "scheduled", "manual", or ""
}

// SessionManager handles session logging and tracking using session-latest.log approach
// All active messages are written to memory/chat/session-latest.log
// On reset, this file is moved to the appropriate daily folder with proper numbering
type SessionManager struct {
	mu             sync.RWMutex
	latestLogPath  string        // Path to session-latest.log
	memoryBasePath string        // Base path for memory storage
	state          SessionState  // Current session state
	logger         *utils.Logger // Logger for session operations
}

// NewSessionManager creates a new session manager
func NewSessionManager(memoryBasePath string) *SessionManager {
	latestLogPath := filepath.Join(memoryBasePath, "chat", "session-latest.log")
	
	return &SessionManager{
		memoryBasePath: memoryBasePath,
		latestLogPath:  latestLogPath,
		logger:         utils.NewLogger("info"),
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

// AppendToSessionLog appends a message to session-latest.log
// This is the only file that receives new messages during active conversation
func (sm *SessionManager) AppendToSessionLog(role, content string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()

	// Ensure chat folder exists
	chatFolderPath := filepath.Join(sm.memoryBasePath, "chat")
	if err := os.MkdirAll(chatFolderPath, 0755); err != nil {
		return fmt.Errorf("failed to create chat folder: %w", err)
	}

	// Format the log entry with timestamp
	timestamp := now.Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s: %s\n\n", timestamp, role, content)

	// Append to session-latest.log
	f, err := os.OpenFile(sm.latestLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open session-latest.log: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(logEntry); err != nil {
		return fmt.Errorf("failed to write to session-latest.log: %w", err)
	}

	// Update session state
	if !sm.state.IsActive {
		sm.state.IsActive = true
		sm.state.StartTime = now
		sm.logger.InfoWithComponent("SessionManager", "Session started",
			"start_time", now.Format("2006-01-02 15:04:05"),
		)
	}
	
	sm.state.LastActivity = now
	sm.state.MessageCount++
	// Approximate token count: ~4 characters per token
	sm.state.TokenCount += len(content) / 4

	return nil
}

// ArchiveLatestSession moves session-latest.log to the appropriate daily folder
// It determines the date from the first message timestamp in the file
// Returns the path where the session was archived, or empty string if no session exists
func (sm *SessionManager) ArchiveLatestSession() (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if session-latest.log exists
	if _, err := os.Stat(sm.latestLogPath); os.IsNotExist(err) {
		sm.logger.InfoWithComponent("SessionManager", "No session-latest.log to archive")
		return "", nil
	}

	// Read the file to find the first message timestamp
	firstMessageDate, err := sm.extractFirstMessageDate(sm.latestLogPath)
	if err != nil {
		return "", fmt.Errorf("failed to extract first message date: %w", err)
	}

	if firstMessageDate == "" {
		sm.logger.WarnWithComponent("SessionManager", "session-latest.log is empty, removing it")
		if err := os.Remove(sm.latestLogPath); err != nil {
			return "", fmt.Errorf("failed to remove empty session-latest.log: %w", err)
		}
		return "", nil
	}

	// Create daily folder if it doesn't exist
	dailyFolderPath := filepath.Join(sm.memoryBasePath, "chat", firstMessageDate)
	if err := os.MkdirAll(dailyFolderPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create daily folder: %w", err)
	}

	// Find the next available session number in that folder
	sessionNum, err := sm.getNextSessionNumber(dailyFolderPath)
	if err != nil {
		return "", fmt.Errorf("failed to get next session number: %w", err)
	}

	// Create destination path
	destPath := filepath.Join(dailyFolderPath, fmt.Sprintf("session-%03d.log", sessionNum))

	// Move the file
	if err := os.Rename(sm.latestLogPath, destPath); err != nil {
		return "", fmt.Errorf("failed to move session-latest.log to %s: %w", destPath, err)
	}

	sm.logger.InfoWithComponent("SessionManager", "Session archived",
		"from", sm.latestLogPath,
		"to", destPath,
		"date", firstMessageDate,
		"session_number", sessionNum,
	)

	return destPath, nil
}

// extractFirstMessageDate reads the first line of a log file and extracts the date
// Returns date in YYYY-MM-DD format, or empty string if file is empty or invalid
func (sm *SessionManager) extractFirstMessageDate(logPath string) (string, error) {
	file, err := os.Open(logPath)
	if err != nil {
		return "", fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Look for timestamp pattern: [YYYY-MM-DD HH:MM:SS]
		if strings.HasPrefix(line, "[") && len(line) > 20 {
			// Extract date part: [2026-03-03 19:38:48]
			timestampEnd := strings.Index(line, "]")
			if timestampEnd > 0 {
				timestamp := line[1:timestampEnd]
				// Extract just the date part (YYYY-MM-DD)
				if len(timestamp) >= 10 {
					return timestamp[:10], nil
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading log file: %w", err)
	}

	return "", nil // Empty file
}

// getNextSessionNumber finds the next available session number in a daily folder
// Returns 1 if no sessions exist, otherwise returns max + 1
func (sm *SessionManager) getNextSessionNumber(dailyFolderPath string) (int, error) {
	entries, err := os.ReadDir(dailyFolderPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read daily folder: %w", err)
	}

	maxSessionNum := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Look for session-NNN.log files
		name := entry.Name()
		if strings.HasPrefix(name, "session-") && strings.HasSuffix(name, ".log") {
			// Extract number from session-NNN.log
			var num int
			if _, err := fmt.Sscanf(name, "session-%d.log", &num); err == nil {
				if num > maxSessionNum {
					maxSessionNum = num
				}
			}
		}
	}

	return maxSessionNum + 1, nil
}

// PerformManualSessionReset handles user-initiated /reset command
// Archives session-latest.log to appropriate daily folder WITHOUT generating summaries
// Summaries are only generated during scheduled maintenance
func (sm *SessionManager) PerformManualSessionReset() error {
	sm.logger.InfoWithComponent("SessionManager", "Manual session reset initiated")

	// Archive the current session
	archivedPath, err := sm.ArchiveLatestSession()
	if err != nil {
		return fmt.Errorf("failed to archive session: %w", err)
	}

	if archivedPath == "" {
		sm.logger.InfoWithComponent("SessionManager", "No active session to archive")
	} else {
		sm.logger.InfoWithComponent("SessionManager", "Session archived successfully",
			"archived_path", archivedPath,
		)
	}

	// Reset session state
	sm.mu.Lock()
	sm.state.IsActive = false
	sm.state.SessionNumber = 0
	sm.state.StartTime = time.Time{}
	sm.state.MessageCount = 0
	sm.state.TokenCount = 0
	sm.state.HasBeenReset = true
	sm.state.LastResetTime = time.Now()
	sm.state.LastResetType = "manual"
	sm.mu.Unlock()

	sm.logger.InfoWithComponent("SessionManager", "Manual session reset complete")

	return nil
}

// PerformScheduledSessionReset handles scheduled 4:00 AM maintenance reset
// Archives session-latest.log and returns the date folder that needs summarization
// The caller (scheduler) is responsible for:
// 1. Sending maintenance message to Telegram
// 2. Calling GenerateDailySummary for the returned date
// 3. Calling topic extraction
func (sm *SessionManager) PerformScheduledSessionReset() (string, error) {
	sm.logger.InfoWithComponent("SessionManager", "Scheduled session reset initiated")

	// Archive the current session
	archivedPath, err := sm.ArchiveLatestSession()
	if err != nil {
		return "", fmt.Errorf("failed to archive session: %w", err)
	}

	var dateToSummarize string
	if archivedPath != "" {
		// Extract date from archived path: memory/chat/2026-03-03/session-001.log
		parts := strings.Split(archivedPath, string(filepath.Separator))
		for i, part := range parts {
			if part == "chat" && i+1 < len(parts) {
				dateToSummarize = parts[i+1]
				break
			}
		}
		
		sm.logger.InfoWithComponent("SessionManager", "Session archived successfully",
			"archived_path", archivedPath,
			"date_to_summarize", dateToSummarize,
		)
	} else {
		sm.logger.InfoWithComponent("SessionManager", "No active session to archive")
	}

	// Reset session state
	sm.mu.Lock()
	sm.state.IsActive = false
	sm.state.SessionNumber = 0
	sm.state.StartTime = time.Time{}
	sm.state.MessageCount = 0
	sm.state.TokenCount = 0
	sm.state.HasBeenReset = true
	sm.state.LastResetTime = time.Now()
	sm.state.LastResetType = "scheduled"
	sm.mu.Unlock()

	sm.logger.InfoWithComponent("SessionManager", "Scheduled session reset complete",
		"date_to_summarize", dateToSummarize,
	)

	return dateToSummarize, nil
}

// GetCurrentSessionPath returns the path to session-latest.log
func (sm *SessionManager) GetCurrentSessionPath() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.latestLogPath
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

// DEPRECATED METHODS - Kept for backward compatibility, will be removed in future versions

// GetCurrentSessionNumber is deprecated - session numbers are only assigned during archival
func (sm *SessionManager) GetCurrentSessionNumber() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.SessionNumber
}

// GetCurrentDate is deprecated - dates are determined from message timestamps during archival
func (sm *SessionManager) GetCurrentDate() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if sm.state.StartTime.IsZero() {
		return ""
	}
	return sm.state.StartTime.Format("2006-01-02")
}

// IncrementSession is deprecated - session numbering is handled automatically during archival
func (sm *SessionManager) IncrementSession() error {
	sm.logger.WarnWithComponent("SessionManager", "IncrementSession is deprecated and does nothing")
	return nil
}

// PerformSessionReset is deprecated - use PerformManualSessionReset or PerformScheduledSessionReset
func (sm *SessionManager) PerformSessionReset(triggerReason string) error {
	sm.logger.WarnWithComponent("SessionManager", "PerformSessionReset is deprecated",
		"trigger_reason", triggerReason,
		"use_instead", "PerformManualSessionReset or PerformScheduledSessionReset",
	)
	
	if triggerReason == "scheduled" || triggerReason == "scheduled_reset" {
		_, err := sm.PerformScheduledSessionReset()
		return err
	}
	
	return sm.PerformManualSessionReset()
}
