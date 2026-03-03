package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"simple-telegram-chatbot/pkg/utils"
)

func TestNewMemoryManager(t *testing.T) {
	tmpDir := t.TempDir()
	logger := utils.NewLogger("info")

	sessionManager := NewSessionManager(tmpDir)
	summaryManager := NewSummaryManager(tmpDir, sessionManager, 50000, nil, nil)
	topicManager := NewTopicManager(tmpDir, 100*1024)
	notesManager := NewNotesManager(tmpDir, true, 30, 7, 7)

	mm := NewMemoryManager(tmpDir, sessionManager, summaryManager, topicManager, notesManager, logger)

	if mm == nil {
		t.Fatal("NewMemoryManager returned nil")
	}

	if mm.memoryBasePath != tmpDir {
		t.Errorf("Expected memoryBasePath %s, got %s", tmpDir, mm.memoryBasePath)
	}

	if mm.sessionManager != sessionManager {
		t.Error("SessionManager not set correctly")
	}

	if mm.summaryManager != summaryManager {
		t.Error("SummaryManager not set correctly")
	}

	if mm.topicManager != topicManager {
		t.Error("TopicManager not set correctly")
	}

	if mm.notesManager != notesManager {
		t.Error("NotesManager not set correctly")
	}
}

func TestInitializeDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	logger := utils.NewLogger("info")

	// Create a temporary agent directory for MEMORY.md
	agentDir := filepath.Join(filepath.Dir(tmpDir), "agent")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatalf("Failed to create agent directory: %v", err)
	}
	defer os.RemoveAll(agentDir)

	sessionManager := NewSessionManager(tmpDir)
	summaryManager := NewSummaryManager(tmpDir, sessionManager, 50000, nil, nil)
	topicManager := NewTopicManager(tmpDir, 100*1024)
	notesManager := NewNotesManager(tmpDir, true, 30, 7, 7)

	mm := NewMemoryManager(tmpDir, sessionManager, summaryManager, topicManager, notesManager, logger)

	// Initialize directories
	if err := mm.InitializeDirectories(); err != nil {
		t.Fatalf("InitializeDirectories failed: %v", err)
	}

	// Verify main memory directory exists
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Memory directory was not created")
	}

	// Verify chat directory exists
	chatPath := filepath.Join(tmpDir, "chat")
	if _, err := os.Stat(chatPath); os.IsNotExist(err) {
		t.Error("Chat directory was not created")
	}

	// Verify topics directory exists
	topicsPath := filepath.Join(tmpDir, "topics")
	if _, err := os.Stat(topicsPath); os.IsNotExist(err) {
		t.Error("Topics directory was not created")
	}

	// Verify notes directory exists
	notesPath := filepath.Join(tmpDir, "notes")
	if _, err := os.Stat(notesPath); os.IsNotExist(err) {
		t.Error("Notes directory was not created")
	}

	// Verify notes subdirectories exist
	notesSubdirs := []string{"tasks", "ideas", "reflections", "scratchpad"}
	for _, subdir := range notesSubdirs {
		subdirPath := filepath.Join(notesPath, subdir)
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			t.Errorf("Notes subdirectory %s was not created", subdir)
		}
	}

	// Verify MEMORY.md was created
	memoryMdPath := filepath.Join(agentDir, "MEMORY.md")
	if _, err := os.Stat(memoryMdPath); os.IsNotExist(err) {
		t.Error("MEMORY.md was not created")
	}

	// Verify MEMORY.md content
	memoryContent, err := os.ReadFile(memoryMdPath)
	if err != nil {
		t.Fatalf("Failed to read MEMORY.md: %v", err)
	}

	memoryStr := string(memoryContent)
	if !strings.Contains(memoryStr, "# Memory Index") {
		t.Error("MEMORY.md does not contain expected header")
	}
	if !strings.Contains(memoryStr, "## Current Context") {
		t.Error("MEMORY.md does not contain Current Context section")
	}
	if !strings.Contains(memoryStr, "## Chat History Structure") {
		t.Error("MEMORY.md does not contain Chat History Structure section")
	}
	if !strings.Contains(memoryStr, "## Topics Knowledge Base") {
		t.Error("MEMORY.md does not contain Topics Knowledge Base section")
	}

	// Verify notes/index.md was created
	notesIndexPath := filepath.Join(notesPath, "index.md")
	if _, err := os.Stat(notesIndexPath); os.IsNotExist(err) {
		t.Error("notes/index.md was not created")
	}

	// Verify notes/index.md content
	notesContent, err := os.ReadFile(notesIndexPath)
	if err != nil {
		t.Fatalf("Failed to read notes/index.md: %v", err)
	}

	notesStr := string(notesContent)
	if !strings.Contains(notesStr, "# Agent Notes Index") {
		t.Error("notes/index.md does not contain expected header")
	}
	if !strings.Contains(notesStr, "## Tasks (0)") {
		t.Error("notes/index.md does not contain Tasks section")
	}
	if !strings.Contains(notesStr, "## Ideas (0)") {
		t.Error("notes/index.md does not contain Ideas section")
	}
	if !strings.Contains(notesStr, "## Reflections (0)") {
		t.Error("notes/index.md does not contain Reflections section")
	}
	if !strings.Contains(notesStr, "## Scratchpad (0)") {
		t.Error("notes/index.md does not contain Scratchpad section")
	}
}

