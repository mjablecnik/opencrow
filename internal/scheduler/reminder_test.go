package scheduler

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// MockTelegramSender implements TelegramSender for testing
type MockTelegramSender struct {
	sentMessages []struct {
		chatID int64
		text   string
	}
	shouldFail bool
}

func (m *MockTelegramSender) SendMessage(chatID int64, text string) error {
	if m.shouldFail {
		return fmt.Errorf("mock telegram send failure")
	}
	m.sentMessages = append(m.sentMessages, struct {
		chatID int64
		text   string
	}{chatID, text})
	return nil
}

// MockSessionLogger implements SessionLogger for testing
type MockSessionLogger struct {
	loggedMessages []struct {
		role    string
		content string
	}
	shouldFail bool
}

func (m *MockSessionLogger) AppendToSessionLog(role, content string) error {
	if m.shouldFail {
		return fmt.Errorf("mock session log failure")
	}
	m.loggedMessages = append(m.loggedMessages, struct {
		role    string
		content string
	}{role, content})
	return nil
}

func TestReminderExecution(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "cron.json")

	// Create scheduler
	scheduler := NewCronScheduler(configPath, nil)

	// Create mock dependencies
	mockTelegram := &MockTelegramSender{}
	mockSession := &MockSessionLogger{}

	// Set dependencies
	scheduler.SetTelegramSender(mockTelegram)
	scheduler.SetSessionLogger(mockSession)

	// Create a reminder job
	job := &JobInfo{
		Name:     "test_reminder",
		TaskType: "reminder",
		Message:  "Test reminder message",
		ChatID:   123456789,
		Enabled:  true,
		Status:   JobStatusEnabled,
	}

	// Execute the reminder
	err := scheduler.executeReminder(job)
	if err != nil {
		t.Fatalf("executeReminder failed: %v", err)
	}

	// Verify Telegram message was sent
	if len(mockTelegram.sentMessages) != 1 {
		t.Fatalf("Expected 1 Telegram message, got %d", len(mockTelegram.sentMessages))
	}

	sentMsg := mockTelegram.sentMessages[0]
	if sentMsg.chatID != 123456789 {
		t.Errorf("Expected chatID 123456789, got %d", sentMsg.chatID)
	}
	if sentMsg.text != "Test reminder message" {
		t.Errorf("Expected message 'Test reminder message', got '%s'", sentMsg.text)
	}

	// Verify session log was written
	if len(mockSession.loggedMessages) != 1 {
		t.Fatalf("Expected 1 session log entry, got %d", len(mockSession.loggedMessages))
	}

	loggedMsg := mockSession.loggedMessages[0]
	if loggedMsg.role != "Assistant (Cron)" {
		t.Errorf("Expected role 'Assistant (Cron)', got '%s'", loggedMsg.role)
	}
	if loggedMsg.content != "Test reminder message" {
		t.Errorf("Expected content 'Test reminder message', got '%s'", loggedMsg.content)
	}
}

func TestReminderExecution_MissingDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "cron.json")

	// Create scheduler without setting dependencies
	scheduler := NewCronScheduler(configPath, nil)

	job := &JobInfo{
		Name:     "test_reminder",
		TaskType: "reminder",
		Message:  "Test message",
		ChatID:   123456789,
	}

	// Should fail without telegram sender
	err := scheduler.executeReminder(job)
	if err == nil {
		t.Fatal("Expected error when telegram sender not configured")
	}
	if err.Error() != "telegram sender not configured" {
		t.Errorf("Expected 'telegram sender not configured' error, got: %v", err)
	}

	// Set telegram sender but not session logger
	scheduler.SetTelegramSender(&MockTelegramSender{})

	err = scheduler.executeReminder(job)
	if err == nil {
		t.Fatal("Expected error when session logger not configured")
	}
	if err.Error() != "session logger not configured" {
		t.Errorf("Expected 'session logger not configured' error, got: %v", err)
	}
}

func TestReminderExecution_MissingFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "cron.json")

	scheduler := NewCronScheduler(configPath, nil)
	scheduler.SetTelegramSender(&MockTelegramSender{})
	scheduler.SetSessionLogger(&MockSessionLogger{})

	// Test missing message
	job := &JobInfo{
		Name:     "test_reminder",
		TaskType: "reminder",
		Message:  "",
		ChatID:   123456789,
	}

	err := scheduler.executeReminder(job)
	if err == nil {
		t.Fatal("Expected error when message is missing")
	}
	if err.Error() != "reminder job missing message" {
		t.Errorf("Expected 'reminder job missing message' error, got: %v", err)
	}

	// Test missing chat_id
	job.Message = "Test message"
	job.ChatID = 0

	err = scheduler.executeReminder(job)
	if err == nil {
		t.Fatal("Expected error when chat_id is missing")
	}
	if err.Error() != "reminder job missing chat_id" {
		t.Errorf("Expected 'reminder job missing chat_id' error, got: %v", err)
	}
}

