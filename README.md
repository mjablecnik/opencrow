# OpenCrow Telegram Chatbot

A Telegram bot implemented in Go that provides conversational AI capabilities through OpenRouter's LLM integration. The bot features a comprehensive memory and scheduling system that maintains conversation history, organizes domain-specific knowledge, and executes scheduled tasks automatically.

## Features

- **Telegram Integration**: Direct message support with automatic retry logic and exponential backoff
- **OpenRouter LLM**: Conversational AI powered by configurable LLM models via OpenRouter API
- **Memory System**: Hierarchical conversation memory with permanent log preservation
  - Session logging with daily, weekly, and quarterly organization
  - Automatic summarization at multiple levels
  - Topic-based knowledge extraction and management
  - Agent notes with automatic cleanup
- **Scheduling System**: Cron-based task scheduler for automated maintenance
  - Daily maintenance cascade (summarization, topic extraction, cleanup)
  - Weekly and quarterly reorganization
  - Reminder and notification support
- **Identity System**: Customizable bot behavior through identity files (IDENTITY.md, PERSONALITY.md, SOUL.md, USER.md, MEMORY.md)
- **Tool Execution**: Framework for executing tools with shell command support
- **Docker Deployment**: Containerized deployment with volume persistence for configuration, logs, and memory
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
| `MEMORY_TOKEN_THRESHOLD` | Token limit for emergency summarization | `50000` |
| `TOPIC_SIZE_THRESHOLD` | Topic file size threshold for subdivision (bytes) | `102400` (100KB) |
| `NOTES_CLEANUP_ENABLED` | Enable automatic notes cleanup | `true` |
| `NOTES_MAX_AGE_DAYS` | Maximum age for notes without modifications | `30` |
| `NOTES_COMPLETED_RETENTION_DAYS` | Retention days for completed notes | `7` |
| `NOTES_SCRATCHPAD_MAX_AGE_DAYS` | Maximum age for scratchpad notes | `7` |
| `DAILY_MAINTENANCE_TIME` | Cron schedule for daily maintenance | `0 4 * * *` (4:00 AM) |

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
│   ├── USER.md                     # User preferences and context
│   └── MEMORY.md                   # Memory index (auto-generated)
├── memory/                         # Memory system data (created at runtime)
│   ├── chat/                       # Conversation logs
│   │   ├── 2024-01-15/            # Daily folder
│   │   │   ├── session-001.log
│   │   │   ├── session-001-summary.md
│   │   │   └── daily-summary.md
│   │   ├── week-03-2024/          # Weekly folder
│   │   │   ├── 2024-01-15/
│   │   │   └── summary.md
│   │   └── Q1-2024/               # Quarterly folder
│   │       ├── week-01-2024/
│   │       └── summary.md
│   ├── topics/                     # Domain-specific knowledge
│   │   ├── Programming.md
│   │   ├── Psychology.md
│   │   └── Food.md
│   └── notes/                      # Agent's private notes
│       ├── index.md
│       ├── tasks/
│       ├── ideas/
│       ├── reflections/
│       └── scratchpad/
├── config/                         # Configuration (created at runtime)
│   ├── cron.json                   # Cron job configurations
│   └── cron_history.json           # Execution history
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
│   ├── memory/                     # Memory management system
│   │   ├── manager.go              # Memory coordinator
│   │   ├── session.go              # Session logging
│   │   ├── summary.go              # Summarization
│   │   ├── topics.go               # Topic management
│   │   ├── notes.go                # Notes management
│   │   ├── reorganize.go           # Hierarchical reorganization
│   │   └── context.go              # Context retrieval
│   ├── scheduler/                  # Cron scheduling system
│   │   ├── cron.go                 # Scheduler implementation
│   │   └── jobs.go                 # Job definitions
│   ├── session/                    # In-memory session manager
│   └── tools/                      # Tool executor and tools
│       ├── executor.go             # Tool execution framework
│       ├── shell.go                # Shell command tool
│       ├── cron_tool.go            # Cron management tool
│       ├── memory_tool.go          # Memory summary tool
│       ├── topic_tool.go           # Topic knowledge tool
│       ├── chatlog_tool.go         # Chat log search tool
│       └── notes_tool.go           # Notes management tool
├── pkg/
│   └── utils/                      # Logging and file utilities
├── .env.example                    # Environment variable template
├── docker-compose.yml              # Docker Compose configuration
├── Dockerfile                      # Docker image definition
├── go.mod                          # Go module definition
└── README.md                       # This file
```

### Important Directories

- **agent/**: Contains identity files that define the bot's behavior and personality. These files are loaded on startup and included in every LLM request. Includes MEMORY.md which is auto-generated by the memory system.
- **memory/**: Stores all conversation history, domain knowledge, and agent notes. This directory is automatically created and managed by the memory system.
  - **chat/**: Hierarchical conversation logs organized by day, week, and quarter
  - **topics/**: Domain-specific knowledge files (Programming, Psychology, Food, etc.)
  - **notes/**: Agent's private working notes organized by category
- **config/**: Stores cron job configurations and execution history. Created automatically at runtime.
- **logs/**: Stores application logs. In Docker deployments, this directory is mounted as a volume for persistence.

## Memory System

The bot features a comprehensive memory system that permanently preserves all conversations while organizing them hierarchically for efficient access.

### Conversation Memory

**Session Logging:**
- Each conversation is logged to session files (session-001.log, session-002.log, etc.)
- Sessions are organized in daily folders (YYYY-MM-DD format)
- All session logs are preserved permanently and never deleted
- Messages include timestamps: `[YYYY-MM-DD HH:MM:SS] Role: Content`

**Hierarchical Organization:**
- **Daily folders**: Created automatically for each day
- **Weekly folders**: Created every Monday, containing the previous 7 daily folders
- **Quarterly folders**: Created on the first day of each quarter, containing all week folders from the completed quarter

**Summarization:**
- **Session summaries**: Generated when sessions are manually reset
- **Daily summaries**: Generated at 4:00 AM, summarizing all sessions from the previous day
- **Weekly summaries**: Generated on Mondays, summarizing all daily summaries from the week
- **Quarterly summaries**: Generated on the first day of each quarter, summarizing all weekly summaries

**Token-Based Summarization:**
- When conversation token usage exceeds `MEMORY_TOKEN_THRESHOLD` (default: 50,000), the system automatically:
  1. Sends notification: "Performing memory summarization due to high token usage..."
  2. Generates a summary of the current session
  3. Clears the in-memory context
  4. Inserts the summary at the beginning of the session
  5. Continues the conversation seamlessly without session reset

### Topic-Based Knowledge

The system automatically extracts domain-specific knowledge from conversations and organizes it into topic files:

**Automatic Topic Extraction:**
- Runs during all summarization operations (daily, weekly, quarterly, session, token-based)
- Uses LLM to identify relevant domain knowledge
- Only creates/updates topic files when relevant knowledge is found
- Logs "no relevant domain knowledge found" when appropriate

**Topic Organization:**
- Topics stored as markdown files: `Programming.md`, `Psychology.md`, `Food.md`, etc.
- When a topic file exceeds `TOPIC_SIZE_THRESHOLD` (default: 100KB), it's automatically subdivided into a folder structure
- Supports hierarchical organization for large topics

**Manual Topic Writing:**
- Users can explicitly request to remember information: "Remember that I hate onions"
- Agent immediately writes to the appropriate topic file
- No need to wait for scheduled summarization

**Supported Topic Domains:**
- Programming: Code examples, configurations, preferences, mistakes to avoid
- Psychology: Conversation patterns, user preferences, communication style
- Food: Dietary tracking, preferences, meal plans
- Sport-Health: Exercise routines, health information, medical notes
- Politics: User's views and preferences
- Custom topics: System dynamically creates new topics as needed

### Agent Notes

The bot maintains private working notes for complex tasks, ideas, and temporary calculations:

**Note Categories:**
- **tasks/**: Task planning and execution notes
- **ideas/**: Ideas and brainstorming
- **reflections/**: Reflections on conversations and learning
- **scratchpad/**: Temporary calculations and working notes

**Note Metadata:**
- Each note includes frontmatter with: created date, last_modified date, status (in_progress/completed/archived), auto_delete flag
- Notes are indexed in `memory/notes/index.md`

**Automatic Cleanup:**
- Notes older than `NOTES_MAX_AGE_DAYS` (default: 30) without modifications are deleted
- Completed notes older than `NOTES_COMPLETED_RETENTION_DAYS` (default: 7) are deleted
- Scratchpad notes older than `NOTES_SCRATCHPAD_MAX_AGE_DAYS` (default: 7) are deleted
- Notes referenced in MEMORY.md or topic files are preserved regardless of age
- Notes with `auto_delete: false` are never automatically deleted
- Cleanup can be disabled with `NOTES_CLEANUP_ENABLED=false`

### MEMORY.md Index

The system maintains `agent/MEMORY.md` as an index file that provides:
- Current session status
- Recent summary information
- Chat history structure overview
- Topics knowledge base listing
- Agent notes summary
- Memory statistics

This file is automatically updated by the memory system and included in the bot's context.

## Scheduling System

The bot includes a cron-based scheduler for automated maintenance and reminders.

### Daily Maintenance Cascade

Runs at 4:00 AM daily (configurable via `DAILY_MAINTENANCE_TIME`):

1. **Daily Summary**: Generates summary of all sessions from the previous day
2. **Topic Extraction**: Extracts domain knowledge from the daily summary
3. **Notes Cleanup**: Removes old notes based on retention policies
4. **Weekly Operations** (Mondays only):
   - Weekly Reorganization: Moves daily folders into week folder
   - Weekly Summary: Generates summary from daily summaries
5. **Quarterly Operations** (First Monday of quarter only):
   - Quarterly Reorganization: Moves week folders into quarter folder
   - Quarterly Summary: Generates summary from weekly summaries
6. **Session Reset**: Clears in-memory context and starts fresh session

**Cascade Behavior:**
- Operations execute in strict sequence
- If any operation fails, the cascade aborts to prevent data inconsistency
- All operations are logged with timestamps and status

### Reminders and Notifications

The scheduler supports reminder tasks that send messages via Telegram:

**Recurring Reminders:**
- Use cron expressions: `0 8 * * *` (daily at 8 AM), `0 */2 * * *` (every 2 hours)
- Support time ranges: `0 13-20 * * *` (every hour from 1 PM to 8 PM)
- Support step values: `*/15 * * * *` (every 15 minutes)

**One-Time Reminders:**
- Execute once at a specific timestamp
- Automatically deleted after execution

**Job Lifecycle:**
- `starts_at`: Job doesn't execute until this time
- `paused_until`: Temporarily suspend job until this time
- `expires_at`: Job stops and is deleted at this time

**Cron Messages:**
- Reminder messages are sent via Telegram
- Logged to session logs as: `[timestamp] Assistant (Cron): <message>`
- Users can respond to cron messages naturally

### Agent Tools

The bot provides several tools that the LLM can use programmatically:

**Cron Management Tool:**
- Create, remove, enable, disable, list cron jobs
- Create recurring and one-time reminders
- Pause, resume, and extend job expiration
- View execution history

**Memory Summary Tool:**
- Retrieve daily, weekly, quarterly summaries
- Query summaries by date range
- Access structured summary data with metadata

**Topic Knowledge Tool:**
- Retrieve specific topic files
- List all available topics
- Search across topics by keyword
- Write, create, or append to topic files immediately

**Chat Log Search Tool:**
- Search raw session logs by keyword
- Filter by date range
- Retrieve complete session logs
- Returns excerpts with context (use sparingly, prefer summaries)

**Notes Management Tool:**
- Create, read, update, delete notes
- List notes by category or status
- Automatically manage metadata and index

## Identity Files

The bot's behavior is influenced by five identity files located in the `agent/` directory. These files are loaded on startup and included in the system context for every LLM request.

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

### MEMORY.md

Auto-generated index file that provides an overview of the memory system (created automatically by the bot).

**Example:**
```markdown
# Memory Index

