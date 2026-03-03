package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"simple-telegram-chatbot/internal/memory"
)

// TopicKnowledgeTool provides LLM-friendly methods for accessing and writing topic knowledge
type TopicKnowledgeTool struct {
	topicManager *memory.TopicManager
}

// TopicToolResult represents the result of a topic tool operation with structured data
type TopicToolResult struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// TopicData represents structured topic information
type TopicData struct {
	Name         string    `json:"name"`
	FilePath     string    `json:"file_path"`
	Content      string    `json:"content"`
	LastModified time.Time `json:"last_modified"`
	TokenCount   int       `json:"token_count"`
}

// TopicInfo represents topic metadata without full content
type TopicInfo struct {
	Name         string    `json:"name"`
	FilePath     string    `json:"file_path"`
	LastModified time.Time `json:"last_modified"`
	Size         int64     `json:"size"`
	Subdivided   bool      `json:"subdivided"`
}

// TopicSearchResult represents a search result with context
type TopicSearchResult struct {
	TopicName string `json:"topic_name"`
	FilePath  string `json:"file_path"`
	Excerpt   string `json:"excerpt"`
	Context   string `json:"context"`
}

// toToolResult converts TopicToolResult to ToolResult by encoding data as JSON
func (r *TopicToolResult) toToolResult() ToolResult {
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

// NewTopicKnowledgeTool creates a new TopicKnowledgeTool instance
func NewTopicKnowledgeTool(topicManager *memory.TopicManager) *TopicKnowledgeTool {
	return &TopicKnowledgeTool{
		topicManager: topicManager,
	}
}

// GetTopic retrieves a specific topic file by name
func (t *TopicKnowledgeTool) GetTopic(name string) *TopicToolResult {
	if name == "" {
		return &TopicToolResult{
			Success: false,
			Message: "Topic name cannot be empty",
		}
	}
	
	// First, try to get from MEMORY.md
	memoryIndexManager := memory.NewMemoryIndexManager(t.topicManager.GetMemoryBasePath())
	content, err := memoryIndexManager.GetTopicFromMemory(name)
	if err == nil {
		// Topic found in MEMORY.md
		return &TopicToolResult{
			Success: true,
			Message: fmt.Sprintf("Successfully retrieved topic from MEMORY.md: %s", name),
			Data: TopicData{
				Name:         name,
				FilePath:     "memory/MEMORY.md",
				Content:      content,
				LastModified: time.Now(), // MEMORY.md is always current
				TokenCount:   len(content) / 4,
			},
		}
	}
	
	// Get topic info from TopicManager (separate file)
	topics, err := t.topicManager.ListTopics()
	if err != nil {
		return &TopicToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to list topics: %v", err),
		}
	}
	
	// Find the requested topic
	var topicInfo *memory.TopicInfo
	for i := range topics {
		if topics[i].Name == name {
			topicInfo = &topics[i]
			break
		}
	}
	
	if topicInfo == nil {
		return &TopicToolResult{
			Success: false,
			Message: fmt.Sprintf("Topic not found: %s", name),
		}
	}
	
	// Read topic content
	contentBytes, err := os.ReadFile(topicInfo.FilePath)
	if err != nil {
		return &TopicToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to read topic file: %v", err),
		}
	}
	
	return &TopicToolResult{
		Success: true,
		Message: fmt.Sprintf("Successfully retrieved topic: %s", name),
		Data: TopicData{
			Name:         topicInfo.Name,
			FilePath:     topicInfo.FilePath,
			Content:      string(contentBytes),
			LastModified: topicInfo.LastUpdated,
			TokenCount:   len(string(contentBytes)) / 4, // Rough estimate
		},
	}
}

// ListTopics returns a list of all available topics
func (t *TopicKnowledgeTool) ListTopics() *TopicToolResult {
	topics, err := t.topicManager.ListTopics()
	if err != nil {
		return &TopicToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to list topics: %v", err),
		}
	}
	
	// Convert to TopicInfo format
	topicInfoList := make([]TopicInfo, len(topics))
	for i, topic := range topics {
		topicInfoList[i] = TopicInfo{
			Name:         topic.Name,
			FilePath:     topic.FilePath,
			LastModified: topic.LastUpdated,
			Size:         topic.Size,
			Subdivided:   topic.Subdivided,
		}
	}
	
	return &TopicToolResult{
		Success: true,
		Message: fmt.Sprintf("Successfully listed all topics. Found %d topics.", len(topicInfoList)),
		Data: map[string]interface{}{
			"topics": topicInfoList,
			"count":  len(topicInfoList),
		},
	}
}

// SearchTopics searches across all topic files for a query string
func (t *TopicKnowledgeTool) SearchTopics(query string) *TopicToolResult {
	if query == "" {
		return &TopicToolResult{
			Success: false,
			Message: "Search query cannot be empty",
		}
	}
	
	results, err := t.topicManager.SearchTopics(query)
	if err != nil {
		return &TopicToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to search topics: %v", err),
		}
	}
	
	// Convert to TopicSearchResult format
	searchResults := make([]TopicSearchResult, len(results))
	for i, result := range results {
		searchResults[i] = TopicSearchResult{
			TopicName: result.TopicName,
			FilePath:  result.FilePath,
			Excerpt:   result.Excerpt,
			Context:   result.Context,
		}
	}
	
	return &TopicToolResult{
		Success: true,
		Message: fmt.Sprintf("Successfully searched topics for: '%s'. Found %d results.", query, len(searchResults)),
		Data: map[string]interface{}{
			"query":   query,
			"results": searchResults,
			"count":   len(searchResults),
		},
	}
}

