package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"simple-telegram-chatbot/pkg/utils"
)

// TopicInfo represents metadata about a topic file
type TopicInfo struct {
	Name         string
	FilePath     string
	LastUpdated  time.Time
	Size         int64
	Subdivided   bool
	Subfolders   []string
}

// TopicSearchResult represents a search result from topic files
type TopicSearchResult struct {
	TopicName string
	FilePath  string
	Excerpt   string
	Context   string
}

// TopicManager handles topic extraction, file management, and hierarchical structure
type TopicManager struct {
	mu                 sync.RWMutex
	memoryBasePath     string
	topicSizeThreshold int64
	logger             *utils.Logger
}

// NewTopicManager creates a new topic manager
func NewTopicManager(memoryBasePath string, topicSizeThreshold int64) *TopicManager {
	return &TopicManager{
		memoryBasePath:     memoryBasePath,
		topicSizeThreshold: topicSizeThreshold,
		logger:             utils.NewLogger("info"),
	}
}

// ExtractTopics extracts topics from content using the provided topic extractions
// This method receives the LLM-extracted topics and writes them to files
func (tm *TopicManager) ExtractTopics(topicExtractions []TopicExtraction) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Filter topics that should be written
	topicsToWrite := []TopicExtraction{}
	for _, topic := range topicExtractions {
		if topic.ShouldWrite {
			topicsToWrite = append(topicsToWrite, topic)
		}
	}

	if len(topicsToWrite) == 0 {
		tm.logger.InfoWithComponent("TopicManager", "No relevant domain knowledge found for topic extraction",
			"timestamp", timestamp,
		)
		return nil
	}

	// Log topics identified
	tm.logger.InfoWithComponent("TopicManager", "Topic extraction completed",
		"topics_count", len(topicsToWrite),
		"timestamp", timestamp,
	)

	// Write each topic to file
	for _, topic := range topicsToWrite {
		tm.logger.InfoWithComponent("TopicManager", "Processing topic",
			"topic_name", topic.TopicName,
			"confidence", fmt.Sprintf("%.2f", topic.Confidence),
		)

		// Check if topic file exists
		topicPath := tm.getTopicFilePath(topic.TopicName)
		exists := tm.topicFileExists(topicPath)

		if exists {
			// Update existing topic file
			if err := tm.updateTopicFileInternal(topic.TopicName, topic.Content); err != nil {
				tm.logger.ErrorWithDetails("TopicManager", "Failed to update topic file", err,
					"topic_name", topic.TopicName,
					"file_path", topicPath,
				)
				continue
			}
			tm.logger.InfoWithComponent("TopicManager", "Topic file updated",
				"topic_name", topic.TopicName,
				"file_path", topicPath,
			)
		} else {
			// Create new topic file
			if err := tm.createTopicFileInternal(topic.TopicName, topic.Content); err != nil {
				tm.logger.ErrorWithDetails("TopicManager", "Failed to create topic file", err,
					"topic_name", topic.TopicName,
					"file_path", topicPath,
				)
				continue
			}
			tm.logger.InfoWithComponent("TopicManager", "Topic file created",
				"topic_name", topic.TopicName,
				"file_path", topicPath,
			)
		}

		// Check if topic file needs subdivision
		if err := tm.checkAndSubdivideIfNeeded(topic.TopicName); err != nil {
			tm.logger.WarnWithComponent("TopicManager", "Failed to check topic subdivision",
				"topic_name", topic.TopicName,
				"error", err.Error(),
			)
		}
	}

	return nil
}

// CreateTopicFile creates a new topic file with the given content
func (tm *TopicManager) CreateTopicFile(topicName, content string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	return tm.createTopicFileInternal(topicName, content)
}

// createTopicFileInternal is the internal implementation without locking
func (tm *TopicManager) createTopicFileInternal(topicName, content string) error {
	topicPath := tm.getTopicFilePath(topicName)

	// Ensure topics directory exists
	topicsDir := filepath.Join(tm.memoryBasePath, "topics")
	if err := os.MkdirAll(topicsDir, 0755); err != nil {
		return fmt.Errorf("failed to create topics directory: %w", err)
	}

	// Check if file already exists
	if _, err := os.Stat(topicPath); err == nil {
		return fmt.Errorf("topic file already exists: %s", topicName)
	}

	// Format content with header and timestamp
	formattedContent := tm.formatTopicContent(topicName, content, true)

	// Write topic file
	if err := os.WriteFile(topicPath, []byte(formattedContent), 0644); err != nil {
		return fmt.Errorf("failed to write topic file: %w", err)
	}

	tm.logger.InfoWithComponent("TopicManager", "Topic file created",
		"topic_name", topicName,
		"file_path", topicPath,
		"size", len(formattedContent),
	)

	return nil
}

