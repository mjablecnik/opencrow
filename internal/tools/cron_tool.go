package tools

import (
	"encoding/json"
	"fmt"
	"time"

	"simple-telegram-chatbot/internal/scheduler"
)

// CronManagementTool provides LLM-friendly methods for managing cron jobs
type CronManagementTool struct {
	scheduler *scheduler.CronScheduler
}

// CronToolResult represents the result of a cron tool operation with structured data
type CronToolResult struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// toToolResult converts CronToolResult to ToolResult by encoding data as JSON
func (r *CronToolResult) toToolResult() ToolResult {
	if !r.Success {
		return ToolResult{
			Success: false,
			Error:   r.Message,
		}
	}

	// Encode the full result as JSON for the Output field
	output, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to encode result: %v", err),
		}
	}

	return ToolResult{
		Success: true,
		Output:  string(output),
	}
}

// NewCronManagementTool creates a new CronManagementTool instance
func NewCronManagementTool(scheduler *scheduler.CronScheduler) *CronManagementTool {
	return &CronManagementTool{
		scheduler: scheduler,
	}
}

// AddJob adds a new cron job with the specified schedule and task type
func (t *CronManagementTool) AddJob(name, schedule, taskType string) *CronToolResult {
	// Validate cron expression
	if err := t.scheduler.ValidateCronExpression(schedule); err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Invalid cron expression: %v", err),
		}
	}

	// Add the job
	if err := t.scheduler.AddJob(name, schedule, nil); err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to add job: %v", err),
		}
	}

	// Get job info to return next execution time
	job, err := t.scheduler.GetJob(name)
	if err != nil {
		return &CronToolResult{
			Success: true,
			Message: fmt.Sprintf("Job '%s' added successfully", name),
		}
	}

	return &CronToolResult{
		Success: true,
		Message: fmt.Sprintf("Job '%s' added successfully. Next execution: %s", name, job.NextRun.Format(time.RFC3339)),
		Data: map[string]interface{}{
			"name":      job.Name,
			"schedule":  job.Schedule,
			"task_type": job.TaskType,
			"enabled":   job.Enabled,
			"next_run":  job.NextRun.Format(time.RFC3339),
		},
	}
}

// RemoveJob removes a cron job by name
func (t *CronManagementTool) RemoveJob(name string) *CronToolResult {
	// Check if job exists
	_, err := t.scheduler.GetJob(name)
	if err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Job '%s' not found", name),
		}
	}

	// Remove the job
	if err := t.scheduler.RemoveJob(name); err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to remove job: %v", err),
		}
	}

	return &CronToolResult{
		Success: true,
		Message: fmt.Sprintf("Job '%s' removed successfully", name),
	}
}

// EnableJob enables a disabled cron job
func (t *CronManagementTool) EnableJob(name string) *CronToolResult {
	// Check if job exists
	job, err := t.scheduler.GetJob(name)
	if err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Job '%s' not found", name),
		}
	}

	// Check if already enabled
	if job.Enabled {
		return &CronToolResult{
			Success: true,
			Message: fmt.Sprintf("Job '%s' is already enabled", name),
			Data: map[string]interface{}{
				"name":     job.Name,
				"enabled":  job.Enabled,
				"next_run": job.NextRun.Format(time.RFC3339),
			},
		}
	}

	// Enable the job
	if err := t.scheduler.EnableJob(name); err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to enable job: %v", err),
		}
	}

	// Get updated job info
	job, _ = t.scheduler.GetJob(name)

	return &CronToolResult{
		Success: true,
		Message: fmt.Sprintf("Job '%s' enabled successfully. Next execution: %s", name, job.NextRun.Format(time.RFC3339)),
		Data: map[string]interface{}{
			"name":     job.Name,
			"enabled":  job.Enabled,
			"next_run": job.NextRun.Format(time.RFC3339),
		},
	}
}

// DisableJob disables an enabled cron job
func (t *CronManagementTool) DisableJob(name string) *CronToolResult {
	// Check if job exists
	job, err := t.scheduler.GetJob(name)
	if err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Job '%s' not found", name),
		}
	}

	// Check if already disabled
	if !job.Enabled {
		return &CronToolResult{
			Success: true,
			Message: fmt.Sprintf("Job '%s' is already disabled", name),
			Data: map[string]interface{}{
				"name":    job.Name,
				"enabled": job.Enabled,
			},
		}
	}

	// Disable the job
	if err := t.scheduler.DisableJob(name); err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to disable job: %v", err),
		}
	}

	return &CronToolResult{
		Success: true,
		Message: fmt.Sprintf("Job '%s' disabled successfully", name),
		Data: map[string]interface{}{
			"name":    name,
			"enabled": false,
		},
	}
}

