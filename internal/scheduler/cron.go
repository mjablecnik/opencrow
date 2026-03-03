package scheduler

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// JobStatus represents the current status of a cron job
type JobStatus string

const (
	JobStatusEnabled  JobStatus = "enabled"
	JobStatusDisabled JobStatus = "disabled"
	JobStatusRunning  JobStatus = "running"
	JobStatusSuccess  JobStatus = "success"
	JobStatusFailed   JobStatus = "failed"
)

// JobInfo contains information about a scheduled job
type JobInfo struct {
	Name        string     `json:"name"`
	Schedule    string     `json:"schedule,omitempty"`     // Cron expression for recurring jobs
	ExecuteAt   *time.Time `json:"execute_at,omitempty"`   // Timestamp for one-time jobs
	TaskType    string     `json:"task_type"`
	Enabled     bool       `json:"enabled"`
	Status      JobStatus  `json:"status"`
	LastRun     time.Time  `json:"last_run,omitempty"`
	NextRun     time.Time  `json:"next_run,omitempty"`
	LastError   string     `json:"last_error,omitempty"`
	StartsAt    *time.Time `json:"starts_at,omitempty"`    // When job should begin executing
	PausedUntil *time.Time `json:"paused_until,omitempty"` // Temporarily suspend until this time
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`   // When job should stop and be deleted
	Message     string     `json:"message,omitempty"`      // Message content for reminder tasks
	ChatID      int64      `json:"chat_id,omitempty"`      // Telegram chat ID for reminder tasks
	Description string     `json:"description,omitempty"`
}

// TelegramSender defines the interface for sending Telegram messages
type TelegramSender interface {
	SendMessage(chatID int64, text string) error
}

// SessionLogger defines the interface for logging messages to session logs
type SessionLogger interface {
	AppendToSessionLog(role, content string) error
}

// CronScheduler manages scheduled tasks using cron expressions
type CronScheduler struct {
	cron           *cron.Cron
	jobs           map[string]*JobInfo
	cronIDs        map[string]cron.EntryID
	configPath     string
	mu             sync.RWMutex
	logger         *log.Logger
	telegramSender TelegramSender // For sending reminder messages
	sessionLogger  SessionLogger  // For logging cron messages to session logs
}

// NewCronScheduler creates a new CronScheduler instance
func NewCronScheduler(configPath string, logger *log.Logger) *CronScheduler {
	if logger == nil {
		logger = log.New(os.Stdout, "[SCHEDULER] ", log.LstdFlags)
	}

	return &CronScheduler{
		cron:           cron.New(),
		jobs:           make(map[string]*JobInfo),
		cronIDs:        make(map[string]cron.EntryID),
		configPath:     configPath,
		logger:         logger,
		telegramSender: nil, // Will be set via SetTelegramSender
		sessionLogger:  nil, // Will be set via SetSessionLogger
	}
}

// SetTelegramSender sets the Telegram sender for reminder messages
func (s *CronScheduler) SetTelegramSender(sender TelegramSender) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.telegramSender = sender
}

// SetSessionLogger sets the session logger for logging cron messages
func (s *CronScheduler) SetSessionLogger(logger SessionLogger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionLogger = logger
}

// Start initializes the cron scheduler and begins executing scheduled jobs
func (s *CronScheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Println("Starting cron scheduler...")

	// Load jobs from config/cron.json
	if err := s.loadJobsInternal(); err != nil {
		// If config doesn't exist, create default jobs
		if os.IsNotExist(err) {
			s.logger.Println("Config file not found, creating default jobs...")
			if err := s.createDefaultJobs(); err != nil {
				return fmt.Errorf("failed to create default jobs: %w", err)
			}
			// Save the default jobs
			if err := s.saveJobsInternal(); err != nil {
				return fmt.Errorf("failed to save default jobs: %w", err)
			}
		} else {
			return fmt.Errorf("failed to load jobs: %w", err)
		}
	}

	// Schedule all enabled jobs
	for name, job := range s.jobs {
		if !job.Enabled {
			s.logger.Printf("Skipping disabled job: %s", name)
			continue
		}

		// Check if job has a starts_at time and hasn't started yet
		if job.StartsAt != nil && time.Now().Before(*job.StartsAt) {
			s.logger.Printf("Job %s not yet started (starts at %s)", name, job.StartsAt.Format(time.RFC3339))
			continue
		}

		// Check if job is expired
		if job.ExpiresAt != nil && time.Now().After(*job.ExpiresAt) {
			s.logger.Printf("Job %s has expired, skipping", name)
			continue
		}

		// Schedule the job
		if err := s.scheduleJob(name, job); err != nil {
			s.logger.Printf("Failed to schedule job %s: %v", name, err)
			continue
		}

		s.logger.Printf("Scheduled job: %s (schedule: %s, type: %s)", name, job.Schedule, job.TaskType)
	}

	// Start the cron scheduler
	s.cron.Start()
	s.logger.Println("Cron scheduler started successfully")

	return nil
}

// Stop gracefully stops the cron scheduler
func (s *CronScheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Println("Stopping cron scheduler...")
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Println("Cron scheduler stopped")

	return nil
}

// scheduleJob schedules a single job with the cron scheduler
func (s *CronScheduler) scheduleJob(name string, job *JobInfo) error {
	// Create a wrapper function that handles job execution
	jobFunc := func() {
		s.executeJob(name)
	}

	// Handle one-time jobs with execute_at timestamp
	if job.ExecuteAt != nil {
		// Calculate duration until execution
		now := time.Now()
		duration := job.ExecuteAt.Sub(now)
		
		if duration <= 0 {
			// Job should have already executed, skip it
			s.logger.Printf("One-time job %s has past execution time, skipping", name)
			return nil
		}

		// Schedule the job to run once after the duration
		s.logger.Printf("Scheduling one-time job %s to execute at %s (in %v)", 
			name, job.ExecuteAt.Format(time.RFC3339), duration)
		
		// Use a goroutine with timer for one-time execution
		go func() {
			timer := time.NewTimer(duration)
			<-timer.C
			s.executeJob(name)
		}()
		
		// Update next run time
		job.NextRun = *job.ExecuteAt
		return nil
	}

	// Handle recurring jobs with cron schedule
	if job.Schedule == "" {
		return fmt.Errorf("job must have either schedule or execute_at")
	}

	// Add the job to the cron scheduler
	entryID, err := s.cron.AddFunc(job.Schedule, jobFunc)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	// Store the entry ID for later reference
	s.cronIDs[name] = entryID

	// Update next run time
	entry := s.cron.Entry(entryID)
	job.NextRun = entry.Next

	return nil
}

// executeJob executes a scheduled job
func (s *CronScheduler) executeJob(name string) {
	s.mu.Lock()
	job, exists := s.jobs[name]
	if !exists {
		s.mu.Unlock()
		s.logger.Printf("Job %s not found", name)
		return
	}

	// Check if job is paused
	if job.PausedUntil != nil && time.Now().Before(*job.PausedUntil) {
		s.mu.Unlock()
		s.logger.Printf("Job %s is paused until %s", name, job.PausedUntil.Format(time.RFC3339))
		return
	}

	// Check if job has expired
	if job.ExpiresAt != nil && time.Now().After(*job.ExpiresAt) {
		s.mu.Unlock()
		s.logger.Printf("Job %s has expired, removing", name)
		// Remove the expired job
		if err := s.RemoveJob(name); err != nil {
			s.logger.Printf("Failed to remove expired job %s: %v", name, err)
		}
		return
	}

	// Update job status
	job.Status = JobStatusRunning
	job.LastRun = time.Now()
	s.mu.Unlock()

	s.logger.Printf("Executing job: %s (type: %s)", name, job.TaskType)

	// Execute the job based on its type
	var err error
	switch job.TaskType {
	case "maintenance_cascade":
		err = s.executeDailyMaintenanceCascade()
	case "reminder":
		err = s.executeReminder(job)
	default:
		err = fmt.Errorf("unknown task type: %s", job.TaskType)
	}

	// Update job status based on execution result
	s.mu.Lock()
	if err != nil {
		job.Status = JobStatusFailed
		job.LastError = err.Error()
		s.logger.Printf("Job %s failed: %v", name, err)
	} else {
		job.Status = JobStatusSuccess
		job.LastError = ""
		s.logger.Printf("Job %s completed successfully", name)
	}

	// Check if this is a one-time job that should be auto-deleted
	if job.ExecuteAt != nil {
		s.logger.Printf("One-time job %s completed, removing from scheduler", name)
		s.mu.Unlock()
		
		// Remove the job (this will acquire the lock internally)
		if removeErr := s.RemoveJob(name); removeErr != nil {
			s.logger.Printf("Failed to remove one-time job %s: %v", name, removeErr)
		}
		return
	}

	// Update next run time for recurring jobs
	if entryID, exists := s.cronIDs[name]; exists {
		entry := s.cron.Entry(entryID)
		job.NextRun = entry.Next
	}

	// Save jobs after execution
	if err := s.saveJobsInternal(); err != nil {
		s.logger.Printf("Failed to save jobs after execution: %v", err)
	}
	s.mu.Unlock()
}

// executeDailyMaintenanceCascade executes the daily maintenance cascade
func (s *CronScheduler) executeDailyMaintenanceCascade() error {
	// TODO: Implement cascade execution
	// This will be implemented in a later task
	s.logger.Println("Daily maintenance cascade execution (placeholder)")
	return nil
}

// executeReminder executes a reminder task
func (s *CronScheduler) executeReminder(job *JobInfo) error {
	// Validate that we have the required dependencies
	if s.telegramSender == nil {
		return fmt.Errorf("telegram sender not configured")
	}
	
	if s.sessionLogger == nil {
		return fmt.Errorf("session logger not configured")
	}

	// Validate job has required fields
	if job.Message == "" {
		return fmt.Errorf("reminder job missing message")
	}
	
	if job.ChatID == 0 {
		return fmt.Errorf("reminder job missing chat_id")
	}

	s.logger.Printf("Sending reminder to chat %d: %s", job.ChatID, job.Message)

	// Send the reminder message via Telegram
	if err := s.telegramSender.SendMessage(job.ChatID, job.Message); err != nil {
		return fmt.Errorf("failed to send reminder message: %w", err)
	}

	// Log the message to session logs with "Assistant (Cron):" prefix
	if err := s.sessionLogger.AppendToSessionLog("Assistant (Cron)", job.Message); err != nil {
		// Log the error but don't fail the reminder - the message was sent successfully
		s.logger.Printf("Warning: Failed to log cron message to session log: %v", err)
	}

	s.logger.Printf("Reminder sent successfully to chat %d", job.ChatID)
	return nil
}

// loadJobsInternal loads jobs from the config file (internal, no lock)
func (s *CronScheduler) loadJobsInternal() error {
	// Check if config file exists
	if _, err := os.Stat(s.configPath); os.IsNotExist(err) {
		return err
	}

	// Read config file
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config struct {
		Version string     `json:"version"`
		Jobs    []*JobInfo `json:"jobs"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Load jobs into map
	s.jobs = make(map[string]*JobInfo)
	for _, job := range config.Jobs {
		s.jobs[job.Name] = job
	}

	s.logger.Printf("Loaded %d jobs from config", len(s.jobs))
	return nil
}

