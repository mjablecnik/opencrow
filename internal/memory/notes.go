package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"simple-telegram-chatbot/pkg/utils"
)

// Note represents an agent note with metadata
type Note struct {
	Path         string    // Full file path
	Category     string    // Category: tasks, ideas, reflections, scratchpad
	Name         string    // Note name (without .md extension)
	Created      time.Time // Creation timestamp
	LastModified time.Time // Last modification timestamp
	Status       string    // Status: in_progress, completed, archived
	AutoDelete   bool      // Whether note should be auto-deleted
	Content      string    // Note content (without frontmatter)
}

// NotesManager handles agent notes management with automatic cleanup
type NotesManager struct {
	mu                          sync.RWMutex
	memoryBasePath              string
	notesCleanupEnabled         bool
	notesMaxAgeDays             int
	notesCompletedRetentionDays int
	notesScratchpadMaxAgeDays   int
	logger                      *utils.Logger
}

// NewNotesManager creates a new notes manager
func NewNotesManager(memoryBasePath string, cleanupEnabled bool, maxAgeDays, completedRetentionDays, scratchpadMaxAgeDays int) *NotesManager {
	return &NotesManager{
		memoryBasePath:              memoryBasePath,
		notesCleanupEnabled:         cleanupEnabled,
		notesMaxAgeDays:             maxAgeDays,
		notesCompletedRetentionDays: completedRetentionDays,
		notesScratchpadMaxAgeDays:   scratchpadMaxAgeDays,
		logger:                      utils.NewLogger("info"),
	}
}

// CreateNote creates a new note with frontmatter metadata
func (nm *NotesManager) CreateNote(category, name, content string, autoDelete bool) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Validate category
	validCategories := map[string]bool{
		"tasks":       true,
		"ideas":       true,
		"reflections": true,
		"scratchpad":  true,
	}
	if !validCategories[category] {
		return fmt.Errorf("invalid category: %s (must be tasks, ideas, reflections, or scratchpad)", category)
	}

	// Ensure notes directory structure exists
	categoryPath := filepath.Join(nm.memoryBasePath, "notes", category)
	if err := os.MkdirAll(categoryPath, 0755); err != nil {
		return fmt.Errorf("failed to create category directory: %w", err)
	}

	// Create note file path
	notePath := filepath.Join(categoryPath, name+".md")

	// Check if note already exists
	if _, err := os.Stat(notePath); err == nil {
		return fmt.Errorf("note already exists: %s", notePath)
	}

	// Create note with frontmatter
	now := time.Now()
	frontmatter := fmt.Sprintf(`---
created: %s
last_modified: %s
status: in_progress
auto_delete: %t
---

`, now.Format(time.RFC3339), now.Format(time.RFC3339), autoDelete)

	fullContent := frontmatter + content

	// Write note to file
	if err := os.WriteFile(notePath, []byte(fullContent), 0644); err != nil {
		return fmt.Errorf("failed to write note file: %w", err)
	}

	// Update notes index
	if err := nm.updateNotesIndexInternal(); err != nil {
		nm.logger.WarnWithComponent("NotesManager", "Failed to update notes index after creating note",
			"note_name", name,
			"category", category,
			"error", err.Error(),
		)
	}

	nm.logger.InfoWithComponent("NotesManager", "Note created successfully",
		"note_name", name,
		"category", category,
		"file_path", notePath,
		"auto_delete", autoDelete,
	)

	return nil
}