// UpdateTopicFile updates an existing topic file with new content
func (tm *TopicManager) UpdateTopicFile(topicName, content string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	return tm.updateTopicFileInternal(topicName, content)
}

// updateTopicFileInternal is the internal implementation without locking
func (tm *TopicManager) updateTopicFileInternal(topicName, content string) error {
	topicPath := tm.getTopicFilePath(topicName)

	// Check if file exists
	if _, err := os.Stat(topicPath); os.IsNotExist(err) {
		return fmt.Errorf("topic file does not exist: %s", topicName)
	}

	// Read existing content
	existingContent, err := os.ReadFile(topicPath)
	if err != nil {
		return fmt.Errorf("failed to read existing topic file: %w", err)
	}

	// Merge new content with existing content
	mergedContent := tm.mergeTopicContent(string(existingContent), content)

	// Format content with updated timestamp
	formattedContent := tm.formatTopicContent(topicName, mergedContent, false)

	// Write updated topic file
	if err := os.WriteFile(topicPath, []byte(formattedContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated topic file: %w", err)
	}

	tm.logger.InfoWithComponent("TopicManager", "Topic file updated",
		"topic_name", topicName,
		"file_path", topicPath,
		"size", len(formattedContent),
	)

	return nil
}

// AppendToTopicFile appends content to an existing topic file
func (tm *TopicManager) AppendToTopicFile(topicName, content string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	topicPath := tm.getTopicFilePath(topicName)

	// Check if file exists
	if _, err := os.Stat(topicPath); os.IsNotExist(err) {
		return fmt.Errorf("topic file does not exist: %s", topicName)
	}

	// Read existing content
	existingContent, err := os.ReadFile(topicPath)
	if err != nil {
		return fmt.Errorf("failed to read existing topic file: %w", err)
	}

	// Append new content
	appendedContent := string(existingContent) + "\n\n" + content

	// Update timestamp in header
	updatedContent := tm.updateTimestamp(appendedContent)

	// Write updated topic file
	if err := os.WriteFile(topicPath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write appended topic file: %w", err)
	}

	tm.logger.InfoWithComponent("TopicManager", "Content appended to topic file",
		"topic_name", topicName,
		"file_path", topicPath,
		"size", len(updatedContent),
	)

	return nil
}

// SubdivideTopic subdivides a large topic file into multiple specialized files
func (tm *TopicManager) SubdivideTopic(topicName string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	topicPath := tm.getTopicFilePath(topicName)

	// Check if file exists
	fileInfo, err := os.Stat(topicPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("topic file does not exist: %s", topicName)
	}

	// Check if file is large enough to subdivide
	if fileInfo.Size() < tm.topicSizeThreshold {
		return fmt.Errorf("topic file is not large enough to subdivide: %s (size: %d, threshold: %d)",
			topicName, fileInfo.Size(), tm.topicSizeThreshold)
	}

	// Create topic subfolder
	topicsDir := filepath.Join(tm.memoryBasePath, "topics")
	subfolderPath := filepath.Join(topicsDir, topicName)
	if err := os.MkdirAll(subfolderPath, 0755); err != nil {
		return fmt.Errorf("failed to create topic subfolder: %w", err)
	}

	// Read existing content
	content, err := os.ReadFile(topicPath)
	if err != nil {
		return fmt.Errorf("failed to read topic file for subdivision: %w", err)
	}

	// For now, move the entire content to a "General.md" file in the subfolder
	// In a more sophisticated implementation, this could use LLM to intelligently split content
	generalPath := filepath.Join(subfolderPath, "General.md")
	if err := os.WriteFile(generalPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write subdivided topic file: %w", err)
	}

	// Remove the original file
	if err := os.Remove(topicPath); err != nil {
		return fmt.Errorf("failed to remove original topic file: %w", err)
	}

	tm.logger.InfoWithComponent("TopicManager", "Topic subdivided",
		"topic_name", topicName,
		"original_path", topicPath,
		"subfolder_path", subfolderPath,
	)

	return nil
}

// ListTopics returns a list of all available topics
func (tm *TopicManager) ListTopics() ([]TopicInfo, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	topicsDir := filepath.Join(tm.memoryBasePath, "topics")

	// Check if topics directory exists
	if _, err := os.Stat(topicsDir); os.IsNotExist(err) {
		return []TopicInfo{}, nil
	}

	var topics []TopicInfo

	// Read topics directory
	entries, err := os.ReadDir(topicsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read topics directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// This is a subdivided topic
			topicInfo, err := tm.getSubdividedTopicInfo(entry.Name())
			if err != nil {
				tm.logger.WarnWithComponent("TopicManager", "Failed to get subdivided topic info",
					"topic_name", entry.Name(),
					"error", err.Error(),
				)
				continue
			}
			topics = append(topics, topicInfo)
		} else if strings.HasSuffix(entry.Name(), ".md") {
			// This is a regular topic file
			topicName := strings.TrimSuffix(entry.Name(), ".md")
			topicPath := filepath.Join(topicsDir, entry.Name())

			fileInfo, err := os.Stat(topicPath)
			if err != nil {
				tm.logger.WarnWithComponent("TopicManager", "Failed to stat topic file",
					"topic_name", topicName,
					"error", err.Error(),
				)
				continue
			}

			topics = append(topics, TopicInfo{
				Name:        topicName,
				FilePath:    topicPath,
				LastUpdated: fileInfo.ModTime(),
				Size:        fileInfo.Size(),
				Subdivided:  false,
				Subfolders:  []string{},
			})
		}
	}

	return topics, nil
}

