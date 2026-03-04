package llm

import (
	"testing"

	"simple-telegram-chatbot/internal/agent"
	"simple-telegram-chatbot/internal/session"
	"simple-telegram-chatbot/internal/tools"
	"simple-telegram-chatbot/pkg/utils"
)

// TestNewOpenRouterClient verifies client initialization
func TestNewOpenRouterClient(t *testing.T) {
	logger := utils.NewLogger("info")
	ag := agent.NewAgent("../../../agent", logger)
	sm := session.NewSessionManager()
	
	client := NewOpenRouterClient(
		"test-api-key",
		"test-model",
		ag,
		sm,
		nil, // toolExecutor not needed for this test
		logger,
	)
	
	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}
	
	if client.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey to be 'test-api-key', got '%s'", client.apiKey)
	}
	
	if client.modelName != "test-model" {
		t.Errorf("Expected modelName to be 'test-model', got '%s'", client.modelName)
	}
}

// TestBuildSystemContext verifies system context building from identity files
func TestBuildSystemContext(t *testing.T) {
	logger := utils.NewLogger("info")
	ag := agent.NewAgent("../../../agent", logger)
	sm := session.NewSessionManager()
	
	client := NewOpenRouterClient(
		"test-api-key",
		"test-model",
		ag,
		sm,
		nil, // toolExecutor not needed for this test
		logger,
	)
	
	identityContext := agent.IdentityContext{
		Identity:    "Test Identity",
		Personality: "Test Personality",
		Soul:        "Test Soul",
		User:        "Test User",
	}
	
	systemContext := client.buildSystemContext(identityContext)
	
	// Verify all sections are included
	if systemContext == "" {
		t.Fatal("Expected non-empty system context")
	}
	
	// Check that all identity components are present
	expectedSubstrings := []string{
		"# Identity",
		"Test Identity",
		"# Personality",
		"Test Personality",
		"# Soul",
		"Test Soul",
		"# User Context",
		"Test User",
	}
	
	for _, expected := range expectedSubstrings {
		if !contains(systemContext, expected) {
			t.Errorf("Expected system context to contain '%s'", expected)
		}
	}
}

// TestAssembleContextWithEmptyHistory verifies context assembly with no history
func TestAssembleContextWithEmptyHistory(t *testing.T) {
	logger := utils.NewLogger("info")
	ag := agent.NewAgent("../../../agent", logger)
	sm := session.NewSessionManager()
	
	// Load identity files
	_, err := ag.LoadIdentityFiles()
	if err != nil {
		t.Skipf("Skipping test - identity files not available: %v", err)
	}
	
	client := NewOpenRouterClient(
		"test-api-key",
		"test-model",
		ag,
		sm,
		nil, // toolExecutor not needed for this test
		logger,
	)
	
	chatID := int64(12345)
	userMessage := "Hello, bot!"
	
	messages, systemContext, err := client.AssembleContext(chatID, userMessage)
	if err != nil {
		t.Fatalf("AssembleContext failed: %v", err)
	}
	
	// Should have system message + user message (no history)
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}
	
	// First message should be system
	if messages[0].Role != "system" {
		t.Errorf("Expected first message role to be 'system', got '%s'", messages[0].Role)
	}
	
	if messages[0].Content != systemContext {
		t.Error("Expected first message content to match system context")
	}
	
	// Second message should be user
	if messages[1].Role != "user" {
		t.Errorf("Expected second message role to be 'user', got '%s'", messages[1].Role)
	}
	
	if messages[1].Content != userMessage {
		t.Errorf("Expected second message content to be '%s', got '%s'", userMessage, messages[1].Content)
	}
}

// TestAssembleContextWithHistory verifies context assembly with conversation history
func TestAssembleContextWithHistory(t *testing.T) {
	logger := utils.NewLogger("info")
	ag := agent.NewAgent("../../../agent", logger)
	sm := session.NewSessionManager()
	
	// Load identity files
	_, err := ag.LoadIdentityFiles()
	if err != nil {
		t.Skipf("Skipping test - identity files not available: %v", err)
	}
	
	client := NewOpenRouterClient(
		"test-api-key",
		"test-model",
		ag,
		sm,
		nil, // toolExecutor not needed for this test
		logger,
	)
	
	chatID := int64(12345)
	
	// Add some history
	sm.AppendMessage(chatID, "user", "First message")
	sm.AppendMessage(chatID, "assistant", "First response")
	sm.AppendMessage(chatID, "user", "Second message")
	sm.AppendMessage(chatID, "assistant", "Second response")
	
	userMessage := "Third message"
	
	messages, _, err := client.AssembleContext(chatID, userMessage)
	if err != nil {
		t.Fatalf("AssembleContext failed: %v", err)
	}
	
	// Should have: system + 4 history messages + current user message = 6 total
	expectedCount := 6
	if len(messages) != expectedCount {
		t.Errorf("Expected %d messages, got %d", expectedCount, len(messages))
	}
	
	// Verify message order
	if messages[0].Role != "system" {
		t.Error("First message should be system")
	}
	
	if messages[1].Role != "user" || messages[1].Content != "First message" {
		t.Error("Second message should be first user message from history")
	}
	
	if messages[2].Role != "assistant" || messages[2].Content != "First response" {
		t.Error("Third message should be first assistant response from history")
	}
	
	if messages[len(messages)-1].Role != "user" || messages[len(messages)-1].Content != userMessage {
		t.Error("Last message should be current user message")
	}
}

