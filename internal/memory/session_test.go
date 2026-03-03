package memory

import (
	"fmt"
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

func TestPerformSessionReset(t *testing.T) {
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

	// Perform session reset
	err = sm.PerformSessionReset("scheduled_summarization")
	if err != nil {
		t.Fatalf("PerformSessionReset failed: %v", err)
	}

	// Verify session number was incremented
	if sm.GetCurrentSessionNumber() != 2 {
		t.Errorf("Expected session number to be 2 after reset, got %d", sm.GetCurrentSessionNumber())
	}

	// Verify session path was cleared
	if sm.GetCurrentSessionPath() != "" {
		t.Errorf("Expected session path to be cleared after reset, got %s", sm.GetCurrentSessionPath())
	}

	// Append to new session
	err = sm.AppendToSessionLog("User", "Second session message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed after reset: %v", err)
	}

	// Verify second session file was created
	currentDate := time.Now().Format("2006-01-02")
	session2Path := filepath.Join(tmpDir, "chat", currentDate, "session-002.log")
	if _, err := os.Stat(session2Path); os.IsNotExist(err) {
		t.Errorf("Second session log file was not created after reset: %s", session2Path)
	}

	// Verify content
	content, err := os.ReadFile(session2Path)
	if err != nil {
		t.Fatalf("Failed to read second session log: %v", err)
	}

	if !strings.Contains(string(content), "Second session message") {
		t.Errorf("Second session log does not contain expected message")
	}

	// Verify first session file still exists
	session1Path := filepath.Join(tmpDir, "chat", currentDate, "session-001.log")
	if _, err := os.Stat(session1Path); os.IsNotExist(err) {
		t.Errorf("First session log file should still exist: %s", session1Path)
	}
}

func TestPerformSessionReset_NewDay(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Set up a session for a previous date
	sm.currentDate = "2024-01-01"
	sm.currentSessionNum = 5
	sm.currentLogPath = "/some/old/path"

	// Perform session reset (should reset to session 1 for current date)
	err := sm.PerformSessionReset("scheduled_summarization")
	if err != nil {
		t.Fatalf("PerformSessionReset failed: %v", err)
	}

	// Verify session number was reset to 1 (not incremented from 5)
	if sm.GetCurrentSessionNumber() != 1 {
		t.Errorf("Expected session number to be 1 on new day, got %d", sm.GetCurrentSessionNumber())
	}

	// Verify current date was updated
	currentDate := time.Now().Format("2006-01-02")
	if sm.GetCurrentDate() != currentDate {
		t.Errorf("Expected current date to be %s, got %s", currentDate, sm.GetCurrentDate())
	}

	// Verify session path was cleared
	if sm.GetCurrentSessionPath() != "" {
		t.Errorf("Expected session path to be cleared, got %s", sm.GetCurrentSessionPath())
	}
}

