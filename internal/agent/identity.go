package agent

import (
	"fmt"
	"path/filepath"

	"simple-telegram-chatbot/pkg/utils"
)

// IdentityFiles holds the content of all identity files
type IdentityFiles struct {
	Identity    string
	Personality string
	Soul        string
	User        string
	Tools       string // Tool usage guidelines
}

// IdentityContext provides identity content for LLM context
type IdentityContext struct {
	Identity    string
	Personality string
	Soul        string
	User        string
	Tools       string // Tool usage guidelines
}

// Agent manages identity files and provides context to the LLM
type Agent struct {
	identityDir string
	files       *IdentityFiles
	logger      *utils.Logger
}

// NewAgent creates a new Agent instance
func NewAgent(identityDir string, logger *utils.Logger) *Agent {
	return &Agent{
		identityDir: identityDir,
		logger:      logger,
	}
}

// LoadIdentityFiles reads all four identity files from the agent directory
func (a *Agent) LoadIdentityFiles() (*IdentityFiles, error) {
	a.logger.InfoWithComponent("Agent", "Loading identity files", "directory", a.identityDir)

	files := &IdentityFiles{}

	// Read IDENTITY.md
	identityPath := filepath.Join(a.identityDir, "IDENTITY.md")
	identity, err := utils.ReadFile(identityPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read IDENTITY.md: %w", err)
	}
	files.Identity = identity

	// Read PERSONALITY.md
	personalityPath := filepath.Join(a.identityDir, "PERSONALITY.md")
	personality, err := utils.ReadFile(personalityPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PERSONALITY.md: %w", err)
	}
	files.Personality = personality

	// Read SOUL.md
	soulPath := filepath.Join(a.identityDir, "SOUL.md")
	soul, err := utils.ReadFile(soulPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SOUL.md: %w", err)
	}
	files.Soul = soul

	// Read USER.md
	userPath := filepath.Join(a.identityDir, "USER.md")
	user, err := utils.ReadFile(userPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read USER.md: %w", err)
	}
	files.User = user

	// Read TOOLS.md
	toolsPath := filepath.Join(a.identityDir, "TOOLS.md")
	tools, err := utils.ReadFile(toolsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read TOOLS.md: %w", err)
	}
	files.Tools = tools

	a.files = files
	a.logger.InfoWithComponent("Agent", "Successfully loaded all identity files")

	return files, nil
}

// ValidateIdentityFiles checks that all required identity files exist
func (a *Agent) ValidateIdentityFiles() error {
	a.logger.InfoWithComponent("Agent", "Validating identity files", "directory", a.identityDir)

	requiredFiles := []string{
		"IDENTITY.md",
		"PERSONALITY.md",
		"SOUL.md",
		"USER.md",
		"TOOLS.md",
	}

	missingFiles := []string{}

	for _, filename := range requiredFiles {
		filePath := filepath.Join(a.identityDir, filename)
		if !utils.FileExists(filePath) {
			missingFiles = append(missingFiles, filename)
		}
	}

	if len(missingFiles) > 0 {
		return fmt.Errorf("missing required identity files: %v", missingFiles)
	}

	a.logger.InfoWithComponent("Agent", "All identity files validated successfully")
	return nil
}

// GetIdentityContext returns the loaded identity content for LLM context
func (a *Agent) GetIdentityContext() IdentityContext {
	if a.files == nil {
		a.logger.WarnWithComponent("Agent", "GetIdentityContext called before LoadIdentityFiles")
		return IdentityContext{}
	}

	return IdentityContext{
		Identity:    a.files.Identity,
		Personality: a.files.Personality,
		Soul:        a.files.Soul,
		User:        a.files.User,
		Tools:       a.files.Tools,
	}
}
