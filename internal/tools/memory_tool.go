package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"simple-telegram-chatbot/internal/memory"
)

// MemorySummaryTool provides LLM-friendly methods for retrieving memory summaries
type MemorySummaryTool struct {
	summaryManager *memory.SummaryManager
}

// MemoryToolResult represents the result of a memory tool operation with structured data
type MemoryToolResult struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SummaryData represents structured summary information
type SummaryData struct {
	Text       string `json:"text"`
	DateRange  string `json:"date_range"`
	Level      string `json:"level"` // "daily", "weekly", "quarterly"
	FilePath   string `json:"file_path"`
	TokenCount int    `json:"token_count"`
}

// toToolResult converts MemoryToolResult to ToolResult by encoding data as JSON
func (r *MemoryToolResult) toToolResult() ToolResult {
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

// NewMemorySummaryTool creates a new MemorySummaryTool instance
func NewMemorySummaryTool(summaryManager *memory.SummaryManager) *MemorySummaryTool {
	return &MemorySummaryTool{
		summaryManager: summaryManager,
	}
}

// GetDailySummary retrieves the daily summary for a specific date
func (t *MemorySummaryTool) GetDailySummary(date time.Time) *MemoryToolResult {
	dateStr := date.Format("2006-01-02")
	filePath := fmt.Sprintf("workplace/memory/chat/%s/daily-summary.md", dateStr)
	
	// Read the summary file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return &MemoryToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to read daily summary for %s: %v", dateStr, err),
		}
	}
	
	return &MemoryToolResult{
		Success: true,
		Message: fmt.Sprintf("Successfully retrieved daily summary for %s", dateStr),
		Data: SummaryData{
			Text:       string(content),
			DateRange:  dateStr,
			Level:      "daily",
			FilePath:   filePath,
			TokenCount: len(string(content)) / 4, // Rough estimate
		},
	}
}

// GetWeeklySummary retrieves the weekly summary for a specific week
func (t *MemorySummaryTool) GetWeeklySummary(weekNum, year int) *MemoryToolResult {
	weekFolder := fmt.Sprintf("week-%02d-%d", weekNum, year)
	filePath := fmt.Sprintf("workplace/memory/chat/%s/summary.md", weekFolder)
	
	// Read the summary file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return &MemoryToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to read weekly summary for week %d of %d: %v", weekNum, year, err),
		}
	}
	
	return &MemoryToolResult{
		Success: true,
		Message: fmt.Sprintf("Successfully retrieved weekly summary for week %d of %d", weekNum, year),
		Data: SummaryData{
			Text:       string(content),
			DateRange:  fmt.Sprintf("Week %d, %d", weekNum, year),
			Level:      "weekly",
			FilePath:   filePath,
			TokenCount: len(string(content)) / 4, // Rough estimate
		},
	}
}

// GetQuarterlySummary retrieves the quarterly summary for a specific quarter
func (t *MemorySummaryTool) GetQuarterlySummary(quarter, year int) *MemoryToolResult {
	if quarter < 1 || quarter > 4 {
		return &MemoryToolResult{
			Success: false,
			Message: fmt.Sprintf("Invalid quarter: %d. Must be between 1 and 4.", quarter),
		}
	}
	
	quarterFolder := fmt.Sprintf("Q%d-%d", quarter, year)
	filePath := fmt.Sprintf("workplace/memory/chat/%s/summary.md", quarterFolder)
	
	// Read the summary file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return &MemoryToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to read quarterly summary for Q%d %d: %v", quarter, year, err),
		}
	}
	
	return &MemoryToolResult{
		Success: true,
		Message: fmt.Sprintf("Successfully retrieved quarterly summary for Q%d %d", quarter, year),
		Data: SummaryData{
			Text:       string(content),
			DateRange:  fmt.Sprintf("Q%d %d", quarter, year),
			Level:      "quarterly",
			FilePath:   filePath,
			TokenCount: len(string(content)) / 4, // Rough estimate
		},
	}
}

// GetSummariesInRange retrieves all summaries within a date range
func (t *MemorySummaryTool) GetSummariesInRange(startDate, endDate time.Time) *MemoryToolResult {
	// Validate date range
	if endDate.Before(startDate) {
		return &MemoryToolResult{
			Success: false,
			Message: "End date must be after start date",
		}
	}
	
	dateRange := fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	
	// Collect all daily summaries in range
	var summaries []SummaryData
	currentDate := startDate
	for !currentDate.After(endDate) {
		dateStr := currentDate.Format("2006-01-02")
		filePath := fmt.Sprintf("workplace/memory/chat/%s/daily-summary.md", dateStr)
		
		// Try to read the summary file
		content, err := os.ReadFile(filePath)
		if err == nil {
			summaries = append(summaries, SummaryData{
				Text:       string(content),
				DateRange:  dateStr,
				Level:      "daily",
				FilePath:   filePath,
				TokenCount: len(string(content)) / 4,
			})
		}
		
		// Move to next day
		currentDate = currentDate.AddDate(0, 0, 1)
	}
	
	return &MemoryToolResult{
		Success: true,
		Message: fmt.Sprintf("Successfully retrieved %d summaries from %s", len(summaries), dateRange),
		Data: map[string]interface{}{
			"date_range": dateRange,
			"summaries":  summaries,
			"count":      len(summaries),
		},
	}
}

// Name returns the tool name
func (t *MemorySummaryTool) Name() string {
	return "memory_summary"
}

// Description returns the tool description
func (t *MemorySummaryTool) Description() string {
	return "Retrieve memory summaries for specific dates, weeks, quarters, or date ranges"
}

// Execute implements the Tool interface
func (t *MemorySummaryTool) Execute(params map[string]interface{}) (ToolResult, error) {
	// Determine which method to call based on parameters
	if dateStr, ok := params["date"].(string); ok {
		// Parse date and call GetDailySummary
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return ToolResult{
				Success: false,
				Error:   fmt.Sprintf("Invalid date format: %v", err),
			}, err
		}
		result := t.GetDailySummary(date)
		return result.toToolResult(), nil
	}
	
	if weekNum, ok := params["week"].(float64); ok {
		year, ok := params["year"].(float64)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Year parameter required for weekly summary",
			}, fmt.Errorf("year parameter required")
		}
		result := t.GetWeeklySummary(int(weekNum), int(year))
		return result.toToolResult(), nil
	}
	
	if quarter, ok := params["quarter"].(float64); ok {
		year, ok := params["year"].(float64)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Year parameter required for quarterly summary",
			}, fmt.Errorf("year parameter required")
		}
		result := t.GetQuarterlySummary(int(quarter), int(year))
		return result.toToolResult(), nil
	}
	
	if startDateStr, ok := params["start_date"].(string); ok {
		endDateStr, ok := params["end_date"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "End date parameter required for date range query",
			}, fmt.Errorf("end_date parameter required")
		}
		
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return ToolResult{
				Success: false,
				Error:   fmt.Sprintf("Invalid start date format: %v", err),
			}, err
		}
		
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return ToolResult{
				Success: false,
				Error:   fmt.Sprintf("Invalid end date format: %v", err),
			}, err
		}
		
		result := t.GetSummariesInRange(startDate, endDate)
		return result.toToolResult(), nil
	}
	
	return ToolResult{
		Success: false,
		Error:   "Invalid parameters. Provide either: date, week+year, quarter+year, or start_date+end_date",
	}, fmt.Errorf("invalid parameters")
}
