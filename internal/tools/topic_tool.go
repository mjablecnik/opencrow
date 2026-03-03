package tools

import (
	"encoding/json"
	"fmt"
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
	// TODO: Integrate with Memory Manager's GetTopic method (task 16)
	
	if name == "" {
		return &TopicToolResult{
			Success: false,
			Message: "Topic name cannot be empty",
		}
	}
	
	filePath := fmt.Sprintf("memory/topics/%s.md", name)
	
	return &TopicToolResult{
		Success: true,
		Message: fmt.Sprintf("[PLACEHOLDER] Would retrieve topic: %s", name),
		Data: TopicData{
			Name:         name,
			FilePath:     filePath,
			Content:      fmt.Sprintf("Placeholder content for topic '%s'. This will be populated by Memory Manager.", name),
			LastModified: time.Now(),
			TokenCount:   0, // Will be calculated by Memory Manager
		},
	}
}

// ListTopics returns a list of all available topics
func (t *TopicKnowledgeTool) ListTopics() *TopicToolResult {
	// TODO: Integrate with Memory Manager's ListTopics method (task 16)
	
	// Placeholder: Would scan memory/topics/ directory
	topics := []TopicInfo{
		{
			Name:         "Programming",
			FilePath:     "memory/topics/Programming.md",
			LastModified: time.Now(),
			Size:         15360, // 15KB placeholder
			Subdivided:   false,
		},
		{
			Name:         "Psychology",
			FilePath:     "memory/topics/Psychology.md",
			LastModified: time.Now(),
			Size:         8192, // 8KB placeholder
			Subdivided:   false,
		},
	}
	
	return &TopicToolResult{
		Success: true,
		Message: fmt.Sprintf("[PLACEHOLDER] Would list all topics. Found %d topics.", len(topics)),
		Data: map[string]interface{}{
			"topics": topics,
			"count":  len(topics),
		},
	}
}

// SearchTopics searches across all topic files for a query string
func (t *TopicKnowledgeTool) SearchTopics(query string) *TopicToolResult {
	// TODO: Integrate with Memory Manager's SearchTopics method (task 16)
	
	if query == "" {
		return &TopicToolResult{
			Success: false,
			Message: "Search query cannot be empty",
		}
	}
	
	// Placeholder: Would search through all topic files
	results := []TopicSearchResult{
		{
			TopicName: "Programming",
			FilePath:  "memory/topics/Programming.md",
			Excerpt:   fmt.Sprintf("Placeholder excerpt containing '%s'...", query),
			Context:   "Surrounding context would be provided here",
		},
	}
	
	return &TopicToolResult{
		Success: true,
		Message: fmt.Sprintf("[PLACEHOLDER] Would search topics for: '%s'. Found %d results.", query, len(results)),
		Data: map[string]interface{}{
			"query":   query,
			"results": results,
			"count":   len(results),
		},
	}
}

// WriteTopic writes or updates a topic file with new content (replaces existing content)
func (t *TopicKnowledgeTool) WriteTopic(name, content string) *TopicToolResult {
	// TODO: Integrate with Memory Manager's UpdateTopicFile method (task 16)
	
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
	
	filePath := fmt.Sprintf("memory/topics/%s.md", name)
	
	return &TopicToolResult{
		Success: true,
		Message: fmt.Sprintf("[PLACEHOLDER] Would write/update topic '%s' at %s", name, filePath),
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
	// TODO: Integrate with Memory Manager's CreateTopicFile method (task 16)
	
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
	
	filePath := fmt.Sprintf("memory/topics/%s.md", name)
	
	// Placeholder: Would check if file exists and fail if it does
	return &TopicToolResult{
		Success: true,
		Message: fmt.Sprintf("[PLACEHOLDER] Would create new topic '%s' at %s", name, filePath),
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
	// TODO: Integrate with Memory Manager's AppendToTopicFile method (task 16)
	
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
	
	filePath := fmt.Sprintf("memory/topics/%s.md", name)
	
	// Placeholder: Would check if file exists and append to it
	return &TopicToolResult{
		Success: true,
		Message: fmt.Sprintf("[PLACEHOLDER] Would append to topic '%s' at %s", name, filePath),
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
