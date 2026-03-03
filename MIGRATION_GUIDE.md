# Memory and Scheduling System Migration Guide

This guide explains how to initialize and configure the Memory and Scheduling System for OpenCrow.

## Table of Contents

1. [System Overview](#system-overview)
2. [Initializing the Memory System](#initializing-the-memory-system)
3. [Configuring Cron Jobs](#configuring-cron-jobs)
4. [Using Agent Tools](#using-agent-tools)
5. [Troubleshooting](#troubleshooting)

## System Overview

The Memory and Scheduling System adds three major capabilities to OpenCrow:

1. **Memory Management**: Permanent conversation logging with hierarchical organization (daily → weekly → quarterly)
2. **Scheduling**: Cron-based task scheduler for automated maintenance and reminders
3. **Agent Tools**: Programmatic interfaces for the LLM to access memory and manage scheduling

### Key Components

- **Memory Manager**: Coordinates session logging, summarization, topic extraction, and notes management
- **Cron Scheduler**: Executes scheduled tasks including daily maintenance and reminders
- **Agent Tools**: Five tools for cron management, memory access, topic knowledge, chat log search, and notes management

## Initializing the Memory System

The memory system initializes automatically on first startup. No manual setup is required.

### Automatic Initialization

When the bot starts for the first time, it automatically creates:

```
memory/
├── chat/                    # Conversation logs
├── topics/                  # Domain-specific knowledge
└── notes/                   # Agent's private notes
    ├── tasks/
    ├── ideas/
    ├── reflections/
    └── scratchpad/

agent/
└── MEMORY.md               # Memory index (auto-generated)

memory/notes/
└── index.md                # Notes index (auto-generated)
```

### Verification

To verify the memory system initialized correctly:

1. Start the bot:
   ```bash
   docker-compose up -d
   ```

2. Check the logs for initialization messages:
   ```bash
   docker-compose logs bot | grep "Memory"
   ```

   You should see:
   ```
   [MemoryManager] Initializing memory directory structure
   [MemoryManager] Created directory: memory/chat
   [MemoryManager] Created directory: memory/topics
   [MemoryManager] Created directory: memory/notes
   [MemoryManager] Created MEMORY.md with initial template
   [MemoryManager] Memory directory structure initialized successfully
   ```

3. Verify the directory structure:
   ```bash
   ls -la memory/
   ls -la memory/notes/
   ls -la agent/MEMORY.md
   ```

### Configuration

Configure the memory system via environment variables in `.env`:

```bash
# Memory System Configuration
MEMORY_TOKEN_THRESHOLD=50000              # Token limit for emergency summarization
TOPIC_SIZE_THRESHOLD=102400               # Topic file size threshold (100KB)
NOTES_CLEANUP_ENABLED=true                # Enable automatic notes cleanup
NOTES_MAX_AGE_DAYS=30                     # Max age for notes without modifications
NOTES_COMPLETED_RETENTION_DAYS=7          # Retention for completed notes
NOTES_SCRATCHPAD_MAX_AGE_DAYS=7           # Max age for scratchpad notes
DAILY_MAINTENANCE_TIME=0 4 * * *          # Cron schedule for daily maintenance
```

### First Conversation

After initialization, start a conversation with the bot:

1. Send a message via Telegram
2. The bot creates the first session log: `memory/chat/YYYY-MM-DD/session-001.log`
3. MEMORY.md is updated with current session information

## Configuring Cron Jobs

The cron scheduler initializes automatically with a default daily maintenance job.

### Default Jobs

On first startup, the scheduler creates:

**Daily Maintenance Cascade** (`daily_maintenance_cascade`):
- Schedule: `0 4 * * *` (4:00 AM daily, configurable via `DAILY_MAINTENANCE_TIME`)
- Task Type: `maintenance_cascade`
- Operations:
  1. Generate daily summary
  2. Extract topics from summary
  3. Clean up old notes
  4. Weekly reorganization (Mondays only)
  5. Weekly summary (Mondays only)
  6. Quarterly reorganization (First Monday of quarter only)
  7. Quarterly summary (First Monday of quarter only)
  8. Session reset

### Viewing Cron Configuration

The cron configuration is stored in `config/cron.json`:

```bash
cat config/cron.json
```

Example output:
```json
{
  "version": "1.0",
  "jobs": [
    {
      "name": "daily_maintenance_cascade",
      "schedule": "0 4 * * *",
      "task_type": "maintenance_cascade",
      "enabled": true,
      "status": "enabled",
      "description": "Execute daily maintenance cascade"
    }
  ]
}
```

### Viewing Execution History

The execution history is stored in `config/cron_history.json`:

```bash
cat config/cron_history.json
```

Example output:
```json
{
  "version": "1.0",
  "max_entries": 1000,
  "entries": [
    {
      "timestamp": "2024-01-15T04:00:00Z",
      "job_name": "daily_maintenance_cascade",
      "task_type": "maintenance_cascade",
      "status": "success",
      "operations_completed": [
        "daily_summary",
        "topic_extraction",
        "notes_cleanup",
        "session_reset"
      ]
    }
  ]
}
```

### Creating Custom Reminders

You can create reminders by asking the bot:

**Daily Reminder:**
```
User: Remind me every day at 8 AM to take vitamins
Bot: I've created a daily reminder for 8:00 AM. You'll receive: "Time to take your vitamins!"
```

**Hourly Reminder with Time Range:**
```
User: Remind me every hour from 1 PM to 8 PM to drink water
Bot: I've set up hourly reminders from 13:00 to 20:00. You'll receive: "Time to drink water!"
```

**One-Time Reminder:**
```
User: Remind me tomorrow at 3 PM about the meeting
Bot: I've set a one-time reminder for tomorrow at 15:00. You'll receive: "Reminder: Meeting!"
```

**Frequent Reminders Before Event:**
```
User: Remind me every 10 minutes starting 30 minutes before my 3 PM meeting tomorrow
Bot: I've set up reminders every 10 minutes from 14:30 to 15:00 tomorrow.
```

### Managing Reminders

**List Reminders:**
```
User: What reminders do I have?
Bot: You have 2 active reminders:
1. Daily vitamins - Every day at 8:00 AM
2. Hydration - Every hour from 13:00 to 20:00
```

**Pause Reminder:**
```
User: I took my medication
Bot: I've paused your medication reminder until tomorrow.
```

**Resume Reminder:**
```
User: Resume my workout reminders
Bot: Your workout reminders have been resumed.
```

**Cancel Reminder:**
```
User: Cancel the meeting reminder
Bot: I've removed the meeting reminder.
```

### Cron Expression Reference

Common cron expression patterns:

| Expression | Description |
|------------|-------------|
| `0 8 * * *` | Every day at 8:00 AM |
| `0 */2 * * *` | Every 2 hours |
| `*/15 * * * *` | Every 15 minutes |
| `0 9 * * 1` | Every Monday at 9:00 AM |
| `0 9 * * 1-5` | Every weekday at 9:00 AM |
| `0 13-20 * * *` | Every hour from 1 PM to 8 PM |
| `*/10 14-16 * * *` | Every 10 minutes between 2 PM and 5 PM |
| `0 0 1 * *` | First day of every month at midnight |

Format: `minute hour day month weekday`
- minute: 0-59
- hour: 0-23
- day: 1-31
- month: 1-12
- weekday: 0-6 (0 = Sunday)

Special characters:
- `*`: Any value
- `-`: Range (e.g., `1-5`)
- `,`: List (e.g., `1,3,5`)
- `/`: Step (e.g., `*/10` = every 10)

## Using Agent Tools

The bot provides five agent tools that the LLM can use programmatically during conversations.

### 1. Cron Management Tool

Allows the LLM to manage cron jobs autonomously.

**Capabilities:**
- Add, remove, enable, disable jobs
- Create recurring and one-time reminders
- Pause, resume, extend job expiration
- View execution history

**Example Usage:**
```
User: Set up a reminder to take medication every 4 hours
Bot: [Uses Cron Management Tool to create recurring reminder]
     I've set up a reminder every 4 hours. You'll receive: "Time to take your medication!"
```

### 2. Memory Summary Tool

Retrieves conversation summaries for specific date ranges.

**Capabilities:**
- Get daily, weekly, quarterly summaries
- Query summaries by date range
- Access structured summary data with metadata

**Example Usage:**
```
User: What did we discuss last week?
Bot: [Uses Memory Summary Tool to retrieve weekly summary]
     Last week we discussed Docker deployment, Go best practices, and project planning.
```

### 3. Topic Knowledge Tool

Accesses and writes domain-specific knowledge files.

**Capabilities:**
- Retrieve specific topic files
- List all available topics
- Search across topics by keyword
- Write, create, or append to topic files immediately

**Example Usage:**
```
User: Remember that I prefer Docker over Kubernetes for small projects
Bot: [Uses Topic Knowledge Tool to write to Programming.md]
     I've saved that to Programming.md: You prefer Docker over Kubernetes for small projects.
```

### 4. Chat Log Search Tool

Searches raw conversation logs (use sparingly, prefer summaries).

**Capabilities:**
- Search by keyword or phrase
- Filter by date range
- Retrieve complete session logs
- Returns excerpts with context

**Example Usage:**
```
User: When did I mention the database migration issue?
Bot: [Uses Chat Log Search Tool to find specific conversation]
     You mentioned the database migration issue on 2024-01-10 at 14:30.
```

### 5. Notes Management Tool

Manages the agent's private working notes.

**Capabilities:**
- Create, read, update, delete notes
- List notes by category or status
- Automatically manage metadata and index

**Example Usage:**
```
Bot: [Internally uses Notes Management Tool]
     I've created a task note to research Docker security best practices for our next conversation.
```

## Troubleshooting

### Memory System Issues

**Problem**: Memory directory not created

**Solution**:
1. Check file permissions:
   ```bash
   ls -la memory/
   ```
2. Verify the bot has write access to the directory
3. Check logs for initialization errors:
   ```bash
   docker-compose logs bot | grep -i error
   ```

**Problem**: MEMORY.md not updating

**Solution**:
1. Check if the file exists:
   ```bash
   ls -la agent/MEMORY.md
   ```
2. Verify file permissions (should be writable)
3. Check logs for update errors

**Problem**: Session logs not being created

**Solution**:
1. Verify memory/chat/ directory exists
2. Check logs for session manager errors
3. Ensure the bot received and processed messages

### Cron Scheduler Issues

**Problem**: Daily maintenance not running

**Solution**:
1. Check cron configuration:
   ```bash
   cat config/cron.json
   ```
2. Verify the job is enabled: `"enabled": true`
3. Check the schedule is correct: `"schedule": "0 4 * * *"`
4. Review execution history:
   ```bash
   cat config/cron_history.json
   ```
5. Check logs for scheduler errors:
   ```bash
   docker-compose logs bot | grep -i scheduler
   ```

**Problem**: Reminders not sending

**Solution**:
1. Verify Telegram sender is configured
2. Check session logger is configured
3. Review reminder job configuration in cron.json
4. Check execution history for errors
5. Verify chat_id is correct in the job configuration

**Problem**: Cron jobs not persisting after restart

**Solution**:
1. Verify config/ directory is mounted as a volume in docker-compose.yml
2. Check file permissions on config/cron.json
3. Ensure the scheduler saves jobs after modifications

### Topic Extraction Issues

**Problem**: Topics not being extracted

**Solution**:
1. Verify LLM client is configured correctly
2. Check if conversations contain domain-specific knowledge
3. Review logs for topic extraction operations:
   ```bash
   docker-compose logs bot | grep -i topic
   ```
4. Ensure topic extraction runs during summarization

**Problem**: Topic files too large

**Solution**:
1. Adjust TOPIC_SIZE_THRESHOLD in .env:
   ```bash
   TOPIC_SIZE_THRESHOLD=51200  # 50KB instead of 100KB
   ```
2. The system will automatically subdivide large topics

### Notes Cleanup Issues

**Problem**: Notes not being cleaned up

**Solution**:
1. Verify NOTES_CLEANUP_ENABLED=true in .env
2. Check retention day settings:
   ```bash
   NOTES_MAX_AGE_DAYS=30
   NOTES_COMPLETED_RETENTION_DAYS=7
   NOTES_SCRATCHPAD_MAX_AGE_DAYS=7
   ```
3. Review logs for cleanup operations:
   ```bash
   docker-compose logs bot | grep -i cleanup
   ```

**Problem**: Important notes being deleted

**Solution**:
1. Set auto_delete to false in note frontmatter:
   ```markdown
   ---
   auto_delete: false
   ---
   ```
2. Reference the note in MEMORY.md or a topic file
3. Adjust retention day settings to be more lenient

### General Debugging

**Enable Debug Logging:**
```bash
# In .env file
LOG_LEVEL=debug
```

**View Real-Time Logs:**
```bash
docker-compose logs -f bot
```

**Check Memory System Status:**
```bash
# View directory structure
tree memory/

# Check file counts
find memory/chat -name "*.log" | wc -l
find memory/topics -name "*.md" | wc -l
find memory/notes -name "*.md" | wc -l
```

**Check Cron System Status:**
```bash
# View cron configuration
cat config/cron.json | jq .

# View execution history
cat config/cron_history.json | jq '.entries | .[-10:]'  # Last 10 entries
```

**Restart the Bot:**
```bash
docker-compose restart bot
```

**Clean Restart (preserves memory and config):**
```bash
docker-compose down
docker-compose up -d
```

## Best Practices

### Memory Management

1. **Let the system work automatically**: The memory system is designed to run autonomously
2. **Use manual topic writing for important information**: When users explicitly want to remember something
3. **Monitor disk usage**: Conversation logs accumulate over time
4. **Backup memory directory**: Regularly backup the memory/ directory to prevent data loss

### Scheduling

1. **Use descriptive job names**: Include purpose and date for one-time jobs
2. **Set appropriate expiration**: For event-based reminders
3. **Test cron expressions**: Use online cron expression testers before creating jobs
4. **Monitor execution history**: Regularly check for failed jobs

### Agent Tools

1. **Prefer summaries over raw logs**: Summaries are more efficient for token usage
2. **Use topic files for domain knowledge**: More organized than searching logs
3. **Let the agent manage notes**: The agent knows when to create and clean up notes
4. **Trust automatic cleanup**: The system preserves referenced and important notes

## Migration Checklist

- [ ] Verify memory directory structure created
- [ ] Confirm MEMORY.md exists in agent/ directory
- [ ] Check cron.json created with default jobs
- [ ] Test first conversation creates session log
- [ ] Verify daily maintenance runs at 4:00 AM
- [ ] Test manual topic writing
- [ ] Create a test reminder
- [ ] Verify reminder sends and logs to session
- [ ] Check notes directory structure
- [ ] Review execution history after first maintenance run
- [ ] Backup memory/ and config/ directories
- [ ] Document any custom configuration changes

## Support

For issues not covered in this guide:

1. Check the main README.md for general troubleshooting
2. Review the design document: `.kiro/specs/memory-and-scheduling-system/design.md`
3. Check the requirements document: `.kiro/specs/memory-and-scheduling-system/requirements.md`
4. Enable debug logging and review detailed logs
5. Open an issue on the project repository

---

**Last Updated**: 2024-01-15
**Version**: 1.0.0