// TestExtractGeneratedText verifies text extraction from API response
func TestExtractGeneratedText(t *testing.T) {
	logger := utils.NewLogger("info")
	ag := agent.NewAgent("../../../agent", logger)
	sm := session.NewSessionManager()
	
	client := NewOpenRouterClient(
		"test-api-key",
		"test-model",
		ag,
		sm,
		nil, // toolExecutor not needed for this test
		logger,
	)
	
	tests := []struct {
		name        string
		response    OpenRouterResponse
		expected    string
		shouldError bool
	}{
		{
			name: "valid response",
			response: OpenRouterResponse{
				Choices: []Choice{
					{
						Message: Message{
							Role:    "assistant",
							Content: "Hello, user!",
						},
					},
				},
			},
			expected:    "Hello, user!",
			shouldError: false,
		},
		{
			name: "empty choices",
			response: OpenRouterResponse{
				Choices: []Choice{},
			},
			expected:    "",
			shouldError: true,
		},
		{
			name: "empty content",
			response: OpenRouterResponse{
				Choices: []Choice{
					{
						Message: Message{
							Role:    "assistant",
							Content: "",
						},
					},
				},
			},
			expected:    "I executed the command, but it returned no output. This could mean the data wasn't found or the command needs adjustment.",
			shouldError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.extractGeneratedText(tt.response)
			
			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected '%s', got '%s'", tt.expected, result)
				}
			}
		})
	}
}

