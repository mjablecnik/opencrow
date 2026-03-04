# Scheduling System

The bot includes a cron-based scheduler for automated maintenance and reminders.

## Daily Maintenance Cascade

Runs at 4:00 AM daily (configurable via `DAILY_MAINTENANCE_TIME` environment variable).

### Execution Sequence

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

### Cascade Behavior

- Operations execute in strict sequence
- If any operation fails, the cascade aborts to prevent data inconsistency
- All operations are logged with timestamps and status

## Reminders and Notifications

The scheduler supports reminder tasks that send messages via Telegram.

### Recurring Reminders

Use cron expressions for recurring reminders:

- `0 8 * * *` - Daily at 8:00 AM
- `0 */2 * * *` - Every 2 hours
- `*/15 * * * *` - Every 15 minutes
- `0 9 * * 1` - Every Monday at 9:00 AM
- `0 13-20 * * *` - Every hour from 1 PM to 8 PM

### One-Time Reminders

- Execute once at a specific timestamp
- Automatically deleted after execution

### Job Lifecycle

- **starts_at**: Job doesn't execute until this time
- **paused_until**: Temporarily suspend job until this time
- **expires_at**: Job stops and is deleted at this time

### Cron Messages

- Reminder messages are sent via Telegram
- Logged to session logs as: `[timestamp] Assistant (Cron): <message>`
- Users can respond to cron messages naturally

## Agent Tools

The bot provides several tools that the LLM can use programmatically.

### Cron Management Tool

- Create, remove, enable, disable, list cron jobs
- Create recurring and one-time reminders
- Pause, resume, and extend job expiration
- View execution history

### Memory Summary Tool

- Retrieve daily, weekly, quarterly summaries
- Query summaries by date range
- Access structured summary data with metadata

### Topic Knowledge Tool

- Retrieve specific topic files
- List all available topics
- Search across topics by keyword
- Write, create, or append to topic files immediately

### Chat Log Search Tool

- Search raw session logs by keyword
- Filter by date range
- Retrieve complete session logs
- Returns excerpts with context (use sparingly, prefer summaries)

### Notes Management Tool

- Create, read, update, delete notes
- List notes by category or status
- Automatically manage metadata and index

## Usage Examples

### Creating Reminders

**Daily reminder:**
```
"Remind me every day at 8 AM to take vitamins"
```

**Hourly reminder with time range:**
```
"Remind me every hour from 1 PM to 8 PM to drink water"
```

**One-time reminder:**
```
"Remind me tomorrow at 3 PM about the meeting"
```

**Reminder with expiration:**
```
"Remind me every day at 9 AM to exercise, but stop after 30 days"
```

### Managing Reminders

**List all reminders:**
```
"What reminders do I have?"
```

**Pause a reminder:**
```
"Pause my medication reminder"
```

**Resume a reminder:**
```
"Resume my workout reminders"
```

**Cancel a reminder:**
```
"Cancel the meeting reminder"
```

## Cron Expression Format

Standard cron expression format with 5 fields:

```
* * * * *
│ │ │ │ │
│ │ │ │ └─── Day of week (0-6, Sunday=0)
│ │ │ └───── Month (1-12)
│ │ └─────── Day of month (1-31)
│ └───────── Hour (0-23)
└─────────── Minute (0-59)
```

### Special Characters

- `*` - Any value
- `,` - Value list separator (e.g., `1,3,5`)
- `-` - Range of values (e.g., `1-5`)
- `/` - Step values (e.g., `*/15` = every 15 minutes)

### Examples

- `0 8 * * *` - Every day at 8:00 AM
- `0 */2 * * *` - Every 2 hours
- `*/15 * * * *` - Every 15 minutes
- `0 9 * * 1` - Every Monday at 9:00 AM
- `0 13-20 * * *` - Every hour from 1 PM to 8 PM
- `30 14 * * 1-5` - Every weekday at 2:30 PM