// LoadJobs loads jobs from the config file (public, with lock)
func (s *CronScheduler) LoadJobs() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadJobsInternal()
}

// saveJobsInternal saves jobs to the config file (internal, no lock)
func (s *CronScheduler) saveJobsInternal() error {
	// Ensure config directory exists
	configDir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Convert jobs map to slice
	jobSlice := make([]*JobInfo, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobSlice = append(jobSlice, job)
	}

	// Create config structure
	config := struct {
		Version string     `json:"version"`
		Jobs    []*JobInfo `json:"jobs"`
	}{
		Version: "1.0",
		Jobs:    jobSlice,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SaveJobs saves jobs to the config file (public, with lock)
func (s *CronScheduler) SaveJobs() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveJobsInternal()
}

// createDefaultJobs creates default jobs on first run
func (s *CronScheduler) createDefaultJobs() error {
	// Create default daily maintenance cascade job
	defaultJob := &JobInfo{
		Name:        "daily_maintenance_cascade",
		Schedule:    "0 4 * * *", // 4:00 AM daily
		TaskType:    "maintenance_cascade",
		Enabled:     true,
		Status:      JobStatusEnabled,
		Description: "Execute daily maintenance cascade: summary, topic extraction, notes cleanup, weekly/quarterly ops (if applicable), session reset",
	}

	s.jobs[defaultJob.Name] = defaultJob
	s.logger.Println("Created default daily maintenance cascade job")

	return nil
}

// ValidateCronExpression validates a cron expression
func (s *CronScheduler) ValidateCronExpression(expr string) error {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(expr)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}
	return nil
}

