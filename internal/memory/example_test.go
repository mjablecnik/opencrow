package memory_test

import (
	"fmt"
	"os"
	"path/filepath"
	"simple-telegram-chatbot/internal/memory"
	"time"
)

// ExampleSessionManager demonstrates basic usage of the SessionManager
func ExampleSessionManager() {
	// Create a temporary directory for this example
	tmpDir := "/tmp/opencrow-example"
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// Create a new session manager
	sm := memory.NewSessionManager(tmpDir)

	// Log a conversation
	sm.AppendToSessionLog("User", "Hello, how are you?")
	sm.AppendToSessionLog("Assistant", "I'm doing well, thank you! How can I help you today?")
	sm.AppendToSessionLog("User", "Can you help me with Go programming?")
	sm.AppendToSessionLog("Assistant", "Of course! I'd be happy to help with Go programming.")

	// Check current session info
	fmt.Printf("Current session number: %d\n", sm.GetCurrentSessionNumber())
	fmt.Printf("Current date: %s\n", sm.GetCurrentDate())

	// Read the log file to show the format
	currentDate := time.Now().Format("2006-01-02")
	logPath := filepath.Join(tmpDir, "chat", currentDate, "session-001.log")
	content, _ := os.ReadFile(logPath)
	fmt.Println("\nSession log content:")
	fmt.Println(string(content))

	// Simulate a session reset
	sm.IncrementSession()
	fmt.Printf("\nAfter session reset, new session number: %d\n", sm.GetCurrentSessionNumber())

	// Log to the new session
	sm.AppendToSessionLog("User", "Starting a new session")
	sm.AppendToSessionLog("Assistant", "Welcome back! This is session 2.")

	// Output format will vary based on timestamp, so we don't include it in the example output
}
