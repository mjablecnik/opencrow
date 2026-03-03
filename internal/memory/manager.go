package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"simple-telegram-chatbot/pkg/utils"
)

// MemoryManager coordinates all memory operations including session logging,
// summarization, topic extraction, hierarchical reorganization, and notes management
type MemoryManager struct {
	memoryBasePath  string
	sessionManager  *SessionManager
	summaryManager  *SummaryManager
	topicManager    *TopicManager
	notesManager    *NotesManager
	logger          *utils.Logger
}

// NewMemoryManager creates a new memory manager coordinator
func NewMemoryManager(
	memoryBasePath string,
	sessionManager *SessionManager,
	summaryManager *SummaryManager,
	topicManager *TopicManager,
	notesManager *NotesManager,
	logger *utils.Logger,
) *MemoryManager {
	return &MemoryManager{
		memoryBasePath: memoryBasePath,
		sessionManager: sessionManager,
		summaryManager: summaryManager,
		topicManager:   topicManager,
		notesManager:   notesManager,
		logger:         logger,
	}
}

// InitializeDirectories creates the memory directory structure and initial files
// This method is called on bot startup to ensure all required directories exist
func (mm *MemoryManager) InitializeDirectories() error {
	mm.logger.InfoWithComponent("MemoryManager", "Initializing memory directory structure",
		"memory_base_path", mm.memoryBasePath,
		"timestamp", time.Now().Format("2006-01-02 15:04:05"),
	)

	// Create main memory directory
	if err := mm.ensureDirectory(mm.memoryBasePath); err != nil {
		return fmt.Errorf("failed to create memory directory: %w", err)
	}

	// Create chat directory
	chatPath := filepath.Join(mm.memoryBasePath, "chat")
	if err := mm.ensureDirectory(chatPath); err != nil {
		return fmt.Errorf("failed to create chat directory: %w", err)
	}

	// Create topics directory
	topicsPath := filepath.Join(mm.memoryBasePath, "topics")
	if err := mm.ensureDirectory(topicsPath); err != nil {
		return fmt.Errorf("failed to create topics directory: %w", err)
	}

	// Create notes directory
	notesPath := filepath.Join(mm.memoryBasePath, "notes")
	if err := mm.ensureDirectory(notesPath); err != nil {
		return fmt.Errorf("failed to create notes directory: %w", err)
	}

	// Create notes subdirectories
	notesSubdirs := []string{"tasks", "ideas", "reflections", "scratchpad"}
	for _, subdir := range notesSubdirs {
		subdirPath := filepath.Join(notesPath, subdir)
		if err := mm.ensureDirectory(subdirPath); err != nil {
			return fmt.Errorf("failed to create notes subdirectory %s: %w", subdir, err)
		}
	}

	// Create MEMORY.md if it doesn't exist
	memoryMdPath := filepath.Join(filepath.Dir(mm.memoryBasePath), "agent", "MEMORY.md")
	if err := mm.ensureMemoryMd(memoryMdPath); err != nil {
		return fmt.Errorf("failed to create MEMORY.md: %w", err)
	}

	// Create notes/index.md if it doesn't exist
	notesIndexPath := filepath.Join(notesPath, "index.md")
	if err := mm.ensureNotesIndex(notesIndexPath); err != nil {
		return fmt.Errorf("failed to create notes/index.md: %w", err)
	}

	mm.logger.InfoWithComponent("MemoryManager", "Memory directory structure initialized successfully",
		"timestamp", time.Now().Format("2006-01-02 15:04:05"),
	)

	return nil
}

// ensureDirectory creates a directory if it doesn't exist
func (mm *MemoryManager) ensureDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
		mm.logger.InfoWithComponent("MemoryManager", "Created directory",
			"path", path,
			"timestamp", time.Now().Format("2006-01-02 15:04:05"),
		)
	}
	return nil
}

// ensureMemoryMd creates MEMORY.md with initial template content if it doesn't exist
func (mm *MemoryManager) ensureMemoryMd(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		template := mm.getMemoryMdTemplate()
		if err := os.WriteFile(path, []byte(template), 0644); err != nil {
			return err
		}
		mm.logger.InfoWithComponent("MemoryManager", "Created MEMORY.md with initial template",
			"path", path,
			"timestamp", time.Now().Format("2006-01-02 15:04:05"),
		)
	}
	return nil
}

// ensureNotesIndex creates notes/index.md with initial template content if it doesn't exist
func (mm *MemoryManager) ensureNotesIndex(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		template := mm.getNotesIndexTemplate()
		if err := os.WriteFile(path, []byte(template), 0644); err != nil {
			return err
		}
		mm.logger.InfoWithComponent("MemoryManager", "Created notes/index.md with initial template",
			"path", path,
			"timestamp", time.Now().Format("2006-01-02 15:04:05"),
		)
	}
	return nil
}

// getMemoryMdTemplate returns the initial template content for MEMORY.md
func (mm *MemoryManager) getMemoryMdTemplate() string {
	now := time.Now()
	return fmt.Sprintf(`# Memory Index

**Last Updated:** %s

## Current Context

### Active Session
- Session: Not started
- Started: N/A
- Messages: 0
- Topic: N/A

### Recent Summary
Last daily summary: None
Last weekly summary: None
Last quarterly summary: None

## Chat History Structure

### Current Quarter: Q%d %d
No chat history yet.

### Previous Quarters
None

## Topics Knowledge Base

No topics created yet. Topics will be automatically extracted during conversations and summarization.

## Agent Notes

See memory/notes/index.md for agent's private working notes.

---

*This file is automatically updated by the memory system.*
`, now.Format("2006-01-02 15:04:05"), (now.Month()-1)/3+1, now.Year())
}

// getNotesIndexTemplate returns the initial template content for notes/index.md
func (mm *MemoryManager) getNotesIndexTemplate() string {
	now := time.Now()
	return fmt.Sprintf(`# Agent Notes Index

**Last Updated:** %s

**Total Notes:** 0

## Tasks (0)

*No notes in this category*

## Ideas (0)

*No notes in this category*

## Reflections (0)

*No notes in this category*

## Scratchpad (0)

*No notes in this category*

---

*This file is automatically updated by the notes management system.*
`, now.Format("2006-01-02 15:04:05"))
}
