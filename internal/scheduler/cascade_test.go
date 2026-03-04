package scheduler

import (
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// MockMemorySessionManager implements MemorySessionManager for testing
type MockMemorySessionManager struct {
	performScheduledResetCalled bool
	performScheduledResetError  error
	dateToReturn                string
}

func (m *MockMemorySessionManager) PerformScheduledSessionReset() (string, error) {
	m.performScheduledResetCalled = true
	if m.performScheduledResetError != nil {
		return "", m.performScheduledResetError
	}
	return m.dateToReturn, nil
}

func TestExecuteDailyMaintenanceCascade_BasicFlow(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cron.json")

	// Create scheduler
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	scheduler := NewCronScheduler(configPath, logger)

	// Configure mock session manager
	mockSessionManager := &MockMemorySessionManager{
		dateToReturn: "2024-03-04",
	}
	scheduler.SetMemorySessionManager(mockSessionManager)

	// Execute cascade
	err := scheduler.ExecuteDailyMaintenanceCascade()
	if err != nil {
		t.Fatalf("Expected cascade to succeed, got error: %v", err)
	}

	// Verify history was logged
	history, err := scheduler.GetExecutionHistory("", 0)
	if err != nil {
		t.Fatalf("Failed to get execution history: %v", err)
	}

	// Should have one successful cascade entry
	if len(history) != 1 {
		t.Fatalf("Expected 1 history entry, got %d", len(history))
	}

	entry := history[0]
	if entry.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", entry.Status)
	}

	if entry.TaskType != "maintenance_cascade" {
		t.Errorf("Expected task_type 'maintenance_cascade', got '%s'", entry.TaskType)
	}

	// Verify operations completed
	expectedOps := []string{
		"daily_summary",
		"topic_extraction",
		"notes_cleanup",
		"session_reset",
	}

	if len(entry.OperationsCompleted) != len(expectedOps) {
		t.Errorf("Expected %d operations, got %d", len(expectedOps), len(entry.OperationsCompleted))
	}

	for i, op := range expectedOps {
		if i >= len(entry.OperationsCompleted) {
			t.Errorf("Missing operation: %s", op)
			continue
		}
		if entry.OperationsCompleted[i] != op {
			t.Errorf("Expected operation %d to be '%s', got '%s'", i, op, entry.OperationsCompleted[i])
		}
	}
}

func TestExecuteDailyMaintenanceCascade_MondayFlow(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cron.json")

	// Create scheduler
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	scheduler := NewCronScheduler(configPath, logger)

	// Configure mock session manager
	mockSessionManager := &MockMemorySessionManager{
		dateToReturn: "2024-03-04",
	}
	scheduler.SetMemorySessionManager(mockSessionManager)

	// Mock the day check to simulate Monday
	// Note: This test will only verify weekly ops if run on Monday
	// For a complete test, we'd need to inject time or make the methods testable

	// Execute cascade
	err := scheduler.ExecuteDailyMaintenanceCascade()
	if err != nil {
		t.Fatalf("Expected cascade to succeed, got error: %v", err)
	}

	// Verify history was logged
	history, err := scheduler.GetExecutionHistory("", 0)
	if err != nil {
		t.Fatalf("Failed to get execution history: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("Expected 1 history entry, got %d", len(history))
	}

	entry := history[0]
	if entry.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", entry.Status)
	}

	// Check if weekly operations were included (depends on actual day)
	isMonday := time.Now().Weekday() == time.Monday
	hasWeeklyOps := false
	for _, op := range entry.OperationsCompleted {
		if op == "weekly_reorganization" || op == "weekly_summary" {
			hasWeeklyOps = true
			break
		}
	}

	if isMonday && !hasWeeklyOps {
		t.Error("Expected weekly operations on Monday, but they were not executed")
	}

	if !isMonday && hasWeeklyOps {
		t.Error("Did not expect weekly operations on non-Monday, but they were executed")
	}
}

func TestShouldExecuteWeeklyOps(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cron.json")
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	scheduler := NewCronScheduler(configPath, logger)

	result := scheduler.ShouldExecuteWeeklyOps()

	// Should return true only on Monday
	expectedResult := time.Now().Weekday() == time.Monday

	if result != expectedResult {
		t.Errorf("Expected ShouldExecuteWeeklyOps to return %v, got %v", expectedResult, result)
	}
}

func TestShouldExecuteQuarterlyOps(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cron.json")
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	scheduler := NewCronScheduler(configPath, logger)

	result := scheduler.ShouldExecuteQuarterlyOps()

	// Should return true only on first Monday of quarter months (Jan, Apr, Jul, Oct)
	now := time.Now()
	month := now.Month()
	day := now.Day()
	weekday := now.Weekday()

	isQuarterMonth := month == time.January || month == time.April || month == time.July || month == time.October
	isFirstMonday := weekday == time.Monday && day <= 7

	expectedResult := isQuarterMonth && isFirstMonday

	if result != expectedResult {
		t.Errorf("Expected ShouldExecuteQuarterlyOps to return %v, got %v (month=%v, day=%d, weekday=%v)",
			expectedResult, result, month, day, weekday)
	}
}

func TestCascadeOperationSequencing(t *testing.T) {
	// This test verifies that operations are executed in the correct order
	// by checking the operations_completed array in the history

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cron.json")
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	scheduler := NewCronScheduler(configPath, logger)

	// Configure mock session manager
	mockSessionManager := &MockMemorySessionManager{
		dateToReturn: "2024-03-04",
	}
	scheduler.SetMemorySessionManager(mockSessionManager)

	// Execute cascade
	err := scheduler.ExecuteDailyMaintenanceCascade()
	if err != nil {
		t.Fatalf("Expected cascade to succeed, got error: %v", err)
	}

	// Get history
	history, err := scheduler.GetExecutionHistory("", 0)
	if err != nil {
		t.Fatalf("Failed to get execution history: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("Expected 1 history entry, got %d", len(history))
	}

	ops := history[0].OperationsCompleted

	// Verify order of operations
	// Daily operations always come first
	if len(ops) < 4 {
		t.Fatalf("Expected at least 4 operations, got %d", len(ops))
	}

	if ops[0] != "daily_summary" {
		t.Errorf("Expected first operation to be 'daily_summary', got '%s'", ops[0])
	}

	if ops[1] != "topic_extraction" {
		t.Errorf("Expected second operation to be 'topic_extraction', got '%s'", ops[1])
	}

	if ops[2] != "notes_cleanup" {
		t.Errorf("Expected third operation to be 'notes_cleanup', got '%s'", ops[2])
	}

	// Last operation should always be session_reset
	lastOp := ops[len(ops)-1]
	if lastOp != "session_reset" {
		t.Errorf("Expected last operation to be 'session_reset', got '%s'", lastOp)
	}

	// If weekly operations are present, they should come before session_reset
	hasWeeklyReorg := false
	hasWeeklySummary := false
	weeklyReorgIndex := -1
	weeklySummaryIndex := -1

	for i, op := range ops {
		if op == "weekly_reorganization" {
			hasWeeklyReorg = true
			weeklyReorgIndex = i
		}
		if op == "weekly_summary" {
			hasWeeklySummary = true
			weeklySummaryIndex = i
		}
	}

	if hasWeeklyReorg && hasWeeklySummary {
		if weeklyReorgIndex >= weeklySummaryIndex {
			t.Error("weekly_reorganization should come before weekly_summary")
		}
		if weeklySummaryIndex >= len(ops)-1 {
			t.Error("weekly_summary should come before session_reset")
		}
	}
}
