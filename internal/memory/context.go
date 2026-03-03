package memory

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Message represents a single message in the conversation
type Message struct {
	Role      string
	Content   string
	Timestamp time.Time
}

// ContextManager handles context retrieval operations
type ContextManager struct {
	memoryBasePath string
	sessionManager *SessionManager
}

// NewContextManager creates a new context manager
func NewContextManager(memoryBasePath string, sessionManager *SessionManager) *ContextManager {
	return &ContextManager{
		memoryBasePath: memoryBasePath,
		sessionManager: sessionManager,
	}
}

// GetRecentHistory retrieves recent messages from session logs
// chatID is currently unused but kept for interface compatibility
// limit specifies the maximum number of messages to retrieve
func (cm *ContextManager) GetRecentHistory(chatID int64, limit int) ([]Message, error) {
	var messages []Message

	// Get current date and session info
	currentDate := cm.sessionManager.GetCurrentDate()
	if currentDate == "" {
		// No session started yet
		return messages, nil
	}

	// Read current session log
	currentLogPath := cm.sessionManager.GetCurrentSessionPath()
	if currentLogPath != "" {
		sessionMessages, err := cm.readSessionLog(currentLogPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read current session log: %w", err)
		}
		messages = append(messages, sessionMessages...)
	}

	// If we need more messages, read previous sessions from today
	if len(messages) < limit {
		dailyFolderPath := filepath.Join(cm.memoryBasePath, "chat", currentDate)
		additionalMessages, err := cm.readPreviousSessionsFromDay(dailyFolderPath, cm.sessionManager.GetCurrentSessionNumber(), limit-len(messages))
		if err != nil {
			// Log error but don't fail - we have at least current session
			fmt.Printf("Warning: failed to read previous sessions: %v\n", err)
		} else {
			// Prepend older messages
			messages = append(additionalMessages, messages...)
		}
	}

	// Trim to limit
	if len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}

	return messages, nil
}

// GetRelevantTopics identifies relevant topic files based on conversation context
// Returns a list of topic file paths that match keywords in the conversation
func (cm *ContextManager) GetRelevantTopics(conversationContext string) ([]string, error) {
	topicsPath := filepath.Join(cm.memoryBasePath, "topics")

	// Check if topics directory exists
	if _, err := os.Stat(topicsPath); os.IsNotExist(err) {
		return []string{}, nil
	}

	// Extract keywords from conversation context (simple approach)
	keywords := cm.extractKeywords(conversationContext)
	if len(keywords) == 0 {
		return []string{}, nil
	}

	// Find matching topic files
	var relevantTopics []string
	err := filepath.Walk(topicsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		// Check if topic name matches any keyword
		topicName := strings.TrimSuffix(info.Name(), ".md")
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(topicName), strings.ToLower(keyword)) {
				relevantTopics = append(relevantTopics, path)
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk topics directory: %w", err)
	}

	return relevantTopics, nil
}

// GetCurrentTokenUsage calculates token count for in-memory session context
// Uses a simple approximation: ~4 characters per token
func (cm *ContextManager) GetCurrentTokenUsage() (int, error) {
	// Get current session log path
	currentLogPath := cm.sessionManager.GetCurrentSessionPath()
	if currentLogPath == "" {
		return 0, nil
	}

	// Check if file exists
	if _, err := os.Stat(currentLogPath); os.IsNotExist(err) {
		return 0, nil
	}

	// Read file content
	content, err := os.ReadFile(currentLogPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read session log: %w", err)
	}

	// Approximate token count: ~4 characters per token
	tokenCount := len(content) / 4

	return tokenCount, nil
}

// readSessionLog reads and parses a session log file
func (cm *ContextManager) readSessionLog(logPath string) ([]Message, error) {
	file, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Message{}, nil
		}
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var messages []Message
	scanner := bufio.NewScanner(file)

	// Regex to parse log entries: [YYYY-MM-DD HH:MM:SS] Role: Content
	logPattern := regexp.MustCompile(`^\[(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\] ([^:]+): (.*)$`)

	var currentMessage *Message

	for scanner.Scan() {
		line := scanner.Text()

		// Check if this is a new message entry
		if matches := logPattern.FindStringSubmatch(line); matches != nil {
			// Save previous message if exists
			if currentMessage != nil {
				messages = append(messages, *currentMessage)
			}

			// Parse timestamp
			timestamp, err := time.Parse("2006-01-02 15:04:05", matches[1])
			if err != nil {
				// Use current time if parsing fails
				timestamp = time.Now()
			}

			// Create new message
			currentMessage = &Message{
				Role:      matches[2],
				Content:   matches[3],
				Timestamp: timestamp,
			}
		} else if currentMessage != nil && line != "" {
			// Continuation of previous message content
			currentMessage.Content += "\n" + line
		}
	}

	// Add last message
	if currentMessage != nil {
		messages = append(messages, *currentMessage)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return messages, nil
}

// readPreviousSessionsFromDay reads previous session logs from the same day
func (cm *ContextManager) readPreviousSessionsFromDay(dailyFolderPath string, currentSessionNum int, limit int) ([]Message, error) {
	var allMessages []Message

	// Read sessions in reverse order (most recent first)
	for sessionNum := currentSessionNum - 1; sessionNum >= 1 && len(allMessages) < limit; sessionNum-- {
		sessionLogPath := filepath.Join(dailyFolderPath, fmt.Sprintf("session-%03d.log", sessionNum))

		// Check if file exists
		if _, err := os.Stat(sessionLogPath); os.IsNotExist(err) {
			continue
		}

		messages, err := cm.readSessionLog(sessionLogPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read session %d: %w", sessionNum, err)
		}

		// Prepend messages (older sessions first)
		allMessages = append(messages, allMessages...)
	}

	return allMessages, nil
}

// extractKeywords extracts potential topic keywords from conversation context
// This is a simple implementation that looks for capitalized words and common domains
func (cm *ContextManager) extractKeywords(context string) []string {
	// Common topic domains to look for
	commonTopics := []string{
		"programming", "docker", "go", "golang", "python", "javascript",
		"psychology", "food", "health", "sport", "politics",
		"database", "sql", "api", "web", "security",
	}

	var keywords []string
	contextLower := strings.ToLower(context)

	// Check for common topics
	for _, topic := range commonTopics {
		if strings.Contains(contextLower, topic) {
			keywords = append(keywords, topic)
		}
	}

	// Extract capitalized words (potential proper nouns/topics)
	words := strings.Fields(context)
	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:")

		// Check if word starts with capital letter and is longer than 3 chars
		if len(word) > 3 && word[0] >= 'A' && word[0] <= 'Z' {
			keywords = append(keywords, word)
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueKeywords []string
	for _, keyword := range keywords {
		keywordLower := strings.ToLower(keyword)
		if !seen[keywordLower] {
			seen[keywordLower] = true
			uniqueKeywords = append(uniqueKeywords, keyword)
		}
	}

	return uniqueKeywords
}
