package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SummaryManager handles all summarization operations
type SummaryManager struct {
	memoryBasePath    string
	sessionManager    *SessionManager
	tokenThreshold    int
	llmClient         LLMClient      // Interface for LLM integration
	topicManager      *TopicManager  // Topic management for extraction
}

// LLMClient interface for generating summaries using LLM
type LLMClient interface {
	GenerateSummary(content string, summaryType string) (string, error)
	ExtractTopics(content string, existingTopics []string) ([]TopicExtraction, error)
}

// TopicExtraction represents extracted topic information
type TopicExtraction struct {
	TopicName   string
	Content     string
	Confidence  float64
	ShouldWrite bool // true if relevant domain knowledge found
}

// NewSummaryManager creates a new summary manager
func NewSummaryManager(memoryBasePath string, sessionManager *SessionManager, tokenThreshold int, llmClient LLMClient, topicManager *TopicManager) *SummaryManager {
	return &SummaryManager{
		memoryBasePath: memoryBasePath,
		sessionManager: sessionManager,
		tokenThreshold: tokenThreshold,
		llmClient:      llmClient,
		topicManager:   topicManager,
	}
}

// GenerateDailySummary generates a summary of all sessions from a specific date
// Creates daily-summary.md in the daily folder
func (sm *SummaryManager) GenerateDailySummary(date time.Time) error {
	dateStr := date.Format("2006-01-02")
	dailyFolderPath := filepath.Join(sm.memoryBasePath, "chat", dateStr)

	// Check if daily folder exists
	if _, err := os.Stat(dailyFolderPath); os.IsNotExist(err) {
		return fmt.Errorf("daily folder does not exist: %s", dailyFolderPath)
	}

	// Read all session logs from the daily folder
	sessionLogs, err := sm.readAllSessionLogs(dailyFolderPath)
	if err != nil {
		return fmt.Errorf("failed to read session logs: %w", err)
	}

	if len(sessionLogs) == 0 {
		return fmt.Errorf("no session logs found for date: %s", dateStr)
	}

	// Generate summary using LLM
	summary, err := sm.llmClient.GenerateSummary(sessionLogs, "daily")
	if err != nil {
		return fmt.Errorf("failed to generate daily summary: %w", err)
	}

	// Write summary to daily-summary.md
	summaryPath := filepath.Join(dailyFolderPath, "daily-summary.md")
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		return fmt.Errorf("failed to write daily summary: %w", err)
	}

	return nil
}

// GenerateWeeklySummary generates a summary of all daily summaries from a specific week
// Creates summary.md in the week folder
func (sm *SummaryManager) GenerateWeeklySummary(weekNum int, year int) error {
	weekFolderName := fmt.Sprintf("week-%02d-%d", weekNum, year)
	weekFolderPath := filepath.Join(sm.memoryBasePath, "chat", weekFolderName)

	// Check if week folder exists
	if _, err := os.Stat(weekFolderPath); os.IsNotExist(err) {
		return fmt.Errorf("week folder does not exist: %s", weekFolderPath)
	}

	// Read all daily summaries from the week folder
	dailySummaries, err := sm.readAllDailySummaries(weekFolderPath)
	if err != nil {
		return fmt.Errorf("failed to read daily summaries: %w", err)
	}

	if len(dailySummaries) == 0 {
		return fmt.Errorf("no daily summaries found for week: %s", weekFolderName)
	}

	// Generate summary using LLM
	summary, err := sm.llmClient.GenerateSummary(dailySummaries, "weekly")
	if err != nil {
		return fmt.Errorf("failed to generate weekly summary: %w", err)
	}

	// Write summary to summary.md
	summaryPath := filepath.Join(weekFolderPath, "summary.md")
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		return fmt.Errorf("failed to write weekly summary: %w", err)
	}

	return nil
}

