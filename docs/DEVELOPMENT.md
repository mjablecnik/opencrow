# Development Guide

## Project Structure

The project follows standard Go project layout:

- `cmd/`: Application entry points
- `internal/`: Private application code
- `pkg/`: Public library code
- `workplace/`: Runtime data directory (Docker volume)
  - `agent/`: Identity configuration files
  - `memory/`: Conversation logs, topics, and notes
  - `config/`: Cron configurations
  - `logs/`: Runtime logs
- `scripts/`: Deployment and utility scripts
- `docs/`: Documentation files

## Local Development Setup

### Prerequisites

- Go 1.22 or higher
- Telegram Bot Token (from [@BotFather](https://t.me/botfather))
- OpenRouter API Key (from [OpenRouter](https://openrouter.ai/))

### Installation

1. Navigate to the project directory:
   ```bash
   cd opencrow/
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env and add your credentials
   ```

4. Build the bot:
   ```bash
   go build -o bin/bot ./cmd/bot
   ```

5. Run the bot:
   ```bash
   # Load environment variables
   export $(cat .env | xargs)
   
   # Run the bot
   ./bin/bot
   ```

## Testing

The project uses both unit tests and property-based tests.

### Test Types

- **Unit Tests**: Test specific examples and edge cases
- **Property Tests**: Test universal properties across many inputs (using gopter)

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific package
go test ./internal/session/

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...

# Run property-based tests
go test -v ./internal/session/ -run Property
go test -v ./internal/tools/ -run Property
```

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out
```

## Adding New Features

### Adding New Tools

To add a new tool that the LLM can use:

1. Create a new file in `internal/tools/`:
   ```go
   package tools
   
   type MyTool struct {
       // dependencies
   }
   
   func (t *MyTool) Name() string {
       return "my_tool"
   }
   
   func (t *MyTool) Description() string {
       return "Description of what the tool does"
   }
   
   func (t *MyTool) Execute(params map[string]interface{}) (ToolResult, error) {
       // Implementation
       return ToolResult{
           Success: true,
           Output:  "Result",
       }, nil
   }
   ```

2. Register the tool in `cmd/bot/main.go`:
   ```go
   myTool := &tools.MyTool{}
   toolExecutor.RegisterTool("my_tool", myTool)
   ```

3. The LLM will automatically have access to the new tool

### Adding New Identity Files

To add a new identity file:

1. Create the file in `workplace/agent/`
2. Update the agent loader in `internal/agent/identity.go`
3. Update the entrypoint script to copy default files
4. Document the new file in `docs/IDENTITY_FILES.md`

### Modifying Memory System

The memory system is in `internal/memory/`:

- `manager.go`: Main memory coordinator
- `session.go`: Session logging
- `summary.go`: Summarization logic
- `topics.go`: Topic extraction and management
- `notes.go`: Notes management
- `reorganize.go`: Hierarchical reorganization
- `context.go`: Context retrieval
- `memory_index.go`: MEMORY.md generation

## Code Style

### Go Best Practices

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Use `golint` for linting
- Write clear, self-documenting code
- Add comments for exported functions and types

### Naming Conventions

- Use camelCase for variables and functions
- Use PascalCase for exported types and functions
- Use descriptive names (avoid abbreviations)
- Use English for all code elements

### Error Handling

```go
// Good: Return errors, don't panic
func doSomething() error {
    if err := validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    return nil
}

// Bad: Panic on errors
func doSomething() {
    if err := validate(); err != nil {
        panic(err)
    }
}
```

### Logging

```go
// Use structured logging
logger.Info("Processing message",
    "chatID", chatID,
    "messageID", messageID,
)

// Include context in errors
return fmt.Errorf("failed to process message %d: %w", messageID, err)
```

## Building and Deployment

### Building Docker Image

```bash
# Build the image
docker build -t opencrow:latest .

# Run the container
docker-compose up -d
```

### Building for Production

```bash
# Build with optimizations
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/bot ./cmd/bot

# Build for different platforms
GOOS=darwin GOARCH=amd64 go build -o bin/bot-darwin ./cmd/bot
GOOS=linux GOARCH=amd64 go build -o bin/bot-linux ./cmd/bot
```

## Debugging

### Enable Debug Logging

```bash
# In .env file
LOG_LEVEL=debug
```

### Using Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Run with debugger
dlv debug ./cmd/bot
```

### Inspecting Runtime State

```bash
# Access running container
docker exec -it simple-telegram-chatbot sh

# Check memory usage
docker stats simple-telegram-chatbot

# View logs
docker-compose logs -f bot
```

## Performance Optimization

### Memory Management

- Sessions are stored in-memory and grow with conversation length
- Use token-based summarization to limit context size
- Restart bot periodically to clear in-memory state

### Token Usage

- Monitor token usage in logs
- Adjust `MEMORY_TOKEN_THRESHOLD` to balance context vs. cost
- Use summaries instead of full chat logs when possible

### Database Considerations

Currently, the bot uses file system storage. For high-volume deployments, consider:

- Moving to a database for session storage
- Implementing caching for frequently accessed topics
- Using message queues for async processing

## Contributing

### Before Submitting

1. Ensure all tests pass: `go test ./...`
2. Format code: `gofmt -w .`
3. Check for common issues: `go vet ./...`
4. Update documentation if needed
5. Write clear commit messages

### Commit Message Format

```
type(scope): brief description

Detailed explanation of changes if needed.

Fixes #123
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

### Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Write/update tests
5. Update documentation
6. Submit pull request

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Telegram Bot API](https://core.telegram.org/bots/api)
- [OpenRouter API](https://openrouter.ai/docs)
- [Docker Documentation](https://docs.docker.com/)