// AddJob adds a new job to the scheduler
func (s *CronScheduler) AddJob(name string, schedule string, task func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if job already exists
	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job %s already exists", name)
	}

	// Validate cron expression
	if err := s.ValidateCronExpression(schedule); err != nil {
		return err
	}

	// Create new job
	job := &JobInfo{
		Name:     name,
		Schedule: schedule,
		Enabled:  true,
		Status:   JobStatusEnabled,
	}

	s.jobs[name] = job

	// Save jobs
	if err := s.saveJobsInternal(); err != nil {
		return fmt.Errorf("failed to save jobs: %w", err)
	}

	return nil
}

// RemoveJob removes a job from the scheduler
func (s *CronScheduler) RemoveJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if job exists
	if _, exists := s.jobs[name]; !exists {
		return fmt.Errorf("job %s not found", name)
	}

	// Remove from cron scheduler if scheduled
	if entryID, exists := s.cronIDs[name]; exists {
		s.cron.Remove(entryID)
		delete(s.cronIDs, name)
	}

	// Remove from jobs map
	delete(s.jobs, name)

	// Save jobs
	if err := s.saveJobsInternal(); err != nil {
		return fmt.Errorf("failed to save jobs: %w", err)
	}

	s.logger.Printf("Removed job: %s", name)
	return nil
}

