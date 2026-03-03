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

// MemorySessionManager defines the interface for session management operations
type MemorySessionManager interface {
	PerformScheduledSessionReset() (string, error) // Returns date folder to summarize
}

// SummaryManager defines the interface for summary generation operations
type SummaryManager interface {
	GenerateDailySummary(date time.Time) error
	ExtractTopicsFromContent(content string) error
}

// HistoryEntry represents a single execution history entry
type HistoryEntry struct {
	Timestamp           time.Time `json:"timestamp"`
	JobName             string    `json:"job_name"`
	TaskType            string    `json:"task_type"`
	Status              string    `json:"status"` // "success", "failed", "deleted"
	Error               string    `json:"error,omitempty"`
	MessageSent         string    `json:"message_sent,omitempty"`
	ChatID              int64     `json:"chat_id,omitempty"`
	OperationsCompleted []string  `json:"operations_completed,omitempty"`
	Deleted             bool      `json:"deleted,omitempty"`
	DeletionReason      string    `json:"deletion_reason,omitempty"`
}

// CronScheduler manages scheduled tasks using cron expressions
type CronScheduler struct {
	cron                 *cron.Cron
	jobs                 map[string]*JobInfo
	cronIDs              map[string]cron.EntryID
	configPath           string
	historyPath          string
	history              []HistoryEntry
	mu                   sync.RWMutex
	logger               *log.Logger
	telegramSender       TelegramSender       // For sending reminder messages
	sessionLogger        SessionLogger        // For logging cron messages to session logs
	memorySessionManager MemorySessionManager // For session reset operations
	summaryManager       SummaryManager       // For daily summary generation
	maintenanceChatID    int64                // Chat ID to send maintenance messages to
}

