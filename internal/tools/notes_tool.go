package tools

import (
	"encoding/json"
	"fmt"
	"time"
)

// NotesManagementTool provides LLM-friendly methods for managing agent notes
type NotesManagementTool struct {
	// TODO: Add Memory Manager reference when implemented (task 16)
	// memoryManager *memory.Manager
}

// NotesToolResult represents the result of a notes tool operation with structured data
type NotesToolResult struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NoteData represents structured note information
type NoteData struct {
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	FilePath     string    `json:"file_path"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"last_modified"`
	Status       string    `json:"status"` // "in_progress", "completed", "archived"
	AutoDelete   bool      `json:"auto_delete"`
	Content      string    `json:"content"`
}

// toToolResult converts NotesToolResult to ToolResult by encoding data as JSON
func (r *NotesToolResult) toToolResult() ToolResult {
	if !r.Success {
		return ToolResult{
			Success: false,
			Error:   r.Message,
		}
	}

	// Encode the full result as JSON for the Output field
	output, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to encode result: %v", err),
		}
	}

	return ToolResult{
		Success: true,
		Output:  string(output),
	}
}

// NewNotesManagementTool creates a new NotesManagementTool instance
func NewNotesManagementTool() *NotesManagementTool {
	return &NotesManagementTool{
		// TODO: Initialize with Memory Manager reference
	}
}

// CreateNote creates a new note in the specified category
func (t *NotesManagementTool) CreateNote(category, name, content string, autoDelete bool) *NotesToolResult {
	// TODO: Integrate with Memory Manager's CreateNote method (task 16)
	
	// Validate category
	validCategories := map[string]bool{
		"tasks":       true,
		"ideas":       true,
		"reflections": true,
		"scratchpad":  true,
	}
	
	if !validCategories[category] {
		return &NotesToolResult{
			Success: false,
			Message: fmt.Sprintf("Invalid category: %s. Valid categories: tasks, ideas, reflections, scratchpad", category),
		}
	}
	
	if name == "" {
		return &NotesToolResult{
			Success: false,
			Message: "Note name cannot be empty",
		}
	}
	
	if content == "" {
		return &NotesToolResult{
			Success: false,
			Message: "Note content cannot be empty",
		}
	}
	
	filePath := fmt.Sprintf("memory/notes/%s/%s.md", category, name)
	now := time.Now()
	
	return &NotesToolResult{
		Success: true,
		Message: fmt.Sprintf("[PLACEHOLDER] Would create note '%s' in category '%s' at %s", name, category, filePath),
		Data: NoteData{
			Name:         name,
			Category:     category,
			FilePath:     filePath,
			Created:      now,
			LastModified: now,
			Status:       "in_progress",
			AutoDelete:   autoDelete,
			Content:      content,
		},
	}
}

// ReadNote retrieves a note by identifier (category/name or full path)
func (t *NotesManagementTool) ReadNote(identifier string) *NotesToolResult {
	// TODO: Integrate with Memory Manager's ReadNote method (task 16)
	
	if identifier == "" {
		return &NotesToolResult{
			Success: false,
			Message: "Note identifier cannot be empty",
		}
	}
	
	// Placeholder: Would parse identifier and read the note file
	return &NotesToolResult{
		Success: true,
		Message: fmt.Sprintf("[PLACEHOLDER] Would read note: %s", identifier),
		Data: NoteData{
			Name:         "example-note",
			Category:     "tasks",
			FilePath:     fmt.Sprintf("memory/notes/%s", identifier),
			Created:      time.Now().Add(-24 * time.Hour),
			LastModified: time.Now(),
			Status:       "in_progress",
			AutoDelete:   true,
			Content:      "Placeholder note content. This will be populated by Memory Manager.",
		},
	}
}

// UpdateNote updates an existing note's content, status, or auto_delete flag
func (t *NotesManagementTool) UpdateNote(identifier, content, status string, autoDelete bool) *NotesToolResult {
	// TODO: Integrate with Memory Manager's UpdateNote method (task 16)
	
	if identifier == "" {
		return &NotesToolResult{
			Success: false,
			Message: "Note identifier cannot be empty",
		}
	}
	
	// Validate status if provided
	if status != "" {
		validStatuses := map[string]bool{
			"in_progress": true,
			"completed":   true,
			"archived":    true,
		}
		
		if !validStatuses[status] {
			return &NotesToolResult{
				Success: false,
				Message: fmt.Sprintf("Invalid status: %s. Valid statuses: in_progress, completed, archived", status),
			}
		}
	}
	
	updates := []string{}
	if content != "" {
		updates = append(updates, "content")
	}
	if status != "" {
		updates = append(updates, fmt.Sprintf("status to '%s'", status))
	}
	updates = append(updates, fmt.Sprintf("auto_delete to %v", autoDelete))
	
	return &NotesToolResult{
		Success: true,
		Message: fmt.Sprintf("[PLACEHOLDER] Would update note '%s': %v", identifier, updates),
		Data: map[string]interface{}{
			"identifier":    identifier,
			"updated_fields": updates,
			"timestamp":     time.Now().Format(time.RFC3339),
		},
	}
}

// DeleteNote deletes a note by identifier
func (t *NotesManagementTool) DeleteNote(identifier string) *NotesToolResult {
	// TODO: Integrate with Memory Manager's DeleteNote method (task 16)
	
	if identifier == "" {
		return &NotesToolResult{
			Success: false,
			Message: "Note identifier cannot be empty",
		}
	}
	
	return &NotesToolResult{
		Success: true,
		Message: fmt.Sprintf("[PLACEHOLDER] Would delete note: %s and update notes/index.md", identifier),
		Data: map[string]interface{}{
			"identifier": identifier,
			"deleted":    true,
			"timestamp":  time.Now().Format(time.RFC3339),
		},
	}
}