// ListJobs returns a list of all cron jobs
func (t *CronManagementTool) ListJobs() *CronToolResult {
	jobs, err := t.scheduler.ListJobs()
	if err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to list jobs: %v", err),
		}
	}

	if len(jobs) == 0 {
		return &CronToolResult{
			Success: true,
			Message: "No jobs configured",
			Data:    []interface{}{},
		}
	}

	// Format job data for LLM consumption
	jobList := make([]map[string]interface{}, 0, len(jobs))
	for _, job := range jobs {
		jobData := map[string]interface{}{
			"name":        job.Name,
			"task_type":   job.TaskType,
			"enabled":     job.Enabled,
			"status":      string(job.Status),
			"description": job.Description,
		}

		// Add schedule or execute_at
		if job.Schedule != "" {
			jobData["schedule"] = job.Schedule
			if !job.NextRun.IsZero() {
				jobData["next_run"] = job.NextRun.Format(time.RFC3339)
			}
		} else if job.ExecuteAt != nil {
			jobData["execute_at"] = job.ExecuteAt.Format(time.RFC3339)
		}

		// Add lifecycle fields if present
		if job.StartsAt != nil {
			jobData["starts_at"] = job.StartsAt.Format(time.RFC3339)
		}
		if job.PausedUntil != nil {
			jobData["paused_until"] = job.PausedUntil.Format(time.RFC3339)
		}
		if job.ExpiresAt != nil {
			jobData["expires_at"] = job.ExpiresAt.Format(time.RFC3339)
		}

		// Add reminder-specific fields
		if job.TaskType == "reminder" && job.Message != "" {
			jobData["message"] = job.Message
			jobData["chat_id"] = job.ChatID
		}

		// Add last run info
		if !job.LastRun.IsZero() {
			jobData["last_run"] = job.LastRun.Format(time.RFC3339)
		}
		if job.LastError != "" {
			jobData["last_error"] = job.LastError
		}

		jobList = append(jobList, jobData)
	}

	return &CronToolResult{
		Success: true,
		Message: fmt.Sprintf("Found %d job(s)", len(jobs)),
		Data:    jobList,
	}
}

// GetJobInfo returns detailed information about a specific job
func (t *CronManagementTool) GetJobInfo(name string) *CronToolResult {
	job, err := t.scheduler.GetJob(name)
	if err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Job '%s' not found", name),
		}
	}

	jobData := map[string]interface{}{
		"name":        job.Name,
		"task_type":   job.TaskType,
		"enabled":     job.Enabled,
		"status":      string(job.Status),
		"description": job.Description,
	}

	// Add schedule or execute_at
	if job.Schedule != "" {
		jobData["schedule"] = job.Schedule
		if !job.NextRun.IsZero() {
			jobData["next_run"] = job.NextRun.Format(time.RFC3339)
		}
	} else if job.ExecuteAt != nil {
		jobData["execute_at"] = job.ExecuteAt.Format(time.RFC3339)
	}

	// Add lifecycle fields if present
	if job.StartsAt != nil {
		jobData["starts_at"] = job.StartsAt.Format(time.RFC3339)
	}
	if job.PausedUntil != nil {
		jobData["paused_until"] = job.PausedUntil.Format(time.RFC3339)
	}
	if job.ExpiresAt != nil {
		jobData["expires_at"] = job.ExpiresAt.Format(time.RFC3339)
	}

	// Add reminder-specific fields
	if job.TaskType == "reminder" && job.Message != "" {
		jobData["message"] = job.Message
		jobData["chat_id"] = job.ChatID
	}

	// Add last run info
	if !job.LastRun.IsZero() {
		jobData["last_run"] = job.LastRun.Format(time.RFC3339)
	}
	if job.LastError != "" {
		jobData["last_error"] = job.LastError
	}

	return &CronToolResult{
		Success: true,
		Message: fmt.Sprintf("Job '%s' details retrieved", name),
		Data:    jobData,
	}
}

