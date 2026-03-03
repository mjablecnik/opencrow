package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"simple-telegram-chatbot/internal/agent"
	"simple-telegram-chatbot/internal/session"
	"simple-telegram-chatbot/internal/tools"
	"simple-telegram-chatbot/pkg/utils"
)

const (
	openRouterAPIURL = "https://openrouter.ai/api/v1/chat/completions"
	requestTimeout   = 60 * time.Second
)

// ToolExecutorInterface defines the interface for tool execution
// This allows for testing with mocks while using the real tools.ToolExecutor in production
type ToolExecutorInterface interface {
	ListTools() []tools.ToolInfo
	ExecuteTool(name string, params map[string]interface{}) (tools.ToolResult, error)
}

// Message represents a message in the OpenRouter API format
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// OpenRouterRequest represents the request payload for OpenRouter API
type OpenRouterRequest struct {
	Model    string           `json:"model"`
	Messages []Message        `json:"messages"`
	Tools    []ToolDefinition `json:"tools,omitempty"`
}

// ToolDefinition represents a tool that can be called by the LLM
type ToolDefinition struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition describes a function tool
type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// OpenRouterResponse represents the response from OpenRouter API
type OpenRouterResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
	Error   *APIError `json:"error,omitempty"`
}

// Choice represents a single choice in the API response
type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// ToolCall represents a tool execution request from the LLM
type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function represents the function call details
type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// APIError represents an error from the OpenRouter API
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// OpenRouterClient manages communication with the OpenRouter API
type OpenRouterClient struct {
	apiKey         string
	modelName      string
	agent          *agent.Agent
	sessionManager *session.SessionManager
	toolExecutor   ToolExecutorInterface
	httpClient     *http.Client
	logger         *utils.Logger
}

// NewOpenRouterClient creates a new OpenRouter client instance
func NewOpenRouterClient(
	apiKey string,
	modelName string,
	agent *agent.Agent,
	sessionManager *session.SessionManager,
	toolExecutor ToolExecutorInterface,
	logger *utils.Logger,
) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey:         apiKey,
		modelName:      modelName,
		agent:          agent,
		sessionManager: sessionManager,
		toolExecutor:   toolExecutor,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
		logger: logger,
	}
}

// LoadIdentityFiles loads identity files from the agent component
func (c *OpenRouterClient) LoadIdentityFiles() (agent.IdentityContext, error) {
	c.logger.DebugWithComponent("OpenRouterClient", "Loading identity files from agent")
	
	identityContext := c.agent.GetIdentityContext()
	
	if identityContext.Identity == "" {
		return agent.IdentityContext{}, fmt.Errorf("identity files not loaded in agent")
	}
	
	c.logger.DebugWithComponent("OpenRouterClient", "Successfully loaded identity files")
	return identityContext, nil
}

