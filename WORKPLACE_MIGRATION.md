# Workplace Directory Migration Guide

## Overview

This document describes the migration from multiple mounted directories to a single `workplace/` directory in the Docker container.

## Changes Made

### 1. Directory Structure

Created new `workplace/` directory with the following structure:

```
opencrow/
├── workplace/              # ← NEW: Single mount point for Docker
│   ├── agent/             # Bot identity and personality files
│   │   ├── IDENTITY.md
│   │   ├── PERSONALITY.md
│   │   ├── SOUL.md
│   │   ├── TOOLS.md
│   │   └── USER.md
│   ├── config/            # Runtime configuration
│   │   ├── cron.json
│   │   └── cron_history.json
│   ├── memory/            # Conversation memory
│   │   ├── chat/
│   │   ├── notes/
│   │   ├── topics/
│   │   └── MEMORY.md
│   └── logs/              # Application logs
```

### 2. Docker Configuration Updates

#### docker-compose.yml
- **Before:** 4 separate volume mounts
  ```yaml
  volumes:
    - ./agent:/app/agent:ro
    - ./logs:/app/logs
    - ./memory:/app/memory
    - ./config:/app/config
  ```

- **After:** Single volume mount
  ```yaml
  volumes:
    - ./workplace:/app/workplace
  ```

#### Dockerfile
- **Before:** Multiple volume declarations
  ```dockerfile
  RUN mkdir -p /app/agent /app/logs /app/memory /app/config
  VOLUME ["/app/agent", "/app/logs", "/app/memory", "/app/config"]
  ```

- **After:** Single volume declaration
  ```dockerfile
  RUN mkdir -p /app/workplace
  VOLUME ["/app/workplace"]
  ```

### 3. Code Changes Required

The following files need path updates to use `workplace/` prefix:

#### cmd/bot/main.go
```go
// Current paths
agentDir := "agent"
memoryBasePath := "memory"
cronConfigPath := "config/cron.json"

// Should be changed to:
agentDir := "workplace/agent"
memoryBasePath := "workplace/memory"
cronConfigPath := "workplace/config/cron.json"
```

#### internal/tools/notes_tool.go
```go
// Line 106 - Current:
filePath := fmt.Sprintf("memory/notes/%s/%s.md", category, name)

// Should be:
filePath := fmt.Sprintf("workplace/memory/notes/%s/%s.md", category, name)
```

#### internal/tools/memory_tool.go
```go
// Line 67 - Current:
filePath := fmt.Sprintf("memory/chat/%s/daily-summary.md", dateStr)

// Should be:
filePath := fmt.Sprintf("workplace/memory/chat/%s/daily-summary.md", dateStr)

// Line 94 - Current:
filePath := fmt.Sprintf("memory/chat/%s/summary.md", weekFolder)

// Should be:
filePath := fmt.Sprintf("workplace/memory/chat/%s/summary.md", weekFolder)

// Line 128 - Current:
filePath := fmt.Sprintf("memory/chat/%s/summary.md", quarterFolder)

// Should be:
filePath := fmt.Sprintf("workplace/memory/chat/%s/summary.md", quarterFolder)

// Line 169 - Current:
filePath := fmt.Sprintf("memory/chat/%s/daily-summary.md", dateStr)

// Should be:
filePath := fmt.Sprintf("workplace/memory/chat/%s/daily-summary.md", dateStr)
```

#### internal/tools/topic_tool.go
```go
// Line 100 - Current:
FilePath: "memory/MEMORY.md",

// Should be:
FilePath: "workplace/memory/MEMORY.md",

// Line 250 - Current:
filePath := fmt.Sprintf("memory/topics/%s.md", name)

// Should be:
filePath := fmt.Sprintf("workplace/memory/topics/%s.md", name)

// Line 288 - Current:
filePath := fmt.Sprintf("memory/topics/%s.md", name)

// Should be:
filePath := fmt.Sprintf("workplace/memory/topics/%s.md", name)

// Line 326 - Current:
filePath := fmt.Sprintf("memory/topics/%s.md", name)

// Should be:
filePath := fmt.Sprintf("workplace/memory/topics/%s.md", name)
```

#### internal/scheduler/cron.go
```go
// Line 644 - Current:
summaryPath := fmt.Sprintf("memory/chat/%s/daily-summary.md", dateToSummarize)

// Should be:
summaryPath := fmt.Sprintf("workplace/memory/chat/%s/daily-summary.md", dateToSummarize)
```

### 4. Benefits of This Change

1. **Simpler Docker Configuration**
   - Only one volume mount instead of four
   - Easier to understand and maintain

2. **Easier Backup and Migration**
   - Single directory contains all runtime data
   - Simple to backup: `tar -czf backup.tar.gz workplace/`
   - Simple to restore: `tar -xzf backup.tar.gz`

3. **Better Organization**
   - Clear separation between source code and runtime data
   - All persistent data in one location

4. **Improved Security**
   - Minimal surface area for Docker mounts
   - Easier to set permissions on single directory

5. **Simplified Development**
   - Easier to reset/clean development environment
   - Clear understanding of what data persists

## Migration Steps

### For Existing Deployments

1. **Stop the running container:**
   ```bash
   docker-compose down
   ```

2. **Create workplace directory and migrate data:**
   ```bash
   mkdir -p workplace
   cp -r agent workplace/
   cp -r config workplace/
   cp -r memory workplace/
   cp -r logs workplace/
   ```

3. **Update code paths** (as described in section 3 above)

4. **Rebuild and start:**
   ```bash
   docker-compose build
   docker-compose up -d
   ```

### For New Deployments

1. **Clone repository**
2. **Workplace directory already exists with structure**
3. **Configure .env file**
4. **Build and run:**
   ```bash
   docker-compose up -d
   ```

## Rollback Plan

If issues occur, you can rollback by:

1. Stop container: `docker-compose down`
2. Restore old docker-compose.yml and Dockerfile
3. Copy data back to original locations:
   ```bash
   cp -r workplace/agent ./
   cp -r workplace/config ./
   cp -r workplace/memory ./
   cp -r workplace/logs ./
   ```
4. Rebuild and restart: `docker-compose up -d --build`

## Testing Checklist

After migration, verify:

- [ ] Bot starts successfully
- [ ] Agent identity files are loaded
- [ ] Memory system reads/writes correctly
- [ ] Cron jobs execute properly
- [ ] Logs are written to workplace/logs/
- [ ] Notes management works
- [ ] Topic management works
- [ ] Chat log search works
- [ ] Scheduled tasks execute
- [ ] Session reset works

## Notes

- The old `agent/`, `config/`, `memory/`, and `logs/` directories can be kept as backups
- Consider adding `workplace/` to `.gitignore` if not already present
- The `workplace/.gitkeep` file ensures the directory structure is preserved in git