// CreateRecurringReminder creates a recurring reminder with optional lifecycle parameters
func (t *CronManagementTool) CreateRecurringReminder(name, schedule, message string, chatID int64, startsAt, expiresAt *time.Time) *CronToolResult {
	// Validate inputs
	if name == "" {
		return &CronToolResult{
			Success: false,
			Message: "Job name cannot be empty",
		}
	}

	if message == "" {
		return &CronToolResult{
			Success: false,
			Message: "Reminder message cannot be empty",
		}
	}

	// Validate cron expression
	if err := t.scheduler.ValidateCronExpression(schedule); err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Invalid cron expression: %v", err),
		}
	}

	// Validate lifecycle timestamps
	now := time.Now()
	if startsAt != nil && startsAt.Before(now) {
		return &CronToolResult{
			Success: false,
			Message: "starts_at must be in the future",
		}
	}

	if expiresAt != nil && expiresAt.Before(now) {
		return &CronToolResult{
			Success: false,
			Message: "expires_at must be in the future",
		}
	}

	if startsAt != nil && expiresAt != nil && expiresAt.Before(*startsAt) {
		return &CronToolResult{
			Success: false,
			Message: "expires_at must be after starts_at",
		}
	}

	// Create the reminder job directly with all required fields
	// This ensures task_type is set from the beginning
	if err := t.scheduler.AddReminderJob(name, schedule, message, chatID, startsAt, expiresAt); err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create reminder: %v", err),
		}
	}

	// Get the job to return details
	job, err := t.scheduler.GetJob(name)
	if err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to retrieve created job: %v", err),
		}
	}

	// Build response data
	responseData := map[string]interface{}{
		"name":      job.Name,
		"schedule":  job.Schedule,
		"task_type": job.TaskType,
		"message":   job.Message,
		"chat_id":   job.ChatID,
		"enabled":   job.Enabled,
		"next_run":  job.NextRun.Format(time.RFC3339),
	}

	if startsAt != nil {
		responseData["starts_at"] = startsAt.Format(time.RFC3339)
	}
	if expiresAt != nil {
		responseData["expires_at"] = expiresAt.Format(time.RFC3339)
	}

	return &CronToolResult{
		Success: true,
		Message: fmt.Sprintf("Recurring reminder '%s' created successfully. Next execution: %s", name, job.NextRun.Format(time.RFC3339)),
		Data:    responseData,
	}
}

// CreateOneTimeReminder creates a one-time reminder that executes at a specific timestamp
func (t *CronManagementTool) CreateOneTimeReminder(name string, executeAt time.Time, message string, chatID int64) *CronToolResult {
	// Validate inputs
	if name == "" {
		return &CronToolResult{
			Success: false,
			Message: "Job name cannot be empty",
		}
	}

	if message == "" {
		return &CronToolResult{
			Success: false,
			Message: "Reminder message cannot be empty",
		}
	}

	// Validate execute_at is in the future
	if executeAt.Before(time.Now()) {
		return &CronToolResult{
			Success: false,
			Message: "execute_at must be in the future",
		}
	}

	// Create a cron expression that matches the specific time
	// Format: "minute hour day month weekday"
	cronExpr := fmt.Sprintf("%d %d %d %d *",
		executeAt.Minute(),
		executeAt.Hour(),
		executeAt.Day(),
		int(executeAt.Month()),
	)

	// Create the one-time reminder job directly with all required fields
	if err := t.scheduler.AddOneTimeReminderJob(name, cronExpr, executeAt, message, chatID); err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create one-time reminder: %v", err),
		}
	}

	// Get the job to return details
	job, err := t.scheduler.GetJob(name)
	if err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to retrieve created job: %v", err),
		}
	}

	return &CronToolResult{
		Success: true,
		Message: fmt.Sprintf("One-time reminder '%s' created successfully. Will execute at: %s", name, executeAt.Format(time.RFC3339)),
		Data: map[string]interface{}{
			"name":       job.Name,
			"task_type":  job.TaskType,
			"execute_at": executeAt.Format(time.RFC3339),
			"message":    job.Message,
			"chat_id":    job.ChatID,
			"enabled":    job.Enabled,
		},
	}
}

