package tools

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"
)

// ShellTool executes shell commands on the local system
type ShellTool struct {
	timeout time.Duration
}

// NewShellTool creates a new ShellTool with the specified timeout
func NewShellTool(timeout time.Duration) *ShellTool {
	return &ShellTool{
		timeout: timeout,
	}
}

// Name returns the tool name
func (s *ShellTool) Name() string {
	return "shell_tool"
}

// Description returns the tool description
func (s *ShellTool) Description() string {
	return "Executes shell commands on the local system and returns stdout, stderr, and exit code"
}

// Execute runs a shell command with timeout and captures output
func (s *ShellTool) Execute(params map[string]interface{}) (ToolResult, error) {
	// Extract command from parameters
	commandInterface, ok := params["command"]
	if !ok {
		return ToolResult{
			Success: false,
			Error:   "missing required parameter: command",
		}, fmt.Errorf("missing required parameter: command")
	}

	command, ok := commandInterface.(string)
	if !ok {
		return ToolResult{
			Success: false,
			Error:   "command parameter must be a string",
		}, fmt.Errorf("command parameter must be a string")
	}

	// Log command execution
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[ShellTool] [%s] Executing command: %s", timestamp, command)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	// Execute command using sh -c to support shell features
	cmd := exec.CommandContext(ctx, "sh", "-c", command)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	err := cmd.Run()

	// Check for timeout first
	if ctx.Err() == context.DeadlineExceeded {
		// Timeout occurred
		log.Printf("[ShellTool] [%s] Command timed out after %v: %s | Exit code: -1",
			time.Now().Format("2006-01-02 15:04:05"), s.timeout, command)
		return ToolResult{
			Success: false,
			Output:  stdout.String(),
			Error:   fmt.Sprintf("command timed out after %v\nstderr: %s", s.timeout, stderr.String()),
		}, fmt.Errorf("command timed out after %v", s.timeout)
	}

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// Other execution error
			exitCode = -1
		}
	}

	// Determine success based on exit code
	success := exitCode == 0

	// Log execution result
	log.Printf("[ShellTool] [%s] Command completed: %s | Exit code: %d",
		time.Now().Format("2006-01-02 15:04:05"), command, exitCode)

	// Build result
	result := ToolResult{
		Success: success,
		Output:  stdout.String(),
	}

	if !success {
		result.Error = fmt.Sprintf("exit code: %d\nstderr: %s", exitCode, stderr.String())
	}

	return result, nil
}
