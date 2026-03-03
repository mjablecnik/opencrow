package tools

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ChatLogSearchTool provides LLM-friendly methods for searching raw chat logs
type ChatLogSearchTool struct {
	memoryBasePath string
}

// ChatLogToolResult represents the result of a chat log tool operation with structured data
type ChatLogToolResult struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// LogSearchResult represents a search result from chat logs
type LogSearchResult struct {
	Date       time.Time `json:"date"`
	SessionID  string    `json:"session_id"`
	Excerpt    string    `json:"excerpt"`
	Context    string    `json:"context"`
	FilePath   string    `json:"file_path"`
	TokenCount int       `json:"token_count"`
}

// toToolResult converts ChatLogToolResult to ToolResult by encoding data as JSON
func (r *ChatLogToolResult) toToolResult() ToolResult {
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

// NewChatLogSearchTool creates a new ChatLogSearchTool instance
func NewChatLogSearchTool(memoryBasePath string) *ChatLogSearchTool {
	return &ChatLogSearchTool{
		memoryBasePath: memoryBasePath,
	}
}

// SearchLogs searches chat logs for a query string within an optional date range
func (t *ChatLogSearchTool) SearchLogs(query string, startDate, endDate time.Time, maxResults int) *ChatLogToolResult {
	if query == "" {
		return &ChatLogToolResult{
			Success: false,
			Message: "Search query cannot be empty",
		}
	}
	
	// Validate date range
	if !endDate.IsZero() && endDate.Before(startDate) {
		return &ChatLogToolResult{
			Success: false,
			Message: "End date must be after start date",
		}
	}
	
	// Set default max results if not provided
	if maxResults <= 0 {
		maxResults = 10 // Default limit to prevent token overflow
	}
	
	// Limit max results to prevent excessive token usage
	if maxResults > 100 {
		maxResults = 100
	}
	
	dateRange := "all time"
	if !startDate.IsZero() && !endDate.IsZero() {
		dateRange = fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	} else if !startDate.IsZero() {
		dateRange = fmt.Sprintf("from %s", startDate.Format("2006-01-02"))
	} else if !endDate.IsZero() {
		dateRange = fmt.Sprintf("until %s", endDate.Format("2006-01-02"))
	}
	
	// Search through chat logs
	var results []LogSearchResult
	chatPath := filepath.Join(t.memoryBasePath, "chat")
	
	// Walk through all daily folders
	err := filepath.Walk(chatPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		
		// Only process .log files
		if info.IsDir() || !strings.HasSuffix(path, ".log") {
			return nil
		}
		
		// Check if we've reached max results
		if len(results) >= maxResults {
			return filepath.SkipAll
		}
		
		// Search in this file
		fileResults, err := t.searchInFile(path, query, maxResults-len(results))
		if err == nil {
			results = append(results, fileResults...)
		}
		
		return nil
	})
	
	if err != nil {
		return &ChatLogToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to search chat logs: %v", err),
		}
	}
	
	// Add efficiency warning
	warning := "⚠️ EFFICIENCY NOTE: Searching raw chat logs is token-intensive. Consider using memory summaries (memory_summary tool) for better efficiency when possible."
	
	return &ChatLogToolResult{
		Success: true,
		Message: fmt.Sprintf("Successfully searched chat logs for '%s' in %s. Found %d results (limited to %d). %s", query, dateRange, len(results), maxResults, warning),
		Data: map[string]interface{}{
			"query":       query,
			"date_range":  dateRange,
			"results":     results,
			"count":       len(results),
			"max_results": maxResults,
			"warning":     warning,
		},
	}
}

// searchInFile searches for query in a single log file
func (t *ChatLogSearchTool) searchInFile(filePath, query string, maxResults int) ([]LogSearchResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var results []LogSearchResult
	scanner := bufio.NewScanner(file)
	var currentContext []string
	lineNum := 0
	
	for scanner.Scan() && len(results) < maxResults {
		line := scanner.Text()
		lineNum++
		
		// Keep last 5 lines for context
		currentContext = append(currentContext, line)
		if len(currentContext) > 5 {
			currentContext = currentContext[1:]
		}
		
		// Check if line contains query (case-insensitive)
		if strings.Contains(strings.ToLower(line), strings.ToLower(query)) {
			// Extract date and session from file path
			parts := strings.Split(filePath, string(filepath.Separator))
			var dateStr, sessionID string
			for i, part := range parts {
				if len(part) == 10 && strings.Count(part, "-") == 2 {
					dateStr = part
					if i+1 < len(parts) {
						sessionID = strings.TrimSuffix(parts[i+1], ".log")
					}
					break
				}
			}
			
			date, _ := time.Parse("2006-01-02", dateStr)
			
			results = append(results, LogSearchResult{
				Date:       date,
				SessionID:  sessionID,
				Excerpt:    line,
				Context:    strings.Join(currentContext, "\n"),
				FilePath:   filePath,
				TokenCount: len(line) / 4,
			})
		}
	}
	
	return results, scanner.Err()
}

