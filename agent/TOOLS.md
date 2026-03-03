# Tool Usage Guidelines

## Critical Rules for Tool Usage

### Cron Management Tool

When creating reminder jobs using the `cron_management` tool, you MUST follow these rules:

#### Creating Recurring Reminders

**ALWAYS use the `create_recurring_reminder` action** - this ensures all required fields are set correctly.

**Required parameters:**
- `name`: Unique job name (use descriptive names like "depakine_hourly_check")
- `schedule`: Cron expression (e.g., "0 14-23 * * *" for hourly from 14:00 to 23:00)
- `message`: The reminder message to send
- `chat_id`: The Telegram chat ID (use the current chat ID from context)

**Optional parameters:**
- `starts_at`: When the reminder should start (RFC3339 format)
- `expires_at`: When the reminder should stop (RFC3339 format)

**Example:**
```json
{
  "action": "create_recurring_reminder",
  "name": "medication_reminder",
  "schedule": "0 9,13,18 * * *",
  "message": "Time to take your medication!",
  "chat_id": 1100684093,
  "expires_at": "2026-12-31T23:59:59Z"
}
```

#### Creating One-Time Reminders

**ALWAYS use the `create_onetime_reminder` action** for reminders that should execute only once.

**Required parameters:**
- `name`: Unique job name
- `execute_at`: Exact timestamp when to execute (RFC3339 format)
- `message`: The reminder message to send
- `chat_id`: The Telegram chat ID

**Example:**
```json
{
  "action": "create_onetime_reminder",
  "name": "appointment_reminder",
  "execute_at": "2026-03-05T14:30:00Z",
  "message": "Doctor appointment in 30 minutes!",
  "chat_id": 1100684093
}
```

#### NEVER Use Generic `add` Action for Reminders

❌ **WRONG:**
```json
{
  "action": "add",
  "name": "my_reminder",
  "schedule": "0 9 * * *",
  "task_type": ""
}
```

This will create a job with empty `task_type` which will fail to execute!

✅ **CORRECT:**
```json
{
  "action": "create_recurring_reminder",
  "name": "my_reminder",
  "schedule": "0 9 * * *",
  "message": "Good morning!",
  "chat_id": 1100684093
}
```

## Why This Matters

The scheduler needs to know what type of task to execute. If `task_type` is empty or missing, the job will fail with "unknown task type" error. Using the specialized actions (`create_recurring_reminder` and `create_onetime_reminder`) ensures the `task_type` is automatically set to "reminder" and all required fields are populated correctly.

## Other Tool Guidelines

### Shell Tool

- Always validate commands before execution
- Never execute destructive commands without user confirmation
- Provide clear output and error messages

### Memory Tools

- Use memory summaries to recall past conversations
- Search chat logs only when necessary (token-intensive)
- Keep topic knowledge organized and up-to-date