// EnableJob enables a disabled job
func (s *CronScheduler) EnableJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job %s not found", name)
	}

	if job.Enabled {
		return fmt.Errorf("job %s is already enabled", name)
	}

	job.Enabled = true
	job.Status = JobStatusEnabled

	// Save jobs
	if err := s.saveJobsInternal(); err != nil {
		return fmt.Errorf("failed to save jobs: %w", err)
	}

	s.logger.Printf("Enabled job: %s", name)
	return nil
}

// DisableJob disables an enabled job
func (s *CronScheduler) DisableJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job %s not found", name)
	}

	if !job.Enabled {
		return fmt.Errorf("job %s is already disabled", name)
	}

	job.Enabled = false
	job.Status = JobStatusDisabled

	// Remove from cron scheduler if scheduled
	if entryID, exists := s.cronIDs[name]; exists {
		s.cron.Remove(entryID)
		delete(s.cronIDs, name)
	}

	// Save jobs
	if err := s.saveJobsInternal(); err != nil {
		return fmt.Errorf("failed to save jobs: %w", err)
	}

	s.logger.Printf("Disabled job: %s", name)
	return nil
}

// ListJobs returns a list of all jobs
func (s *CronScheduler) ListJobs() ([]JobInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]JobInfo, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, *job)
	}

	return jobs, nil
}

// GetJob returns information about a specific job
func (s *CronScheduler) GetJob(name string) (*JobInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[name]
	if !exists {
		return nil, fmt.Errorf("job %s not found", name)
	}

	// Return a copy to prevent external modification
	jobCopy := *job
	return &jobCopy, nil
}

// ExecuteDailyMaintenanceCascade executes the daily maintenance cascade
func (s *CronScheduler) ExecuteDailyMaintenanceCascade() error {
	return s.executeDailyMaintenanceCascade()
}

// ShouldExecuteWeeklyOps determines if weekly operations should be executed
func (s *CronScheduler) ShouldExecuteWeeklyOps() bool {
	// Weekly operations run on Monday
	return time.Now().Weekday() == time.Monday
}

// ShouldExecuteQuarterlyOps determines if quarterly operations should be executed
func (s *CronScheduler) ShouldExecuteQuarterlyOps() bool {
	now := time.Now()
	// Quarterly operations run on the first day of a new quarter (Jan 1, Apr 1, Jul 1, Oct 1)
	month := now.Month()
	day := now.Day()

	return day == 1 && (month == time.January || month == time.April || month == time.July || month == time.October)
}