func TestPerformSessionReset_WithTriggerReason(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Append to first session
	err := sm.AppendToSessionLog("User", "Test message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	// Test different trigger reasons
	testReasons := []string{
		"scheduled_summarization",
		"manual_reset",
		"token_limit_exceeded",
	}

	for _, reason := range testReasons {
		initialSessionNum := sm.GetCurrentSessionNumber()

		err = sm.PerformSessionReset(reason)
		if err != nil {
			t.Fatalf("PerformSessionReset failed with reason '%s': %v", reason, err)
		}

		// Verify session was incremented
		if sm.GetCurrentSessionNumber() != initialSessionNum+1 {
			t.Errorf("Expected session number to be %d after reset with reason '%s', got %d",
				initialSessionNum+1, reason, sm.GetCurrentSessionNumber())
		}

		// Append message to continue testing
		err = sm.AppendToSessionLog("User", "Message after reset")
		if err != nil {
			t.Fatalf("AppendToSessionLog failed after reset: %v", err)
		}
	}
}

// MockSummaryManager implements SummaryManagerInterface for testing
type MockSummaryManager struct {
	generateSessionSummaryCalled bool
	extractTopicsCalled          bool
	generateSessionSummaryError  error
	extractTopicsError           error
	summaryContent               string
}

func (m *MockSummaryManager) GenerateSessionSummary() error {
	m.generateSessionSummaryCalled = true
	return m.generateSessionSummaryError
}

func (m *MockSummaryManager) ExtractTopicsFromContent(content string) error {
	m.extractTopicsCalled = true
	return m.extractTopicsError
}

func TestPerformManualSessionReset(t *testing.T) {
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

	// Create mock summary manager
	mockSummary := &MockSummaryManager{
		summaryContent: "# Session Summary\n\nThis is a test summary.",
	}

	// Create the session summary file that GenerateSessionSummary would create
	currentDate := time.Now().Format("2006-01-02")
	dailyFolderPath := filepath.Join(tmpDir, "chat", currentDate)
	summaryPath := filepath.Join(dailyFolderPath, "session-001-summary.md")
	err = os.WriteFile(summaryPath, []byte(mockSummary.summaryContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock summary file: %v", err)
	}

	// Perform manual session reset
	err = sm.PerformManualSessionReset(mockSummary)
	if err != nil {
		t.Fatalf("PerformManualSessionReset failed: %v", err)
	}

	// Verify GenerateSessionSummary was called
	if !mockSummary.generateSessionSummaryCalled {
		t.Error("Expected GenerateSessionSummary to be called")
	}

	// Verify ExtractTopicsFromContent was called
	if !mockSummary.extractTopicsCalled {
		t.Error("Expected ExtractTopicsFromContent to be called")
	}

	// Verify session number was incremented
	if sm.GetCurrentSessionNumber() != 2 {
		t.Errorf("Expected session number to be 2 after manual reset, got %d", sm.GetCurrentSessionNumber())
	}

	// Verify session path was cleared
	if sm.GetCurrentSessionPath() != "" {
		t.Errorf("Expected session path to be cleared after manual reset, got %s", sm.GetCurrentSessionPath())
	}

	// Verify session summary file exists
	if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
		t.Errorf("Session summary file should exist: %s", summaryPath)
	}

	// Append to new session
	err = sm.AppendToSessionLog("User", "Second session message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed after manual reset: %v", err)
	}

	// Verify second session file was created
	session2Path := filepath.Join(tmpDir, "chat", currentDate, "session-002.log")
	if _, err := os.Stat(session2Path); os.IsNotExist(err) {
		t.Errorf("Second session log file was not created after manual reset: %s", session2Path)
	}
}

