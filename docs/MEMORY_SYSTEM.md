# Memory System

The bot features a comprehensive memory system that permanently preserves all conversations while organizing them hierarchically for efficient access.

## Conversation Memory

### Session Logging

- Each conversation is logged to session files (session-001.log, session-002.log, etc.)
- Sessions are organized in daily folders (YYYY-MM-DD format)
- All session logs are preserved permanently and never deleted
- Messages include timestamps: `[YYYY-MM-DD HH:MM:SS] Role: Content`

### Hierarchical Organization

- **Daily folders**: Created automatically for each day
- **Weekly folders**: Created every Monday, containing the previous 7 daily folders
- **Quarterly folders**: Created on the first day of each quarter, containing all week folders from the completed quarter

### Summarization

- **Session summaries**: Generated when sessions are manually reset
- **Daily summaries**: Generated at 4:00 AM, summarizing all sessions from the previous day
- **Weekly summaries**: Generated on Mondays, summarizing all daily summaries from the week
- **Quarterly summaries**: Generated on the first day of each quarter, summarizing all weekly summaries

### Token-Based Summarization

When conversation token usage exceeds `MEMORY_TOKEN_THRESHOLD` (default: 50,000), the system automatically:

1. Sends notification: "Performing memory summarization due to high token usage..."
2. Generates a summary of the current session
3. Clears the in-memory context
4. Inserts the summary at the beginning of the session
5. Continues the conversation seamlessly without session reset

## Topic-Based Knowledge

The system automatically extracts domain-specific knowledge from conversations and organizes it into topic files.

### Automatic Topic Extraction

- Runs during all summarization operations (daily, weekly, quarterly, session, token-based)
- Uses LLM to identify relevant domain knowledge
- Only creates/updates topic files when relevant knowledge is found
- Logs "no relevant domain knowledge found" when appropriate

### Topic Organization

- Topics stored as markdown files: `Programming.md`, `Psychology.md`, `Food.md`, etc.
- When a topic file exceeds `TOPIC_SIZE_THRESHOLD` (default: 100KB), it's automatically subdivided into a folder structure
- Supports hierarchical organization for large topics

### Manual Topic Writing

- Users can explicitly request to remember information: "Remember that I hate onions"
- Agent immediately writes to the appropriate topic file
- No need to wait for scheduled summarization

### Supported Topic Domains

- **Programming**: Code examples, configurations, preferences, mistakes to avoid
- **Psychology**: Conversation patterns, user preferences, communication style
- **Food**: Dietary tracking, preferences, meal plans
- **Sport-Health**: Exercise routines, health information, medical notes
- **Politics**: User's views and preferences
- **Custom topics**: System dynamically creates new topics as needed

## Agent Notes

The bot maintains private working notes for complex tasks, ideas, and temporary calculations.

### Note Categories

- **tasks/**: Task planning and execution notes
- **ideas/**: Ideas and brainstorming
- **reflections/**: Reflections on conversations and learning
- **scratchpad/**: Temporary calculations and working notes

### Note Metadata

- Each note includes frontmatter with: created date, last_modified date, status (in_progress/completed/archived), auto_delete flag
- Notes are indexed in `memory/notes/index.md`

### Automatic Cleanup

- Notes older than `NOTES_MAX_AGE_DAYS` (default: 30) without modifications are deleted
- Completed notes older than `NOTES_COMPLETED_RETENTION_DAYS` (default: 7) are deleted
- Scratchpad notes older than `NOTES_SCRATCHPAD_MAX_AGE_DAYS` (default: 7) are deleted
- Notes referenced in MEMORY.md or topic files are preserved regardless of age
- Notes with `auto_delete: false` are never automatically deleted
- Cleanup can be disabled with `NOTES_CLEANUP_ENABLED=false`

## MEMORY.md Index

The system maintains `workplace/agent/MEMORY.md` as an index file that provides:

- Current session status
- Recent summary information
- Chat history structure overview
- Topics knowledge base listing
- Agent notes summary
- Memory statistics

This file is automatically updated by the memory system and included in the bot's context.

### Example MEMORY.md

```markdown
# Memory Index

**Last Updated:** 2026-03-03 10:30:00

## Current Context

### Active Session
- Session: 003
- Started: 2026-03-03 09:00:00
- Messages: 15
- Topic: Docker deployment

### Recent Summary
Last daily summary: 2026-03-02
Last weekly summary: Week 09, 2026
Last quarterly summary: Q1 2026

## Chat History Structure

### Current Quarter: Q1 2026
- **Week 09 (2026-03-03 to 2026-03-09)**
  - 2026-03-03: 3 sessions, Docker focus
  - Summary: memory/chat/week-09-2026/summary.md

## Topics Knowledge Base

### Programming
- **Docker** (memory/topics/Programming/Docker.md)
  - Last updated: 2026-03-03
  - Coverage: Container basics, Go deployment

### Psychology
- **Conversation Patterns** (memory/topics/Psychology.md)
  - Last updated: 2026-03-01
  - Coverage: User preferences, learning style

## Agent Notes

Active notes: 2
- Tasks: 1 in progress
- Ideas: 1 in progress
```