// GetSessionLog retrieves a complete session log for a specific date and session number
func (t *ChatLogSearchTool) GetSessionLog(date time.Time, sessionNum int) *ChatLogToolResult {
	if sessionNum <= 0 {
		return &ChatLogToolResult{
			Success: false,
			Message: "Session number must be positive",
		}
	}
	
	dateStr := date.Format("2006-01-02")
	sessionID := fmt.Sprintf("session-%03d", sessionNum)
	filePath := filepath.Join(t.memoryBasePath, "chat", dateStr, sessionID+".log")
	
	// Read the session log file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return &ChatLogToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to read session log for %s, session %d: %v", dateStr, sessionNum, err),
		}
	}
	
	logContent := string(content)
	
	// Add efficiency warning
	warning := "⚠️ EFFICIENCY NOTE: Retrieving full session logs is token-intensive. Consider using session summaries (memory_summary tool) for better efficiency when possible."
	
	return &ChatLogToolResult{
		Success: true,
		Message: fmt.Sprintf("Successfully retrieved session log for %s, session %d. %s", dateStr, sessionNum, warning),
		Data: map[string]interface{}{
			"date":        dateStr,
			"session_id":  sessionID,
			"file_path":   filePath,
			"content":     logContent,
			"token_count": len(logContent) / 4, // Rough estimate
			"warning":     warning,
		},
	}
}

// Name returns the tool name
func (t *ChatLogSearchTool) Name() string {
	return "chatlog_search"
}

// Description returns the tool description
func (t *ChatLogSearchTool) Description() string {
	return "Search raw chat logs or retrieve complete session logs. WARNING: Token-intensive operation. Prefer memory summaries when possible."
}

// Execute implements the Tool interface
func (t *ChatLogSearchTool) Execute(params map[string]interface{}) (ToolResult, error) {
	// Determine which method to call based on operation parameter
	operation, ok := params["operation"].(string)
	if !ok {
		return ToolResult{
			Success: false,
			Error:   "Operation parameter required. Valid operations: search, get_session",
		}, fmt.Errorf("operation parameter required")
	}
	
	switch operation {
	case "search":
		query, ok := params["query"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Query parameter required for search operation",
			}, fmt.Errorf("query parameter required")
		}
		
		// Parse optional date range
		var startDate, endDate time.Time
		if startDateStr, ok := params["start_date"].(string); ok {
			var err error
			startDate, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				return ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Invalid start_date format: %v", err),
				}, err
			}
		}
		
		if endDateStr, ok := params["end_date"].(string); ok {
			var err error
			endDate, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				return ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Invalid end_date format: %v", err),
				}, err
			}
		}
		
		// Parse optional max_results
		maxResults := 50 // Default
		if maxResultsFloat, ok := params["max_results"].(float64); ok {
			maxResults = int(maxResultsFloat)
		}
		
		result := t.SearchLogs(query, startDate, endDate, maxResults)
		return result.toToolResult(), nil
		
	case "get_session":
		dateStr, ok := params["date"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Date parameter required for get_session operation",
			}, fmt.Errorf("date parameter required")
		}
		
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return ToolResult{
				Success: false,
				Error:   fmt.Sprintf("Invalid date format: %v", err),
			}, err
		}
		
		sessionNum, ok := params["session_num"].(float64)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Session number parameter required for get_session operation",
			}, fmt.Errorf("session_num parameter required")
		}
		
		result := t.GetSessionLog(date, int(sessionNum))
		return result.toToolResult(), nil
		
	default:
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Invalid operation: %s. Valid operations: search, get_session", operation),
		}, fmt.Errorf("invalid operation: %s", operation)
	}
}