// ReadNote reads a note by its path or identifier (category/name)
func (nm *NotesManager) ReadNote(identifier string) (*Note, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	// Determine if identifier is a full path or category/name format
	var notePath string
	if strings.HasPrefix(identifier, nm.memoryBasePath) {
		notePath = identifier
	} else {
		// Assume format is category/name or just name
		parts := strings.Split(identifier, "/")
		if len(parts) == 2 {
			// category/name format
			notePath = filepath.Join(nm.memoryBasePath, "notes", parts[0], parts[1]+".md")
		} else {
			// Search for note in all categories
			categories := []string{"tasks", "ideas", "reflections", "scratchpad"}
			found := false
			for _, cat := range categories {
				testPath := filepath.Join(nm.memoryBasePath, "notes", cat, identifier+".md")
				if _, err := os.Stat(testPath); err == nil {
					notePath = testPath
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("note not found: %s", identifier)
			}
		}
	}

	// Read note file
	fileContent, err := os.ReadFile(notePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read note file: %w", err)
	}

	// Parse note
	note, err := nm.parseNoteFile(notePath, string(fileContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse note: %w", err)
	}

	return note, nil
}

// UpdateNote updates an existing note's content, status, or auto_delete flag
func (nm *NotesManager) UpdateNote(identifier, content, status string, autoDelete bool) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Read existing note
	nm.mu.Unlock()
	existingNote, err := nm.ReadNote(identifier)
	nm.mu.Lock()
	if err != nil {
		return fmt.Errorf("failed to read existing note: %w", err)
	}

	// Validate status if provided
	if status != "" {
		validStatuses := map[string]bool{
			"in_progress": true,
			"completed":   true,
			"archived":    true,
		}
		if !validStatuses[status] {
			return fmt.Errorf("invalid status: %s (must be in_progress, completed, or archived)", status)
		}
	} else {
		status = existingNote.Status
	}

	// Update metadata
	now := time.Now()
	frontmatter := fmt.Sprintf(`---
created: %s
last_modified: %s
status: %s
auto_delete: %t
---

`, existingNote.Created.Format(time.RFC3339), now.Format(time.RFC3339), status, autoDelete)

	fullContent := frontmatter + content

	// Write updated note
	if err := os.WriteFile(existingNote.Path, []byte(fullContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated note: %w", err)
	}

	nm.logger.InfoWithComponent("NotesManager", "Note updated successfully",
		"note_name", existingNote.Name,
		"category", existingNote.Category,
		"status", status,
		"auto_delete", autoDelete,
	)

	return nil
}

// DeleteNote deletes a note and updates the index
func (nm *NotesManager) DeleteNote(identifier string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Read note to get full path
	nm.mu.Unlock()
	note, err := nm.ReadNote(identifier)
	nm.mu.Lock()
	if err != nil {
		return fmt.Errorf("failed to read note for deletion: %w", err)
	}

	// Delete note file
	if err := os.Remove(note.Path); err != nil {
		return fmt.Errorf("failed to delete note file: %w", err)
	}

	// Update notes index
	if err := nm.updateNotesIndexInternal(); err != nil {
		nm.logger.WarnWithComponent("NotesManager", "Failed to update notes index after deleting note",
			"note_name", note.Name,
			"category", note.Category,
			"error", err.Error(),
		)
	}

	nm.logger.InfoWithComponent("NotesManager", "Note deleted successfully",
		"note_name", note.Name,
		"category", note.Category,
		"file_path", note.Path,
	)

	return nil
}

// ListNotes lists all notes, optionally filtered by category and/or status
func (nm *NotesManager) ListNotes(category, status string) ([]Note, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	var notes []Note
	categories := []string{"tasks", "ideas", "reflections", "scratchpad"}

	// Filter categories if specified
	if category != "" {
		categories = []string{category}
	}

	// Iterate through categories
	for _, cat := range categories {
		categoryPath := filepath.Join(nm.memoryBasePath, "notes", cat)

		// Check if category directory exists
		if _, err := os.Stat(categoryPath); os.IsNotExist(err) {
			continue
		}

		// Read directory entries
		entries, err := os.ReadDir(categoryPath)
		if err != nil {
			nm.logger.WarnWithComponent("NotesManager", "Failed to read category directory",
				"category", cat,
				"error", err.Error(),
			)
			continue
		}

		// Process each note file
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			notePath := filepath.Join(categoryPath, entry.Name())
			fileContent, err := os.ReadFile(notePath)
			if err != nil {
				nm.logger.WarnWithComponent("NotesManager", "Failed to read note file",
					"file_path", notePath,
					"error", err.Error(),
				)
				continue
			}

			note, err := nm.parseNoteFile(notePath, string(fileContent))
			if err != nil {
				nm.logger.WarnWithComponent("NotesManager", "Failed to parse note file",
					"file_path", notePath,
					"error", err.Error(),
				)
				continue
			}

			// Filter by status if specified
			if status != "" && note.Status != status {
				continue
			}

			notes = append(notes, *note)
		}
	}

	return notes, nil
}

// CleanupNotes performs automatic cleanup of old notes based on retention policies
func (nm *NotesManager) CleanupNotes() error {
	if !nm.notesCleanupEnabled {
		nm.logger.InfoWithComponent("NotesManager", "Notes cleanup is disabled, skipping")
		return nil
	}

	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.logger.InfoWithComponent("NotesManager", "Starting notes cleanup")

	now := time.Now()
	deletedCount := 0
	preservedCount := 0

	// Get all notes
	nm.mu.Unlock()
	allNotes, err := nm.ListNotes("", "")
	nm.mu.Lock()
	if err != nil {
		return fmt.Errorf("failed to list notes for cleanup: %w", err)
	}

	// Check each note for cleanup eligibility
	for _, note := range allNotes {
		shouldDelete := false
		reason := ""

		// Skip if auto_delete is false
		if !note.AutoDelete {
			preservedCount++
			continue
		}

		// Check if note is referenced in MEMORY.md or topic files
		if nm.isNoteReferencedInternal(note.Name) {
			preservedCount++
			continue
		}

		// Apply retention policies based on category and status
		daysSinceModified := int(now.Sub(note.LastModified).Hours() / 24)

		switch note.Category {
		case "scratchpad":
			if daysSinceModified > nm.notesScratchpadMaxAgeDays {
				shouldDelete = true
				reason = fmt.Sprintf("scratchpad note older than %d days", nm.notesScratchpadMaxAgeDays)
			}
		default:
			if note.Status == "completed" && daysSinceModified > nm.notesCompletedRetentionDays {
				shouldDelete = true
				reason = fmt.Sprintf("completed note older than %d days", nm.notesCompletedRetentionDays)
			} else if daysSinceModified > nm.notesMaxAgeDays {
				shouldDelete = true
				reason = fmt.Sprintf("note older than %d days without modifications", nm.notesMaxAgeDays)
			}
		}

		// Delete note if eligible
		if shouldDelete {
			if err := os.Remove(note.Path); err != nil {
				nm.logger.ErrorWithDetails("NotesManager", "Failed to delete note during cleanup", err,
					"note_name", note.Name,
					"category", note.Category,
					"file_path", note.Path,
				)
				continue
			}

			nm.logger.InfoWithComponent("NotesManager", "Note deleted during cleanup",
				"note_name", note.Name,
				"category", note.Category,
				"file_path", note.Path,
				"reason", reason,
				"days_since_modified", daysSinceModified,
			)
			deletedCount++
		} else {
			preservedCount++
		}
	}

	// Update notes index if any notes were deleted
	if deletedCount > 0 {
		if err := nm.updateNotesIndexInternal(); err != nil {
			nm.logger.WarnWithComponent("NotesManager", "Failed to update notes index after cleanup",
				"error", err.Error(),
			)
		}
	}

	nm.logger.InfoWithComponent("NotesManager", "Notes cleanup completed",
		"deleted_count", deletedCount,
		"preserved_count", preservedCount,
	)

	return nil
}

// parseNoteFile parses a note file and extracts metadata and content
func (nm *NotesManager) parseNoteFile(filePath, fileContent string) (*Note, error) {
	// Extract category and name from path
	relPath, err := filepath.Rel(filepath.Join(nm.memoryBasePath, "notes"), filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %w", err)
	}

	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid note path structure: %s", filePath)
	}

	category := parts[0]
	name := strings.TrimSuffix(parts[1], ".md")

	// Parse frontmatter
	lines := strings.Split(fileContent, "\n")
	if len(lines) < 3 || lines[0] != "---" {
		return nil, fmt.Errorf("invalid note format: missing frontmatter")
	}

	// Find end of frontmatter
	frontmatterEnd := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			frontmatterEnd = i
			break
		}
	}

	if frontmatterEnd == -1 {
		return nil, fmt.Errorf("invalid note format: frontmatter not closed")
	}

	// Parse frontmatter fields
	note := &Note{
		Path:     filePath,
		Category: category,
		Name:     name,
		Status:   "in_progress",
		AutoDelete: true,
	}

	for i := 1; i < frontmatterEnd; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "created":
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				note.Created = t
			}
		case "last_modified":
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				note.LastModified = t
			}
		case "status":
			note.Status = value
		case "auto_delete":
			note.AutoDelete = value == "true"
		}
	}

	// Extract content (everything after frontmatter)
	if frontmatterEnd+1 < len(lines) {
		note.Content = strings.TrimSpace(strings.Join(lines[frontmatterEnd+1:], "\n"))
	}

	return note, nil
}