// GenerateQuarterlySummary generates a summary of all weekly summaries from a specific quarter
// Creates summary.md in the quarter folder
func (sm *SummaryManager) GenerateQuarterlySummary(quarter int, year int) error {
	quarterFolderName := fmt.Sprintf("Q%d-%d", quarter, year)
	quarterFolderPath := filepath.Join(sm.memoryBasePath, "chat", quarterFolderName)

	// Check if quarter folder exists
	if _, err := os.Stat(quarterFolderPath); os.IsNotExist(err) {
		return fmt.Errorf("quarter folder does not exist: %s", quarterFolderPath)
	}

	// Read all weekly summaries from the quarter folder
	weeklySummaries, err := sm.readAllWeeklySummaries(quarterFolderPath)
	if err != nil {
		return fmt.Errorf("failed to read weekly summaries: %w", err)
	}

	if len(weeklySummaries) == 0 {
		return fmt.Errorf("no weekly summaries found for quarter: %s", quarterFolderName)
	}

	// Generate summary using LLM
	summary, err := sm.llmClient.GenerateSummary(weeklySummaries, "quarterly")
	if err != nil {
		return fmt.Errorf("failed to generate quarterly summary: %w", err)
	}

	// Write summary to summary.md
	summaryPath := filepath.Join(quarterFolderPath, "summary.md")
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		return fmt.Errorf("failed to write quarterly summary: %w", err)
	}

	return nil
}

// GenerateSessionSummary generates a summary of the current session
// Creates session-NNN-summary.md corresponding to the session number
func (sm *SummaryManager) GenerateSessionSummary() error {
	currentSessionPath := sm.sessionManager.GetCurrentSessionPath()
	if currentSessionPath == "" {
		return fmt.Errorf("no active session to summarize")
	}

	// Read the current session log
	sessionContent, err := os.ReadFile(currentSessionPath)
	if err != nil {
		return fmt.Errorf("failed to read session log: %w", err)
	}

	if len(sessionContent) == 0 {
		return fmt.Errorf("session log is empty")
	}

	// Generate summary using LLM
	summary, err := sm.llmClient.GenerateSummary(string(sessionContent), "session")
	if err != nil {
		return fmt.Errorf("failed to generate session summary: %w", err)
	}

	// Create session summary file path
	sessionNum := sm.sessionManager.GetCurrentSessionNumber()
	currentDate := sm.sessionManager.GetCurrentDate()
	dailyFolderPath := filepath.Join(sm.memoryBasePath, "chat", currentDate)
	summaryPath := filepath.Join(dailyFolderPath, fmt.Sprintf("session-%03d-summary.md", sessionNum))

	// Write summary to session-NNN-summary.md
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		return fmt.Errorf("failed to write session summary: %w", err)
	}

	return nil
}

// ShouldTriggerTokenBasedSummarization checks if token usage exceeds threshold
func (sm *SummaryManager) ShouldTriggerTokenBasedSummarization(currentTokenUsage int) bool {
	return currentTokenUsage > sm.tokenThreshold
}

// PerformTokenBasedSummarization performs emergency summarization when token limit is reached
// This maintains session continuity by clearing context and inserting summary
func (sm *SummaryManager) PerformTokenBasedSummarization() error {
	// Generate session summary
	if err := sm.GenerateSessionSummary(); err != nil {
		return fmt.Errorf("failed to generate session summary during token-based summarization: %w", err)
	}

	// Read the generated summary
	sessionNum := sm.sessionManager.GetCurrentSessionNumber()
	currentDate := sm.sessionManager.GetCurrentDate()
	dailyFolderPath := filepath.Join(sm.memoryBasePath, "chat", currentDate)
	summaryPath := filepath.Join(dailyFolderPath, fmt.Sprintf("session-%03d-summary.md", sessionNum))

	summaryContent, err := os.ReadFile(summaryPath)
	if err != nil {
		return fmt.Errorf("failed to read generated summary: %w", err)
	}

	// Note: The actual session context clearing and summary insertion
	// will be handled by the calling code (e.g., LLM client or bot handler)
	// This method just ensures the summary is generated and available

	// Log the token-based summarization
	fmt.Printf("[%s] Token-based summarization completed for session %d\n",
		time.Now().Format("2006-01-02 15:04:05"), sessionNum)

	// Return the summary content for the caller to use
	_ = summaryContent // Will be used by caller

	return nil
}