// ListNotes lists notes filtered by category and/or status
func (t *NotesManagementTool) ListNotes(category, status string) *NotesToolResult {
	// TODO: Integrate with Memory Manager's ListNotes method (task 16)
	
	// Validate category if provided
	if category != "" {
		validCategories := map[string]bool{
			"tasks":       true,
			"ideas":       true,
			"reflections": true,
			"scratchpad":  true,
		}
		
		if !validCategories[category] {
			return &NotesToolResult{
				Success: false,
				Message: fmt.Sprintf("Invalid category: %s. Valid categories: tasks, ideas, reflections, scratchpad", category),
			}
		}
	}
	
	// Validate status if provided
	if status != "" {
		validStatuses := map[string]bool{
			"in_progress": true,
			"completed":   true,
			"archived":    true,
		}
		
		if !validStatuses[status] {
			return &NotesToolResult{
				Success: false,
				Message: fmt.Sprintf("Invalid status: %s. Valid statuses: in_progress, completed, archived", status),
			}
		}
	}
	
	// Build filter description
	filters := []string{}
	if category != "" {
		filters = append(filters, fmt.Sprintf("category='%s'", category))
	}
	if status != "" {
		filters = append(filters, fmt.Sprintf("status='%s'", status))
	}
	filterDesc := "all notes"
	if len(filters) > 0 {
		filterDesc = fmt.Sprintf("notes with %v", filters)
	}
	
	// Placeholder: Would scan notes directory and filter
	notes := []NoteData{
		{
			Name:         "example-task",
			Category:     "tasks",
			FilePath:     "memory/notes/tasks/example-task.md",
			Created:      time.Now().Add(-48 * time.Hour),
			LastModified: time.Now().Add(-24 * time.Hour),
			Status:       "in_progress",
			AutoDelete:   true,
			Content:      "", // Content not included in list view
		},
		{
			Name:         "example-idea",
			Category:     "ideas",
			FilePath:     "memory/notes/ideas/example-idea.md",
			Created:      time.Now().Add(-72 * time.Hour),
			LastModified: time.Now().Add(-48 * time.Hour),
			Status:       "in_progress",
			AutoDelete:   true,
			Content:      "", // Content not included in list view
		},
	}
	
	return &NotesToolResult{
		Success: true,
		Message: fmt.Sprintf("[PLACEHOLDER] Would list %s. Found %d notes.", filterDesc, len(notes)),
		Data: map[string]interface{}{
			"notes":    notes,
			"count":    len(notes),
			"category": category,
			"status":   status,
		},
	}
}

// Name returns the tool name
func (t *NotesManagementTool) Name() string {
	return "notes_management"
}

// Description returns the tool description
func (t *NotesManagementTool) Description() string {
	return "Manage agent notes for tasks, ideas, reflections, and scratchpad. Supports create, read, update, delete, and list operations."
}

// Execute implements the Tool interface
func (t *NotesManagementTool) Execute(params map[string]interface{}) (ToolResult, error) {
	// Determine which method to call based on operation parameter
	operation, ok := params["operation"].(string)
	if !ok {
		return ToolResult{
			Success: false,
			Error:   "Operation parameter required. Valid operations: create, read, update, delete, list",
		}, fmt.Errorf("operation parameter required")
	}
	
	switch operation {
	case "create":
		category, ok := params["category"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Category parameter required for create operation",
			}, fmt.Errorf("category parameter required")
		}
		
		name, ok := params["name"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Name parameter required for create operation",
			}, fmt.Errorf("name parameter required")
		}
		
		content, ok := params["content"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Content parameter required for create operation",
			}, fmt.Errorf("content parameter required")
		}
		
		// Parse auto_delete (default to true)
		autoDelete := true
		if autoDeleteBool, ok := params["auto_delete"].(bool); ok {
			autoDelete = autoDeleteBool
		}
		
		result := t.CreateNote(category, name, content, autoDelete)
		return result.toToolResult(), nil
		
	case "read":
		identifier, ok := params["identifier"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Identifier parameter required for read operation",
			}, fmt.Errorf("identifier parameter required")
		}
		
		result := t.ReadNote(identifier)
		return result.toToolResult(), nil
		
	case "update":
		identifier, ok := params["identifier"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Identifier parameter required for update operation",
			}, fmt.Errorf("identifier parameter required")
		}
		
		// All update fields are optional
		content, _ := params["content"].(string)
		status, _ := params["status"].(string)
		
		// Parse auto_delete (default to current value, but we need to pass something)
		autoDelete := true
		if autoDeleteBool, ok := params["auto_delete"].(bool); ok {
			autoDelete = autoDeleteBool
		}
		
		result := t.UpdateNote(identifier, content, status, autoDelete)
		return result.toToolResult(), nil
		
	case "delete":
		identifier, ok := params["identifier"].(string)
		if !ok {
			return ToolResult{
				Success: false,
				Error:   "Identifier parameter required for delete operation",
			}, fmt.Errorf("identifier parameter required")
		}
		
		result := t.DeleteNote(identifier)
		return result.toToolResult(), nil
		
	case "list":
		// Both parameters are optional for list
		category, _ := params["category"].(string)
		status, _ := params["status"].(string)
		
		result := t.ListNotes(category, status)
		return result.toToolResult(), nil
		
	default:
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Invalid operation: %s. Valid operations: create, read, update, delete, list", operation),
		}, fmt.Errorf("invalid operation: %s", operation)
	}
}
