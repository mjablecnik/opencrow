package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetRecentHistory(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create session manager
	sm := NewSessionManager(tempDir)

	// Create context manager
	cm := NewContextManager(tempDir, sm)

	// Test 1: Empty history (no session started)
	messages, err := cm.GetRecentHistory(123456789, 10)
	if err != nil {
		t.Fatalf("GetRecentHistory failed: %v", err)
	}
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(messages))
	}

	// Test 2: Add some messages and retrieve them
	if err := sm.AppendToSessionLog("User", "Hello, how are you?"); err != nil {
		t.Fatalf("Failed to append message: %v", err)
	}
	if err := sm.AppendToSessionLog("Assistant", "I'm doing well, thank you!"); err != nil {
		t.Fatalf("Failed to append message: %v", err)
	}
	if err := sm.AppendToSessionLog("User", "Can you help me with Docker?"); err != nil {
		t.Fatalf("Failed to append message: %v", err)
	}

	messages, err = cm.GetRecentHistory(123456789, 10)
	if err != nil {
		t.Fatalf("GetRecentHistory failed: %v", err)
	}
	if len(messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(messages))
	}

	// Verify message content
	if messages[0].Role != "User" || messages[0].Content != "Hello, how are you?" {
		t.Errorf("First message incorrect: %+v", messages[0])
	}
	if messages[1].Role != "Assistant" || messages[1].Content != "I'm doing well, thank you!" {
		t.Errorf("Second message incorrect: %+v", messages[1])
	}
	if messages[2].Role != "User" || messages[2].Content != "Can you help me with Docker?" {
		t.Errorf("Third message incorrect: %+v", messages[2])
	}

	// Test 3: Limit messages
	messages, err = cm.GetRecentHistory(123456789, 2)
	if err != nil {
		t.Fatalf("GetRecentHistory failed: %v", err)
	}
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages (limited), got %d", len(messages))
	}
	// Should get the most recent 2 messages
	if messages[0].Role != "Assistant" {
		t.Errorf("Expected Assistant message first when limited, got %s", messages[0].Role)
	}
}

func TestGetRecentHistoryMultipleSessions(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create session manager
	sm := NewSessionManager(tempDir)

	// Create context manager
	cm := NewContextManager(tempDir, sm)

	// Add messages to first session
	if err := sm.AppendToSessionLog("User", "Message 1"); err != nil {
		t.Fatalf("Failed to append message: %v", err)
	}
	if err := sm.AppendToSessionLog("Assistant", "Response 1"); err != nil {
		t.Fatalf("Failed to append message: %v", err)
	}

	// Increment session (simulate session reset)
	if err := sm.IncrementSession(); err != nil {
		t.Fatalf("Failed to increment session: %v", err)
	}

	// Add messages to second session
	if err := sm.AppendToSessionLog("User", "Message 2"); err != nil {
		t.Fatalf("Failed to append message: %v", err)
	}
	if err := sm.AppendToSessionLog("Assistant", "Response 2"); err != nil {
		t.Fatalf("Failed to append message: %v", err)
	}

	// Get recent history (should include both sessions)
	messages, err := cm.GetRecentHistory(123456789, 10)
	if err != nil {
		t.Fatalf("GetRecentHistory failed: %v", err)
	}

	// Should have 4 messages total
	if len(messages) != 4 {
		t.Errorf("Expected 4 messages from 2 sessions, got %d", len(messages))
	}

	// Verify order (oldest to newest)
	if messages[0].Content != "Message 1" {
		t.Errorf("Expected first message from session 1, got: %s", messages[0].Content)
	}
	if messages[3].Content != "Response 2" {
		t.Errorf("Expected last message from session 2, got: %s", messages[3].Content)
	}
}

func TestGetRelevantTopics(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create session manager
	sm := NewSessionManager(tempDir)

	// Create context manager
	cm := NewContextManager(tempDir, sm)

	// Test 1: No topics directory
	topics, err := cm.GetRelevantTopics("I need help with Docker")
	if err != nil {
		t.Fatalf("GetRelevantTopics failed: %v", err)
	}
	if len(topics) != 0 {
		t.Errorf("Expected 0 topics (no directory), got %d", len(topics))
	}

	// Test 2: Create topics directory and files
	topicsDir := filepath.Join(tempDir, "topics")
	if err := os.MkdirAll(topicsDir, 0755); err != nil {
		t.Fatalf("Failed to create topics directory: %v", err)
	}

	// Create some topic files
	topicFiles := map[string]string{
		"Programming.md": "# Programming\n\nGeneral programming knowledge",
		"Docker.md":      "# Docker\n\nContainer knowledge",
		"Psychology.md":  "# Psychology\n\nUser preferences",
		"Food.md":        "# Food\n\nDietary information",
	}

	for filename, content := range topicFiles {
		path := filepath.Join(topicsDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create topic file %s: %v", filename, err)
		}
	}

	// Test 3: Find relevant topics - Docker
	topics, err = cm.GetRelevantTopics("I need help with Docker containers")
	if err != nil {
		t.Fatalf("GetRelevantTopics failed: %v", err)
	}
	if len(topics) != 1 {
		t.Errorf("Expected 1 topic (Docker), got %d", len(topics))
	}
	if len(topics) > 0 && filepath.Base(topics[0]) != "Docker.md" {
		t.Errorf("Expected Docker.md, got %s", filepath.Base(topics[0]))
	}

	// Test 4: Find relevant topics - Programming
	topics, err = cm.GetRelevantTopics("Can you help me with programming in Go?")
	if err != nil {
		t.Fatalf("GetRelevantTopics failed: %v", err)
	}
	if len(topics) != 1 {
		t.Errorf("Expected 1 topic (Programming), got %d", len(topics))
	}

	// Test 5: No matching topics
	topics, err = cm.GetRelevantTopics("What's the weather like?")
	if err != nil {
		t.Fatalf("GetRelevantTopics failed: %v", err)
	}
	if len(topics) != 0 {
		t.Errorf("Expected 0 topics (no match), got %d", len(topics))
	}

	// Test 6: Multiple matching topics
	topics, err = cm.GetRelevantTopics("I want to learn programming and psychology")
	if err != nil {
		t.Fatalf("GetRelevantTopics failed: %v", err)
	}
	if len(topics) != 2 {
		t.Errorf("Expected 2 topics (Programming and Psychology), got %d", len(topics))
	}
}