// UpdateReminderMessage updates the message content of an existing reminder
func (t *CronManagementTool) UpdateReminderMessage(name, newMessage string) *CronToolResult {
	// Validate inputs
	if name == "" {
		return &CronToolResult{
			Success: false,
			Message: "Job name cannot be empty",
		}
	}

	if newMessage == "" {
		return &CronToolResult{
			Success: false,
			Message: "New message cannot be empty",
		}
	}

	// Get the job
	job, err := t.scheduler.GetJob(name)
	if err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Job '%s' not found", name),
		}
	}

	// Verify it's a reminder job
	if job.TaskType != "reminder" {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Job '%s' is not a reminder (task_type: %s)", name, job.TaskType),
		}
	}

	// Update the message
	oldMessage := job.Message
	job.Message = newMessage

	// Save the updated configuration
	if err := t.scheduler.SaveJobs(); err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to save updated message: %v", err),
		}
	}

	return &CronToolResult{
		Success: true,
		Message: fmt.Sprintf("Reminder message updated successfully for job '%s'", name),
		Data: map[string]interface{}{
			"name":        name,
			"old_message": oldMessage,
			"new_message": newMessage,
		},
	}
}

// PauseJob temporarily suspends a job until the specified time
func (t *CronManagementTool) PauseJob(name string, pausedUntil time.Time) *CronToolResult {
	// Validate inputs
	if name == "" {
		return &CronToolResult{
			Success: false,
			Message: "Job name cannot be empty",
		}
	}

	// Validate pausedUntil is in the future
	if pausedUntil.Before(time.Now()) {
		return &CronToolResult{
			Success: false,
			Message: "paused_until must be in the future",
		}
	}

	// Get the job
	job, err := t.scheduler.GetJob(name)
	if err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Job '%s' not found", name),
		}
	}

	// Check if job is enabled
	if !job.Enabled {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Job '%s' is disabled. Enable it first before pausing.", name),
		}
	}

	// Update the paused_until field
	job.PausedUntil = &pausedUntil

	// Save the updated configuration
	if err := t.scheduler.SaveJobs(); err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to save pause configuration: %v", err),
		}
	}

	return &CronToolResult{
		Success: true,
		Message: fmt.Sprintf("Job '%s' paused until %s. It will automatically resume after this time.", name, pausedUntil.Format(time.RFC3339)),
		Data: map[string]interface{}{
			"name":          name,
			"paused_until":  pausedUntil.Format(time.RFC3339),
			"will_resume":   true,
		},
	}
}

// ResumeJob immediately resumes a paused job
func (t *CronManagementTool) ResumeJob(name string) *CronToolResult {
	// Validate inputs
	if name == "" {
		return &CronToolResult{
			Success: false,
			Message: "Job name cannot be empty",
		}
	}

	// Get the job
	job, err := t.scheduler.GetJob(name)
	if err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Job '%s' not found", name),
		}
	}

	// Check if job is actually paused
	if job.PausedUntil == nil {
		return &CronToolResult{
			Success: true,
			Message: fmt.Sprintf("Job '%s' is not paused", name),
			Data: map[string]interface{}{
				"name":   name,
				"paused": false,
			},
		}
	}

	// Clear the paused_until field
	job.PausedUntil = nil

	// Save the updated configuration
	if err := t.scheduler.SaveJobs(); err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to save resume configuration: %v", err),
		}
	}

	return &CronToolResult{
		Success: true,
		Message: fmt.Sprintf("Job '%s' resumed successfully. Next execution: %s", name, job.NextRun.Format(time.RFC3339)),
		Data: map[string]interface{}{
			"name":     name,
			"paused":   false,
			"next_run": job.NextRun.Format(time.RFC3339),
		},
	}
}

// ExtendExpiration extends or sets the expiration time for a job
func (t *CronManagementTool) ExtendExpiration(name string, newExpiresAt time.Time) *CronToolResult {
	// Validate inputs
	if name == "" {
		return &CronToolResult{
			Success: false,
			Message: "Job name cannot be empty",
		}
	}

	// Validate newExpiresAt is in the future
	if newExpiresAt.Before(time.Now()) {
		return &CronToolResult{
			Success: false,
			Message: "new expires_at must be in the future",
		}
	}

	// Get the job
	job, err := t.scheduler.GetJob(name)
	if err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Job '%s' not found", name),
		}
	}

	// Validate against starts_at if present
	if job.StartsAt != nil && newExpiresAt.Before(*job.StartsAt) {
		return &CronToolResult{
			Success: false,
			Message: "new expires_at must be after starts_at",
		}
	}

	// Store old expiration for response
	var oldExpiresAt string
	if job.ExpiresAt != nil {
		oldExpiresAt = job.ExpiresAt.Format(time.RFC3339)
	} else {
		oldExpiresAt = "none"
	}

	// Update the expires_at field
	job.ExpiresAt = &newExpiresAt

	// Save the updated configuration
	if err := t.scheduler.SaveJobs(); err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to save expiration configuration: %v", err),
		}
	}

	return &CronToolResult{
		Success: true,
		Message: fmt.Sprintf("Job '%s' expiration extended to %s", name, newExpiresAt.Format(time.RFC3339)),
		Data: map[string]interface{}{
			"name":            name,
			"old_expires_at":  oldExpiresAt,
			"new_expires_at":  newExpiresAt.Format(time.RFC3339),
		},
	}
}