// SearchTopics searches across all topic files for the given query
func (tm *TopicManager) SearchTopics(query string) ([]TopicSearchResult, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	topicsDir := filepath.Join(tm.memoryBasePath, "topics")

	// Check if topics directory exists
	if _, err := os.Stat(topicsDir); os.IsNotExist(err) {
		return []TopicSearchResult{}, nil
	}

	var results []TopicSearchResult

	// Search in all topic files
	err := filepath.Walk(topicsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			tm.logger.WarnWithComponent("TopicManager", "Failed to read topic file during search",
				"file_path", path,
				"error", err.Error(),
			)
			return nil
		}

		// Search for query in content (case-insensitive)
		contentStr := string(content)
		queryLower := strings.ToLower(query)
		contentLower := strings.ToLower(contentStr)

		if strings.Contains(contentLower, queryLower) {
			// Extract excerpt with context
			excerpt := tm.extractExcerpt(contentStr, query)

			// Get topic name from file path
			topicName := tm.getTopicNameFromPath(path)

			results = append(results, TopicSearchResult{
				TopicName: topicName,
				FilePath:  path,
				Excerpt:   excerpt,
				Context:   "", // Could be enhanced to provide more context
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search topics: %w", err)
	}

	return results, nil
}

// Helper methods

// getTopicFilePath returns the file path for a topic
func (tm *TopicManager) getTopicFilePath(topicName string) string {
	topicsDir := filepath.Join(tm.memoryBasePath, "topics")
	return filepath.Join(topicsDir, topicName+".md")
}

// topicFileExists checks if a topic file exists
func (tm *TopicManager) topicFileExists(topicPath string) bool {
	_, err := os.Stat(topicPath)
	return err == nil
}

// formatTopicContent formats topic content with header and timestamp
func (tm *TopicManager) formatTopicContent(topicName, content string, isNew bool) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var header string
	if isNew {
		header = fmt.Sprintf("# %s\n\n**Created:** %s\n**Last Updated:** %s\n\n", topicName, timestamp, timestamp)
	} else {
		header = fmt.Sprintf("# %s\n\n**Last Updated:** %s\n\n", topicName, timestamp)
	}

	return header + content
}

// mergeTopicContent merges new content with existing content
// This is a simple implementation that appends new content
// A more sophisticated implementation could use LLM to intelligently merge
func (tm *TopicManager) mergeTopicContent(existingContent, newContent string) string {
	// Remove the header from existing content
	lines := strings.Split(existingContent, "\n")
	contentStart := 0
	for i, line := range lines {
		if strings.HasPrefix(line, "**Last Updated:**") {
			contentStart = i + 1
			break
		}
	}

	if contentStart > 0 && contentStart < len(lines) {
		existingContent = strings.Join(lines[contentStart:], "\n")
	}

	// Append new content
	return strings.TrimSpace(existingContent) + "\n\n## Update - " + time.Now().Format("2006-01-02") + "\n\n" + newContent
}

// updateTimestamp updates the "Last Updated" timestamp in the content
func (tm *TopicManager) updateTimestamp(content string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, "**Last Updated:**") {
			lines[i] = fmt.Sprintf("**Last Updated:** %s", timestamp)
			break
		}
	}

	return strings.Join(lines, "\n")
}

