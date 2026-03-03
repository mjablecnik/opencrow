package memory

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ReorganizationManager handles weekly and quarterly memory reorganization
type ReorganizationManager struct {
	llmClient      LLMClient
	mu             sync.RWMutex
	memoryBasePath string
	sessionManager *SessionManager
	topicManager   *TopicManager
}

// NewReorganizationManager creates a new ReorganizationManager instance
func NewReorganizationManager(memoryBasePath string, sessionManager *SessionManager, topicManager *TopicManager, llmClient LLMClient) *ReorganizationManager {
	return &ReorganizationManager{
		memoryBasePath: memoryBasePath,
		sessionManager: sessionManager,
		topicManager:   topicManager,
		llmClient:      llmClient,
	}
}

// PerformWeeklyReorganization creates a week folder, moves daily folders from the completed week,
// and generates a weekly summary
func (rm *ReorganizationManager) PerformWeeklyReorganization() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	log.Println("[ReorganizationManager] Starting weekly reorganization")

	// Calculate the completed week (previous 7 days)
	now := time.Now()
	// Get the start of the completed week (7 days ago)
	weekStart := now.AddDate(0, 0, -7)
	weekEnd := now.AddDate(0, 0, -1)

	// Calculate week number and year for the completed week
	year, week := weekStart.ISOWeek()
	weekFolderName := fmt.Sprintf("week-%02d-%d", week, year)

	chatPath := filepath.Join(rm.memoryBasePath, "chat")
	weekFolderPath := filepath.Join(chatPath, weekFolderName)

	// Create week folder
	if err := os.MkdirAll(weekFolderPath, 0755); err != nil {
		log.Printf("[ReorganizationManager] ERROR: Failed to create week folder %s: %v", weekFolderPath, err)
		return fmt.Errorf("failed to create week folder: %w", err)
	}

	log.Printf("[ReorganizationManager] Created week folder: %s", weekFolderName)

	// Find and move all daily folders from the completed week
	dailyFolders, err := rm.findDailyFoldersInRange(chatPath, weekStart, weekEnd)
	if err != nil {
		log.Printf("[ReorganizationManager] ERROR: Failed to find daily folders: %v", err)
		return fmt.Errorf("failed to find daily folders: %w", err)
	}

	if len(dailyFolders) == 0 {
		log.Printf("[ReorganizationManager] WARNING: No daily folders found for week %s", weekFolderName)
		// Continue to generate summary even if no folders found
	}

	// Move daily folders into week folder
	movedFolders := []string{}
	for _, dailyFolder := range dailyFolders {
		sourcePath := filepath.Join(chatPath, dailyFolder)
		destPath := filepath.Join(weekFolderPath, dailyFolder)

		if err := os.Rename(sourcePath, destPath); err != nil {
			log.Printf("[ReorganizationManager] ERROR: Failed to move folder %s to %s: %v", sourcePath, destPath, err)
			// Continue with other folders even if one fails
			continue
		}

		movedFolders = append(movedFolders, dailyFolder)
		log.Printf("[ReorganizationManager] Moved daily folder: %s -> %s", dailyFolder, weekFolderName)
	}

	// Generate weekly summary from daily summaries
	weeklySummary, err := rm.generateWeeklySummary(weekFolderPath, weekStart, weekEnd, year, week)
	if err != nil {
		log.Printf("[ReorganizationManager] ERROR: Failed to generate weekly summary: %v", err)
		return fmt.Errorf("failed to generate weekly summary: %w", err)
	}

	// Write weekly summary to summary.md
	summaryPath := filepath.Join(weekFolderPath, "summary.md")
	if err := os.WriteFile(summaryPath, []byte(weeklySummary), 0644); err != nil {
		log.Printf("[ReorganizationManager] ERROR: Failed to write weekly summary: %v", err)
		return fmt.Errorf("failed to write weekly summary: %w", err)
	}

	log.Printf("[ReorganizationManager] Weekly reorganization completed successfully")
	log.Printf("[ReorganizationManager] Week: %s, Date range: %s to %s, Folders moved: %d",
		weekFolderName,
		weekStart.Format("2006-01-02"),
		weekEnd.Format("2006-01-02"),
		len(movedFolders))

	return nil
}