// GetExecutionHistory retrieves execution history for a specific job with optional filtering and limiting
func (t *CronManagementTool) GetExecutionHistory(name string, limit int) *CronToolResult {
	// Validate inputs
	if name == "" {
		return &CronToolResult{
			Success: false,
			Message: "Job name cannot be empty",
		}
	}

	// Set default limit if not provided or invalid
	if limit <= 0 {
		limit = 10 // Default to last 10 entries
	}

	// Get execution history from scheduler
	history, err := t.scheduler.GetExecutionHistory(name, limit)
	if err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to retrieve execution history: %v", err),
		}
	}

	if len(history) == 0 {
		return &CronToolResult{
			Success: true,
			Message: fmt.Sprintf("No execution history found for job '%s'", name),
			Data:    []interface{}{},
		}
	}

	// Format history entries for LLM consumption
	historyList := make([]map[string]interface{}, 0, len(history))
	for _, entry := range history {
		historyData := map[string]interface{}{
			"timestamp": entry.Timestamp.Format(time.RFC3339),
			"job_name":  entry.JobName,
			"task_type": entry.TaskType,
			"status":    entry.Status,
		}

		// Add error if present
		if entry.Error != "" {
			historyData["error"] = entry.Error
		}

		// Add reminder-specific fields
		if entry.MessageSent != "" {
			historyData["message_sent"] = entry.MessageSent
			historyData["chat_id"] = entry.ChatID
		}

		// Add cascade operation details
		if len(entry.OperationsCompleted) > 0 {
			historyData["operations_completed"] = entry.OperationsCompleted
		}

		// Add deletion info
		if entry.Deleted {
			historyData["deleted"] = true
			historyData["deletion_reason"] = entry.DeletionReason
		}

		historyList = append(historyList, historyData)
	}

	return &CronToolResult{
		Success: true,
		Message: fmt.Sprintf("Retrieved %d execution history entries for job '%s' (limited to %d)", len(history), name, limit),
		Data:    historyList,
	}
}

// CleanupExpiredJobs removes all one-time jobs that have passed their execution time
func (t *CronManagementTool) CleanupExpiredJobs() *CronToolResult {
	count, err := t.scheduler.CleanupExpiredJobs()
	if err != nil {
		return &CronToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to cleanup expired jobs: %v", err),
		}
	}

	if count == 0 {
		return &CronToolResult{
			Success: true,
			Message: "No expired one-time jobs found",
			Data:    map[string]interface{}{"removed_count": 0},
		}
	}

	return &CronToolResult{
		Success: true,
		Message: fmt.Sprintf("Successfully removed %d expired one-time job(s)", count),
		Data:    map[string]interface{}{"removed_count": count},
	}
}

// Name returns the tool name for registration
func (t *CronManagementTool) Name() string {
	return "cron_management"
}

// Description returns the tool description for LLM
func (t *CronManagementTool) Description() string {
	return "Manage cron jobs for scheduled tasks and reminders. Supports adding, removing, enabling, disabling, and listing jobs. Can create recurring reminders with cron expressions or one-time reminders with specific timestamps."
}