**Last Updated:** 2024-01-15 10:30:00

## Current Context

### Active Session
- Session: 003
- Started: 2024-01-15 09:00:00
- Messages: 15
- Topic: Docker deployment

### Recent Summary
Last daily summary: 2024-01-14
Last weekly summary: Week 02, 2024
Last quarterly summary: Q4 2023

## Chat History Structure

### Current Quarter: Q1 2024
- **Week 03 (2024-01-15 to 2024-01-21)**
  - 2024-01-15: 3 sessions, Docker focus
  - Summary: memory/chat/week-03-2024/summary.md

## Topics Knowledge Base

### Programming
- **Docker** (memory/topics/Programming/Docker.md)
  - Last updated: 2024-01-15
  - Coverage: Container basics, Go deployment

### Psychology
- **Conversation Patterns** (memory/topics/Psychology.md)
  - Last updated: 2024-01-10
  - Coverage: User preferences, learning style

## Agent Notes

Active notes: 2
- Tasks: 1 in progress
- Ideas: 1 in progress
```

### Customizing Identity Files

1. Edit the files in the `agent/` directory to customize bot behavior
2. Restart the bot to load the new identity files
3. The bot will fail to start if IDENTITY.md, PERSONALITY.md, SOUL.md, or USER.md are missing
4. MEMORY.md is auto-generated and should not be manually edited

## Usage

### Basic Interaction

1. Start a conversation with your bot on Telegram
2. Send any message to begin chatting
3. The bot maintains conversation history permanently in the memory system
4. All conversations are logged and organized hierarchically

### Memory Features

**Viewing Memory:**
- Ask the bot about past conversations: "What did we discuss yesterday?"
- The bot can access summaries and search through conversation history
- Topic-based knowledge is automatically organized and accessible

**Manual Topic Writing:**
- Explicitly request to remember information: "Remember that I prefer Docker over Kubernetes"
- The bot immediately writes to the appropriate topic file
- Confirm what was saved: "I've saved that to Programming.md: You prefer Docker over Kubernetes"

**Manual Session Reset:**
- Request a fresh start: "Reset the session" or "Start fresh"
- The bot generates a summary, extracts topics, and starts a new session
- Previous conversation is preserved in session logs and summaries

### Scheduling Reminders

**Create Reminders:**
- "Remind me tomorrow at 3 PM about the meeting"
- "Remind me every day at 8 AM to take vitamins"
- "Remind me every hour from 1 PM to 8 PM to drink water"
- "Remind me every 10 minutes starting 30 minutes before my 3 PM meeting"

**Manage Reminders:**
- "What reminders do I have?"
- "Pause my medication reminder" (pauses until next day)
- "Resume my workout reminders"
- "Cancel the meeting reminder"

**Cron Expression Examples:**
- `0 8 * * *` - Every day at 8:00 AM
- `0 */2 * * *` - Every 2 hours
- `*/15 * * * *` - Every 15 minutes
- `0 9 * * 1` - Every Monday at 9:00 AM
- `0 13-20 * * *` - Every hour from 1 PM to 8 PM

### Session Management

- **Automatic Reset**: Sessions reset at 4:00 AM daily after maintenance completes
- **Manual Reset**: Request "reset session" anytime to start fresh
- **Token-Based Summarization**: When conversation gets too long, the bot automatically summarizes to maintain context
- **Permanent Logs**: All session logs are preserved forever in the memory system

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

The system consists of eight primary components:

1. **Channel (Telegram)**: Handles Telegram Bot API interactions, message polling, and retry logic
2. **OpenRouter Client**: Manages LLM API communication, context assembly, and tool request handling
3. **Agent**: Loads and manages identity files that define bot behavior
4. **Memory Manager**: Orchestrates conversation logging, summarization, topic extraction, and notes management
5. **Cron Scheduler**: Manages scheduled tasks including daily maintenance, reminders, and notifications
6. **Session Manager**: Tracks current session state and manages session logging
7. **Tool Executor**: Executes tools requested by the LLM (shell commands, memory access, cron management)
8. **Context Manager**: Retrieves relevant conversation history and topics for LLM context

### Message Flow

```
User (Telegram)
    ↓
