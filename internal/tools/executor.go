package tools

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Tool interface defines the contract for all tools (local and remote)
type Tool interface {
	Name() string
	Description() string
	Execute(params map[string]interface{}) (ToolResult, error)
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Success bool
	Output  string
	Error   string
}

// ToolInfo provides information about a registered tool
type ToolInfo struct {
	Name        string
	Description string
	Type        string // "local" or "remote"
}

// ToolExecutor manages tool registration and execution
type ToolExecutor struct {
	tools map[string]Tool
	mu    sync.RWMutex
}

// NewToolExecutor creates a new ToolExecutor instance
func NewToolExecutor() *ToolExecutor {
	return &ToolExecutor{
		tools: make(map[string]Tool),
	}
}

// RegisterTool adds a tool to the registry
func (te *ToolExecutor) RegisterTool(name string, tool Tool) error {
	te.mu.Lock()
	defer te.mu.Unlock()

	if _, exists := te.tools[name]; exists {
		return fmt.Errorf("tool %s is already registered", name)
	}

	te.tools[name] = tool
	log.Printf("[ToolExecutor] Registered tool: %s - %s", name, tool.Description())
	return nil
}

// ExecuteTool executes the specified tool with the given parameters
func (te *ToolExecutor) ExecuteTool(name string, params map[string]interface{}) (ToolResult, error) {
	te.mu.RLock()
	tool, exists := te.tools[name]
	te.mu.RUnlock()

	if !exists {
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("tool %s not found", name),
		}, fmt.Errorf("tool %s not found", name)
	}

	// Log tool execution start
	startTime := time.Now()
	log.Printf("[ToolExecutor] [%s] Executing tool: %s", startTime.Format("2006-01-02 15:04:05"), name)

	// Execute the tool
	result, err := tool.Execute(params)

	// Log tool execution result
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	status := "success"
	if !result.Success || err != nil {
		status = "failed"
	}

	log.Printf("[ToolExecutor] [%s] Tool execution completed: %s | Status: %s | Duration: %v",
		endTime.Format("2006-01-02 15:04:05"), name, status, duration)

	if err != nil {
		log.Printf("[ToolExecutor] Tool execution error: %s - %v", name, err)
	}

	return result, err
}

// ListTools returns information about all registered tools
func (te *ToolExecutor) ListTools() []ToolInfo {
	te.mu.RLock()
	defer te.mu.RUnlock()

	tools := make([]ToolInfo, 0, len(te.tools))
	for name, tool := range te.tools {
		// Determine tool type based on implementation
		// For now, we'll default to "local" - this can be enhanced later
		toolType := "local"

		tools = append(tools, ToolInfo{
			Name:        name,
			Description: tool.Description(),
			Type:        toolType,
		})
	}

	return tools
}