// checkAndSubdivideIfNeeded checks if a topic file exceeds the size threshold and subdivides if needed
func (tm *TopicManager) checkAndSubdivideIfNeeded(topicName string) error {
	topicPath := tm.getTopicFilePath(topicName)

	fileInfo, err := os.Stat(topicPath)
	if err != nil {
		return fmt.Errorf("failed to stat topic file: %w", err)
	}

	if fileInfo.Size() >= tm.topicSizeThreshold {
		tm.logger.InfoWithComponent("TopicManager", "Topic file exceeds size threshold, subdividing",
			"topic_name", topicName,
			"size", fileInfo.Size(),
			"threshold", tm.topicSizeThreshold,
		)

		if err := tm.SubdivideTopic(topicName); err != nil {
			return fmt.Errorf("failed to subdivide topic: %w", err)
		}
	}

	return nil
}

// getSubdividedTopicInfo returns TopicInfo for a subdivided topic
func (tm *TopicManager) getSubdividedTopicInfo(topicName string) (TopicInfo, error) {
	topicsDir := filepath.Join(tm.memoryBasePath, "topics")
	subfolderPath := filepath.Join(topicsDir, topicName)

	// Check if subfolder exists
	if _, err := os.Stat(subfolderPath); err != nil {
		return TopicInfo{}, fmt.Errorf("failed to stat topic subfolder: %w", err)
	}

	// List subfolders
	entries, err := os.ReadDir(subfolderPath)
	if err != nil {
		return TopicInfo{}, fmt.Errorf("failed to read topic subfolder: %w", err)
	}

	var subfolders []string
	var totalSize int64
	var lastUpdated time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			subfolders = append(subfolders, entry.Name())
		} else if strings.HasSuffix(entry.Name(), ".md") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			totalSize += info.Size()
			if info.ModTime().After(lastUpdated) {
				lastUpdated = info.ModTime()
			}
		}
	}

	return TopicInfo{
		Name:        topicName,
		FilePath:    subfolderPath,
		LastUpdated: lastUpdated,
		Size:        totalSize,
		Subdivided:  true,
		Subfolders:  subfolders,
	}, nil
}

// extractExcerpt extracts a relevant excerpt from content containing the query
func (tm *TopicManager) extractExcerpt(content, query string) string {
	queryLower := strings.ToLower(query)
	contentLower := strings.ToLower(content)

	// Find the position of the query
	index := strings.Index(contentLower, queryLower)
	if index == -1 {
		// Query not found, return first 200 characters
		if len(content) > 200 {
			return content[:200] + "..."
		}
		return content
	}

	// Extract context around the query (100 characters before and after)
	start := index - 100
	if start < 0 {
		start = 0
	}

	end := index + len(query) + 100
	if end > len(content) {
		end = len(content)
	}

	excerpt := content[start:end]

	// Add ellipsis if needed
	if start > 0 {
		excerpt = "..." + excerpt
	}
	if end < len(content) {
		excerpt = excerpt + "..."
	}

	return excerpt
}

// getTopicNameFromPath extracts the topic name from a file path
func (tm *TopicManager) getTopicNameFromPath(path string) string {
	// Get relative path from topics directory
	topicsDir := filepath.Join(tm.memoryBasePath, "topics")
	relPath, err := filepath.Rel(topicsDir, path)
	if err != nil {
		return filepath.Base(path)
	}

	// Remove .md extension
	relPath = strings.TrimSuffix(relPath, ".md")

	// Replace path separators with slashes for consistency
	relPath = filepath.ToSlash(relPath)

	return relPath
}
