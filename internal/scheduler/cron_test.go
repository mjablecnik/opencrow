package scheduler

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCronScheduler_Start(t *testing.T) {
	// Create a temporary directory for test config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cron.json")

	// Create scheduler
	scheduler := NewCronScheduler(configPath, nil)

	// Test starting scheduler without existing config (should create default jobs)
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Verify default job was created
	scheduler.mu.RLock()
	if len(scheduler.jobs) == 0 {
		t.Error("Expected default jobs to be created")
	}

	// Check if daily maintenance job exists
	job, exists := scheduler.jobs["daily_maintenance_cascade"]
	if !exists {
		t.Error("Expected daily_maintenance_cascade job to exist")
	}

	if job.Schedule != "0 4 * * *" {
		t.Errorf("Expected schedule '0 4 * * *', got '%s'", job.Schedule)
	}

	if !job.Enabled {
		t.Error("Expected job to be enabled")
	}
	scheduler.mu.RUnlock()

	// Verify config file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected config file to be created")
	}

	// Stop the scheduler
	err = scheduler.Stop()
	if err != nil {
		t.Fatalf("Failed to stop scheduler: %v", err)
	}
}

func TestCronScheduler_StartWithExistingConfig(t *testing.T) {
	// Create a temporary directory for test config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cron.json")

	// Create a test config file
	testConfig := `{
		"version": "1.0",
		"jobs": [
			{
				"name": "test_job",
				"schedule": "*/5 * * * *",
				"task_type": "reminder",
				"enabled": true,
				"message": "Test reminder",
				"chat_id": 123456789
			}
		]
	}`

	err := os.WriteFile(configPath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Create scheduler
	scheduler := NewCronScheduler(configPath, nil)

	// Start scheduler
	err = scheduler.Start()
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Verify job was loaded
	scheduler.mu.RLock()
	job, exists := scheduler.jobs["test_job"]
	if !exists {
		t.Error("Expected test_job to be loaded")
	}

	if job.Schedule != "*/5 * * * *" {
		t.Errorf("Expected schedule '*/5 * * * *', got '%s'", job.Schedule)
	}

	if job.Message != "Test reminder" {
		t.Errorf("Expected message 'Test reminder', got '%s'", job.Message)
	}
	scheduler.mu.RUnlock()

	// Stop the scheduler
	err = scheduler.Stop()
	if err != nil {
		t.Fatalf("Failed to stop scheduler: %v", err)
	}
}

func TestCronScheduler_StartWithDisabledJob(t *testing.T) {
	// Create a temporary directory for test config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cron.json")

	// Create a test config file with disabled job
	testConfig := `{
		"version": "1.0",
		"jobs": [
			{
				"name": "disabled_job",
				"schedule": "*/5 * * * *",
				"task_type": "reminder",
				"enabled": false,
				"message": "This should not run"
			}
		]
	}`

	err := os.WriteFile(configPath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Create scheduler
	scheduler := NewCronScheduler(configPath, nil)

	// Start scheduler
	err = scheduler.Start()
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Verify job was loaded but not scheduled
	scheduler.mu.RLock()
	job, exists := scheduler.jobs["disabled_job"]
	if !exists {
		t.Error("Expected disabled_job to be loaded")
	}

	if job.Enabled {
		t.Error("Expected job to be disabled")
	}

	// Check that no cron entry was created for disabled job
	_, scheduled := scheduler.cronIDs["disabled_job"]
	if scheduled {
		t.Error("Expected disabled job not to be scheduled")
	}
	scheduler.mu.RUnlock()

	// Stop the scheduler
	err = scheduler.Stop()
	if err != nil {
		t.Fatalf("Failed to stop scheduler: %v", err)
	}
}

func TestCronScheduler_StartWithFutureStartTime(t *testing.T) {
	// Create a temporary directory for test config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cron.json")

	// Create scheduler
	scheduler := NewCronScheduler(configPath, nil)

	// Add a job with future start time
	futureTime := time.Now().Add(1 * time.Hour)
	job := &JobInfo{
		Name:     "future_job",
		Schedule: "*/5 * * * *",
		TaskType: "reminder",
		Enabled:  true,
		StartsAt: &futureTime,
		Message:  "Future reminder",
	}

	scheduler.jobs["future_job"] = job
	err := scheduler.SaveJobs()
	if err != nil {
		t.Fatalf("Failed to save jobs: %v", err)
	}

	// Start scheduler
	err = scheduler.Start()
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Verify job was loaded but not scheduled (starts in future)
	scheduler.mu.RLock()
	_, scheduled := scheduler.cronIDs["future_job"]
	if scheduled {
		t.Error("Expected future job not to be scheduled yet")
	}
	scheduler.mu.RUnlock()

	// Stop the scheduler
	err = scheduler.Stop()
	if err != nil {
		t.Fatalf("Failed to stop scheduler: %v", err)
	}
}

func TestCronScheduler_StartWithExpiredJob(t *testing.T) {
	// Create a temporary directory for test config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cron.json")

	// Create scheduler
	scheduler := NewCronScheduler(configPath, nil)

	// Add a job with past expiration time
	pastTime := time.Now().Add(-1 * time.Hour)
	job := &JobInfo{
		Name:      "expired_job",
		Schedule:  "*/5 * * * *",
		TaskType:  "reminder",
		Enabled:   true,
		ExpiresAt: &pastTime,
		Message:   "Expired reminder",
	}

	scheduler.jobs["expired_job"] = job
	err := scheduler.SaveJobs()
	if err != nil {
		t.Fatalf("Failed to save jobs: %v", err)
	}

	// Start scheduler
	err = scheduler.Start()
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Verify job was loaded but not scheduled (already expired)
	scheduler.mu.RLock()
	_, scheduled := scheduler.cronIDs["expired_job"]
	if scheduled {
		t.Error("Expected expired job not to be scheduled")
	}
	scheduler.mu.RUnlock()

	// Stop the scheduler
	err = scheduler.Stop()
	if err != nil {
		t.Fatalf("Failed to stop scheduler: %v", err)
	}
}

func TestCronScheduler_ValidateCronExpression(t *testing.T) {
	scheduler := NewCronScheduler("", nil)

	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{"valid daily", "0 4 * * *", false},
		{"valid hourly", "0 * * * *", false},
		{"valid every 5 minutes", "*/5 * * * *", false},
		{"valid specific time", "30 14 * * 1", false},
		{"invalid format", "invalid", true},
		{"invalid field count", "0 4 *", true},
		{"invalid minute", "60 * * * *", true},
		{"invalid hour", "0 25 * * *", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := scheduler.ValidateCronExpression(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCronExpression() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