// TestHandleHTTPError verifies HTTP error handling
func TestHandleHTTPError(t *testing.T) {
	logger := utils.NewLogger("info")
	ag := agent.NewAgent("../../../agent", logger)
	sm := session.NewSessionManager()
	
	client := NewOpenRouterClient(
		"test-api-key",
		"test-model",
		ag,
		sm,
		nil, // toolExecutor not needed for this test
		logger,
	)
	
	tests := []struct {
		name       string
		statusCode int
		body       []byte
		contains   string
	}{
		{
			name:       "unauthorized",
			statusCode: 401,
			body:       []byte("{}"),
			contains:   "authentication failed",
		},
		{
			name:       "rate limit",
			statusCode: 429,
			body:       []byte("{}"),
			contains:   "rate limit exceeded",
		},
		{
			name:       "bad request",
			statusCode: 400,
			body:       []byte("{}"),
			contains:   "invalid request",
		},
		{
			name:       "internal server error",
			statusCode: 500,
			body:       []byte("{}"),
			contains:   "temporarily unavailable",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.handleHTTPError(tt.statusCode, tt.body)
			
			if err == nil {
				t.Fatal("Expected error but got none")
			}
			
			if !contains(err.Error(), tt.contains) {
				t.Errorf("Expected error to contain '%s', got '%s'", tt.contains, err.Error())
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestBuildToolDefinitions verifies tool definitions are built correctly
func TestBuildToolDefinitions(t *testing.T) {
	logger := utils.NewLogger("info")
	ag := agent.NewAgent("../../../agent", logger)
	sm := session.NewSessionManager()
	
	t.Run("no tool executor", func(t *testing.T) {
		client := NewOpenRouterClient(
			"test-api-key",
			"test-model",
			ag,
			sm,
			nil, // no tool executor
			logger,
		)
		
		definitions := client.buildToolDefinitions()
		if definitions != nil {
			t.Errorf("Expected nil definitions when no tool executor, got %d definitions", len(definitions))
		}
	})
	
	t.Run("with shell tool", func(t *testing.T) {
		// Create tool executor and register shell tool
		te := &mockToolExecutor{
			tools: []tools.ToolInfo{
				{
					Name:        "shell_tool",
					Description: "Executes shell commands",
					Type:        "local",
				},
			},
		}
		
		client := NewOpenRouterClient(
			"test-api-key",
			"test-model",
			ag,
			sm,
			te,
			logger,
		)
		
		definitions := client.buildToolDefinitions()
		
		if len(definitions) != 1 {
			t.Fatalf("Expected 1 tool definition, got %d", len(definitions))
		}
		
		def := definitions[0]
		if def.Type != "function" {
			t.Errorf("Expected type 'function', got '%s'", def.Type)
		}
		
		if def.Function.Name != "shell_tool" {
			t.Errorf("Expected name 'shell_tool', got '%s'", def.Function.Name)
		}
		
		if def.Function.Description != "Executes shell commands" {
			t.Errorf("Expected description 'Executes shell commands', got '%s'", def.Function.Description)
		}
		
		// Verify parameters schema
		if def.Function.Parameters == nil {
			t.Fatal("Expected parameters to be defined")
		}
		
		params := def.Function.Parameters
		if params["type"] != "object" {
			t.Errorf("Expected parameters type 'object', got '%v'", params["type"])
		}
	})
}

// TestHandleToolRequest verifies tool request handling
func TestHandleToolRequest(t *testing.T) {
	logger := utils.NewLogger("info")
	ag := agent.NewAgent("../../../agent", logger)
	sm := session.NewSessionManager()
	
	t.Run("successful tool execution", func(t *testing.T) {
		te := &mockToolExecutor{
			executeFunc: func(name string, params map[string]interface{}) (tools.ToolResult, error) {
				if name != "shell_tool" {
					t.Errorf("Expected tool name 'shell_tool', got '%s'", name)
				}
				
				if params["command"] != "echo hello" {
					t.Errorf("Expected command 'echo hello', got '%v'", params["command"])
				}
				
				return tools.ToolResult{
					Success: true,
					Output:  "hello\n",
					Error:   "",
				}, nil
			},
		}
		
		client := NewOpenRouterClient(
			"test-api-key",
			"test-model",
			ag,
			sm,
			te,
			logger,
		)
		
		toolCall := ToolCall{
			ID:   "call_123",
			Type: "function",
			Function: Function{
				Name:      "shell_tool",
				Arguments: `{"command":"echo hello"}`,
			},
		}
		
		result, err := client.HandleToolRequest(toolCall)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if !result.Success {
			t.Error("Expected successful result")
		}
		
		if result.Output != "hello\n" {
			t.Errorf("Expected output 'hello\\n', got '%s'", result.Output)
		}
	})
	
	t.Run("invalid JSON arguments", func(t *testing.T) {
		te := &mockToolExecutor{}
		
		client := NewOpenRouterClient(
			"test-api-key",
			"test-model",
			ag,
			sm,
			te,
			logger,
		)
		
		toolCall := ToolCall{
			ID:   "call_123",
			Type: "function",
			Function: Function{
				Name:      "shell_tool",
				Arguments: `{invalid json}`,
			},
		}
		
		result, err := client.HandleToolRequest(toolCall)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
		
		if result.Success {
			t.Error("Expected unsuccessful result for invalid JSON")
		}
	})
	
	t.Run("tool execution error", func(t *testing.T) {
		te := &mockToolExecutor{
			executeFunc: func(name string, params map[string]interface{}) (tools.ToolResult, error) {
				return tools.ToolResult{
					Success: false,
					Output:  "",
					Error:   "command not found",
				}, nil
			},
		}
		
		client := NewOpenRouterClient(
			"test-api-key",
			"test-model",
			ag,
			sm,
			te,
			logger,
		)
		
		toolCall := ToolCall{
			ID:   "call_123",
			Type: "function",
			Function: Function{
				Name:      "shell_tool",
				Arguments: `{"command":"nonexistent"}`,
			},
		}
		
		result, _ := client.HandleToolRequest(toolCall)
		// err might be nil even if tool execution failed
		
		if result.Success {
			t.Error("Expected unsuccessful result")
		}
		
		if result.Error != "command not found" {
			t.Errorf("Expected error 'command not found', got '%s'", result.Error)
		}
	})
}

// Mock types for testing - these implement the minimal interface needed by OpenRouterClient
type mockToolExecutor struct {
	tools       []tools.ToolInfo
	executeFunc func(name string, params map[string]interface{}) (tools.ToolResult, error)
}

func (m *mockToolExecutor) ListTools() []tools.ToolInfo {
	return m.tools
}

func (m *mockToolExecutor) ExecuteTool(name string, params map[string]interface{}) (tools.ToolResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(name, params)
	}
	return tools.ToolResult{Success: true, Output: "mock output"}, nil
}