// AssembleContext builds the message array from session history and identity files
func (c *OpenRouterClient) AssembleContext(chatID int64, userMessage string) ([]Message, string, error) {
	c.logger.DebugWithComponent("OpenRouterClient", "Assembling context", "chatID", chatID)
	
	// Load identity files
	identityContext, err := c.LoadIdentityFiles()
	if err != nil {
		return nil, "", fmt.Errorf("failed to load identity files: %w", err)
	}
	
	// Build system context from identity files
	systemContext := c.buildSystemContext(identityContext)
	
	// Get conversation history from session manager
	history, err := c.sessionManager.GetHistory(chatID)
	if err != nil {
		c.logger.WarnWithComponent("OpenRouterClient", "Failed to get session history", "chatID", chatID, "error", err)
		history = []session.Message{} // Use empty history on error
	}
	
	// Build messages array
	messages := []Message{
		{
			Role:    "system",
			Content: systemContext,
		},
	}
	
	// Add conversation history
	for _, msg := range history {
		messages = append(messages, Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	
	// Add current user message
	messages = append(messages, Message{
		Role:    "user",
		Content: userMessage,
	})
	
	c.logger.DebugWithComponent("OpenRouterClient", "Context assembled", 
		"chatID", chatID, 
		"historyMessages", len(history),
		"totalMessages", len(messages))
	
	return messages, systemContext, nil
}

// buildSystemContext creates the system context from identity files
func (c *OpenRouterClient) buildSystemContext(identityContext agent.IdentityContext) string {
	var systemContext string
	
	if identityContext.Identity != "" {
		systemContext += "# Identity\n\n" + identityContext.Identity + "\n\n"
	}
	
	if identityContext.Personality != "" {
		systemContext += "# Personality\n\n" + identityContext.Personality + "\n\n"
	}
	
	if identityContext.Soul != "" {
		systemContext += "# Soul\n\n" + identityContext.Soul + "\n\n"
	}
	
	if identityContext.User != "" {
		systemContext += "# User Context\n\n" + identityContext.User + "\n\n"
	}
	
	return systemContext
}

// SendRequest sends a request to the OpenRouter API with full context
func (c *OpenRouterClient) SendRequest(ctx context.Context, chatID int64, userMessage string) (string, error) {
	c.logger.InfoWithComponent("OpenRouterClient", "Sending request to OpenRouter API", 
		"chatID", chatID, 
		"model", c.modelName)
	
	// Assemble context
	messages, _, err := c.AssembleContext(chatID, userMessage)
	if err != nil {
		return "", fmt.Errorf("failed to assemble context: %w", err)
	}
	
	// Build tool definitions
	toolDefinitions := c.buildToolDefinitions()
	
	// Build request payload
	requestPayload := OpenRouterRequest{
		Model:    c.modelName,
		Messages: messages,
		Tools:    toolDefinitions,
	}
	
	// Marshal request to JSON
	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Log request at debug level
	c.logger.DebugWithComponent("OpenRouterClient", "API Request", 
		"model", c.modelName,
		"messageCount", len(messages),
		"toolCount", len(toolDefinitions))
	
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", openRouterAPIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/simple-telegram-chatbot")
	req.Header.Set("X-Title", "Simple Telegram Chatbot")
	
	// Send request
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.ErrorWithComponent("OpenRouterClient", "HTTP request failed", "error", err)
		return "", c.handleAPIError(err)
	}
	defer resp.Body.Close()
	
	duration := time.Since(startTime)
	c.logger.DebugWithComponent("OpenRouterClient", "API response received", 
		"statusCode", resp.StatusCode,
		"duration", duration)
	
	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	
	// Log response at debug level
	c.logger.DebugWithComponent("OpenRouterClient", "API Response body received", 
		"bodyLength", len(responseBody))
	
	// Log full response body for debugging
	c.logger.DebugWithComponent("OpenRouterClient", "API Response body content", 
		"body", string(responseBody))
	
	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		c.logger.ErrorWithComponent("OpenRouterClient", "API returned error status", 
			"statusCode", resp.StatusCode,
			"body", string(responseBody))
		return "", c.handleHTTPError(resp.StatusCode, responseBody)
	}
	
	// Parse response
	var apiResponse OpenRouterResponse
	if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	
	// Check for API error in response
	if apiResponse.Error != nil {
		c.logger.ErrorWithComponent("OpenRouterClient", "API returned error", 
			"errorType", apiResponse.Error.Type,
			"errorMessage", apiResponse.Error.Message)
		return "", fmt.Errorf("API error: %s", apiResponse.Error.Message)
	}
	
	// Debug: Log choices count and structure
	c.logger.DebugWithComponent("OpenRouterClient", "Checking for tool calls", 
		"choicesCount", len(apiResponse.Choices))
	
	if len(apiResponse.Choices) > 0 {
		c.logger.DebugWithComponent("OpenRouterClient", "First choice details", 
			"toolCallsCount", len(apiResponse.Choices[0].Message.ToolCalls),
			"messageContent", apiResponse.Choices[0].Message.Content)
	}
	
	// Check if LLM requested tool execution
	if len(apiResponse.Choices) > 0 && len(apiResponse.Choices[0].Message.ToolCalls) > 0 {
		c.logger.InfoWithComponent("OpenRouterClient", "LLM requested tool execution", 
			"chatID", chatID,
			"toolCallCount", len(apiResponse.Choices[0].Message.ToolCalls))
		
		c.logger.DebugWithComponent("OpenRouterClient", "Tool calls details", 
			"toolCalls", fmt.Sprintf("%+v", apiResponse.Choices[0].Message.ToolCalls))
		
		// Handle tool calls and get final response
		return c.handleToolCalls(ctx, chatID, apiResponse.Choices[0].Message.ToolCalls, messages)
	}
	
	// Extract generated text (no tool calls)
	c.logger.DebugWithComponent("OpenRouterClient", "No tool calls, extracting text content")
	generatedText, err := c.extractGeneratedText(apiResponse)
	if err != nil {
		return "", err
	}
	
	c.logger.InfoWithComponent("OpenRouterClient", "Successfully received response from API", 
		"chatID", chatID,
		"responseLength", len(generatedText),
		"tokensUsed", apiResponse.Usage.TotalTokens)
	
	return generatedText, nil
}

