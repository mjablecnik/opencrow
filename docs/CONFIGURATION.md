# Configuration Guide

## Environment Variables

The bot is configured entirely through environment variables. All variables can be set in the `.env` file or passed directly to the Docker container.

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `TELEGRAM_BOT_TOKEN` | Your Telegram bot token from @BotFather | `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11` |
| `OPENROUTER_API_KEY` | Your OpenRouter API key | `sk-or-v1-...` |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MODEL_NAME` | LLM model to use for responses | `google/gemini-2.5-flash-lite` |
| `SHELL_TIMEOUT` | Timeout for shell command execution | `30s` |
| `LOG_LEVEL` | Logging level: `debug`, `info`, `warn`, `error` | `info` |
| `MEMORY_TOKEN_THRESHOLD` | Token limit for emergency summarization | `50000` |
| `TOPIC_SIZE_THRESHOLD` | Topic file size threshold for subdivision (bytes) | `102400` (100KB) |
| `NOTES_ENABLED` | Enable notes system | `true` |
| `NOTES_CLEANUP_ENABLED` | Enable automatic notes cleanup | `true` |
| `NOTES_MAX_AGE_DAYS` | Maximum age for notes without modifications | `30` |
| `NOTES_COMPLETED_RETENTION_DAYS` | Retention days for completed notes | `7` |
| `NOTES_SCRATCHPAD_MAX_AGE_DAYS` | Maximum age for scratchpad notes | `7` |
| `DAILY_MAINTENANCE_TIME` | Cron schedule for daily maintenance | `0 4 * * *` (4:00 AM) |

## Getting Your Credentials

### Telegram Bot Token

1. Open Telegram and search for [@BotFather](https://t.me/botfather)
2. Send `/newbot` command
3. Follow the prompts to create your bot
4. Copy the bot token provided

### OpenRouter API Key

1. Visit [OpenRouter](https://openrouter.ai/)
2. Sign up or log in
3. Navigate to API Keys section
4. Create a new API key
5. Copy the key (starts with `sk-or-v1-`)

## Directory Structure

```
opencrow/
├── workplace/                      # Runtime data directory (Docker volume mount)
│   ├── agent/                      # Identity and personality files
│   │   ├── IDENTITY.md             # Bot's static technical metadata
│   │   ├── PERSONALITY.md          # Communication style and behavior patterns
│   │   ├── SOUL.md                 # Core beliefs and authentic self
│   │   ├── USER.md                 # User preferences and context
│   │   ├── TOOLS.md                # Tool usage guidelines
│   │   └── MEMORY.md               # Memory index (auto-generated)
│   ├── memory/                     # Memory system data (created at runtime)
│   │   ├── chat/                   # Conversation logs
│   │   │   ├── 2026-03-03/        # Daily folder
│   │   │   │   ├── session-001.log
│   │   │   │   ├── session-001-summary.md
│   │   │   │   └── daily-summary.md
│   │   │   ├── week-09-2026/      # Weekly folder
│   │   │   │   ├── 2026-03-03/
│   │   │   │   └── summary.md
│   │   │   └── Q1-2026/           # Quarterly folder
│   │   │       ├── week-01-2026/
│   │   │       └── summary.md
│   │   ├── topics/                 # Domain-specific knowledge
│   │   │   ├── Programming.md
│   │   │   ├── Psychology.md
│   │   │   └── Food.md
│   │   └── notes/                  # Agent's private notes
│   │       ├── index.md
│   │       ├── tasks/
│   │       ├── ideas/
│   │       ├── reflections/
│   │       └── scratchpad/
│   ├── config/                     # Configuration (created at runtime)
│   │   ├── cron.json               # Cron job configurations
│   │   └── cron_history.json       # Execution history
│   └── logs/                       # Application logs (created at runtime)
│       └── bot.log                 # Main log file
├── cmd/
│   └── bot/
│       └── main.go                 # Application entry point
├── internal/
│   ├── agent/                      # Identity file loading
│   ├── channel/                    # Telegram bot interface
│   ├── config/                     # Configuration management
│   ├── llm/                        # OpenRouter client
│   ├── memory/                     # Memory management system
│   ├── scheduler/                  # Cron scheduling system
│   ├── session/                    # In-memory session manager
│   └── tools/                      # Tool executor and tools
├── pkg/
│   └── utils/                      # Logging and file utilities
├── scripts/
│   ├── entrypoint.sh               # Docker entrypoint script
│   ├── deploy.sh                   # Deployment script
│   └── setup-secrets.sh            # Secrets setup script
├── docs/                           # Documentation
├── .env.example                    # Environment variable template
├── docker-compose.yml              # Docker Compose configuration
├── Dockerfile                      # Docker image definition
├── fly.toml                        # Fly.io deployment configuration
├── go.mod                          # Go module definition
└── README.md                       # Main documentation
```

### Important Directories

- **workplace/**: Root directory for all runtime data, mounted as a Docker volume for persistence
  - **agent/**: Contains identity files that define the bot's behavior and personality. These files are loaded on startup and included in every LLM request. Includes MEMORY.md which is auto-generated by the memory system, and TOOLS.md with tool usage guidelines.
  - **memory/**: Stores all conversation history, domain knowledge, and agent notes. This directory is automatically created and managed by the memory system.
    - **chat/**: Hierarchical conversation logs organized by day, week, and quarter
    - **topics/**: Domain-specific knowledge files (Programming, Psychology, Food, etc.)
    - **notes/**: Agent's private working notes organized by category
  - **config/**: Stores cron job configurations and execution history. Created automatically at runtime.
  - **logs/**: Stores application logs. In Docker deployments, this directory is mounted as a volume for persistence.
