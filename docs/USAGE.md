# Usage Guide

## Basic Interaction

1. Start a conversation with your bot on Telegram
2. Send any message to begin chatting
3. The bot maintains conversation history permanently in the memory system
4. All conversations are logged and organized hierarchically

## Memory Features

### Viewing Memory

Ask the bot about past conversations:

- "What did we discuss yesterday?"
- "What do you remember about my preferences?"
- "Show me my programming notes"

The bot can access summaries and search through conversation history. Topic-based knowledge is automatically organized and accessible.

### Manual Topic Writing

Explicitly request to remember information:

- "Remember that I prefer Docker over Kubernetes"
- "Save this to my programming notes: I use Go 1.22"
- "Remember that I hate onions"

The bot immediately writes to the appropriate topic file and confirms what was saved.

### Manual Session Reset

Request a fresh start:

- "Reset the session"
- "Start fresh"
- "Clear the conversation"

The bot generates a summary, extracts topics, and starts a new session. Previous conversation is preserved in session logs and summaries.

## Scheduling Reminders

### Creating Reminders

**Daily reminders:**
```
"Remind me every day at 8 AM to take vitamins"
"Remind me daily at 9 PM to review my tasks"
```

**Hourly reminders:**
```
"Remind me every 2 hours to drink water"
"Remind me every hour from 1 PM to 8 PM to stretch"
```

**One-time reminders:**
```
"Remind me tomorrow at 3 PM about the meeting"
"Remind me in 30 minutes to check the oven"
```

**Weekly reminders:**
```
"Remind me every Monday at 9 AM about the team meeting"
"Remind me every Friday at 5 PM to submit my timesheet"
```

**Complex reminders:**
```
"Remind me every 10 minutes starting 30 minutes before my 3 PM meeting"
"Remind me every weekday at 2:30 PM to take a break"
```

### Managing Reminders

**List reminders:**
```
"What reminders do I have?"
"Show me all my scheduled tasks"
"List my active reminders"
```

**Pause reminders:**
```
"Pause my medication reminder"
"Temporarily disable the water reminder"
```

**Resume reminders:**
```
"Resume my workout reminders"
"Enable the medication reminder again"
```

**Cancel reminders:**
```
"Cancel the meeting reminder"
"Delete all my reminders"
"Remove the vitamin reminder"
```

## Session Management

### Automatic Reset

Sessions reset at 4:00 AM daily after maintenance completes. This ensures:

- Fresh context for the new day
- Summaries are generated from previous day
- Topics are extracted and organized
- Old notes are cleaned up

### Manual Reset

Request "reset session" anytime to start fresh:

```
"Reset the session"
"Start a new conversation"
"Clear the context"
```

The bot will:
1. Generate a summary of the current session
2. Extract any relevant topics
3. Clear the in-memory context
4. Start a new session

### Token-Based Summarization

When conversation gets too long (exceeds token threshold), the bot automatically:

1. Notifies you: "Performing memory summarization due to high token usage..."
2. Summarizes the current session
3. Clears the context
4. Continues the conversation seamlessly

### Permanent Logs

All session logs are preserved forever in the memory system:

- Located in `workplace/memory/chat/`
- Organized by date (daily folders)
- Rolled up into weekly and quarterly folders
- Never deleted, always accessible

## Working with Topics

### Automatic Topic Management

The bot automatically extracts and organizes domain-specific knowledge:

- **Programming**: Code examples, configurations, preferences
- **Psychology**: Conversation patterns, user preferences
- **Food**: Dietary tracking, preferences, meal plans
- **Sport-Health**: Exercise routines, health information
- **Politics**: User's views and preferences

### Manual Topic Updates

Explicitly save information to topics:

```
"Remember in my programming notes: I prefer tabs over spaces"
"Save to my food preferences: I'm allergic to peanuts"
"Add to my health notes: I take medication at 8 AM and 8 PM"
```

### Viewing Topics

Ask the bot to retrieve topic information:

```
"What do you know about my programming preferences?"
"Show me my food notes"
"What health information do you have about me?"
```

## Using Agent Notes

The bot can create private notes for complex tasks:

```
"Create a note to research Docker security best practices"
"Add an idea: Build a CLI tool for managing reminders"
"Make a reflection note about today's conversation"
```

Notes are automatically cleaned up based on age and status, unless marked as important.

## Tips and Best Practices

1. **Be specific with reminders**: Include exact times and clear messages
2. **Use topics for long-term knowledge**: Explicitly ask the bot to remember important information
3. **Reset sessions when switching contexts**: Start fresh when changing topics or tasks
4. **Review summaries periodically**: Ask the bot about past conversations to see what's been captured
5. **Customize identity files**: Edit the files in `workplace/agent/` to personalize the bot's behavior