func TestInitializeDirectories_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	logger := utils.NewLogger("info")

	// Create a temporary agent directory for MEMORY.md
	agentDir := filepath.Join(filepath.Dir(tmpDir), "agent")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatalf("Failed to create agent directory: %v", err)
	}
	defer os.RemoveAll(agentDir)

	sessionManager := NewSessionManager(tmpDir)
	summaryManager := NewSummaryManager(tmpDir, sessionManager, 50000, nil, nil)
	topicManager := NewTopicManager(tmpDir, 100*1024)
	notesManager := NewNotesManager(tmpDir, true, 30, 7, 7)

	mm := NewMemoryManager(tmpDir, sessionManager, summaryManager, topicManager, notesManager, logger)

	// Initialize directories first time
	if err := mm.InitializeDirectories(); err != nil {
		t.Fatalf("First InitializeDirectories failed: %v", err)
	}

	// Modify MEMORY.md to verify it's not overwritten
	memoryMdPath := filepath.Join(agentDir, "MEMORY.md")
	customContent := "# Custom Memory Content\n\nThis should not be overwritten."
	if err := os.WriteFile(memoryMdPath, []byte(customContent), 0644); err != nil {
		t.Fatalf("Failed to write custom MEMORY.md: %v", err)
	}

	// Initialize directories second time
	if err := mm.InitializeDirectories(); err != nil {
		t.Fatalf("Second InitializeDirectories failed: %v", err)
	}

	// Verify MEMORY.md was not overwritten
	memoryContent, err := os.ReadFile(memoryMdPath)
	if err != nil {
		t.Fatalf("Failed to read MEMORY.md: %v", err)
	}

	if string(memoryContent) != customContent {
		t.Error("MEMORY.md was overwritten on second initialization")
	}

	// Verify all directories still exist
	chatPath := filepath.Join(tmpDir, "chat")
	if _, err := os.Stat(chatPath); os.IsNotExist(err) {
		t.Error("Chat directory does not exist after second initialization")
	}

	topicsPath := filepath.Join(tmpDir, "topics")
	if _, err := os.Stat(topicsPath); os.IsNotExist(err) {
		t.Error("Topics directory does not exist after second initialization")
	}

	notesPath := filepath.Join(tmpDir, "notes")
	if _, err := os.Stat(notesPath); os.IsNotExist(err) {
		t.Error("Notes directory does not exist after second initialization")
	}
}

func TestInitializeDirectories_ExistingStructure(t *testing.T) {
	tmpDir := t.TempDir()
	logger := utils.NewLogger("info")

	// Create a temporary agent directory for MEMORY.md
	agentDir := filepath.Join(filepath.Dir(tmpDir), "agent")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatalf("Failed to create agent directory: %v", err)
	}
	defer os.RemoveAll(agentDir)

	// Pre-create some directories
	chatPath := filepath.Join(tmpDir, "chat")
	if err := os.MkdirAll(chatPath, 0755); err != nil {
		t.Fatalf("Failed to pre-create chat directory: %v", err)
	}

	notesTasksPath := filepath.Join(tmpDir, "notes", "tasks")
	if err := os.MkdirAll(notesTasksPath, 0755); err != nil {
		t.Fatalf("Failed to pre-create notes/tasks directory: %v", err)
	}

	sessionManager := NewSessionManager(tmpDir)
	summaryManager := NewSummaryManager(tmpDir, sessionManager, 50000, nil, nil)
	topicManager := NewTopicManager(tmpDir, 100*1024)
	notesManager := NewNotesManager(tmpDir, true, 30, 7, 7)

	mm := NewMemoryManager(tmpDir, sessionManager, summaryManager, topicManager, notesManager, logger)

	// Initialize directories with existing structure
	if err := mm.InitializeDirectories(); err != nil {
		t.Fatalf("InitializeDirectories failed with existing structure: %v", err)
	}

	// Verify all directories exist (including pre-existing ones)
	if _, err := os.Stat(chatPath); os.IsNotExist(err) {
		t.Error("Pre-existing chat directory was removed")
	}

	topicsPath := filepath.Join(tmpDir, "topics")
	if _, err := os.Stat(topicsPath); os.IsNotExist(err) {
		t.Error("Topics directory was not created")
	}

	notesPath := filepath.Join(tmpDir, "notes")
	if _, err := os.Stat(notesPath); os.IsNotExist(err) {
		t.Error("Notes directory was not created")
	}

	// Verify all notes subdirectories exist
	notesSubdirs := []string{"tasks", "ideas", "reflections", "scratchpad"}
	for _, subdir := range notesSubdirs {
		subdirPath := filepath.Join(notesPath, subdir)
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			t.Errorf("Notes subdirectory %s was not created", subdir)
		}
	}
}
