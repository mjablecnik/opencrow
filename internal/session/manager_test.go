package session

import (
	"testing"
	"time"
)

func TestCreateSession(t *testing.T) {
	sm := NewSessionManager()

	// Test creating a new session
	err := sm.CreateSession(12345)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Verify session exists
	session, err := sm.GetSession(12345)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if session.ChatID != 12345 {
		t.Errorf("Expected ChatID 12345, got %d", session.ChatID)
	}

	if len(session.Messages) != 0 {
		t.Errorf("Expected empty messages, got %d messages", len(session.Messages))
	}

	// Test creating duplicate session
	err = sm.CreateSession(12345)
	if err == nil {
		t.Error("Expected error when creating duplicate session, got nil")
	}
}

func TestAppendMessage(t *testing.T) {
	sm := NewSessionManager()

	// Test auto-create session on first message
	err := sm.AppendMessage(12345, "user", "Hello")
	if err != nil {
		t.Fatalf("Failed to append message: %v", err)
	}

	// Verify message was added
	history, err := sm.GetHistory(12345)
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(history))
	}

	if history[0].Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", history[0].Role)
	}

	if history[0].Content != "Hello" {
		t.Errorf("Expected content 'Hello', got '%s'", history[0].Content)
	}

	// Test appending another message
	err = sm.AppendMessage(12345, "assistant", "Hi there!")
	if err != nil {
		t.Fatalf("Failed to append second message: %v", err)
	}

	history, err = sm.GetHistory(12345)
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(history) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(history))
	}
}

func TestGetHistory(t *testing.T) {
	sm := NewSessionManager()

	// Test getting history for non-existent session
	history, err := sm.GetHistory(99999)
	if err != nil {
		t.Fatalf("Expected no error for non-existent session, got: %v", err)
	}

	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d messages", len(history))
	}

	// Test getting history with messages
	sm.AppendMessage(12345, "user", "Message 1")
	sm.AppendMessage(12345, "assistant", "Message 2")
	sm.AppendMessage(12345, "user", "Message 3")

	history, err = sm.GetHistory(12345)
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(history) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(history))
	}

	// Verify order is preserved
	if history[0].Content != "Message 1" {
		t.Errorf("Expected first message 'Message 1', got '%s'", history[0].Content)
	}
	if history[1].Content != "Message 2" {
		t.Errorf("Expected second message 'Message 2', got '%s'", history[1].Content)
	}
	if history[2].Content != "Message 3" {
		t.Errorf("Expected third message 'Message 3', got '%s'", history[2].Content)
	}
}

func TestClearSession(t *testing.T) {
	sm := NewSessionManager()

	// Create and populate a session
	sm.AppendMessage(12345, "user", "Hello")

	// Clear the session
	err := sm.ClearSession(12345)
	if err != nil {
		t.Fatalf("Failed to clear session: %v", err)
	}

	// Verify session is gone
	_, err = sm.GetSession(12345)
	if err == nil {
		t.Error("Expected error when getting cleared session, got nil")
	}

	// Test clearing non-existent session
	err = sm.ClearSession(99999)
	if err == nil {
		t.Error("Expected error when clearing non-existent session, got nil")
	}
}

func TestSessionIsolation(t *testing.T) {
	sm := NewSessionManager()

	// Create sessions for two different users
	sm.AppendMessage(111, "user", "User 1 message")
	sm.AppendMessage(222, "user", "User 2 message")

	// Verify sessions are separate
	history1, _ := sm.GetHistory(111)
	history2, _ := sm.GetHistory(222)

	if len(history1) != 1 || len(history2) != 1 {
		t.Error("Sessions should have exactly 1 message each")
	}

	if history1[0].Content == history2[0].Content {
		t.Error("Sessions should have different messages")
	}

	if history1[0].Content != "User 1 message" {
		t.Errorf("User 1 session has wrong message: %s", history1[0].Content)
	}

	if history2[0].Content != "User 2 message" {
		t.Errorf("User 2 session has wrong message: %s", history2[0].Content)
	}
}

func TestTimestampFormatting(t *testing.T) {
	sm := NewSessionManager()

	// Add a message
	sm.AppendMessage(12345, "user", "Test message")

	// Get the message
	history, _ := sm.GetHistory(12345)
	if len(history) != 1 {
		t.Fatal("Expected 1 message")
	}

	msg := history[0]

	// Verify timestamp exists and is recent
	if msg.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}

	if time.Since(msg.Timestamp) > time.Second {
		t.Error("Timestamp should be very recent")
	}

	// Test formatted timestamp
	formatted := FormatTimestamp(msg.Timestamp)
	if len(formatted) < 10 {
		t.Errorf("Formatted timestamp seems too short: %s", formatted)
	}

	// Test FormattedMessage
	formattedMsg := msg.FormattedMessage()
	if formattedMsg == "" {
		t.Error("FormattedMessage should not be empty")
	}
}