// NewCronScheduler creates a new CronScheduler instance
func NewCronScheduler(configPath string, logger *log.Logger) *CronScheduler {
	if logger == nil {
		logger = log.New(os.Stdout, "[SCHEDULER] ", log.LstdFlags)
	}

	// Determine history path from config path
	configDir := filepath.Dir(configPath)
	historyPath := filepath.Join(configDir, "cron_history.json")

	return &CronScheduler{
		cron:                 cron.New(),
		jobs:                 make(map[string]*JobInfo),
		cronIDs:              make(map[string]cron.EntryID),
		configPath:           configPath,
		historyPath:          historyPath,
		history:              make([]HistoryEntry, 0),
		logger:               logger,
		telegramSender:       nil, // Will be set via SetTelegramSender
		sessionLogger:        nil, // Will be set via SetSessionLogger
		memorySessionManager: nil, // Will be set via SetMemorySessionManager
		summaryManager:       nil, // Will be set via SetSummaryManager
		maintenanceChatID:    0,   // Will be set via SetMaintenanceChatID
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

// SetMemorySessionManager sets the memory session manager for session reset operations
func (s *CronScheduler) SetMemorySessionManager(manager MemorySessionManager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.memorySessionManager = manager
}

// SetSummaryManager sets the summary manager for daily summary generation
func (s *CronScheduler) SetSummaryManager(manager SummaryManager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.summaryManager = manager
}

// SetMaintenanceChatID sets the chat ID to send maintenance messages to
func (s *CronScheduler) SetMaintenanceChatID(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maintenanceChatID = chatID
}

// Start initializes the cron scheduler and begins executing scheduled jobs
func (s *CronScheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Println("Starting cron scheduler...")

	// Load execution history
	if err := s.loadHistoryInternal(); err != nil {
		s.logger.Printf("Warning: Failed to load execution history: %v", err)
		// Continue anyway - history is not critical for startup
	}

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
		
		// Remove the expired job (will log as "expired" in RemoveJob)
		if err := s.removeJobWithReason(name, "expired"); err != nil {
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
	
	// Create history entry
	historyEntry := HistoryEntry{
		Timestamp: time.Now(),
		JobName:   name,
		TaskType:  job.TaskType,
	}
	
	if err != nil {
		job.Status = JobStatusFailed
		job.LastError = err.Error()
		historyEntry.Status = "failed"
		historyEntry.Error = err.Error()
		s.logger.Printf("Job %s failed: %v", name, err)
	} else {
		job.Status = JobStatusSuccess
		job.LastError = ""
		historyEntry.Status = "success"
		s.logger.Printf("Job %s completed successfully", name)
	}
	
	// Add reminder-specific information to history
	if job.TaskType == "reminder" && err == nil {
		historyEntry.MessageSent = job.Message
		historyEntry.ChatID = job.ChatID
	}
	
	// Append history entry
	if appendErr := s.appendHistoryEntry(historyEntry); appendErr != nil {
		s.logger.Printf("Warning: Failed to append history entry: %v", appendErr)
	}

	// Check if this is a one-time job that should be auto-deleted
	if job.ExecuteAt != nil {
		s.logger.Printf("One-time job %s completed, removing from scheduler", name)
		s.mu.Unlock()
		
		// Remove the job (this will acquire the lock internally and log as "one_time_execution_complete")
		if removeErr := s.removeJobWithReason(name, "one_time_execution_complete"); removeErr != nil {
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
	s.logger.Println("Starting daily maintenance cascade...")
	
	// Track completed operations for history logging
	completedOps := make([]string, 0)
	
	// Determine which operations to execute based on day of week and quarter
	isMonday := s.ShouldExecuteWeeklyOps()
	isFirstMondayOfQuarter := isMonday && s.ShouldExecuteQuarterlyOps()
	
	s.logger.Printf("Cascade conditions - Monday: %v, First Monday of Quarter: %v", isMonday, isFirstMondayOfQuarter)
	
	// Operation 1: Daily summary generation
	s.logger.Println("Step 1: Generating daily summary...")
	if err := s.executeDailySummary(); err != nil {
		s.logger.Printf("ERROR: Daily summary failed: %v", err)
		s.logCascadeFailure(completedOps, "daily_summary", err)
		return fmt.Errorf("cascade aborted at daily_summary: %w", err)
	}
	completedOps = append(completedOps, "daily_summary")
	s.logger.Println("Daily summary completed successfully")
	
	// Operation 2: Topic extraction from daily summary
	s.logger.Println("Step 2: Extracting topics from daily summary...")
	if err := s.executeTopicExtraction(); err != nil {
		s.logger.Printf("ERROR: Topic extraction failed: %v", err)
		s.logCascadeFailure(completedOps, "topic_extraction", err)
		return fmt.Errorf("cascade aborted at topic_extraction: %w", err)
	}
	completedOps = append(completedOps, "topic_extraction")
	s.logger.Println("Topic extraction completed successfully")
	
	// Operation 3: Notes cleanup
	s.logger.Println("Step 3: Cleaning up old notes...")
	if err := s.executeNotesCleanup(); err != nil {
		s.logger.Printf("ERROR: Notes cleanup failed: %v", err)
		s.logCascadeFailure(completedOps, "notes_cleanup", err)
		return fmt.Errorf("cascade aborted at notes_cleanup: %w", err)
	}
	completedOps = append(completedOps, "notes_cleanup")
	s.logger.Println("Notes cleanup completed successfully")
	
	// Operation 4: Weekly reorganization (if Monday)
	if isMonday {
		s.logger.Println("Step 4: Performing weekly reorganization (Monday)...")
		if err := s.executeWeeklyReorganization(); err != nil {
			s.logger.Printf("ERROR: Weekly reorganization failed: %v", err)
			s.logCascadeFailure(completedOps, "weekly_reorganization", err)
			return fmt.Errorf("cascade aborted at weekly_reorganization: %w", err)
		}
		completedOps = append(completedOps, "weekly_reorganization")
		s.logger.Println("Weekly reorganization completed successfully")
		
		// Operation 5: Weekly summary generation (if Monday)
		s.logger.Println("Step 5: Generating weekly summary (Monday)...")
		if err := s.executeWeeklySummary(); err != nil {
			s.logger.Printf("ERROR: Weekly summary failed: %v", err)
			s.logCascadeFailure(completedOps, "weekly_summary", err)
			return fmt.Errorf("cascade aborted at weekly_summary: %w", err)
		}
		completedOps = append(completedOps, "weekly_summary")
		s.logger.Println("Weekly summary completed successfully")
	} else {
		s.logger.Println("Skipping weekly operations (not Monday)")
	}
	
	// Operation 6: Quarterly reorganization (if first Monday of quarter)
	if isFirstMondayOfQuarter {
		s.logger.Println("Step 6: Performing quarterly reorganization (First Monday of Quarter)...")
		if err := s.executeQuarterlyReorganization(); err != nil {
			s.logger.Printf("ERROR: Quarterly reorganization failed: %v", err)
			s.logCascadeFailure(completedOps, "quarterly_reorganization", err)
			return fmt.Errorf("cascade aborted at quarterly_reorganization: %w", err)
		}
		completedOps = append(completedOps, "quarterly_reorganization")
		s.logger.Println("Quarterly reorganization completed successfully")
		
		// Operation 7: Quarterly summary generation (if first Monday of quarter)
		s.logger.Println("Step 7: Generating quarterly summary (First Monday of Quarter)...")
		if err := s.executeQuarterlySummary(); err != nil {
			s.logger.Printf("ERROR: Quarterly summary failed: %v", err)
			s.logCascadeFailure(completedOps, "quarterly_summary", err)
			return fmt.Errorf("cascade aborted at quarterly_summary: %w", err)
		}
		completedOps = append(completedOps, "quarterly_summary")
		s.logger.Println("Quarterly summary completed successfully")
	} else if isMonday {
		s.logger.Println("Skipping quarterly operations (not first Monday of quarter)")
	}
	
	// Operation 8: Session reset (after all operations complete)
	s.logger.Println("Final Step: Performing session reset...")
	if err := s.executeSessionReset(); err != nil {
		s.logger.Printf("ERROR: Session reset failed: %v", err)
		s.logCascadeFailure(completedOps, "session_reset", err)
		return fmt.Errorf("cascade aborted at session_reset: %w", err)
	}
	completedOps = append(completedOps, "session_reset")
	s.logger.Println("Session reset completed successfully")
	
	// Log successful cascade completion
	s.logger.Printf("Daily maintenance cascade completed successfully. Operations: %v", completedOps)
	s.logCascadeSuccess(completedOps)
	
	return nil
}

// logCascadeSuccess logs a successful cascade execution to history
func (s *CronScheduler) logCascadeSuccess(completedOps []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	entry := HistoryEntry{
		Timestamp:           time.Now(),
		JobName:             "daily_maintenance_cascade",
		TaskType:            "maintenance_cascade",
		Status:              "success",
		OperationsCompleted: completedOps,
	}
	
	if err := s.appendHistoryEntry(entry); err != nil {
		s.logger.Printf("Warning: Failed to log cascade success: %v", err)
	}
}

// logCascadeFailure logs a failed cascade execution to history
func (s *CronScheduler) logCascadeFailure(completedOps []string, failedOp string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	entry := HistoryEntry{
		Timestamp:           time.Now(),
		JobName:             "daily_maintenance_cascade",
		TaskType:            "maintenance_cascade",
		Status:              "failed",
		Error:               fmt.Sprintf("Failed at %s: %v", failedOp, err),
		OperationsCompleted: completedOps,
	}
	
	if appendErr := s.appendHistoryEntry(entry); appendErr != nil {
		s.logger.Printf("Warning: Failed to log cascade failure: %v", appendErr)
	}
}

// Placeholder implementations for Memory Manager operations
// These will be replaced with actual Memory Manager integration in task 16

func (s *CronScheduler) executeDailySummary() error {
	// TODO: Call Memory Manager's GenerateDailySummary method
	s.logger.Println("[PLACEHOLDER] Would generate daily summary from all session logs")
	s.logger.Println("[PLACEHOLDER] Would create daily-summary.md in today's daily folder")
	return nil
}

func (s *CronScheduler) executeTopicExtraction() error {
	// TODO: Call Memory Manager's ExtractTopics method with daily summary content
	s.logger.Println("[PLACEHOLDER] Would extract topics from daily summary")
	s.logger.Println("[PLACEHOLDER] Would update or create topic files if relevant knowledge found")
	return nil
}

func (s *CronScheduler) executeNotesCleanup() error {
	// TODO: Call Memory Manager's CleanupNotes method
	s.logger.Println("[PLACEHOLDER] Would clean up old notes based on age, status, and references")
	s.logger.Println("[PLACEHOLDER] Would respect auto_delete flags and NOTES_CLEANUP_ENABLED config")
	return nil
}

func (s *CronScheduler) executeWeeklyReorganization() error {
	// TODO: Call Memory Manager's PerformWeeklyReorganization method
	s.logger.Println("[PLACEHOLDER] Would create week folder (week-WW-YYYY)")
	s.logger.Println("[PLACEHOLDER] Would move last 7 daily folders into week folder")
	s.logger.Println("[PLACEHOLDER] Would preserve all session logs")
	return nil
}

func (s *CronScheduler) executeWeeklySummary() error {
	// TODO: Call Memory Manager's GenerateWeeklySummary method
	s.logger.Println("[PLACEHOLDER] Would generate weekly summary from daily summaries")
	s.logger.Println("[PLACEHOLDER] Would create summary.md in week folder")
	return nil
}

func (s *CronScheduler) executeQuarterlyReorganization() error {
	// TODO: Call Memory Manager's PerformQuarterlyReorganization method
	s.logger.Println("[PLACEHOLDER] Would create quarter folder (QN-YYYY)")
	s.logger.Println("[PLACEHOLDER] Would move all week folders from completed quarter")
	s.logger.Println("[PLACEHOLDER] Would preserve all session logs")
	return nil
}

func (s *CronScheduler) executeQuarterlySummary() error {
	// TODO: Call Memory Manager's GenerateQuarterlySummary method
	s.logger.Println("[PLACEHOLDER] Would generate quarterly summary from weekly summaries")
	s.logger.Println("[PLACEHOLDER] Would create summary.md in quarter folder")
	return nil
}

func (s *CronScheduler) executeSessionReset() error {
	s.logger.Println("Executing scheduled session reset...")
	
	// Step 1: Send maintenance message to Telegram (if configured)
	if s.telegramSender != nil && s.maintenanceChatID != 0 {
		maintenanceMsg := "🔧 <b>Probíhá denní údržba...</b>\n\nGeneruji souhrn včerejšího dne a resetuji session. Chvilku to potrvá."
		if err := s.telegramSender.SendMessage(s.maintenanceChatID, maintenanceMsg); err != nil {
			s.logger.Printf("Warning: Failed to send maintenance message: %v", err)
			// Continue anyway - this is not critical
		} else {
			s.logger.Println("Maintenance message sent to Telegram")
		}
	}
	
	// Step 2: Perform scheduled session reset (archives session-latest.log)
	if s.memorySessionManager == nil {
		return fmt.Errorf("memory session manager not configured")
	}
	
	dateToSummarize, err := s.memorySessionManager.PerformScheduledSessionReset()
	if err != nil {
		return fmt.Errorf("failed to perform scheduled session reset: %w", err)
	}
	
	s.logger.Printf("Session reset complete. Date to summarize: %s", dateToSummarize)
	
	// Step 3: Generate daily summary if there's a date to summarize
	if dateToSummarize != "" && s.summaryManager != nil {
		s.logger.Printf("Generating daily summary for %s...", dateToSummarize)
		
		// Parse the date string
		date, err := time.Parse("2006-01-02", dateToSummarize)
		if err != nil {
			return fmt.Errorf("failed to parse date %s: %w", dateToSummarize, err)
		}
		
		// Generate daily summary
		if err := s.summaryManager.GenerateDailySummary(date); err != nil {
			return fmt.Errorf("failed to generate daily summary: %w", err)
		}
		
		s.logger.Printf("Daily summary generated successfully for %s", dateToSummarize)
		
		// Step 4: Extract topics from the daily summary
		s.logger.Println("Extracting topics from daily summary...")
		
		// Read the generated summary
		summaryPath := fmt.Sprintf("memory/chat/%s/daily-summary.md", dateToSummarize)
		summaryContent, err := os.ReadFile(summaryPath)
		if err != nil {
			s.logger.Printf("Warning: Failed to read daily summary for topic extraction: %v", err)
			// Continue anyway - topic extraction is not critical
		} else {
			if err := s.summaryManager.ExtractTopicsFromContent(string(summaryContent)); err != nil {
				s.logger.Printf("Warning: Topic extraction failed: %v", err)
				// Continue anyway - topic extraction is not critical
			} else {
				s.logger.Println("Topic extraction completed successfully")
			}
		}
	} else {
		s.logger.Println("No date to summarize (no active session was archived)")
	}
	
	// Step 5: Send completion message to Telegram (if configured)
	if s.telegramSender != nil && s.maintenanceChatID != 0 {
		completionMsg := "✅ <b>Denní údržba dokončena!</b>\n\nSession byla resetována a jsem připraven na nový den."
		if err := s.telegramSender.SendMessage(s.maintenanceChatID, completionMsg); err != nil {
			s.logger.Printf("Warning: Failed to send completion message: %v", err)
			// Continue anyway - this is not critical
		} else {
			s.logger.Println("Completion message sent to Telegram")
		}
	}
	
	s.logger.Println("Scheduled session reset completed successfully")
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

// loadHistoryInternal loads execution history from the history file (internal, no lock)
func (s *CronScheduler) loadHistoryInternal() error {
	// Check if history file exists
	if _, err := os.Stat(s.historyPath); os.IsNotExist(err) {
		// History file doesn't exist yet, start with empty history
		s.history = make([]HistoryEntry, 0)
		return nil
	}

	// Read history file
	data, err := os.ReadFile(s.historyPath)
	if err != nil {
		return fmt.Errorf("failed to read history file: %w", err)
	}

	// Parse JSON
	var historyData struct {
		Version    string         `json:"version"`
		MaxEntries int            `json:"max_entries"`
		Entries    []HistoryEntry `json:"entries"`
	}

	if err := json.Unmarshal(data, &historyData); err != nil {
		return fmt.Errorf("failed to parse history file: %w", err)
	}

	s.history = historyData.Entries
	s.logger.Printf("Loaded %d history entries from file", len(s.history))
	return nil
}

// saveHistoryInternal saves execution history to the history file (internal, no lock)
func (s *CronScheduler) saveHistoryInternal() error {
	// Ensure config directory exists
	configDir := filepath.Dir(s.historyPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Limit history to 1000 most recent entries
	const maxHistoryEntries = 1000
	if len(s.history) > maxHistoryEntries {
		// Keep only the most recent entries
		s.history = s.history[len(s.history)-maxHistoryEntries:]
	}

	// Create history structure
	historyData := struct {
		Version    string         `json:"version"`
		MaxEntries int            `json:"max_entries"`
		Entries    []HistoryEntry `json:"entries"`
	}{
		Version:    "1.0",
		MaxEntries: maxHistoryEntries,
		Entries:    s.history,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(historyData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	// Write to file
	if err := os.WriteFile(s.historyPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// appendHistoryEntry adds a new entry to the execution history
func (s *CronScheduler) appendHistoryEntry(entry HistoryEntry) error {
	s.history = append(s.history, entry)
	
	// Save history to file
	if err := s.saveHistoryInternal(); err != nil {
		s.logger.Printf("Warning: Failed to save history: %v", err)
		return err
	}
	
	return nil
}

// logDeletion logs a job deletion event to the execution history
func (s *CronScheduler) logDeletion(jobName, taskType, reason string) {
	entry := HistoryEntry{
		Timestamp:      time.Now(),
		JobName:        jobName,
		TaskType:       taskType,
		Status:         "deleted",
		Deleted:        true,
		DeletionReason: reason,
	}
	
	if err := s.appendHistoryEntry(entry); err != nil {
		s.logger.Printf("Warning: Failed to log deletion event: %v", err)
	}
}

// GetExecutionHistory returns the execution history, optionally filtered by job name
func (s *CronScheduler) GetExecutionHistory(jobName string, limit int) ([]HistoryEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []HistoryEntry

	// Filter by job name if specified
	if jobName != "" {
		for _, entry := range s.history {
			if entry.JobName == jobName {
				filtered = append(filtered, entry)
			}
		}
	} else {
		filtered = s.history
	}

	// Apply limit if specified
	if limit > 0 && len(filtered) > limit {
		// Return the most recent entries
		filtered = filtered[len(filtered)-limit:]
	}

	return filtered, nil
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

// AddReminderJob adds a new recurring reminder job with all required fields set correctly
func (s *CronScheduler) AddReminderJob(name, schedule, message string, chatID int64, startsAt, expiresAt *time.Time) error {
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

	// Create new reminder job with all fields
	job := &JobInfo{
		Name:        name,
		Schedule:    schedule,
		TaskType:    "reminder",
		Message:     message,
		ChatID:      chatID,
		StartsAt:    startsAt,
		ExpiresAt:   expiresAt,
		Enabled:     true,
		Status:      JobStatusEnabled,
		Description: fmt.Sprintf("Recurring reminder: %s", message),
	}

	s.jobs[name] = job

	// Save jobs
	if err := s.saveJobsInternal(); err != nil {
		return fmt.Errorf("failed to save jobs: %w", err)
	}

	s.logger.Printf("Added recurring reminder job: %s (schedule: %s)", name, schedule)
	return nil
}

// AddOneTimeReminderJob adds a new one-time reminder job
func (s *CronScheduler) AddOneTimeReminderJob(name, schedule string, executeAt time.Time, message string, chatID int64) error {
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

	// Create new one-time reminder job
	job := &JobInfo{
		Name:        name,
		Schedule:    schedule,
		ExecuteAt:   &executeAt,
		TaskType:    "reminder",
		Message:     message,
		ChatID:      chatID,
		Enabled:     true,
		Status:      JobStatusEnabled,
		Description: fmt.Sprintf("One-time reminder: %s", message),
	}

	s.jobs[name] = job

	// Save jobs
	if err := s.saveJobsInternal(); err != nil {
		return fmt.Errorf("failed to save jobs: %w", err)
	}

	s.logger.Printf("Added one-time reminder job: %s (execute at: %s)", name, executeAt.Format(time.RFC3339))
	return nil
}

// removeJobWithReason removes a job from the scheduler with a specific deletion reason
func (s *CronScheduler) removeJobWithReason(name string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if job exists
	job, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job %s not found", name)
	}

	// Log deletion event with the specified reason
	s.logDeletion(name, job.TaskType, reason)

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

	s.logger.Printf("Removed job: %s (reason: %s)", name, reason)
	return nil
}

// RemoveJob removes a job from the scheduler
func (s *CronScheduler) RemoveJob(name string) error {
	return s.removeJobWithReason(name, "manual_deletion")
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
	
	// Quarterly operations run on the first Monday of a new quarter
	// Quarters start on: Jan 1, Apr 1, Jul 1, Oct 1
	month := now.Month()
	
	// Check if we're in a quarter-starting month
	isQuarterMonth := month == time.January || month == time.April || month == time.July || month == time.October
	if !isQuarterMonth {
		return false
	}
	
	// Check if this is the first Monday of the month
	// We need to find the first Monday that occurs in the first 7 days of the quarter month
	day := now.Day()
	weekday := now.Weekday()
	
	// If today is Monday and it's within the first 7 days of the quarter month
	if weekday == time.Monday && day <= 7 {
		return true
	}
	
	return false
}
