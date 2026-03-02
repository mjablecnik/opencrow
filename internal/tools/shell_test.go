package tools

import (
	"strings"
	"testing"
	"time"
)

func TestShellTool_Name(t *testing.T) {
	tool := NewShellTool(30 * time.Second)
	if tool.Name() != "shell_tool" {
		t.Errorf("Expected name 'shell_tool', got '%s'", tool.Name())
	}
}

func TestShellTool_Description(t *testing.T) {
	tool := NewShellTool(30 * time.Second)
	desc := tool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
}

func TestShellTool_Execute_Success(t *testing.T) {
	tool := NewShellTool(30 * time.Second)
	
	params := map[string]interface{}{
		"command": "echo 'Hello World'",
	}
	
	result, err := tool.Execute(params)
	
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if !result.Success {
		t.Error("Expected success to be true")
	}
	
	if !strings.Contains(result.Output, "Hello World") {
		t.Errorf("Expected output to contain 'Hello World', got: %s", result.Output)
	}
}

func TestShellTool_Execute_CaptureStderr(t *testing.T) {
	tool := NewShellTool(30 * time.Second)
	
	params := map[string]interface{}{
		"command": "echo 'error message' >&2",
	}
	
	result, err := tool.Execute(params)
	
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if !result.Success {
		t.Error("Expected success to be true (exit code 0)")
	}
}

func TestShellTool_Execute_FailedCommand(t *testing.T) {
	tool := NewShellTool(30 * time.Second)
	
	params := map[string]interface{}{
		"command": "exit 1",
	}
	
	result, err := tool.Execute(params)
	
	if err != nil {
		t.Errorf("Expected no error from Execute, got: %v", err)
	}
	
	if result.Success {
		t.Error("Expected success to be false for failed command")
	}
	
	if !strings.Contains(result.Error, "exit code: 1") {
		t.Errorf("Expected error to contain exit code, got: %s", result.Error)
	}
}

func TestShellTool_Execute_Timeout(t *testing.T) {
	tool := NewShellTool(1 * time.Second)
	
	params := map[string]interface{}{
		"command": "sleep 5",
	}
	
	result, err := tool.Execute(params)
	
	if err == nil {
		t.Error("Expected timeout error")
	}
	
	if result.Success {
		t.Error("Expected success to be false for timed out command")
	}
	
	if !strings.Contains(result.Error, "timed out") {
		t.Errorf("Expected error to contain 'timed out', got: %s", result.Error)
	}
}

func TestShellTool_Execute_MissingCommand(t *testing.T) {
	tool := NewShellTool(30 * time.Second)
	
	params := map[string]interface{}{}
	
	result, err := tool.Execute(params)
	
	if err == nil {
		t.Error("Expected error for missing command parameter")
	}
	
	if result.Success {
		t.Error("Expected success to be false")
	}
	
	if !strings.Contains(result.Error, "missing required parameter") {
		t.Errorf("Expected error about missing parameter, got: %s", result.Error)
	}
}

func TestShellTool_Execute_InvalidCommandType(t *testing.T) {
	tool := NewShellTool(30 * time.Second)
	
	params := map[string]interface{}{
		"command": 123, // Invalid type
	}
	
	result, err := tool.Execute(params)
	
	if err == nil {
		t.Error("Expected error for invalid command type")
	}
	
	if result.Success {
		t.Error("Expected success to be false")
	}
	
	if !strings.Contains(result.Error, "must be a string") {
		t.Errorf("Expected error about string type, got: %s", result.Error)
	}
}

func TestShellTool_Execute_CapturesBothOutputs(t *testing.T) {
	tool := NewShellTool(30 * time.Second)
	
	params := map[string]interface{}{
		"command": "echo 'stdout message' && echo 'stderr message' >&2 && exit 1",
	}
	
	result, err := tool.Execute(params)
	
	if err != nil {
		t.Errorf("Expected no error from Execute, got: %v", err)
	}
	
	if result.Success {
		t.Error("Expected success to be false")
	}
	
	if !strings.Contains(result.Output, "stdout message") {
		t.Errorf("Expected stdout to be captured, got: %s", result.Output)
	}
	
	if !strings.Contains(result.Error, "stderr message") {
		t.Errorf("Expected stderr to be in error field, got: %s", result.Error)
	}
}