func TestPerformManualSessionReset_SummaryGenerationError(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Append to first session
	err := sm.AppendToSessionLog("User", "Test message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	// Create mock summary manager that returns an error
	mockSummary := &MockSummaryManager{
		generateSessionSummaryError: os.ErrNotExist,
	}

	// Perform manual session reset (should fail)
	err = sm.PerformManualSessionReset(mockSummary)
	if err == nil {
		t.Fatal("Expected PerformManualSessionReset to fail when summary generation fails")
	}

	// Verify error message
	if !strings.Contains(err.Error(), "failed to generate session summary") {
		t.Errorf("Expected error message to mention summary generation failure, got: %v", err)
	}

	// Verify session number was NOT incremented
	if sm.GetCurrentSessionNumber() != 1 {
		t.Errorf("Expected session number to remain 1 after failed reset, got %d", sm.GetCurrentSessionNumber())
	}
}

func TestPerformManualSessionReset_TopicExtractionError(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Append to first session
	err := sm.AppendToSessionLog("User", "Test message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	// Create mock summary manager that fails topic extraction
	mockSummary := &MockSummaryManager{
		summaryContent:     "# Session Summary\n\nTest summary.",
		extractTopicsError: os.ErrPermission,
	}

	// Create the session summary file
	currentDate := time.Now().Format("2006-01-02")
	dailyFolderPath := filepath.Join(tmpDir, "chat", currentDate)
	summaryPath := filepath.Join(dailyFolderPath, "session-001-summary.md")
	err = os.WriteFile(summaryPath, []byte(mockSummary.summaryContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock summary file: %v", err)
	}

	// Perform manual session reset (should succeed despite topic extraction error)
	err = sm.PerformManualSessionReset(mockSummary)
	if err != nil {
		t.Fatalf("PerformManualSessionReset should succeed even if topic extraction fails: %v", err)
	}

	// Verify both methods were called
	if !mockSummary.generateSessionSummaryCalled {
		t.Error("Expected GenerateSessionSummary to be called")
	}
	if !mockSummary.extractTopicsCalled {
		t.Error("Expected ExtractTopicsFromContent to be called")
	}

	// Verify session number was incremented (reset should continue despite topic extraction error)
	if sm.GetCurrentSessionNumber() != 2 {
		t.Errorf("Expected session number to be 2 after manual reset, got %d", sm.GetCurrentSessionNumber())
	}
}

func TestPerformManualSessionReset_MultipleSessions(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	currentDate := time.Now().Format("2006-01-02")
	dailyFolderPath := filepath.Join(tmpDir, "chat", currentDate)

	// Perform multiple manual resets
	for i := 1; i <= 3; i++ {
		// Append message to session
		err := sm.AppendToSessionLog("User", "Message in session")
		if err != nil {
			t.Fatalf("AppendToSessionLog failed: %v", err)
		}

		// Create mock summary manager
		mockSummary := &MockSummaryManager{
			summaryContent: "# Session Summary\n\nSummary for session.",
		}

		// Create the session summary file
		summaryPath := filepath.Join(dailyFolderPath, fmt.Sprintf("session-%03d-summary.md", i))
		err = os.WriteFile(summaryPath, []byte(mockSummary.summaryContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create mock summary file: %v", err)
		}

		// Perform manual session reset
		err = sm.PerformManualSessionReset(mockSummary)
		if err != nil {
			t.Fatalf("PerformManualSessionReset failed on iteration %d: %v", i, err)
		}

		// Verify session number was incremented
		expectedSessionNum := i + 1
		if sm.GetCurrentSessionNumber() != expectedSessionNum {
			t.Errorf("Expected session number to be %d after reset %d, got %d",
				expectedSessionNum, i, sm.GetCurrentSessionNumber())
		}

		// Verify summary file exists
		if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
			t.Errorf("Session summary file should exist: %s", summaryPath)
		}
	}

	// Verify all session log files exist
	for i := 1; i <= 3; i++ {
		sessionPath := filepath.Join(dailyFolderPath, fmt.Sprintf("session-%03d.log", i))
		if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
			t.Errorf("Session log file should exist: %s", sessionPath)
		}
	}

	// Verify all summary files exist
	for i := 1; i <= 3; i++ {
		summaryPath := filepath.Join(dailyFolderPath, fmt.Sprintf("session-%03d-summary.md", i))
		if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
			t.Errorf("Session summary file should exist: %s", summaryPath)
		}
	}
}

func TestSessionStateTracking_InitialState(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Verify initial state
	state := sm.GetSessionState()
	if state.IsActive {
		t.Error("Expected initial session to be inactive")
	}
	if state.SessionNumber != 0 {
		t.Errorf("Expected initial session number to be 0, got %d", state.SessionNumber)
	}
	if state.MessageCount != 0 {
		t.Errorf("Expected initial message count to be 0, got %d", state.MessageCount)
	}
	if state.TokenCount != 0 {
		t.Errorf("Expected initial token count to be 0, got %d", state.TokenCount)
	}
	if state.HasBeenReset {
		t.Error("Expected HasBeenReset to be false initially")
	}
	if state.LastResetType != "" {
		t.Errorf("Expected LastResetType to be empty initially, got %s", state.LastResetType)
	}
}

func TestSessionStateTracking_AfterFirstMessage(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Append first message
	err := sm.AppendToSessionLog("User", "Hello, bot!")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	// Verify session state
	state := sm.GetSessionState()
	if !state.IsActive {
		t.Error("Expected session to be active after first message")
	}
	if state.SessionNumber != 1 {
		t.Errorf("Expected session number to be 1, got %d", state.SessionNumber)
	}
	if state.MessageCount != 1 {
		t.Errorf("Expected message count to be 1, got %d", state.MessageCount)
	}
	if state.TokenCount == 0 {
		t.Error("Expected token count to be greater than 0")
	}
	if state.StartTime.IsZero() {
		t.Error("Expected StartTime to be set")
	}
	if state.LastActivity.IsZero() {
		t.Error("Expected LastActivity to be set")
	}
}

func TestSessionStateTracking_MessageCount(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Append multiple messages
	messages := []string{
		"First message",
		"Second message",
		"Third message",
		"Fourth message",
		"Fifth message",
	}

	for i, msg := range messages {
		err := sm.AppendToSessionLog("User", msg)
		if err != nil {
			t.Fatalf("AppendToSessionLog failed: %v", err)
		}

		// Verify message count increments
		count := sm.GetSessionMessageCount()
		expectedCount := i + 1
		if count != expectedCount {
			t.Errorf("Expected message count to be %d, got %d", expectedCount, count)
		}
	}

	// Verify final state
	state := sm.GetSessionState()
	if state.MessageCount != len(messages) {
		t.Errorf("Expected final message count to be %d, got %d", len(messages), state.MessageCount)
	}
}

func TestSessionStateTracking_TokenCount(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Append message with known length
	message := "This is a test message with exactly 50 characters!"
	err := sm.AppendToSessionLog("User", message)
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	// Verify token count (approximately len/4)
	tokenCount := sm.GetSessionTokenCount()
	expectedTokens := len(message) / 4
	// Allow for rounding differences
	if tokenCount < expectedTokens-1 || tokenCount > expectedTokens+1 {
		t.Errorf("Expected token count to be approximately %d, got %d", expectedTokens, tokenCount)
	}

	// Append another message
	message2 := "Another message with 40 characters here"
	err = sm.AppendToSessionLog("Assistant", message2)
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	// Verify token count increased
	newTokenCount := sm.GetSessionTokenCount()
	expectedNewTokens := (len(message) + len(message2)) / 4
	// Allow for rounding differences
	if newTokenCount < expectedNewTokens-1 || newTokenCount > expectedNewTokens+1 {
		t.Errorf("Expected token count to be approximately %d, got %d", expectedNewTokens, newTokenCount)
	}
}

func TestSessionStateTracking_LastActivity(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Append first message
	before := time.Now()
	err := sm.AppendToSessionLog("User", "First message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}
	after := time.Now()

	// Verify LastActivity is within expected range
	lastActivity := sm.GetLastActivity()
	if lastActivity.Before(before) || lastActivity.After(after) {
		t.Errorf("Expected LastActivity to be between %v and %v, got %v", before, after, lastActivity)
	}

	// Wait a bit and append another message
	time.Sleep(10 * time.Millisecond)
	before2 := time.Now()
	err = sm.AppendToSessionLog("Assistant", "Second message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}
	after2 := time.Now()

	// Verify LastActivity was updated
	newLastActivity := sm.GetLastActivity()
	if newLastActivity.Before(before2) || newLastActivity.After(after2) {
		t.Errorf("Expected LastActivity to be between %v and %v, got %v", before2, after2, newLastActivity)
	}
	if !newLastActivity.After(lastActivity) {
		t.Error("Expected LastActivity to be updated to a later time")
	}
}

func TestSessionStateTracking_AfterReset(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Append messages to first session
	err := sm.AppendToSessionLog("User", "First session message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}
	err = sm.AppendToSessionLog("Assistant", "Response")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	// Verify state before reset
	stateBefore := sm.GetSessionState()
	if !stateBefore.IsActive {
		t.Error("Expected session to be active before reset")
	}
	if stateBefore.MessageCount != 2 {
		t.Errorf("Expected message count to be 2 before reset, got %d", stateBefore.MessageCount)
	}

	// Perform session reset
	err = sm.PerformSessionReset("scheduled")
	if err != nil {
		t.Fatalf("PerformSessionReset failed: %v", err)
	}

	// Verify state after reset
	stateAfter := sm.GetSessionState()
	if stateAfter.IsActive {
		t.Error("Expected session to be inactive after reset")
	}
	if stateAfter.SessionNumber != 2 {
		t.Errorf("Expected session number to be 2 after reset, got %d", stateAfter.SessionNumber)
	}
	if stateAfter.MessageCount != 0 {
		t.Errorf("Expected message count to be 0 after reset, got %d", stateAfter.MessageCount)
	}
	if stateAfter.TokenCount != 0 {
		t.Errorf("Expected token count to be 0 after reset, got %d", stateAfter.TokenCount)
	}
	if !stateAfter.HasBeenReset {
		t.Error("Expected HasBeenReset to be true after reset")
	}
	if stateAfter.LastResetType != "scheduled" {
		t.Errorf("Expected LastResetType to be 'scheduled', got %s", stateAfter.LastResetType)
	}
	if stateAfter.LastResetTime.IsZero() {
		t.Error("Expected LastResetTime to be set")
	}
}

func TestSessionStateTracking_ResetTypes(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	resetTypes := []string{
		"scheduled",
		"manual",
		"token_based",
	}

	for _, resetType := range resetTypes {
		// Append a message
		err := sm.AppendToSessionLog("User", "Test message")
		if err != nil {
			t.Fatalf("AppendToSessionLog failed: %v", err)
		}

		// Perform reset with specific type
		err = sm.PerformSessionReset(resetType)
		if err != nil {
			t.Fatalf("PerformSessionReset failed: %v", err)
		}

		// Verify reset type was recorded
		_, lastResetType := sm.GetLastResetInfo()
		if lastResetType != resetType {
			t.Errorf("Expected LastResetType to be '%s', got '%s'", resetType, lastResetType)
		}

		// Verify HasBeenReset is true
		if !sm.HasBeenReset() {
			t.Error("Expected HasBeenReset to be true")
		}
	}
}

func TestSessionStateTracking_StartTime(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Verify StartTime is zero initially
	startTime := sm.GetSessionStartTime()
	if !startTime.IsZero() {
		t.Error("Expected StartTime to be zero initially")
	}

	// Append first message
	before := time.Now()
	err := sm.AppendToSessionLog("User", "First message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}
	after := time.Now()

	// Verify StartTime is set
	startTime = sm.GetSessionStartTime()
	if startTime.IsZero() {
		t.Error("Expected StartTime to be set after first message")
	}
	if startTime.Before(before) || startTime.After(after) {
		t.Errorf("Expected StartTime to be between %v and %v, got %v", before, after, startTime)
	}

	// Append another message
	time.Sleep(10 * time.Millisecond)
	err = sm.AppendToSessionLog("Assistant", "Second message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	// Verify StartTime didn't change
	newStartTime := sm.GetSessionStartTime()
	if !newStartTime.Equal(startTime) {
		t.Error("Expected StartTime to remain unchanged after subsequent messages")
	}
}

func TestSessionStateTracking_NewDayReset(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Append message
	err := sm.AppendToSessionLog("User", "Message on day 1")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	// Verify state
	state1 := sm.GetSessionState()
	if state1.SessionNumber != 1 {
		t.Errorf("Expected session number to be 1, got %d", state1.SessionNumber)
	}
	if state1.MessageCount != 1 {
		t.Errorf("Expected message count to be 1, got %d", state1.MessageCount)
	}

	// Simulate new day by changing the internal date
	sm.mu.Lock()
	sm.currentDate = "2024-01-01"
	sm.mu.Unlock()

	// Append message (should reset to session 1 for new day)
	err = sm.AppendToSessionLog("User", "Message on day 2")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	// Verify state was reset for new day
	state2 := sm.GetSessionState()
	if state2.SessionNumber != 1 {
		t.Errorf("Expected session number to be 1 on new day, got %d", state2.SessionNumber)
	}
	if state2.MessageCount != 1 {
		t.Errorf("Expected message count to be 1 on new day, got %d", state2.MessageCount)
	}
}

func TestSessionStateTracking_IsSessionActive(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Initially inactive
	if sm.IsSessionActive() {
		t.Error("Expected session to be inactive initially")
	}

	// Active after first message
	err := sm.AppendToSessionLog("User", "Test message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}
	if !sm.IsSessionActive() {
		t.Error("Expected session to be active after first message")
	}

	// Inactive after reset
	err = sm.PerformSessionReset("test")
	if err != nil {
		t.Fatalf("PerformSessionReset failed: %v", err)
	}
	if sm.IsSessionActive() {
		t.Error("Expected session to be inactive after reset")
	}

	// Active again after new message
	err = sm.AppendToSessionLog("User", "New session message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}
	if !sm.IsSessionActive() {
		t.Error("Expected session to be active after new message")
	}
}

func TestSessionStateTracking_GetLastResetInfo(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Initially no reset
	resetTime, resetType := sm.GetLastResetInfo()
	if !resetTime.IsZero() {
		t.Error("Expected LastResetTime to be zero initially")
	}
	if resetType != "" {
		t.Errorf("Expected LastResetType to be empty initially, got %s", resetType)
	}

	// Append message and reset
	err := sm.AppendToSessionLog("User", "Test message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	before := time.Now()
	err = sm.PerformSessionReset("manual")
	if err != nil {
		t.Fatalf("PerformSessionReset failed: %v", err)
	}
	after := time.Now()

	// Verify reset info
	resetTime, resetType = sm.GetLastResetInfo()
	if resetTime.IsZero() {
		t.Error("Expected LastResetTime to be set after reset")
	}
	if resetTime.Before(before) || resetTime.After(after) {
		t.Errorf("Expected LastResetTime to be between %v and %v, got %v", before, after, resetTime)
	}
	if resetType != "manual" {
		t.Errorf("Expected LastResetType to be 'manual', got %s", resetType)
	}
}

func TestSessionStateTracking_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Append initial message
	err := sm.AppendToSessionLog("User", "Initial message")
	if err != nil {
		t.Fatalf("AppendToSessionLog failed: %v", err)
	}

	// Concurrently read session state
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			state := sm.GetSessionState()
			_ = state.IsActive
			_ = sm.IsSessionActive()
			_ = sm.GetSessionMessageCount()
			_ = sm.GetSessionTokenCount()
			_ = sm.GetSessionStartTime()
			_ = sm.GetLastActivity()
			_ = sm.HasBeenReset()
			_, _ = sm.GetLastResetInfo()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify state is still consistent
	state := sm.GetSessionState()
	if !state.IsActive {
		t.Error("Expected session to still be active")
	}
	if state.MessageCount != 1 {
		t.Errorf("Expected message count to be 1, got %d", state.MessageCount)
	}
}