// PerformQuarterlyReorganization creates a quarter folder, moves week folders from the completed quarter,
// and generates a quarterly summary
func (rm *ReorganizationManager) PerformQuarterlyReorganization() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	log.Println("[ReorganizationManager] Starting quarterly reorganization")

	// Calculate the completed quarter
	now := time.Now()
	currentQuarter := (int(now.Month())-1)/3 + 1
	currentYear := now.Year()

	// Determine the completed quarter
	var completedQuarter int
	var completedYear int

	if currentQuarter == 1 {
		// If we're in Q1, the completed quarter is Q4 of previous year
		completedQuarter = 4
		completedYear = currentYear - 1
	} else {
		// Otherwise, it's the previous quarter of current year
		completedQuarter = currentQuarter - 1
		completedYear = currentYear
	}

	quarterFolderName := fmt.Sprintf("Q%d-%d", completedQuarter, completedYear)
	chatPath := filepath.Join(rm.memoryBasePath, "chat")
	quarterFolderPath := filepath.Join(chatPath, quarterFolderName)

	// Create quarter folder
	if err := os.MkdirAll(quarterFolderPath, 0755); err != nil {
		log.Printf("[ReorganizationManager] ERROR: Failed to create quarter folder %s: %v", quarterFolderPath, err)
		return fmt.Errorf("failed to create quarter folder: %w", err)
	}

	log.Printf("[ReorganizationManager] Created quarter folder: %s", quarterFolderName)

	// Calculate date range for the completed quarter
	quarterStart, quarterEnd := rm.getQuarterDateRange(completedQuarter, completedYear)

	// Find and move all week folders from the completed quarter
	weekFolders, err := rm.findWeekFoldersInRange(chatPath, quarterStart, quarterEnd)
	if err != nil {
		log.Printf("[ReorganizationManager] ERROR: Failed to find week folders: %v", err)
		return fmt.Errorf("failed to find week folders: %w", err)
	}

	if len(weekFolders) == 0 {
		log.Printf("[ReorganizationManager] WARNING: No week folders found for quarter %s", quarterFolderName)
		// Continue to generate summary even if no folders found
	}

	// Move week folders into quarter folder
	movedFolders := []string{}
	for _, weekFolder := range weekFolders {
		sourcePath := filepath.Join(chatPath, weekFolder)
		destPath := filepath.Join(quarterFolderPath, weekFolder)

		if err := os.Rename(sourcePath, destPath); err != nil {
			log.Printf("[ReorganizationManager] ERROR: Failed to move folder %s to %s: %v", sourcePath, destPath, err)
			// Continue with other folders even if one fails
			continue
		}

		movedFolders = append(movedFolders, weekFolder)
		log.Printf("[ReorganizationManager] Moved week folder: %s -> %s", weekFolder, quarterFolderName)
	}

	// Generate quarterly summary from weekly summaries
	quarterlySummary, err := rm.generateQuarterlySummary(quarterFolderPath, quarterStart, quarterEnd, completedQuarter, completedYear)
	if err != nil {
		log.Printf("[ReorganizationManager] ERROR: Failed to generate quarterly summary: %v", err)
		return fmt.Errorf("failed to generate quarterly summary: %w", err)
	}

	// Write quarterly summary to summary.md
	summaryPath := filepath.Join(quarterFolderPath, "summary.md")
	if err := os.WriteFile(summaryPath, []byte(quarterlySummary), 0644); err != nil {
		log.Printf("[ReorganizationManager] ERROR: Failed to write quarterly summary: %v", err)
		return fmt.Errorf("failed to write quarterly summary: %w", err)
	}

	log.Printf("[ReorganizationManager] Quarterly reorganization completed successfully")
	log.Printf("[ReorganizationManager] Quarter: %s, Date range: %s to %s, Folders moved: %d",
		quarterFolderName,
		quarterStart.Format("2006-01-02"),
		quarterEnd.Format("2006-01-02"),
		len(movedFolders))

	return nil
}

// findDailyFoldersInRange finds all daily folders (YYYY-MM-DD format) within the specified date range
func (rm *ReorganizationManager) findDailyFoldersInRange(chatPath string, startDate, endDate time.Time) ([]string, error) {
	entries, err := os.ReadDir(chatPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read chat directory: %w", err)
	}

	var dailyFolders []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folderName := entry.Name()
		// Check if folder name matches YYYY-MM-DD format
		folderDate, err := time.Parse("2006-01-02", folderName)
		if err != nil {
			// Not a daily folder, skip
			continue
		}

		// Check if folder date is within range (inclusive)
		if (folderDate.Equal(startDate) || folderDate.After(startDate)) &&
			(folderDate.Equal(endDate) || folderDate.Before(endDate)) {
			dailyFolders = append(dailyFolders, folderName)
		}
	}

	// Sort folders chronologically
	sort.Strings(dailyFolders)

	return dailyFolders, nil
}