// extractGeneratedText extracts the generated text from the API response
func (c *OpenRouterClient) extractGeneratedText(response OpenRouterResponse) (string, error) {
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in API response")
	}
	
	choice := response.Choices[0]
	
	// If content is empty, check if there were tool calls that returned empty results
	// In this case, provide a helpful fallback message
	if choice.Message.Content == "" {
		c.logger.WarnWithComponent("OpenRouterClient", "Empty content in API response, using fallback message")
		return "I executed the command, but it returned no output. This could mean the data wasn't found or the command needs adjustment.", nil
	}
	
	return choice.Message.Content, nil
}

// handleAPIError converts API errors to user-friendly messages
func (c *OpenRouterClient) handleAPIError(err error) error {
	// Network or timeout errors
	return fmt.Errorf("unable to connect to AI service. Please try again later")
}

// handleHTTPError converts HTTP status codes to user-friendly messages
func (c *OpenRouterClient) handleHTTPError(statusCode int, body []byte) error {
	switch statusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("authentication failed. Please check API configuration")
	case http.StatusTooManyRequests:
		return fmt.Errorf("rate limit exceeded. Please try again in a moment")
	case http.StatusBadRequest:
		// Try to parse error message from body
		var apiResponse OpenRouterResponse
		if err := json.Unmarshal(body, &apiResponse); err == nil && apiResponse.Error != nil {
			return fmt.Errorf("request error: %s", apiResponse.Error.Message)
		}
		return fmt.Errorf("invalid request. Please try again")
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return fmt.Errorf("AI service is temporarily unavailable. Please try again later")
	default:
		return fmt.Errorf("unexpected error occurred (status %d). Please try again", statusCode)
	}
}

// buildToolDefinitions creates tool definitions for the API request
func (c *OpenRouterClient) buildToolDefinitions() []ToolDefinition {
	if c.toolExecutor == nil {
		return nil
	}

	toolInfos := c.toolExecutor.ListTools()
	if len(toolInfos) == 0 {
		return nil
	}

	definitions := make([]ToolDefinition, 0, len(toolInfos))
	
	for _, toolInfo := range toolInfos {
		// Build parameters schema based on tool name
		var parameters map[string]interface{}
		
		if toolInfo.Name == "shell_tool" {
			parameters = map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "The shell command to execute",
					},
				},
				"required": []string{"command"},
			}
		} else {
			// Default parameters schema for unknown tools
			parameters = map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			}
		}

		definitions = append(definitions, ToolDefinition{
			Type: "function",
			Function: FunctionDefinition{
				Name:        toolInfo.Name,
				Description: toolInfo.Description,
				Parameters:  parameters,
			},
		})
	}

	c.logger.DebugWithComponent("OpenRouterClient", "Built tool definitions", "count", len(definitions))
	return definitions
}