func TestGetCurrentTokenUsage(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create session manager
	sm := NewSessionManager(tempDir)

	// Create context manager
	cm := NewContextManager(tempDir, sm)

	// Test 1: No session started
	tokens, err := cm.GetCurrentTokenUsage()
	if err != nil {
		t.Fatalf("GetCurrentTokenUsage failed: %v", err)
	}
	if tokens != 0 {
		t.Errorf("Expected 0 tokens (no session), got %d", tokens)
	}

	// Test 2: Add some messages
	if err := sm.AppendToSessionLog("User", "Hello, how are you?"); err != nil {
		t.Fatalf("Failed to append message: %v", err)
	}
	if err := sm.AppendToSessionLog("Assistant", "I'm doing well, thank you! How can I help you today?"); err != nil {
		t.Fatalf("Failed to append message: %v", err)
	}

	tokens, err = cm.GetCurrentTokenUsage()
	if err != nil {
		t.Fatalf("GetCurrentTokenUsage failed: %v", err)
	}
	if tokens <= 0 {
		t.Errorf("Expected positive token count, got %d", tokens)
	}

	// Test 3: Add more messages and verify token count increases
	previousTokens := tokens
	if err := sm.AppendToSessionLog("User", "Can you help me with a very long question about Docker containers and how to deploy them in production with proper security measures?"); err != nil {
		t.Fatalf("Failed to append message: %v", err)
	}

	tokens, err = cm.GetCurrentTokenUsage()
	if err != nil {
		t.Fatalf("GetCurrentTokenUsage failed: %v", err)
	}
	if tokens <= previousTokens {
		t.Errorf("Expected token count to increase, got %d (was %d)", tokens, previousTokens)
	}
}

func TestExtractKeywords(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create session manager
	sm := NewSessionManager(tempDir)

	// Create context manager
	cm := NewContextManager(tempDir, sm)

	tests := []struct {
		name     string
		context  string
		expected []string
	}{
		{
			name:     "Common topics",
			context:  "I need help with Docker and programming",
			expected: []string{"docker", "programming"},
		},
		{
			name:     "Capitalized words",
			context:  "Can you help me with Python and JavaScript?",
			expected: []string{"python", "javascript"},
		},
		{
			name:     "Mixed case",
			context:  "I want to learn about Psychology and Food preferences",
			expected: []string{"psychology", "food"},
		},
		{
			name:     "No keywords",
			context:  "hello there",
			expected: []string{},
		},
		{
			name:     "Multiple occurrences",
			context:  "Docker is great. I love Docker. Docker containers are amazing.",
			expected: []string{"docker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := cm.extractKeywords(tt.context)

			// Check if all expected keywords are present
			for _, expected := range tt.expected {
				found := false
				for _, keyword := range keywords {
					if strings.ToLower(keyword) == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected keyword '%s' not found in %v", expected, keywords)
				}
			}
		})
	}
}

func TestReadSessionLog(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create session manager
	sm := NewSessionManager(tempDir)

	// Create context manager
	cm := NewContextManager(tempDir, sm)

	// Create a test log file
	logContent := `[2024-03-01 14:30:15] User: Hello, how are you?

[2024-03-01 14:30:22] Assistant: I'm doing well, thank you!
How can I help you today?

[2024-03-01 14:32:10] User: Can you help me with Docker?

`

	logPath := filepath.Join(tempDir, "test-session.log")
	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	// Read the log
	messages, err := cm.readSessionLog(logPath)
	if err != nil {
		t.Fatalf("readSessionLog failed: %v", err)
	}

	// Verify message count
	if len(messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(messages))
	}

	// Verify first message
	if messages[0].Role != "User" {
		t.Errorf("Expected first message role 'User', got '%s'", messages[0].Role)
	}
	if messages[0].Content != "Hello, how are you?" {
		t.Errorf("Expected first message content 'Hello, how are you?', got '%s'", messages[0].Content)
	}

	// Verify second message (with multiline content)
	if messages[1].Role != "Assistant" {
		t.Errorf("Expected second message role 'Assistant', got '%s'", messages[1].Role)
	}
	expectedContent := "I'm doing well, thank you!\nHow can I help you today?"
	if messages[1].Content != expectedContent {
		t.Errorf("Expected multiline content, got '%s'", messages[1].Content)
	}

	// Verify timestamps are parsed
	if messages[0].Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp for first message")
	}
}

func TestReadSessionLogNonExistent(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create session manager
	sm := NewSessionManager(tempDir)

	// Create context manager
	cm := NewContextManager(tempDir, sm)

	// Try to read non-existent log
	messages, err := cm.readSessionLog(filepath.Join(tempDir, "nonexistent.log"))
	if err != nil {
		t.Fatalf("readSessionLog should not fail for non-existent file: %v", err)
	}
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages for non-existent file, got %d", len(messages))
	}
}