// findWeekFoldersInRange finds all week folders (week-WW-YYYY format) within the specified date range
func (rm *ReorganizationManager) findWeekFoldersInRange(chatPath string, startDate, endDate time.Time) ([]string, error) {
	entries, err := os.ReadDir(chatPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read chat directory: %w", err)
	}

	var weekFolders []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folderName := entry.Name()
		// Check if folder name matches week-WW-YYYY format
		if !strings.HasPrefix(folderName, "week-") {
			continue
		}

		// Parse week number and year from folder name
		var weekNum, year int
		_, err := fmt.Sscanf(folderName, "week-%d-%d", &weekNum, &year)
		if err != nil {
			// Not a valid week folder, skip
			continue
		}

		// Calculate the date range for this week
		weekStart := rm.getWeekStartDate(weekNum, year)
		weekEnd := weekStart.AddDate(0, 0, 6)

		// Check if week overlaps with the quarter date range
		if rm.dateRangesOverlap(weekStart, weekEnd, startDate, endDate) {
			weekFolders = append(weekFolders, folderName)
		}
	}

	// Sort folders by week number and year
	sort.Slice(weekFolders, func(i, j int) bool {
		var week1, year1, week2, year2 int
		fmt.Sscanf(weekFolders[i], "week-%d-%d", &week1, &year1)
		fmt.Sscanf(weekFolders[j], "week-%d-%d", &week2, &year2)

		if year1 != year2 {
			return year1 < year2
		}
		return week1 < week2
	})

	return weekFolders, nil
}

// generateWeeklySummary generates a summary from all daily summaries in the week folder
func (rm *ReorganizationManager) generateWeeklySummary(weekFolderPath string, startDate, endDate time.Time, year, week int) (string, error) {
	// Read all daily summaries from the week folder
	dailySummaries, err := rm.readAllDailySummariesFromWeek(weekFolderPath)
	if err != nil {
		return "", fmt.Errorf("failed to read daily summaries: %w", err)
	}

	if dailySummaries == "" {
		log.Printf("[ReorganizationManager] WARNING: No daily summaries found in week folder")
		dailySummaries = "No daily summaries available for this week."
	}

	// Use LLM to generate weekly summary
	prompt := fmt.Sprintf(`Generate a comprehensive weekly summary for Week %d, %d (Date Range: %s to %s).

Summarize the following daily summaries into a cohesive weekly overview:

%s

Please provide:
1. Overview of the week
2. Major themes and topics discussed
3. Progress and achievements
4. Topics updated (if any)
5. Looking ahead

Format the summary in markdown.`, week, year, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), dailySummaries)

	summary, err := rm.llmClient.GenerateSummary(prompt, "weekly")
	if err != nil {
		log.Printf("[ReorganizationManager] ERROR: LLM failed to generate weekly summary: %v", err)
		return "", fmt.Errorf("failed to generate weekly summary with LLM: %w", err)
	}

	// Format the final summary
	finalSummary := fmt.Sprintf(`# Weekly Summary - Week %d, %d

**Date Range:** %s to %s

%s
`, week, year, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), summary)

	return finalSummary, nil
}