func TestReminderExecution_TelegramFailure(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "cron.json")

	scheduler := NewCronScheduler(configPath, nil)
	
	mockTelegram := &MockTelegramSender{shouldFail: true}
	mockSession := &MockSessionLogger{}
	
	scheduler.SetTelegramSender(mockTelegram)
	scheduler.SetSessionLogger(mockSession)

	job := &JobInfo{
		Name:     "test_reminder",
		TaskType: "reminder",
		Message:  "Test message",
		ChatID:   123456789,
	}

	err := scheduler.executeReminder(job)
	if err == nil {
		t.Fatal("Expected error when Telegram send fails")
	}

	// Session log should not be written if Telegram send fails
	if len(mockSession.loggedMessages) != 0 {
		t.Errorf("Expected 0 session log entries when Telegram fails, got %d", len(mockSession.loggedMessages))
	}
}

func TestReminderExecution_SessionLogFailure(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "cron.json")

	scheduler := NewCronScheduler(configPath, nil)
	
	mockTelegram := &MockTelegramSender{}
	mockSession := &MockSessionLogger{shouldFail: true}
	
	scheduler.SetTelegramSender(mockTelegram)
	scheduler.SetSessionLogger(mockSession)

	job := &JobInfo{
		Name:     "test_reminder",
		TaskType: "reminder",
		Message:  "Test message",
		ChatID:   123456789,
	}

	// Should succeed even if session log fails (non-critical)
	err := scheduler.executeReminder(job)
	if err != nil {
		t.Fatalf("executeReminder should succeed even if session log fails: %v", err)
	}

	// Telegram message should still be sent
	if len(mockTelegram.sentMessages) != 1 {
		t.Errorf("Expected 1 Telegram message, got %d", len(mockTelegram.sentMessages))
	}
}

func TestOneTimeJobExecution(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "cron.json")

	// Create scheduler
	scheduler := NewCronScheduler(configPath, nil)

	// Create mock dependencies
	mockTelegram := &MockTelegramSender{}
	mockSession := &MockSessionLogger{}

	scheduler.SetTelegramSender(mockTelegram)
	scheduler.SetSessionLogger(mockSession)

	// Create a one-time reminder job that executes in 1 second (longer delay to avoid race)
	executeAt := time.Now().Add(1 * time.Second)
	
	// Manually create the config file with a one-time job
	configContent := fmt.Sprintf(`{
		"version": "1.0",
		"jobs": [
			{
				"name": "one_time_reminder",
				"execute_at": "%s",
				"task_type": "reminder",
				"enabled": true,
				"message": "One-time reminder",
				"chat_id": 123456789,
				"description": "Test one-time reminder"
			}
		]
	}`, executeAt.Format(time.RFC3339))

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Start scheduler
	if err := scheduler.Start(); err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Wait for job to execute (wait longer than execution time)
	time.Sleep(1500 * time.Millisecond)

	// Verify the job was executed
	if len(mockTelegram.sentMessages) != 1 {
		t.Fatalf("Expected 1 Telegram message, got %d", len(mockTelegram.sentMessages))
	}

	// Verify the job was removed after execution
	jobs, err := scheduler.ListJobs()
	if err != nil {
		t.Fatalf("Failed to list jobs: %v", err)
	}

	if len(jobs) != 0 {
		t.Errorf("Expected 0 jobs after one-time execution, got %d", len(jobs))
	}

	// Stop scheduler
	if err := scheduler.Stop(); err != nil {
		t.Fatalf("Failed to stop scheduler: %v", err)
	}
}

func TestOneTimeJobScheduling_PastExecutionTime(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "cron.json")

	scheduler := NewCronScheduler(configPath, nil)

	// Create a one-time job with past execution time
	pastTime := time.Now().Add(-1 * time.Hour)
	
	configContent := fmt.Sprintf(`{
		"version": "1.0",
		"jobs": [
			{
				"name": "past_reminder",
				"execute_at": "%s",
				"task_type": "reminder",
				"enabled": true,
				"message": "Past reminder",
				"chat_id": 123456789
			}
		]
	}`, pastTime.Format(time.RFC3339))

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Start scheduler - should skip the past job
	if err := scheduler.Start(); err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Job should still be in the list (not removed until execution)
	jobs, err := scheduler.ListJobs()
	if err != nil {
		t.Fatalf("Failed to list jobs: %v", err)
	}

	if len(jobs) != 1 {
		t.Errorf("Expected 1 job in list, got %d", len(jobs))
	}

	if err := scheduler.Stop(); err != nil {
		t.Fatalf("Failed to stop scheduler: %v", err)
	}
}
