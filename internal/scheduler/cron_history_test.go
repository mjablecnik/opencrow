package scheduler

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHistoryTracking(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cron.json")
	
	// Create scheduler
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	scheduler := NewCronScheduler(configPath, logger)
	
	// Verify history path is set correctly
	expectedHistoryPath := filepath.Join(tempDir, "cron_history.json")
	if scheduler.historyPath != expectedHistoryPath {
		t.Errorf("Expected history path %s, got %s", expectedHistoryPath, scheduler.historyPath)
	}
	
	// Test appending history entry
	entry := HistoryEntry{
		Timestamp: time.Now(),
		JobName:   "test_job",
		TaskType:  "reminder",
		Status:    "success",
		MessageSent: "Test message",
		ChatID:    123456,
	}
	
	scheduler.mu.Lock()
	err := scheduler.appendHistoryEntry(entry)
	scheduler.mu.Unlock()
	
	if err != nil {
		t.Fatalf("Failed to append history entry: %v", err)
	}
	
	// Verify history file was created
	if _, err := os.Stat(scheduler.historyPath); os.IsNotExist(err) {
		t.Fatal("History file was not created")
	}
	
	// Read and verify history file content
	data, err := os.ReadFile(scheduler.historyPath)
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}
	
	var historyData struct {
		Version    string         `json:"version"`
		MaxEntries int            `json:"max_entries"`
		Entries    []HistoryEntry `json:"entries"`
	}
	
	if err := json.Unmarshal(data, &historyData); err != nil {
		t.Fatalf("Failed to parse history file: %v", err)
	}
	
	if historyData.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", historyData.Version)
	}
	
	if historyData.MaxEntries != 1000 {
		t.Errorf("Expected max entries 1000, got %d", historyData.MaxEntries)
	}
	
	if len(historyData.Entries) != 1 {
		t.Fatalf("Expected 1 history entry, got %d", len(historyData.Entries))
	}
	
	savedEntry := historyData.Entries[0]
	if savedEntry.JobName != "test_job" {
		t.Errorf("Expected job name 'test_job', got '%s'", savedEntry.JobName)
	}
	
	if savedEntry.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", savedEntry.Status)
	}
}

func TestHistorySizeLimiting(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cron.json")
	
	// Create scheduler
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	scheduler := NewCronScheduler(configPath, logger)
	
	// Add more than 1000 entries
	scheduler.mu.Lock()
	for i := 0; i < 1100; i++ {
		entry := HistoryEntry{
			Timestamp: time.Now(),
			JobName:   "test_job",
			TaskType:  "reminder",
			Status:    "success",
		}
		scheduler.history = append(scheduler.history, entry)
	}
	
	// Save history (should trigger size limiting)
	err := scheduler.saveHistoryInternal()
	scheduler.mu.Unlock()
	
	if err != nil {
		t.Fatalf("Failed to save history: %v", err)
	}
	
	// Verify history was limited to 1000 entries
	scheduler.mu.RLock()
	historyLen := len(scheduler.history)
	scheduler.mu.RUnlock()
	
	if historyLen != 1000 {
		t.Errorf("Expected history to be limited to 1000 entries, got %d", historyLen)
	}
}

func TestDeletionLogging(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cron.json")
	
	// Create scheduler
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	scheduler := NewCronScheduler(configPath, logger)
	
	// Log a deletion event
	scheduler.mu.Lock()
	scheduler.logDeletion("test_job", "reminder", "expired")
	scheduler.mu.Unlock()
	
	// Verify deletion was logged
	scheduler.mu.RLock()
	if len(scheduler.history) != 1 {
		t.Fatalf("Expected 1 history entry, got %d", len(scheduler.history))
	}
	
	entry := scheduler.history[0]
	scheduler.mu.RUnlock()
	
	if entry.Status != "deleted" {
		t.Errorf("Expected status 'deleted', got '%s'", entry.Status)
	}
	
	if !entry.Deleted {
		t.Error("Expected Deleted flag to be true")
	}
	
	if entry.DeletionReason != "expired" {
		t.Errorf("Expected deletion reason 'expired', got '%s'", entry.DeletionReason)
	}
}

func TestGetExecutionHistory(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cron.json")
	
	// Create scheduler
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	scheduler := NewCronScheduler(configPath, logger)
	
	// Add multiple history entries
	scheduler.mu.Lock()
	for i := 0; i < 10; i++ {
		entry := HistoryEntry{
			Timestamp: time.Now(),
			JobName:   "job_a",
			TaskType:  "reminder",
			Status:    "success",
		}
		scheduler.history = append(scheduler.history, entry)
	}
	
	for i := 0; i < 5; i++ {
		entry := HistoryEntry{
			Timestamp: time.Now(),
			JobName:   "job_b",
			TaskType:  "maintenance_cascade",
			Status:    "success",
		}
		scheduler.history = append(scheduler.history, entry)
	}
	scheduler.mu.Unlock()
	
	// Test getting all history
	allHistory, err := scheduler.GetExecutionHistory("", 0)
	if err != nil {
		t.Fatalf("Failed to get execution history: %v", err)
	}
	
	if len(allHistory) != 15 {
		t.Errorf("Expected 15 total entries, got %d", len(allHistory))
	}
	
	// Test filtering by job name
	jobAHistory, err := scheduler.GetExecutionHistory("job_a", 0)
	if err != nil {
		t.Fatalf("Failed to get filtered history: %v", err)
	}
	
	if len(jobAHistory) != 10 {
		t.Errorf("Expected 10 entries for job_a, got %d", len(jobAHistory))
	}
	
	// Test limiting results
	limitedHistory, err := scheduler.GetExecutionHistory("", 5)
	if err != nil {
		t.Fatalf("Failed to get limited history: %v", err)
	}
	
	if len(limitedHistory) != 5 {
		t.Errorf("Expected 5 limited entries, got %d", len(limitedHistory))
	}
}