// HandleToolRequest routes a tool execution request to the Tool_Executor
func (c *OpenRouterClient) HandleToolRequest(toolCall ToolCall) (tools.ToolResult, error) {
	c.logger.InfoWithComponent("OpenRouterClient", "Handling tool request", 
		"toolName", toolCall.Function.Name,
		"toolCallID", toolCall.ID)

	// Parse arguments from JSON string
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
		c.logger.ErrorWithComponent("OpenRouterClient", "Failed to parse tool arguments", 
			"error", err,
			"arguments", toolCall.Function.Arguments)
		return tools.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to parse tool arguments: %v", err),
		}, fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	// Execute the tool via Tool_Executor
	result, err := c.toolExecutor.ExecuteTool(toolCall.Function.Name, params)
	
	if err != nil {
		c.logger.WarnWithComponent("OpenRouterClient", "Tool execution failed", 
			"toolName", toolCall.Function.Name,
			"error", err)
	} else {
		c.logger.InfoWithComponent("OpenRouterClient", "Tool execution completed", 
			"toolName", toolCall.Function.Name,
			"success", result.Success)
	}

	return result, err
}

// handleToolCalls processes tool calls from the LLM and returns the final response
func (c *OpenRouterClient) handleToolCalls(ctx context.Context, chatID int64, toolCalls []ToolCall, messages []Message) (string, error) {
	c.logger.InfoWithComponent("OpenRouterClient", "Processing tool calls", 
		"chatID", chatID,
		"toolCallCount", len(toolCalls))

	// Execute each tool call and collect results
	toolMessages := make([]Message, 0, len(toolCalls))
	
	for _, toolCall := range toolCalls {
		result, err := c.HandleToolRequest(toolCall)
		
		// Build tool result message
		var resultContent string
		if err != nil {
			resultContent = fmt.Sprintf("Tool execution error: %s\nError details: %s", 
				toolCall.Function.Name, result.Error)
		} else if result.Success {
			resultContent = fmt.Sprintf("Tool: %s\nOutput: %s", 
				toolCall.Function.Name, result.Output)
		} else {
			resultContent = fmt.Sprintf("Tool: %s\nError: %s", 
				toolCall.Function.Name, result.Error)
		}

		// Add tool result as a message
		toolMessages = append(toolMessages, Message{
			Role:       "tool",
			Content:    resultContent,
			ToolCallID: toolCall.ID,
		})
	}

	// Send tool results back to LLM for final response
	c.logger.DebugWithComponent("OpenRouterClient", "Sending tool results back to LLM", 
		"chatID", chatID,
		"toolResultCount", len(toolMessages))

	// Append tool results to messages
	finalMessages := append(messages, toolMessages...)

	// Build request with tool results
	requestPayload := OpenRouterRequest{
		Model:    c.modelName,
		Messages: finalMessages,
		Tools:    c.buildToolDefinitions(),
	}

	// Marshal and send request
	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal follow-up request: %w", err)
	}

	c.logger.DebugWithComponent("OpenRouterClient", "Follow-up request body", 
		"body", string(requestBody))

	req, err := http.NewRequestWithContext(ctx, "POST", openRouterAPIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create follow-up HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/simple-telegram-chatbot")
	req.Header.Set("X-Title", "Simple Telegram Chatbot")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.ErrorWithComponent("OpenRouterClient", "Follow-up HTTP request failed", "error", err)
		return "", c.handleAPIError(err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read follow-up response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		c.logger.ErrorWithComponent("OpenRouterClient", "Follow-up API returned error status", 
			"statusCode", resp.StatusCode,
			"body", string(responseBody))
		return "", c.handleHTTPError(resp.StatusCode, responseBody)
	}

	// Parse response
	var apiResponse OpenRouterResponse
	if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
		return "", fmt.Errorf("failed to parse follow-up API response: %w", err)
	}

	// Check for API error
	if apiResponse.Error != nil {
		c.logger.ErrorWithComponent("OpenRouterClient", "Follow-up API returned error", 
			"errorType", apiResponse.Error.Type,
			"errorMessage", apiResponse.Error.Message)
		return "", fmt.Errorf("API error: %s", apiResponse.Error.Message)
	}

	// Extract final response
	generatedText, err := c.extractGeneratedText(apiResponse)
	if err != nil {
		return "", err
	}

	c.logger.InfoWithComponent("OpenRouterClient", "Successfully received final response after tool execution", 
		"chatID", chatID,
		"responseLength", len(generatedText))

	return generatedText, nil
}

