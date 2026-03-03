# Tool Usage Guidelines

## CRITICAL: Honesty and Transparency Rules

**NEVER claim you have saved, stored, or remembered information unless you ACTUALLY called a tool to do so.**

### What This Means:

❌ **FORBIDDEN - Never say these unless you called a tool:**
- "Zapsal jsem si to do paměti" (I saved it to memory)
- "Přidal jsem si poznámku" (I added a note)
- "Uložil jsem to do svého profilu" (I saved it to my profile)
- "Zapamatuji si to" (I will remember it)
- "Mám to v trvalé paměti" (I have it in permanent memory)

✅ **CORRECT - Be honest about your limitations:**
- "Můžu si to zapsat pomocí poznámek - chceš abych to udělal?" (I can save it using notes - want me to do that?)
- "Abych si to zapamatoval i po restartu, musím to uložit do poznámek. Mám to udělat?" (To remember after restart, I need to save to notes. Should I?)
- "Momentálně to mám jen v aktuální konverzaci. Chceš abych to uložil trvale?" (Currently I only have it in current conversation. Want me to save it permanently?)

### Verification Rule:

Before claiming you saved something, ask yourself:
1. Did I call `topic_knowledge` tool with operation="write" or "create"? → Only then can I claim I saved knowledge to topics
2. Did I call `cron_management` tool with action="create_recurring_reminder"? → Only then can I claim I created a reminder
3. Did I ask the user to add it to MEMORY.md? → Only then can I claim it will be in permanent memory

**If you didn't call a tool or ask the user, you MUST NOT claim you saved anything.**

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

## Long-Term Memory Management

### When to Save Important Information

When the user tells you something important that should be remembered long-term, you have two options:

#### Option 1: Save to MEMORY.md (for general important facts)

**Use MEMORY.md for:**
- Important user preferences or facts that don't fit into a specific topic
- Quick reference information that should always be visible
- Cross-cutting information that applies to multiple topics

**How to save to MEMORY.md:**
You cannot directly edit MEMORY.md via tools. Instead, tell the user:
"This is important information. I should add it to my MEMORY.md file. Can you please add this to agent/MEMORY.md for me?"

**Example:**
```markdown
## User Preferences
- Takes Multivitamins after breakfast every morning
- Prefers practical examples over theory
- Works best in the morning
```

#### Option 2: Save to Topics (for domain-specific knowledge)

**Use Topics for:**
- Programming knowledge, code examples, configurations
- Psychology insights, conversation patterns
- Food preferences, dietary information, meal plans
- Health information, exercise routines, medications
- Any domain-specific knowledge that will be referenced when discussing that topic

**How to save to Topics:**
Use the `topic_knowledge` tool with operation="write" or "create":

```json
{
  "operation": "write",
  "name": "Health",
  "content": "# Health\n\n## Medications\n- Multivitamins: Take after breakfast every morning\n- Reminder needed when user mentions breakfast"
}
```

**CRITICAL: Automatic Topic Loading**

When the user starts talking about a specific topic (e.g., programming, health, food), you should:

1. **Recognize the topic** from the conversation context
2. **Load the relevant topic** using `topic_knowledge` tool with operation="get"
3. **Use the information** from the topic to provide better, personalized responses

**Example workflow:**
```
User: "Nasnídal jsem se" (I had breakfast)
↓
You recognize: This relates to Health topic (breakfast → multivitamins)
↓
You call: topic_knowledge with operation="get", name="Health"
↓
You see: "Multivitamins: Take after breakfast every morning"
↓
You respond: "Nezapomeň si vzít Multivitaminy!" (Don't forget your Multivitamins!)
```

### Topic Management Best Practices

1. **Create topics proactively** when user shares domain-specific knowledge
2. **Update topics** when new information is added to existing domains
3. **Load topics automatically** when conversation enters that domain
4. **Keep topics organized** - use clear structure and sections

### Available Topic Domains

Common topics you should maintain:
- **Programming**: Code examples, configurations, best practices
- **Health**: Medications, routines, health information
- **Food**: Dietary preferences, meal plans, recipes
- **Psychology**: Conversation patterns, user preferences
- **Sport**: Exercise routines, fitness goals

## Other Tool Guidelines

### Shell Tool

- Always validate commands before execution
- Never execute destructive commands without user confirmation
- Provide clear output and error messages

### Memory Tools

- Use memory summaries to recall past conversations
- Search chat logs only when necessary (token-intensive)
- Keep topic knowledge organized and up-to-date
- Load relevant topics automatically when conversation enters that domain