// Execute implements the Tool interface for generic tool execution
func (t *CronManagementTool) Execute(params map[string]interface{}) (ToolResult, error) {
	action, ok := params["action"].(string)
	if !ok {
		return ToolResult{
			Success: false,
			Error:   "missing or invalid 'action' parameter",
		}, fmt.Errorf("missing or invalid 'action' parameter")
	}

	switch action {
	case "list":
		return t.ListJobs().toToolResult(), nil
		
	case "get":
		name, ok := params["name"].(string)
		if !ok {
			return ToolResult{Success: false, Error: "missing 'name' parameter"}, fmt.Errorf("missing 'name' parameter")
		}
		return t.GetJobInfo(name).toToolResult(), nil
		
	case "add":
		name, _ := params["name"].(string)
		schedule, _ := params["schedule"].(string)
		taskType, _ := params["task_type"].(string)
		return t.AddJob(name, schedule, taskType).toToolResult(), nil
		
	case "remove":
		name, _ := params["name"].(string)
		return t.RemoveJob(name).toToolResult(), nil
		
	case "enable":
		name, _ := params["name"].(string)
		return t.EnableJob(name).toToolResult(), nil
		
	case "disable":
		name, _ := params["name"].(string)
		return t.DisableJob(name).toToolResult(), nil
		
	case "create_recurring_reminder":
		name, _ := params["name"].(string)
		schedule, _ := params["schedule"].(string)
		message, _ := params["message"].(string)
		
		// Get chatID from params or use 0 as placeholder (will be set by scheduler from context)
		var chatID int64
		if chatIDFloat, ok := params["chat_id"].(float64); ok {
			chatID = int64(chatIDFloat)
		} else if chatIDInt, ok := params["chat_id"].(int64); ok {
			chatID = chatIDInt
		}
		// If chatID is 0, the scheduler will need to determine it from context
		
		var startsAt, expiresAt *time.Time
		if startsAtStr, ok := params["starts_at"].(string); ok && startsAtStr != "" {
			if parsed, err := time.Parse(time.RFC3339, startsAtStr); err == nil {
				startsAt = &parsed
			}
		}
		if expiresAtStr, ok := params["expires_at"].(string); ok && expiresAtStr != "" {
			if parsed, err := time.Parse(time.RFC3339, expiresAtStr); err == nil {
				expiresAt = &parsed
			}
		}
		
		return t.CreateRecurringReminder(name, schedule, message, chatID, startsAt, expiresAt).toToolResult(), nil
		
	case "create_onetime_reminder":
		name, _ := params["name"].(string)
		message, _ := params["message"].(string)
		executeAtStr, ok := params["execute_at"].(string)
		if !ok {
			return ToolResult{Success: false, Error: "missing 'execute_at' parameter"}, fmt.Errorf("missing 'execute_at' parameter")
		}
		
		executeAt, err := time.Parse(time.RFC3339, executeAtStr)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("invalid execute_at format: %v", err)}, err
		}
		
		// Get chatID from params or use 0 as placeholder
		var chatID int64
		if chatIDFloat, ok := params["chat_id"].(float64); ok {
			chatID = int64(chatIDFloat)
		} else if chatIDInt, ok := params["chat_id"].(int64); ok {
			chatID = chatIDInt
		}
		
		return t.CreateOneTimeReminder(name, executeAt, message, chatID).toToolResult(), nil
		
	case "pause":
		name, _ := params["name"].(string)
		pausedUntilStr, ok := params["paused_until"].(string)
		if !ok {
			return ToolResult{Success: false, Error: "missing 'paused_until' parameter"}, fmt.Errorf("missing 'paused_until' parameter")
		}
		
		pausedUntil, err := time.Parse(time.RFC3339, pausedUntilStr)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("invalid paused_until format: %v", err)}, err
		}
		
		return t.PauseJob(name, pausedUntil).toToolResult(), nil
		
	case "resume":
		name, _ := params["name"].(string)
		return t.ResumeJob(name).toToolResult(), nil
		
	case "extend_expiration":
		name, _ := params["name"].(string)
		newExpiresAtStr, ok := params["new_expires_at"].(string)
		if !ok {
			return ToolResult{Success: false, Error: "missing 'new_expires_at' parameter"}, fmt.Errorf("missing 'new_expires_at' parameter")
		}
		
		newExpiresAt, err := time.Parse(time.RFC3339, newExpiresAtStr)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("invalid new_expires_at format: %v", err)}, err
		}
		
		return t.ExtendExpiration(name, newExpiresAt).toToolResult(), nil
		
	case "get_history":
		name, _ := params["name"].(string)
		limit := 10 // default
		if limitFloat, ok := params["limit"].(float64); ok {
			limit = int(limitFloat)
		} else if limitInt, ok := params["limit"].(int); ok {
			limit = limitInt
		}
		
		return t.GetExecutionHistory(name, limit).toToolResult(), nil
		
	case "cleanup_expired":
		return t.CleanupExpiredJobs().toToolResult(), nil
		
	default:
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("unknown action: %s", action),
		}, fmt.Errorf("unknown action: %s", action)
	}
}