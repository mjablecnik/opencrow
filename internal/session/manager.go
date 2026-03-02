package session

import (
	"fmt"
	"sync"
	"time"
)

// Message represents a single message in a conversation
type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Session represents a conversation session with a user
type Session struct {
	ChatID    int64
	StartTime time.Time
	Messages  []Message
	mu        sync.RWMutex
}

// SessionManager manages in-memory conversation sessions
type SessionManager struct {
	sessions map[int64]*Session
	mu       sync.RWMutex
}

// NewSessionManager creates a new SessionManager instance
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[int64]*Session),
	}
}

// CreateSession creates a new in-memory session for a user
func (sm *SessionManager) CreateSession(chatID int64) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sessions[chatID]; exists {
		return fmt.Errorf("session already exists for chatID %d", chatID)
	}

	sm.sessions[chatID] = &Session{
		ChatID:    chatID,
		StartTime: time.Now(),
		Messages:  make([]Message, 0),
	}

	return nil
}

// GetSession retrieves a session for a given chatID
func (sm *SessionManager) GetSession(chatID int64) (*Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[chatID]
	if !exists {
		return nil, fmt.Errorf("session not found for chatID %d", chatID)
	}

	return session, nil
}

// AppendMessage adds a message to the in-memory session
func (sm *SessionManager) AppendMessage(chatID int64, role string, content string) error {
	sm.mu.RLock()
	session, exists := sm.sessions[chatID]
	sm.mu.RUnlock()

	if !exists {
		// Auto-create session if it doesn't exist
		if err := sm.CreateSession(chatID); err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
		sm.mu.RLock()
		session = sm.sessions[chatID]
		sm.mu.RUnlock()
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	message := Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}

	session.Messages = append(session.Messages, message)
	return nil
}

// GetHistory retrieves the conversation history for a given chatID
func (sm *SessionManager) GetHistory(chatID int64) ([]Message, error) {
	sm.mu.RLock()
	session, exists := sm.sessions[chatID]
	sm.mu.RUnlock()

	if !exists {
		return []Message{}, nil // Return empty history for non-existent sessions
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	// Return a copy of the messages to prevent external modification
	messagesCopy := make([]Message, len(session.Messages))
	copy(messagesCopy, session.Messages)

	return messagesCopy, nil
}

// ClearSession removes a session from memory
func (sm *SessionManager) ClearSession(chatID int64) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sessions[chatID]; !exists {
		return fmt.Errorf("session not found for chatID %d", chatID)
	}

	delete(sm.sessions, chatID)
	return nil
}

// FormatTimestamp formats a timestamp in the required format [YYYY-MM-DD HH:MM:SS]
func FormatTimestamp(t time.Time) string {
	return fmt.Sprintf("[%s]", t.Format("2006-01-02 15:04:05"))
}

// FormattedMessage returns a message with formatted timestamp
func (m *Message) FormattedMessage() string {
	return fmt.Sprintf("%s %s: %s", FormatTimestamp(m.Timestamp), m.Role, m.Content)
}
