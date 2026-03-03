package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager("/tmp/test-memory")
	if sm == nil {
		t.Fatal("NewSessionManager returned nil")
	}
	if sm.memoryBasePath != "/tmp/test-memory" {
		t.Errorf("Expected memoryBasePath to be /tmp/test-memory, got %s", sm.memoryBasePath)
	}
}

func TestAppendToSessionLog(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Append a message
	err := sm.AppendToSessionLog("User", "Hello, bot!")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	// Verify daily folder was created
	currentDate := time.Now().Format("2006-01-02")
	dailyFolderPath := filepath.Join(tmpDir, "chat", currentDate)
	if _, err := os.Stat(dailyFolderPath); os.IsNotExist(err) {
		t.Errorf("Daily folder was not created: %s", dailyFolderPath)
	}

	// Verify session log file was created
	sessionLogPath := filepath.Join(dailyFolderPath, "session-001.log")
	if _, err := os.Stat(sessionLogPath); os.IsNotExist(err) {
		t.Errorf("Session log file was not created: %s", sessionLogPath)
	}

	// Read the log file and verify content
	content, err := os.ReadFile(sessionLogPath)
	if err != nil {
		t.Fatalf("Failed to read session log: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "User: Hello, bot!") {
		t.Errorf("Log content does not contain expected message. Got: %s", logContent)
	}

	// Verify timestamp format [YYYY-MM-DD HH:MM:SS]
	if !strings.Contains(logContent, "["+currentDate) {
		t.Errorf("Log content does not contain expected timestamp format. Got: %s", logContent)
	}
}

func TestAppendToSessionLog_MultipleMessages(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Append multiple messages
	messages := []struct {
		role    string
		content string
	}{
		{"User", "First message"},
		{"Assistant", "First response"},
		{"User", "Second message"},
		{"Assistant", "Second response"},
	}

	for _, msg := range messages {
		err := sm.AppendToSessionLog(msg.role, msg.content)
		if err != nil {
			t.Fatalf("AppendToSessionLog failed: %v", err)
		}
	}

	// Read the log file
	currentDate := time.Now().Format("2006-01-02")
	sessionLogPath := filepath.Join(tmpDir, "chat", currentDate, "session-001.log")
	content, err := os.ReadFile(sessionLogPath)
	if err != nil {
		t.Fatalf("Failed to read session log: %v", err)
	}

	logContent := string(content)

	// Verify all messages are present
	for _, msg := range messages {
		expected := msg.role + ": " + msg.content
		if !strings.Contains(logContent, expected) {
			t.Errorf("Log content does not contain expected message: %s", expected)
		}
	}
}

func TestGetCurrentSessionNumber(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Initial session number should be 0 (not set yet)
	if sm.GetCurrentSessionNumber() != 0 {
		t.Errorf("Expected initial session number to be 0, got %d", sm.GetCurrentSessionNumber())
	}

	// After appending a message, session number should be 1
	err := sm.AppendToSessionLog("User", "Test message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	if sm.GetCurrentSessionNumber() != 1 {
		t.Errorf("Expected session number to be 1, got %d", sm.GetCurrentSessionNumber())
	}
}

func TestGetCurrentSessionPath(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Initial path should be empty
	if sm.GetCurrentSessionPath() != "" {
		t.Errorf("Expected initial session path to be empty, got %s", sm.GetCurrentSessionPath())
	}

	// After appending a message, path should be set
	err := sm.AppendToSessionLog("User", "Test message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	path := sm.GetCurrentSessionPath()
	if path == "" {
		t.Error("Expected session path to be set after appending message")
	}

	// Verify path format
	currentDate := time.Now().Format("2006-01-02")
	expectedPath := filepath.Join(tmpDir, "chat", currentDate, "session-001.log")
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}
}

func TestIncrementSession(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Append to first session
	err := sm.AppendToSessionLog("User", "First session message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	if sm.GetCurrentSessionNumber() != 1 {
		t.Errorf("Expected session number to be 1, got %d", sm.GetCurrentSessionNumber())
	}

	// Increment session
	err = sm.IncrementSession()
	if err != nil {
		t.Fatalf("IncrementSession failed: %v", err)
	}

	if sm.GetCurrentSessionNumber() != 2 {
		t.Errorf("Expected session number to be 2 after increment, got %d", sm.GetCurrentSessionNumber())
	}

	// Append to second session
	err = sm.AppendToSessionLog("User", "Second session message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	// Verify second session file was created
	currentDate := time.Now().Format("2006-01-02")
	session2Path := filepath.Join(tmpDir, "chat", currentDate, "session-002.log")
	if _, err := os.Stat(session2Path); os.IsNotExist(err) {
		t.Errorf("Second session log file was not created: %s", session2Path)
	}

	// Verify content
	content, err := os.ReadFile(session2Path)
	if err != nil {
		t.Fatalf("Failed to read second session log: %v", err)
	}

	if !strings.Contains(string(content), "Second session message") {
		t.Errorf("Second session log does not contain expected message")
	}
}

func TestSessionNumberResetOnNewDay(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Set up a session for a previous date
	sm.currentDate = "2024-01-01"
	sm.currentSessionNum = 5

	// Append a message (should reset to session 1 for current date)
	err := sm.AppendToSessionLog("User", "New day message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	// Verify session number was reset to 1
	if sm.GetCurrentSessionNumber() != 1 {
		t.Errorf("Expected session number to be 1 on new day, got %d", sm.GetCurrentSessionNumber())
	}

	// Verify current date was updated
	currentDate := time.Now().Format("2006-01-02")
	if sm.GetCurrentDate() != currentDate {
		t.Errorf("Expected current date to be %s, got %s", currentDate, sm.GetCurrentDate())
	}
}

func TestSessionLogFormat(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Append a message
	err := sm.AppendToSessionLog("User", "Test message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	// Read the log file
	currentDate := time.Now().Format("2006-01-02")
	sessionLogPath := filepath.Join(tmpDir, "chat", currentDate, "session-001.log")
	content, err := os.ReadFile(sessionLogPath)
	if err != nil {
		t.Fatalf("Failed to read session log: %v", err)
	}

	logContent := string(content)

	// Verify format: [YYYY-MM-DD HH:MM:SS] Role: Content
	// Should have timestamp in brackets
	if !strings.HasPrefix(logContent, "[") {
		t.Error("Log entry should start with timestamp in brackets")
	}

	// Should have role prefix
	if !strings.Contains(logContent, "User:") {
		t.Error("Log entry should contain role prefix")
	}

	// Should have blank line after entry
	if !strings.HasSuffix(logContent, "\n\n") {
		t.Error("Log entry should end with blank line")
	}
}

func TestConcurrentAppends(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Append messages concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			err := sm.AppendToSessionLog("User", "Concurrent message")
			if err != nil {
				t.Errorf("Concurrent append failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify log file exists and has content
	currentDate := time.Now().Format("2006-01-02")
	sessionLogPath := filepath.Join(tmpDir, "chat", currentDate, "session-001.log")
	content, err := os.ReadFile(sessionLogPath)
	if err != nil {
		t.Fatalf("Failed to read session log: %v", err)
	}

	// Count occurrences of "Concurrent message"
	count := strings.Count(string(content), "Concurrent message")
	if count != 10 {
		t.Errorf("Expected 10 concurrent messages, got %d", count)
	}
}