// ExtractTopics extracts domain-specific topics from conversation content using LLM
// Returns a list of TopicExtraction with ShouldWrite=true if relevant domain knowledge is found
// Retries up to 3 times with exponential backoff on failures
func (c *OpenRouterClient) ExtractTopics(content string, existingTopics []string) ([]TopicExtraction, error) {
	c.logger.InfoWithComponent("OpenRouterClient", "Extracting topics from content",
		"contentLength", len(content),
		"existingTopicsCount", len(existingTopics))

	// Build the prompt for topic extraction
	prompt := c.buildTopicExtractionPrompt(content, existingTopics)

	// Retry logic with exponential backoff
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			// Exponential backoff: 1s, 2s, 4s
			backoffDuration := time.Duration(1<<uint(attempt-2)) * time.Second
			c.logger.InfoWithComponent("OpenRouterClient", "Retrying topic extraction",
				"attempt", attempt,
				"backoff", backoffDuration)
			time.Sleep(backoffDuration)
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		// Build request payload
		messages := []Message{
			{
				Role:    "system",
				Content: "You are an expert at analyzing conversations and extracting domain-specific knowledge. Your task is to identify relevant topics and extract meaningful insights.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		}

		requestPayload := OpenRouterRequest{
			Model:    c.modelName,
			Messages: messages,
		}

		// Marshal request to JSON
		requestBody, err := json.Marshal(requestPayload)
		if err != nil {
			lastErr = fmt.Errorf("failed to marshal topic extraction request: %w", err)
			c.logger.ErrorWithComponent("OpenRouterClient", "Failed to marshal request", "error", err)
			continue
		}

		// Create HTTP request
		req, err := http.NewRequestWithContext(ctx, "POST", openRouterAPIURL, bytes.NewBuffer(requestBody))
		if err != nil {
			lastErr = fmt.Errorf("failed to create HTTP request: %w", err)
			c.logger.ErrorWithComponent("OpenRouterClient", "Failed to create HTTP request", "error", err)
			continue
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("HTTP-Referer", "https://github.com/simple-telegram-chatbot")
		req.Header.Set("X-Title", "Simple Telegram Chatbot")

		// Send request
		startTime := time.Now()
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %w", err)
			c.logger.ErrorWithComponent("OpenRouterClient", "HTTP request failed", "error", err, "attempt", attempt)
			continue
		}

		duration := time.Since(startTime)
		c.logger.DebugWithComponent("OpenRouterClient", "Topic extraction API response received",
			"statusCode", resp.StatusCode,
			"duration", duration,
			"attempt", attempt)

		// Read response body
		responseBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			c.logger.ErrorWithComponent("OpenRouterClient", "Failed to read response body", "error", err)
			continue
		}

		// Check for HTTP errors
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API returned error status: %d", resp.StatusCode)
			c.logger.ErrorWithComponent("OpenRouterClient", "API returned error status",
				"statusCode", resp.StatusCode,
				"body", string(responseBody))
			continue
		}

		// Parse response
		var apiResponse OpenRouterResponse
		if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
			lastErr = fmt.Errorf("failed to parse API response: %w", err)
			c.logger.ErrorWithComponent("OpenRouterClient", "Failed to parse API response", "error", err)
			continue
		}

		// Check for API error in response
		if apiResponse.Error != nil {
			lastErr = fmt.Errorf("API error: %s", apiResponse.Error.Message)
			c.logger.ErrorWithComponent("OpenRouterClient", "API returned error",
				"errorType", apiResponse.Error.Type,
				"errorMessage", apiResponse.Error.Message)
			continue
		}

		// Extract generated text
		generatedText, err := c.extractGeneratedText(apiResponse)
		if err != nil {
			lastErr = fmt.Errorf("failed to extract generated text: %w", err)
			c.logger.ErrorWithComponent("OpenRouterClient", "Failed to extract generated text", "error", err)
			continue
		}

		// Parse topics from the response
		topics, err := c.parseTopicExtractionResponse(generatedText)
		if err != nil {
			lastErr = fmt.Errorf("failed to parse topic extraction response: %w", err)
			c.logger.ErrorWithComponent("OpenRouterClient", "Failed to parse topic extraction response", "error", err)
			continue
		}

		c.logger.InfoWithComponent("OpenRouterClient", "Successfully extracted topics",
			"topicsCount", len(topics),
			"attempt", attempt,
			"tokensUsed", apiResponse.Usage.TotalTokens)

		return topics, nil
	}

	// All retries failed
	c.logger.ErrorWithComponent("OpenRouterClient", "All topic extraction retries failed",
		"maxRetries", maxRetries,
		"lastError", lastErr)

	// Return empty list instead of error to allow graceful degradation
	return []TopicExtraction{}, nil
}

