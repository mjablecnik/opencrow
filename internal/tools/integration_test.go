package tools

import (
	"strings"
	"testing"
	"time"
)

// TestShellToolIntegration verifies Shell Tool works with ToolExecutor
func TestShellToolIntegration(t *testing.T) {
	// Create executor
	executor := NewToolExecutor()
	
	// Create and register shell tool
	shellTool := NewShellTool(30 * time.Second)
	err := executor.RegisterTool("shell_tool", shellTool)
	if err != nil {
		t.Fatalf("Failed to register shell tool: %v", err)
	}
	
	// Verify tool is listed
	tools := executor.ListTools()
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
	
	if tools[0].Name != "shell_tool" {
		t.Errorf("Expected tool name 'shell_tool', got '%s'", tools[0].Name)
	}
	
	// Execute a simple command
	params := map[string]interface{}{
		"command": "echo 'Integration test'",
	}
	
	result, err := executor.ExecuteTool("shell_tool", params)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if !result.Success {
		t.Error("Expected successful execution")
	}
	
	if !strings.Contains(result.Output, "Integration test") {
		t.Errorf("Expected output to contain 'Integration test', got: %s", result.Output)
	}
}

// TestShellToolWithConfigTimeout verifies Shell Tool respects configured timeout
func TestShellToolWithConfigTimeout(t *testing.T) {
	// Create shell tool with short timeout
	shellTool := NewShellTool(500 * time.Millisecond)
	
	// Execute command that exceeds timeout
	params := map[string]interface{}{
		"command": "sleep 2",
	}
	
	result, err := shellTool.Execute(params)
	
	if err == nil {
		t.Error("Expected timeout error")
	}
	
	if result.Success {
		t.Error("Expected failure due to timeout")
	}
	
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout error message, got: %v", err)
	}
}