// updateNotesIndexInternal updates the notes/index.md file with current notes
// Must be called with lock held
func (nm *NotesManager) updateNotesIndexInternal() error {
	indexPath := filepath.Join(nm.memoryBasePath, "notes", "index.md")

	// Get all notes
	nm.mu.Unlock()
	allNotes, err := nm.ListNotes("", "")
	nm.mu.Lock()
	if err != nil {
		return fmt.Errorf("failed to list notes for index update: %w", err)
	}

	// Group notes by category
	notesByCategory := make(map[string][]Note)
	for _, note := range allNotes {
		notesByCategory[note.Category] = append(notesByCategory[note.Category], note)
	}

	// Build index content
	var indexContent strings.Builder
	indexContent.WriteString("# Agent Notes Index\n\n")
	indexContent.WriteString(fmt.Sprintf("**Last Updated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	indexContent.WriteString(fmt.Sprintf("**Total Notes:** %d\n\n", len(allNotes)))

	categories := []string{"tasks", "ideas", "reflections", "scratchpad"}
	for _, category := range categories {
		notes := notesByCategory[category]
		indexContent.WriteString(fmt.Sprintf("## %s (%d)\n\n", strings.Title(category), len(notes)))

		if len(notes) == 0 {
			indexContent.WriteString("*No notes in this category*\n\n")
			continue
		}

		for _, note := range notes {
			indexContent.WriteString(fmt.Sprintf("- **%s** (Status: %s, Modified: %s)\n",
				note.Name,
				note.Status,
				note.LastModified.Format("2006-01-02"),
			))
			indexContent.WriteString(fmt.Sprintf("  - Path: `%s`\n", note.Path))
			if !note.AutoDelete {
				indexContent.WriteString("  - Auto-delete: disabled\n")
			}
		}
		indexContent.WriteString("\n")
	}

	// Write index file
	if err := os.WriteFile(indexPath, []byte(indexContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to write notes index: %w", err)
	}

	return nil
}

// isNoteReferencedInternal checks if a note is referenced in MEMORY.md or topic files
// Must be called with lock held
func (nm *NotesManager) isNoteReferencedInternal(noteName string) bool {
	// Check MEMORY.md
	memoryPath := filepath.Join(nm.memoryBasePath, "..", "agent", "MEMORY.md")
	if content, err := os.ReadFile(memoryPath); err == nil {
		if strings.Contains(string(content), noteName) {
			return true
		}
	}

	// Check topic files
	topicsPath := filepath.Join(nm.memoryBasePath, "topics")
	if entries, err := os.ReadDir(topicsPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if strings.HasSuffix(entry.Name(), ".md") {
				topicPath := filepath.Join(topicsPath, entry.Name())
				if content, err := os.ReadFile(topicPath); err == nil {
					if strings.Contains(string(content), noteName) {
						return true
					}
				}
			}
		}
	}

	return false
}