// buildTopicExtractionPrompt creates the prompt for topic extraction
func (c *OpenRouterClient) buildTopicExtractionPrompt(content string, existingTopics []string) string {
	var prompt strings.Builder

	prompt.WriteString("Analyze the following conversation content and extract domain-specific knowledge that should be preserved.\n\n")
	prompt.WriteString("**Task:** Identify relevant topics and extract meaningful insights, lessons learned, preferences, or important information.\n\n")
	prompt.WriteString("**Supported Domains:** Programming, Psychology, Food, Sport-Health, Politics, and other relevant domains.\n\n")

	if len(existingTopics) > 0 {
		prompt.WriteString("**Existing Topics:** ")
		prompt.WriteString(strings.Join(existingTopics, ", "))
		prompt.WriteString("\n\n")
		prompt.WriteString("You can update existing topics or create new ones as needed.\n\n")
	}

	prompt.WriteString("**Instructions:**\n")
	prompt.WriteString("1. Identify if there is any relevant domain-specific knowledge worth preserving\n")
	prompt.WriteString("2. If NO relevant domain knowledge is found, respond with: NO_TOPICS_FOUND\n")
	prompt.WriteString("3. If relevant knowledge IS found, extract topics in the following JSON format:\n\n")
	prompt.WriteString("```json\n")
	prompt.WriteString("[\n")
	prompt.WriteString("  {\n")
	prompt.WriteString("    \"topic_name\": \"TopicName\",\n")
	prompt.WriteString("    \"content\": \"Detailed content about this topic...\",\n")
	prompt.WriteString("    \"confidence\": 0.85,\n")
	prompt.WriteString("    \"should_write\": true\n")
	prompt.WriteString("  }\n")
	prompt.WriteString("]\n")
	prompt.WriteString("```\n\n")
	prompt.WriteString("**Guidelines:**\n")
	prompt.WriteString("- topic_name: Use descriptive names like \"Programming\", \"Docker\", \"Psychology\", \"Food-Preferences\"\n")
	prompt.WriteString("- content: Extract specific insights, code examples, preferences, lessons learned\n")
	prompt.WriteString("- confidence: 0.0 to 1.0 (how confident you are this is relevant domain knowledge)\n")
	prompt.WriteString("- should_write: true only if this contains meaningful domain knowledge worth preserving\n\n")
	prompt.WriteString("**Conversation Content:**\n\n")
	prompt.WriteString(content)
	prompt.WriteString("\n\n")
	prompt.WriteString("**Response:** Provide either NO_TOPICS_FOUND or the JSON array of topics.\n")

	return prompt.String()
}