Channel (Telegram) ←→ OpenRouter Client ←→ OpenRouter API
    ↓                         ↓                    ↓
Memory Manager           Tool Executor        Agent Tools
    ↓                         ↓                    ↓
Session Manager          Cron Scheduler      (cron, memory, topic,
    ↓                         ↓                chatlog, notes)
File System              Config Files
(memory/chat/)           (config/cron.json)
(memory/topics/)
(memory/notes/)
    ↑
Agent (Identity Files + MEMORY.md)
```

1. User sends message via Telegram
2. Channel receives and validates message
3. Memory Manager logs message to session log
4. Context Manager retrieves relevant history, summaries, and topics
5. Agent provides identity files including MEMORY.md
6. OpenRouter Client assembles context and sends to API
7. If LLM requests tool execution, Tool Executor handles it
8. Response returns through the chain to user
9. Memory Manager logs the exchange

Scheduled operations flow (cascade execution at 4:00 AM):
1. Cron Scheduler triggers daily maintenance task
2. Memory Manager generates daily summary from all session logs
3. Memory Manager performs topic extraction from summary
4. Memory Manager performs notes cleanup
5. If Monday: Memory Manager performs weekly reorganization and summary
6. If first Monday of quarter: Memory Manager performs quarterly reorganization and summary
7. Session Manager performs session reset after all operations complete

### Data Persistence

- **Permanent**: Session logs, summaries, topic files, MEMORY.md (memory/ directory)
- **Persistent**: Cron configurations, execution history (config/ directory)
- **Persistent**: Identity files (agent/ directory), application logs (logs/ directory)
- **Ephemeral**: In-memory session context (cleared on session reset)
- **Docker Volumes**: agent/, memory/, config/, and logs/ directories are mounted as volumes

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

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

Copyright (C) 2026 Martin Jablečník

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

## Contributing

Contributions are welcome! Please ensure:

1. All code is written in English (variable names, comments, documentation)
2. Tests pass before submitting: `go test ./...`
3. Code follows Go best practices and conventions
4. Documentation is updated for new features

## Support

For issues, questions, or contributions, please open an issue on the project repository.
