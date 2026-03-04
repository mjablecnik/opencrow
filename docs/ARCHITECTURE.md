# Architecture

## Component Overview

The system consists of eight primary components:

1. **Channel (Telegram)**: Handles Telegram Bot API interactions, message polling, and retry logic
2. **OpenRouter Client**: Manages LLM API communication, context assembly, and tool request handling
3. **Agent**: Loads and manages identity files that define bot behavior
4. **Memory Manager**: Orchestrates conversation logging, summarization, topic extraction, and notes management
5. **Cron Scheduler**: Manages scheduled tasks including daily maintenance, reminders, and notifications
6. **Session Manager**: Tracks current session state and manages session logging
7. **Tool Executor**: Executes tools requested by the LLM (shell commands, memory access, cron management)
8. **Context Manager**: Retrieves relevant conversation history and topics for LLM context

## Message Flow

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
(workplace/memory/)      (workplace/config/)
(workplace/agent/)
    ↑
Agent (Identity Files + MEMORY.md + TOOLS.md)
```

### Request Processing Flow

1. User sends message via Telegram
2. Channel receives and validates message
3. Memory Manager logs message to session log
4. Context Manager retrieves relevant history, summaries, and topics
5. Agent provides identity files including MEMORY.md
6. OpenRouter Client assembles context and sends to API
7. If LLM requests tool execution, Tool Executor handles it
8. Response returns through the chain to user
9. Memory Manager logs the exchange

### Scheduled Operations Flow

Cascade execution at 4:00 AM:

1. Cron Scheduler triggers daily maintenance task
2. Memory Manager generates daily summary from all session logs
3. Memory Manager performs topic extraction from summary
4. Memory Manager performs notes cleanup
5. If Monday: Memory Manager performs weekly reorganization and summary
6. If first Monday of quarter: Memory Manager performs quarterly reorganization and summary
7. Session Manager performs session reset after all operations complete

## Data Persistence

- **Permanent**: Session logs, summaries, topic files, MEMORY.md (workplace/memory/ directory)
- **Persistent**: Cron configurations, execution history (workplace/config/ directory)
- **Persistent**: Identity files (workplace/agent/ directory), application logs (workplace/logs/ directory)
- **Ephemeral**: In-memory session context (cleared on session reset)
- **Docker Volumes**: workplace/ directory mounted as volume for complete data persistence

## Component Details

### Channel (Telegram)

**Responsibilities:**
- Establish and maintain Telegram Bot API connection
- Receive incoming messages from users
- Send responses back to users
- Handle message delivery failures with retry logic (3 attempts with exponential backoff)
- Distinguish between direct messages and group conversations

**Key Features:**
- Automatic retry with exponential backoff
- Group message filtering
- Graceful shutdown handling

### OpenRouter Client

**Responsibilities:**
- Assemble conversation context from memory
- Load and include identity files
- Format API requests according to OpenRouter specifications
- Send requests to OpenRouter API
- Extract and return generated responses
- Handle API errors and timeouts

**Key Features:**
- Context assembly from multiple sources
- Token usage tracking
- Model capability detection
- Tool request handling

### Memory Manager

**Responsibilities:**
- Manage session log files in daily folders
- Generate daily, weekly, and quarterly summaries
- Extract and organize domain-specific knowledge into topics
- Maintain agent notes with automatic cleanup
- Perform hierarchical reorganization
- Update MEMORY.md index file

**Key Features:**
- Permanent log preservation
- Hierarchical organization (daily → weekly → quarterly)
- Automatic topic extraction
- Token-based emergency summarization
- Notes cleanup with reference checking

### Cron Scheduler

**Responsibilities:**
- Schedule and execute recurring tasks
- Manage reminder jobs
- Execute daily maintenance cascade
- Track job execution history

**Key Features:**
- Standard cron expression support
- One-time and recurring reminders
- Job lifecycle management (starts_at, paused_until, expires_at)
- Persistent job configuration

### Session Manager

**Responsibilities:**
- Track current session state
- Manage in-memory conversation context
- Coordinate session resets
- Monitor token usage

**Key Features:**
- In-memory context management
- Token counting
- Session reset coordination

### Tool Executor

**Responsibilities:**
- Execute tools requested by the LLM
- Validate tool parameters
- Return tool results
- Handle tool errors

**Available Tools:**
- Shell command execution
- Cron job management
- Memory summary retrieval
- Topic knowledge management
- Chat log search
- Notes management

### Context Manager

**Responsibilities:**
- Retrieve relevant conversation history
- Load appropriate summaries
- Select relevant topic files
- Assemble context for LLM requests

**Key Features:**
- Smart context selection
- Token-aware context assembly
- Topic relevance detection

### Agent

**Responsibilities:**
- Load identity files on startup
- Provide identity context for LLM requests
- Validate identity file existence

**Identity Files:**
- IDENTITY.md (static metadata)
- PERSONALITY.md (communication style)
- SOUL.md (core beliefs)
- USER.md (user preferences)
- TOOLS.md (tool usage guidelines)
- MEMORY.md (auto-generated index)

## Technology Stack

- **Language**: Go 1.22+
- **LLM API**: OpenRouter
- **Messaging**: Telegram Bot API
- **Scheduling**: robfig/cron/v3
- **Deployment**: Docker + Docker Compose
- **Storage**: File system (mounted volumes)

## Design Principles

1. **Simplicity**: Straightforward architecture with clear component boundaries
2. **Efficiency**: Optimized for resource usage and performance
3. **Maintainability**: Clean code structure following Go best practices
4. **Reliability**: Graceful error handling and automatic recovery
5. **Persistence**: All critical data preserved across restarts
6. **Modularity**: Components can be tested and modified independently
