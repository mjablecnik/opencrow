# Simple Telegram Chatbot

A minimal viable Telegram bot implemented in Go that provides conversational AI capabilities through OpenRouter's LLM integration. The bot maintains in-memory session history, supports identity and personality files to influence behavior, and includes basic tool execution capabilities.

## Features

- **Telegram Integration**: Direct message support with automatic retry logic and exponential backoff
- **OpenRouter LLM**: Conversational AI powered by configurable LLM models via OpenRouter API
- **In-Memory Sessions**: Maintains conversation history during runtime (ephemeral, resets on restart)
- **Identity System**: Customizable bot behavior through identity files (IDENTITY.md, PERSONALITY.md, SOUL.md, USER.md)
- **Tool Execution**: Framework for executing tools with shell command support
- **Docker Deployment**: Containerized deployment with volume persistence for configuration and logs
- **Graceful Shutdown**: Handles shutdown signals properly to complete in-flight messages
- **Comprehensive Logging**: Configurable log levels with structured error reporting

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Telegram Bot Token (from [@BotFather](https://t.me/botfather))
- OpenRouter API Key (from [OpenRouter](https://openrouter.ai/))

### Docker Deployment (Recommended)

1. Clone the repository and navigate to the project directory:
   ```bash
   cd opencrow/
   ```

2. Copy the environment template and configure your credentials:
   ```bash
   cp .env.example .env
   # Edit .env and add your TELEGRAM_BOT_TOKEN and OPENROUTER_API_KEY
   ```

3. Customize identity files (optional):
   ```bash
   # Edit files in agent/ directory to customize bot behavior
   nano agent/IDENTITY.md
   nano agent/PERSONALITY.md
   nano agent/SOUL.md
   nano agent/USER.md
   ```

4. Start the bot:
   ```bash
   docker-compose up -d
   ```

5. View logs:
   ```bash
   docker-compose logs -f bot
   ```

6. Stop the bot:
   ```bash
   docker-compose down
   ```

## Local Development

### Prerequisites

- Go 1.21 or higher
- Telegram Bot Token
- OpenRouter API Key

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

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run property-based tests (requires gopter)
go test -v ./internal/session/
go test -v ./internal/tools/
```

## Configuration

### Environment Variables

The bot is configured entirely through environment variables. All variables can be set in the `.env` file or passed directly to the Docker container.

#### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `TELEGRAM_BOT_TOKEN` | Your Telegram bot token from @BotFather | `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11` |
| `OPENROUTER_API_KEY` | Your OpenRouter API key | `sk-or-v1-...` |

#### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MODEL_NAME` | LLM model to use for responses | `google/gemini-2.5-flash-lite` |
| `SHELL_TIMEOUT` | Timeout for shell command execution (seconds) | `30` |
| `LOG_LEVEL` | Logging level: `debug`, `info`, `warn`, `error` | `info` |

### Getting Your Credentials

**Telegram Bot Token:**
1. Open Telegram and search for [@BotFather](https://t.me/botfather)
2. Send `/newbot` command
3. Follow the prompts to create your bot
4. Copy the bot token provided

**OpenRouter API Key:**
1. Visit [OpenRouter](https://openrouter.ai/)
2. Sign up or log in
3. Navigate to API Keys section
4. Create a new API key
5. Copy the key (starts with `sk-or-v1-`)

## Directory Structure

```
opencrow/
├── agent/                          # Identity and personality files
│   ├── IDENTITY.md                 # Bot's static technical metadata
│   ├── PERSONALITY.md              # Communication style and behavior patterns
│   ├── SOUL.md                     # Core beliefs and authentic self
│   └── USER.md                     # User preferences and context
├── logs/                           # Application logs (created at runtime)
│   └── bot.log                     # Main log file
├── cmd/
│   └── bot/
│       └── main.go                 # Application entry point
├── internal/
│   ├── agent/                      # Identity file loading
│   ├── channel/                    # Telegram bot interface
│   ├── config/                     # Configuration management
│   ├── llm/                        # OpenRouter client
│   ├── session/                    # In-memory session manager
│   └── tools/                      # Tool executor and shell tool
├── pkg/
│   └── utils/                      # Logging and file utilities
├── .env.example                    # Environment variable template
├── docker-compose.yml              # Docker Compose configuration
├── Dockerfile                      # Docker image definition
├── go.mod                          # Go module definition
└── README.md                       # This file
```

### Important Directories

- **agent/**: Contains identity files that define the bot's behavior and personality. These files are loaded on startup and included in every LLM request.
- **logs/**: Stores application logs. In Docker deployments, this directory is mounted as a volume for persistence.

**Note**: Session data is stored in-memory only and is lost when the bot restarts. Only identity files and logs are persisted.

## Identity Files

The bot's behavior is influenced by four identity files located in the `agent/` directory. These files are loaded on startup and included in the system context for every LLM request.

### IDENTITY.md

Contains static technical metadata about the bot.

**Example:**
```markdown
# Bot Identity

**Name:** SimpleTelegramBot
**Created:** 2024-01-15
**Creator:** OpenCrow Team
**Version:** 1.0.0

## Purpose

SimpleTelegramBot is a minimal viable conversational AI assistant that maintains
session history and can execute basic shell commands.

## Core Facts

- Built with Go and OpenRouter API
- Deployed as Docker container
- Maintains in-memory session history
- Supports basic tool execution
```

### PERSONALITY.md

Describes the bot's communication style and behavior patterns.

**Example:**
```markdown
# Personality

## Communication Style

- Clear and concise
- Friendly and helpful
- Professional but approachable
- Patient with questions

## Behavior Patterns

- Provide direct answers when possible
- Ask clarifying questions when needed
- Explain technical concepts simply
- Acknowledge limitations honestly
```

### SOUL.md

Contains the bot's core beliefs, convictions, and authentic self.

**Example:**
```markdown
# Soul - Core Beliefs

## Values

- Honesty and transparency
- Helpfulness and support
- Continuous learning
- Respect for user privacy

## Worldview

I exist to assist users with their tasks and questions. I aim to be helpful
while being honest about my capabilities and limitations.

## Authentic Self

I am a simple chatbot designed to provide conversational AI assistance with
basic tool execution capabilities. I maintain context within sessions and
strive to be consistently helpful.
```

### USER.md

Describes the user's preferences and context (template to be filled in).

**Example:**
```markdown
# User Profile

## Preferences

- Prefers concise technical explanations
- Comfortable with command-line interfaces
- Works primarily with Go and Docker

## Context

- Software developer with 5+ years experience
- Currently building microservices architecture
- Timezone: UTC+1

## Notes

- Prefers examples over lengthy explanations
- Interested in best practices and optimization
```

### Customizing Identity Files

1. Edit the files in the `agent/` directory to customize bot behavior
2. Restart the bot to load the new identity files
3. The bot will fail to start if any identity file is missing

## Usage

### Basic Interaction

1. Start a conversation with your bot on Telegram
2. Send any message to begin chatting
3. The bot maintains conversation history during the session
4. Sessions are ephemeral and reset when the bot restarts

### Tool Execution

The bot can execute shell commands when requested by the LLM:

**Example conversation:**
```
User: Can you list the files in the current directory?

Bot: I'll run the ls command for you.
[Executes: ls -la]

Output:
total 48
drwxr-xr-x  8 user  staff   256 Jan 15 10:30 .
drwxr-xr-x  5 user  staff   160 Jan 14 09:15 ..
-rw-r--r--  1 user  staff  1234 Jan 15 10:30 README.md
...
```

**Security Note**: Shell command execution is powerful and potentially dangerous. Only use this bot in trusted environments and be cautious about what commands you allow it to execute.

### Session Management

- **In-Memory Only**: Sessions are stored in memory and lost on restart
- **Per-User Sessions**: Each Telegram user has a separate conversation history
- **No Persistence**: Unlike file-based systems, sessions do not persist across restarts
- **Timestamps**: All messages include timestamps in format `[YYYY-MM-DD HH:MM:SS]`

## Troubleshooting

### Bot doesn't start

**Problem**: Bot exits immediately after starting

**Solutions**:
1. Check that all required environment variables are set:
   ```bash
   echo $TELEGRAM_BOT_TOKEN
   echo $OPENROUTER_API_KEY
   ```

2. Verify identity files exist:
   ```bash
   ls -la agent/
   # Should show: IDENTITY.md, PERSONALITY.md, SOUL.md, USER.md
   ```

3. Check logs for specific error messages:
   ```bash
   # Docker
   docker-compose logs bot
   
   # Local
   cat logs/bot.log
   ```

### Bot doesn't respond to messages

**Problem**: Bot is running but doesn't reply to Telegram messages

**Solutions**:
1. Verify the bot token is correct:
   - Test the token with Telegram's API: `https://api.telegram.org/bot<YOUR_TOKEN>/getMe`
   - Should return bot information if token is valid

2. Check if the bot is polling for updates:
   ```bash
   # Look for "Started polling" in logs
   docker-compose logs bot | grep -i polling
   ```

3. Ensure the bot isn't blocked by the user on Telegram

4. Check for API errors in logs:
   ```bash
   docker-compose logs bot | grep -i error
   ```

### OpenRouter API errors

**Problem**: Bot receives messages but fails to generate responses

**Solutions**:
1. Verify OpenRouter API key is valid:
   - Check your API key at [OpenRouter Dashboard](https://openrouter.ai/)
   - Ensure you have sufficient credits

2. Check the model name is correct:
   ```bash
   # In .env file
   MODEL_NAME=google/gemini-2.5-flash-lite
   ```

3. Review API error messages in logs:
   ```bash
   docker-compose logs bot | grep -i "openrouter\|api"
   ```

4. Test API connectivity:
   ```bash
   curl -X POST https://openrouter.ai/api/v1/chat/completions \
     -H "Authorization: Bearer $OPENROUTER_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{"model":"google/gemini-2.5-flash-lite","messages":[{"role":"user","content":"test"}]}'
   ```

### Message send failures

**Problem**: Bot generates responses but fails to send them to Telegram

**Solutions**:
1. Check for rate limiting:
   - Telegram has rate limits for bot messages
   - The bot automatically retries with exponential backoff (3 attempts)

2. Verify network connectivity:
   ```bash
   # Test Telegram API connectivity
   curl https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getMe
   ```

3. Check logs for retry attempts:
   ```bash
   docker-compose logs bot | grep -i "retry\|send"
   ```

### Docker volume issues

**Problem**: Identity files or logs not persisting

**Solutions**:
1. Verify volume mounts in docker-compose.yml:
   ```yaml
   volumes:
     - ./agent:/app/agent:ro
     - ./logs:/app/logs
   ```

2. Check file permissions:
   ```bash
   ls -la agent/
   ls -la logs/
   # Files should be readable by the container user (UID 1000)
   ```

3. Recreate volumes:
   ```bash
   docker-compose down -v
   docker-compose up -d
   ```

### High memory usage

**Problem**: Bot consumes excessive memory over time

**Solutions**:
1. Sessions are stored in-memory and grow with conversation length
2. Restart the bot periodically to clear sessions:
   ```bash
   docker-compose restart bot
   ```

3. Monitor memory usage:
   ```bash
   docker stats simple-telegram-chatbot
   ```

### Shell command timeouts

**Problem**: Shell commands fail with timeout errors

**Solutions**:
1. Increase the timeout value:
   ```bash
   # In .env file
   SHELL_TIMEOUT=60  # Increase to 60 seconds
   ```

2. Avoid long-running commands:
   - Shell commands should complete quickly
   - Use background processes for long operations

3. Check command execution in logs:
   ```bash
   docker-compose logs bot | grep -i "shell\|timeout"
   ```

### Debugging tips

1. **Enable debug logging**:
   ```bash
   # In .env file
   LOG_LEVEL=debug
   ```

2. **Check container status**:
   ```bash
   docker-compose ps
   docker-compose logs --tail=100 bot
   ```

3. **Inspect running container**:
   ```bash
   docker exec -it simple-telegram-chatbot sh
   ```

4. **Test components individually**:
   ```bash
   # Run specific tests
   go test -v ./internal/channel/
   go test -v ./internal/llm/
   ```

## Architecture

### Component Overview

The system consists of five primary components:

1. **Channel (Telegram)**: Handles Telegram Bot API interactions, message polling, and retry logic
2. **OpenRouter Client**: Manages LLM API communication, context assembly, and tool request handling
3. **Agent**: Loads and manages identity files that define bot behavior
4. **Session Manager**: Maintains in-memory conversation history per user
5. **Tool Executor**: Executes tools requested by the LLM (currently supports shell commands)

### Message Flow

```
User (Telegram)
    ↓
Channel (Telegram) ←→ OpenRouter Client ←→ OpenRouter API
    ↓                         ↓
Session Manager           Tool Executor
    ↓                         ↓
In-Memory Storage        Shell Tool
    ↑
Agent (Identity Files)
```

1. User sends message via Telegram
2. Channel receives and validates message
3. Session Manager retrieves conversation history
4. Agent provides identity files
5. OpenRouter Client assembles context and sends to API
6. If LLM requests tool execution, Tool Executor handles it
7. Response returns through the chain to user
8. Session Manager stores exchange in memory

### Data Persistence

- **Ephemeral**: Session data (in-memory only, lost on restart)
- **Persistent**: Identity files (agent/ directory), logs (logs/ directory)
- **Docker Volumes**: agent/ and logs/ directories are mounted as volumes

## Development

### Project Structure

The project follows standard Go project layout:

- `cmd/`: Application entry points
- `internal/`: Private application code
- `pkg/`: Public library code
- `agent/`: Identity configuration files
- `logs/`: Runtime logs

### Adding New Tools

To add a new tool:

1. Implement the `Tool` interface in `internal/tools/`:
   ```go
   type Tool interface {
       Name() string
       Description() string
       Execute(params map[string]interface{}) (ToolResult, error)
   }
   ```

2. Register the tool in `cmd/bot/main.go`:
   ```go
   toolExecutor.RegisterTool("my_tool", myTool)
   ```

3. The LLM will automatically have access to the new tool

### Testing

The project uses both unit tests and property-based tests:

- **Unit Tests**: Test specific examples and edge cases
- **Property Tests**: Test universal properties across many inputs (using gopter)

Run tests:
```bash
# All tests
go test ./...

# Specific package
go test ./internal/session/

# With coverage
go test -cover ./...

# Property tests with verbose output
go test -v ./internal/session/ -run Property
```

## License

[To be determined]

## Contributing

Contributions are welcome! Please ensure:

1. All code is written in English (variable names, comments, documentation)
2. Tests pass before submitting: `go test ./...`
3. Code follows Go best practices and conventions
4. Documentation is updated for new features

## Support

For issues, questions, or contributions, please open an issue on the project repository.