// generateQuarterlySummary generates a summary from all weekly summaries in the quarter folder
func (rm *ReorganizationManager) generateQuarterlySummary(quarterFolderPath string, startDate, endDate time.Time, quarter, year int) (string, error) {
	// Read all weekly summaries from the quarter folder
	weeklySummaries, err := rm.readAllWeeklySummariesFromQuarter(quarterFolderPath)
	if err != nil {
		return "", fmt.Errorf("failed to read weekly summaries: %w", err)
	}

	if weeklySummaries == "" {
		log.Printf("[ReorganizationManager] WARNING: No weekly summaries found in quarter folder")
		weeklySummaries = "No weekly summaries available for this quarter."
	}

	// Use LLM to generate quarterly summary
	prompt := fmt.Sprintf(`Generate a comprehensive quarterly summary for Q%d %d (Date Range: %s to %s).

Summarize the following weekly summaries into a cohesive quarterly overview:

%s

Please provide:
1. Overview of the quarter
2. Major milestones and achievements
3. Knowledge domains developed
4. Personal growth and progress
5. Next quarter goals

Format the summary in markdown.`, quarter, year, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), weeklySummaries)

	summary, err := rm.llmClient.GenerateSummary(prompt, "quarterly")
	if err != nil {
		log.Printf("[ReorganizationManager] ERROR: LLM failed to generate quarterly summary: %v", err)
		return "", fmt.Errorf("failed to generate quarterly summary with LLM: %w", err)
	}

	// Format the final summary
	finalSummary := fmt.Sprintf(`# Quarterly Summary - Q%d %d

**Date Range:** %s to %s

%s
`, quarter, year, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), summary)

	return finalSummary, nil
}

// readAllDailySummariesFromWeek reads all daily-summary.md files from daily folders within a week folder
func (rm *ReorganizationManager) readAllDailySummariesFromWeek(weekFolderPath string) (string, error) {
	entries, err := os.ReadDir(weekFolderPath)
	if err != nil {
		return "", fmt.Errorf("failed to read week folder: %w", err)
	}

	var summaries []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if this is a daily folder (YYYY-MM-DD format)
		_, err := time.Parse("2006-01-02", entry.Name())
		if err != nil {
			continue
		}

		// Read daily-summary.md from this folder
		summaryPath := filepath.Join(weekFolderPath, entry.Name(), "daily-summary.md")
		content, err := os.ReadFile(summaryPath)
		if err != nil {
			log.Printf("[ReorganizationManager] WARNING: Failed to read daily summary from %s: %v", summaryPath, err)
			continue
		}

		summaries = append(summaries, string(content))
	}

	return strings.Join(summaries, "\n\n---\n\n"), nil
}

// readAllWeeklySummariesFromQuarter reads all summary.md files from week folders within a quarter folder
func (rm *ReorganizationManager) readAllWeeklySummariesFromQuarter(quarterFolderPath string) (string, error) {
	entries, err := os.ReadDir(quarterFolderPath)
	if err != nil {
		return "", fmt.Errorf("failed to read quarter folder: %w", err)
	}

	var summaries []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if this is a week folder (week-WW-YYYY format)
		if !strings.HasPrefix(entry.Name(), "week-") {
			continue
		}

		// Read summary.md from this folder
		summaryPath := filepath.Join(quarterFolderPath, entry.Name(), "summary.md")
		content, err := os.ReadFile(summaryPath)
		if err != nil {
			log.Printf("[ReorganizationManager] WARNING: Failed to read weekly summary from %s: %v", summaryPath, err)
			continue
		}

		summaries = append(summaries, string(content))
	}

	return strings.Join(summaries, "\n\n---\n\n"), nil
}

// getQuarterDateRange returns the start and end dates for a given quarter
func (rm *ReorganizationManager) getQuarterDateRange(quarter, year int) (time.Time, time.Time) {
	var startMonth time.Month
	switch quarter {
	case 1:
		startMonth = time.January
	case 2:
		startMonth = time.April
	case 3:
		startMonth = time.July
	case 4:
		startMonth = time.October
	default:
		startMonth = time.January
	}

	startDate := time.Date(year, startMonth, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 3, -1) // Last day of the quarter

	return startDate, endDate
}

// getWeekStartDate returns the start date (Monday) for a given ISO week number and year
func (rm *ReorganizationManager) getWeekStartDate(weekNum, year int) time.Time {
	// Start with January 4th of the year (always in week 1)
	jan4 := time.Date(year, time.January, 4, 0, 0, 0, 0, time.UTC)

	// Find the Monday of week 1
	weekday := int(jan4.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}
	mondayWeek1 := jan4.AddDate(0, 0, -(weekday - 1))

	// Calculate the Monday of the target week
	targetMonday := mondayWeek1.AddDate(0, 0, (weekNum-1)*7)

	return targetMonday
}

// dateRangesOverlap checks if two date ranges overlap
func (rm *ReorganizationManager) dateRangesOverlap(start1, end1, start2, end2 time.Time) bool {
	return (start1.Before(end2) || start1.Equal(end2)) && (end1.After(start2) || end1.Equal(start2))
}