// parseTopicExtractionResponse parses the LLM response and extracts topics
func (c *OpenRouterClient) parseTopicExtractionResponse(response string) ([]TopicExtraction, error) {
	// Check if no topics were found
	if strings.Contains(strings.ToUpper(response), "NO_TOPICS_FOUND") {
		c.logger.InfoWithComponent("OpenRouterClient", "No relevant domain knowledge found in content")
		return []TopicExtraction{}, nil
	}

	// Try to extract JSON from the response
	// The response might contain markdown code blocks or plain JSON
	jsonStr := c.extractJSONFromResponse(response)
	if jsonStr == "" {
		c.logger.WarnWithComponent("OpenRouterClient", "No JSON found in topic extraction response",
			"response", response)
		return []TopicExtraction{}, nil
	}

	// Parse JSON array
	var rawTopics []struct {
		TopicName   string  `json:"topic_name"`
		Content     string  `json:"content"`
		Confidence  float64 `json:"confidence"`
		ShouldWrite bool    `json:"should_write"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &rawTopics); err != nil {
		c.logger.ErrorWithComponent("OpenRouterClient", "Failed to parse topic JSON",
			"error", err,
			"json", jsonStr)
		return nil, fmt.Errorf("failed to parse topic JSON: %w", err)
	}

	// Convert to TopicExtraction structs
	topics := make([]TopicExtraction, 0, len(rawTopics))
	for _, raw := range rawTopics {
		topics = append(topics, TopicExtraction{
			TopicName:   raw.TopicName,
			Content:     raw.Content,
			Confidence:  raw.Confidence,
			ShouldWrite: raw.ShouldWrite,
		})
	}

	return topics, nil
}

// extractJSONFromResponse extracts JSON content from a response that might contain markdown
func (c *OpenRouterClient) extractJSONFromResponse(response string) string {
	// Try to find JSON in markdown code blocks
	jsonBlockStart := strings.Index(response, "```json")
	if jsonBlockStart != -1 {
		jsonBlockStart += 7 // Move past ```json
		jsonBlockEnd := strings.Index(response[jsonBlockStart:], "```")
		if jsonBlockEnd != -1 {
			return strings.TrimSpace(response[jsonBlockStart : jsonBlockStart+jsonBlockEnd])
		}
	}

	// Try to find JSON in plain code blocks
	codeBlockStart := strings.Index(response, "```")
	if codeBlockStart != -1 {
		codeBlockStart += 3 // Move past ```
		codeBlockEnd := strings.Index(response[codeBlockStart:], "```")
		if codeBlockEnd != -1 {
			content := strings.TrimSpace(response[codeBlockStart : codeBlockStart+codeBlockEnd])
			// Check if it looks like JSON
			if strings.HasPrefix(content, "[") || strings.HasPrefix(content, "{") {
				return content
			}
		}
	}

	// Try to find raw JSON array
	arrayStart := strings.Index(response, "[")
	if arrayStart != -1 {
		// Find the matching closing bracket
		bracketCount := 0
		for i := arrayStart; i < len(response); i++ {
			if response[i] == '[' {
				bracketCount++
			} else if response[i] == ']' {
				bracketCount--
				if bracketCount == 0 {
					return strings.TrimSpace(response[arrayStart : i+1])
				}
			}
		}
	}

	return ""
}

// TopicExtraction represents extracted topic information
// This struct is defined here to match the memory package's interface
type TopicExtraction struct {
	TopicName   string
	Content     string
	Confidence  float64
	ShouldWrite bool
}

