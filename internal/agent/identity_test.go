package agent

import (
	"os"
	"path/filepath"
	"testing"

	"simple-telegram-chatbot/pkg/utils"
)

func TestValidateIdentityFiles_AllFilesExist(t *testing.T) {
	// Create temporary directory with all identity files
	tmpDir := t.TempDir()
	
	files := []string{"IDENTITY.md", "PERSONALITY.md", "SOUL.md", "USER.md"}
	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	logger := utils.NewLogger("info")
	agent := NewAgent(tmpDir, logger)

	err := agent.ValidateIdentityFiles()
	if err != nil {
		t.Errorf("ValidateIdentityFiles() failed with all files present: %v", err)
	}
}

func TestValidateIdentityFiles_MissingFiles(t *testing.T) {
	// Create temporary directory with only some identity files
	tmpDir := t.TempDir()
	
	// Only create IDENTITY.md and PERSONALITY.md, leave SOUL.md and USER.md missing
	files := []string{"IDENTITY.md", "PERSONALITY.md"}
	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	logger := utils.NewLogger("info")
	agent := NewAgent(tmpDir, logger)

	err := agent.ValidateIdentityFiles()
	if err == nil {
		t.Error("ValidateIdentityFiles() should fail with missing files")
	}
}

func TestLoadIdentityFiles_Success(t *testing.T) {
	// Create temporary directory with all identity files
	tmpDir := t.TempDir()
	
	testContent := map[string]string{
		"IDENTITY.md":    "Identity content",
		"PERSONALITY.md": "Personality content",
		"SOUL.md":        "Soul content",
		"USER.md":        "User content",
	}

	for file, content := range testContent {
		path := filepath.Join(tmpDir, file)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	logger := utils.NewLogger("info")
	agent := NewAgent(tmpDir, logger)

	files, err := agent.LoadIdentityFiles()
	if err != nil {
		t.Fatalf("LoadIdentityFiles() failed: %v", err)
	}

	if files.Identity != testContent["IDENTITY.md"] {
		t.Errorf("Identity content mismatch: got %q, want %q", files.Identity, testContent["IDENTITY.md"])
	}
	if files.Personality != testContent["PERSONALITY.md"] {
		t.Errorf("Personality content mismatch: got %q, want %q", files.Personality, testContent["PERSONALITY.md"])
	}
	if files.Soul != testContent["SOUL.md"] {
		t.Errorf("Soul content mismatch: got %q, want %q", files.Soul, testContent["SOUL.md"])
	}
	if files.User != testContent["USER.md"] {
		t.Errorf("User content mismatch: got %q, want %q", files.User, testContent["USER.md"])
	}
}

func TestLoadIdentityFiles_MissingFile(t *testing.T) {
	// Create temporary directory with only some files
	tmpDir := t.TempDir()
	
	// Only create IDENTITY.md
	path := filepath.Join(tmpDir, "IDENTITY.md")
	if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	logger := utils.NewLogger("info")
	agent := NewAgent(tmpDir, logger)

	_, err := agent.LoadIdentityFiles()
	if err == nil {
		t.Error("LoadIdentityFiles() should fail with missing files")
	}
}

func TestGetIdentityContext(t *testing.T) {
	// Create temporary directory with all identity files
	tmpDir := t.TempDir()
	
	testContent := map[string]string{
		"IDENTITY.md":    "Identity content",
		"PERSONALITY.md": "Personality content",
		"SOUL.md":        "Soul content",
		"USER.md":        "User content",
	}

	for file, content := range testContent {
		path := filepath.Join(tmpDir, file)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	logger := utils.NewLogger("info")
	agent := NewAgent(tmpDir, logger)

	// Load files first
	_, err := agent.LoadIdentityFiles()
	if err != nil {
		t.Fatalf("LoadIdentityFiles() failed: %v", err)
	}

	// Get context
	context := agent.GetIdentityContext()

	if context.Identity != testContent["IDENTITY.md"] {
		t.Errorf("Identity context mismatch: got %q, want %q", context.Identity, testContent["IDENTITY.md"])
	}
	if context.Personality != testContent["PERSONALITY.md"] {
		t.Errorf("Personality context mismatch: got %q, want %q", context.Personality, testContent["PERSONALITY.md"])
	}
	if context.Soul != testContent["SOUL.md"] {
		t.Errorf("Soul context mismatch: got %q, want %q", context.Soul, testContent["SOUL.md"])
	}
	if context.User != testContent["USER.md"] {
		t.Errorf("User context mismatch: got %q, want %q", context.User, testContent["USER.md"])
	}
}

func TestGetIdentityContext_BeforeLoad(t *testing.T) {
	tmpDir := t.TempDir()
	logger := utils.NewLogger("info")
	agent := NewAgent(tmpDir, logger)

	// Call GetIdentityContext before LoadIdentityFiles
	context := agent.GetIdentityContext()

	// Should return empty context
	if context.Identity != "" || context.Personality != "" || context.Soul != "" || context.User != "" {
		t.Error("GetIdentityContext() should return empty context when called before LoadIdentityFiles()")
	}
}
