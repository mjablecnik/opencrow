# OpenCrow Telegram Chatbot

A Telegram bot implemented in Go that provides conversational AI capabilities through OpenRouter's LLM integration. OpenCrow features a comprehensive memory system, scheduled task management, and intelligent conversation organization.

OpenCrow is designed to be more efficient than traditional chatbot implementations, with a focus on simplicity, performance, and maintainability.

## Features

- **Telegram Integration**: Direct message support with automatic retry logic
- **OpenRouter LLM**: Conversational AI powered by configurable models
- **Memory System**: Hierarchical conversation memory with permanent log preservation
  - Daily, weekly, and quarterly organization
  - Automatic summarization at multiple levels
  - Topic-based knowledge extraction
  - Agent notes with automatic cleanup
- **Scheduling System**: Cron-based task scheduler for automated maintenance and reminders
- **Identity System**: Customizable bot behavior through identity files
- **Tool Execution**: Framework for executing tools (shell commands, memory access, cron management)
- **Docker Deployment**: Containerized deployment with volume persistence

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Telegram Bot Token (from [@BotFather](https://t.me/botfather))
- OpenRouter API Key (from [OpenRouter](https://openrouter.ai/))

### Installation

1. Clone and navigate to the project:
   ```bash
   cd opencrow/
   ```

2. Configure credentials:
   ```bash
   cp .env.example .env
   # Edit .env and add your TELEGRAM_BOT_TOKEN and OPENROUTER_API_KEY
   ```

3. Start the bot:
   ```bash
   docker-compose up -d
   ```

4. View logs:
   ```bash
   docker-compose logs -f bot
   ```

Identity files will be created automatically on first run. To customize bot behavior, edit files in `workplace/agent/` and restart the bot.

## Documentation

Comprehensive documentation is available in the `docs/` directory:

- **[Configuration Guide](docs/CONFIGURATION.md)**: Environment variables, credentials, and directory structure
- **[Memory System](docs/MEMORY_SYSTEM.md)**: Conversation logging, summarization, topics, and notes
- **[Identity Files](docs/IDENTITY_FILES.md)**: Customizing bot behavior and personality
- **[Scheduling](docs/SCHEDULING.md)**: Automated maintenance, reminders, and cron jobs
- **[Usage Guide](docs/USAGE.md)**: How to interact with the bot and use its features
- **[Troubleshooting](docs/TROUBLESHOOTING.md)**: Common issues and solutions
- **[Architecture](docs/ARCHITECTURE.md)**: System design and component overview
- **[Development](docs/DEVELOPMENT.md)**: Local development, testing, and contributing

## Key Concepts

### Memory System

OpenCrow maintains permanent conversation logs organized hierarchically:

- **Daily folders**: Created automatically for each day
- **Weekly folders**: Roll up daily folders every Monday
- **Quarterly folders**: Roll up weekly folders at quarter boundaries
- **Topics**: Domain-specific knowledge extracted automatically
- **Notes**: Agent's private working notes with automatic cleanup

All session logs are preserved permanently. Summaries are generated at multiple levels for efficient context retrieval.

### Scheduling

The bot includes a cron-based scheduler for:

- **Daily maintenance**: Runs at 4:00 AM (summarization, topic extraction, cleanup)
- **Reminders**: Recurring and one-time notifications via Telegram
- **Custom tasks**: User-defined scheduled operations

### Identity Files

Bot behavior is defined by files in `workplace/agent/`:

- **IDENTITY.md**: Static technical metadata
- **PERSONALITY.md**: Communication style and behavior
- **SOUL.md**: Core beliefs and values
- **USER.md**: User preferences and context
- **TOOLS.md**: Tool usage guidelines
- **MEMORY.md**: Auto-generated memory index

## Configuration

### Required Environment Variables

| Variable | Description |
|----------|-------------|
| `TELEGRAM_BOT_TOKEN` | Your Telegram bot token from @BotFather |
| `OPENROUTER_API_KEY` | Your OpenRouter API key |

### Optional Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MODEL_NAME` | `google/gemini-2.5-flash-lite` | LLM model to use |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `MEMORY_TOKEN_THRESHOLD` | `50000` | Token limit for emergency summarization |
| `DAILY_MAINTENANCE_TIME` | `0 4 * * *` | Cron schedule for daily maintenance |

See [Configuration Guide](docs/CONFIGURATION.md) for complete list of environment variables.

## Usage Examples

### Basic Interaction

```
User: "What's the weather like?"
Bot: [Responds with conversational AI]
```

### Creating Reminders

```
User: "Remind me every day at 8 AM to take vitamins"
Bot: "I've created a daily reminder for 8:00 AM..."
```

### Managing Memory

```
User: "Remember that I prefer Docker over Kubernetes"
Bot: "I've saved that to Programming.md: You prefer Docker over Kubernetes"
```

See [Usage Guide](docs/USAGE.md) for more examples and detailed instructions.

## Development

### Local Development

```bash
# Install dependencies
go mod download

# Build
go build -o bin/bot ./cmd/bot

# Run tests
go test ./...

# Run with environment variables
export $(cat .env | xargs)
./bin/bot
```

See [Development Guide](docs/DEVELOPMENT.md) for detailed development instructions.

### Project Structure

```
opencrow/
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── agent/             # Identity file loading
│   ├── channel/           # Telegram bot interface
│   ├── llm/               # OpenRouter client
│   ├── memory/            # Memory management system
│   ├── scheduler/         # Cron scheduling
│   ├── session/           # Session manager
│   └── tools/             # Tool executor and tools
├── pkg/                   # Public library code
├── workplace/             # Runtime data (Docker volume)
│   ├── agent/            # Identity files
│   ├── memory/           # Conversation logs and topics
│   ├── config/           # Cron configurations
│   └── logs/             # Application logs
├── docs/                  # Documentation
├── scripts/               # Deployment scripts
├── docker-compose.yml     # Docker Compose configuration
├── Dockerfile             # Docker image definition
└── README.md              # This file
```

## Troubleshooting

### Bot doesn't start

Check logs for errors:
```bash
docker-compose logs bot
```

Verify environment variables are set correctly in `.env` file.

### Bot doesn't respond

Verify bot token is correct:
```bash
curl https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getMe
```

See [Troubleshooting Guide](docs/TROUBLESHOOTING.md) for more solutions.

## Architecture

OpenCrow consists of eight primary components:

1. **Channel (Telegram)**: Message I/O and retry logic
2. **OpenRouter Client**: LLM API communication
3. **Agent**: Identity file management
4. **Memory Manager**: Conversation logging and organization
5. **Cron Scheduler**: Scheduled task execution
6. **Session Manager**: In-memory session state
7. **Tool Executor**: Tool execution framework
8. **Context Manager**: Context retrieval for LLM

See [Architecture Guide](docs/ARCHITECTURE.md) for detailed component descriptions.

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

Copyright (C) 2026 Martin Jablečník

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

## Contributing

Contributions are welcome! Please ensure:

1. All code is written in English
2. Tests pass before submitting: `go test ./...`
3. Code follows Go best practices
4. Documentation is updated for new features

See [Development Guide](docs/DEVELOPMENT.md) for detailed contribution guidelines.

## Support

For issues, questions, or contributions, please open an issue on the project repository.