// ExtractTopicsFromContent extracts topics from the provided content using LLM
// This is called during manual session reset and scheduled summarization
func (sm *SummaryManager) ExtractTopicsFromContent(content string) error {
	if sm.llmClient == nil {
		return fmt.Errorf("LLM client not configured")
	}

	if sm.topicManager == nil {
		return fmt.Errorf("topic manager not configured")
	}

	// Get list of existing topics
	existingTopics := []string{}
	topicInfos, err := sm.topicManager.ListTopics()
	if err == nil {
		for _, info := range topicInfos {
			existingTopics = append(existingTopics, info.Name)
		}
	}

	// Extract topics using LLM
	topics, err := sm.llmClient.ExtractTopics(content, existingTopics)
	if err != nil {
		return fmt.Errorf("failed to extract topics: %w", err)
	}

	// Use TopicManager to write topics to files
	// TopicManager.ExtractTopics handles filtering, logging, and file operations
	if err := sm.topicManager.ExtractTopics(topics); err != nil {
		return fmt.Errorf("failed to write topics: %w", err)
	}

	return nil
}

// readAllSessionLogs reads all session log files from a daily folder
func (sm *SummaryManager) readAllSessionLogs(dailyFolderPath string) (string, error) {
	var allLogs string

	entries, err := os.ReadDir(dailyFolderPath)
	if err != nil {
		return "", fmt.Errorf("failed to read daily folder: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only read session-*.log files
		if filepath.Ext(entry.Name()) == ".log" {
			logPath := filepath.Join(dailyFolderPath, entry.Name())
			content, err := os.ReadFile(logPath)
			if err != nil {
				return "", fmt.Errorf("failed to read log file %s: %w", entry.Name(), err)
			}
			allLogs += string(content) + "\n"
		}
	}

	return allLogs, nil
}

// readAllDailySummaries reads all daily-summary.md files from subdirectories in a week folder
func (sm *SummaryManager) readAllDailySummaries(weekFolderPath string) (string, error) {
	var allSummaries string

	entries, err := os.ReadDir(weekFolderPath)
	if err != nil {
		return "", fmt.Errorf("failed to read week folder: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Read daily-summary.md from each daily folder
		summaryPath := filepath.Join(weekFolderPath, entry.Name(), "daily-summary.md")
		if _, err := os.Stat(summaryPath); err == nil {
			content, err := os.ReadFile(summaryPath)
			if err != nil {
				return "", fmt.Errorf("failed to read daily summary %s: %w", summaryPath, err)
			}
			allSummaries += string(content) + "\n\n"
		}
	}

	return allSummaries, nil
}

// readAllWeeklySummaries reads all summary.md files from subdirectories in a quarter folder
func (sm *SummaryManager) readAllWeeklySummaries(quarterFolderPath string) (string, error) {
	var allSummaries string

	entries, err := os.ReadDir(quarterFolderPath)
	if err != nil {
		return "", fmt.Errorf("failed to read quarter folder: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Read summary.md from each week folder
		summaryPath := filepath.Join(quarterFolderPath, entry.Name(), "summary.md")
		if _, err := os.Stat(summaryPath); err == nil {
			content, err := os.ReadFile(summaryPath)
			if err != nil {
				return "", fmt.Errorf("failed to read weekly summary %s: %w", summaryPath, err)
			}
			allSummaries += string(content) + "\n\n"
		}
	}

	return allSummaries, nil
}