// WriteTopic writes or updates a topic file with new content (replaces existing content)
func (t *TopicKnowledgeTool) WriteTopic(name, content string) *TopicToolResult {
	if name == "" {
		return &TopicToolResult{
			Success: false,
			Message: "Topic name cannot be empty",
		}
	}
	
	if content == "" {
		return &TopicToolResult{
			Success: false,
			Message: "Topic content cannot be empty",
		}
	}
	
	err := t.topicManager.UpdateTopicFile(name, content)
	if err != nil {
		return &TopicToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to write/update topic: %v", err),
		}
	}
	
	filePath := fmt.Sprintf("memory/topics/%s.md", name)
	
	return &TopicToolResult{
		Success: true,
		Message: fmt.Sprintf("Successfully wrote/updated topic '%s' at %s", name, filePath),
		Data: map[string]interface{}{
			"name":      name,
			"file_path": filePath,
			"operation": "write",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}
}

// CreateTopic creates a new topic file (fails if already exists)
func (t *TopicKnowledgeTool) CreateTopic(name, content string) *TopicToolResult {
	if name == "" {
		return &TopicToolResult{
			Success: false,
			Message: "Topic name cannot be empty",
		}
	}
	
	if content == "" {
		return &TopicToolResult{
			Success: false,
			Message: "Topic content cannot be empty",
		}
	}
	
	err := t.topicManager.CreateTopicFile(name, content)
	if err != nil {
		return &TopicToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create topic: %v", err),
		}
	}
	
	filePath := fmt.Sprintf("memory/topics/%s.md", name)
	
	return &TopicToolResult{
		Success: true,
		Message: fmt.Sprintf("Successfully created new topic '%s' at %s", name, filePath),
		Data: map[string]interface{}{
			"name":      name,
			"file_path": filePath,
			"operation": "create",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}
}

// AppendToTopic appends content to an existing topic file
func (t *TopicKnowledgeTool) AppendToTopic(name, content string) *TopicToolResult {
	if name == "" {
		return &TopicToolResult{
			Success: false,
			Message: "Topic name cannot be empty",
		}
	}
	
	if content == "" {
		return &TopicToolResult{
			Success: false,
			Message: "Content to append cannot be empty",
		}
	}
	
	err := t.topicManager.AppendToTopicFile(name, content)
	if err != nil {
		return &TopicToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to append to topic: %v", err),
		}
	}
	
	filePath := fmt.Sprintf("memory/topics/%s.md", name)
	
	return &TopicToolResult{
		Success: true,
		Message: fmt.Sprintf("Successfully appended to topic '%s' at %s", name, filePath),
		Data: map[string]interface{}{
			"name":      name,
			"file_path": filePath,
			"operation": "append",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}
}

// Name returns the tool name
func (t *TopicKnowledgeTool) Name() string {
	return "topic_knowledge"
}

// Description returns the tool description
func (t *TopicKnowledgeTool) Description() string {
	return "Access and write domain-specific knowledge topics. Supports get, list, search, write, create, and append operations."
}

// Execute implements the Tool interface
func (t *TopicKnowledgeTool) Execute(params map[string]interface{}) (ToolResult, error) {
	// Determine which method to call based on operation parameter
	operation, ok := params["operation"].(string)
	if !ok {
		return ToolResult{
			Success: false,
			Error:   "Operation parameter required. Valid operations: get, list, search, write, create, append",
		}, fmt.Errorf("operation parameter required")
	}
	
	switch operation {
	case "get":
		name, ok := params["name"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Name parameter required for get operation",
			}, fmt.Errorf("name parameter required")
		}
		result := t.GetTopic(name)
		return result.toToolResult(), nil
		
	case "list":
		result := t.ListTopics()
		return result.toToolResult(), nil
		
	case "search":
		query, ok := params["query"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Query parameter required for search operation",
			}, fmt.Errorf("query parameter required")
		}
		result := t.SearchTopics(query)
		return result.toToolResult(), nil
		
	case "write":
		name, ok := params["name"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Name parameter required for write operation",
			}, fmt.Errorf("name parameter required")
		}
		content, ok := params["content"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Content parameter required for write operation",
			}, fmt.Errorf("content parameter required")
		}
		result := t.WriteTopic(name, content)
		return result.toToolResult(), nil
		
	case "create":
		name, ok := params["name"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Name parameter required for create operation",
			}, fmt.Errorf("name parameter required")
		}
		content, ok := params["content"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Content parameter required for create operation",
			}, fmt.Errorf("content parameter required")
		}
		result := t.CreateTopic(name, content)
		return result.toToolResult(), nil
		
	case "append":
		name, ok := params["name"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Name parameter required for append operation",
			}, fmt.Errorf("name parameter required")
		}
		content, ok := params["content"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Content parameter required for append operation",
			}, fmt.Errorf("content parameter required")
		}
		result := t.AppendToTopic(name, content)
		return result.toToolResult(), nil
		
	default:
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Invalid operation: %s. Valid operations: get, list, search, write, create, append", operation),
		}, fmt.Errorf("invalid operation: %s", operation)
	}
}
