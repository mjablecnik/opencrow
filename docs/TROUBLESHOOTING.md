# Troubleshooting Guide

## Bot doesn't start

**Problem**: Bot exits immediately after starting

### Solutions

1. Check that all required environment variables are set:
   ```bash
   echo $TELEGRAM_BOT_TOKEN
   echo $OPENROUTER_API_KEY
   ```

2. Check logs for specific error messages:
   ```bash
   # Docker
   docker-compose logs bot
   
   # Local
   cat workplace/logs/bot.log
   ```

3. Verify Docker volume mounts are correct:
   ```bash
   docker-compose config
   ```

## Bot doesn't respond to messages

**Problem**: Bot is running but doesn't reply to Telegram messages

### Solutions

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

## OpenRouter API errors

**Problem**: Bot receives messages but fails to generate responses

### Solutions

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

## Message send failures

**Problem**: Bot generates responses but fails to send them to Telegram

### Solutions

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

## Docker volume issues

**Problem**: Identity files or logs not persisting

### Solutions

1. Verify volume mounts in docker-compose.yml:
   ```yaml
   volumes:
     - ./workplace:/app/workplace
   ```

2. Check file permissions:
   ```bash
   ls -la workplace/
   # Files should be readable by the container user (UID 1000)
   ```

3. Recreate volumes:
   ```bash
   docker-compose down -v
   docker-compose up -d
   ```

## High memory usage

**Problem**: Bot consumes excessive memory over time

### Solutions

1. Sessions are stored in-memory and grow with conversation length

2. Restart the bot periodically to clear sessions:
   ```bash
   docker-compose restart bot
   ```

3. Monitor memory usage:
   ```bash
   docker stats simple-telegram-chatbot
   ```

4. Adjust token threshold to trigger summarization earlier:
   ```bash
   # In .env file
   MEMORY_TOKEN_THRESHOLD=30000  # Lower threshold
   ```

## Shell command timeouts

**Problem**: Shell commands fail with timeout errors

### Solutions

1. Increase the timeout value:
   ```bash
   # In .env file
   SHELL_TIMEOUT=60s  # Increase to 60 seconds
   ```

2. Avoid long-running commands:
   - Shell commands should complete quickly
   - Use background processes for long operations

3. Check command execution in logs:
   ```bash
   docker-compose logs bot | grep -i "shell\|timeout"
   ```

## Memory system issues

**Problem**: Summaries not being generated or topics not being extracted

### Solutions

1. Check daily maintenance logs:
   ```bash
   docker-compose logs bot | grep -i "maintenance\|summary"
   ```

2. Verify maintenance schedule:
   ```bash
   # In .env file
   DAILY_MAINTENANCE_TIME=0 4 * * *
   ```

3. Manually trigger summarization by resetting the session:
   ```
   "Reset the session"
   ```

4. Check for errors in memory operations:
   ```bash
   docker-compose logs bot | grep -i "memory\|topic"
   ```

## Reminder issues

**Problem**: Reminders not being sent or cron jobs not executing

### Solutions

1. Check cron scheduler logs:
   ```bash
   docker-compose logs bot | grep -i "cron\|reminder"
   ```

2. Verify cron configuration:
   ```bash
   cat workplace/config/cron.json
   ```

3. Check cron execution history:
   ```bash
   cat workplace/config/cron_history.json
   ```

4. Test reminder creation:
   ```
   "Remind me in 1 minute to test reminders"
   ```

## Debugging Tips

### Enable Debug Logging

```bash
# In .env file
LOG_LEVEL=debug
```

Restart the bot to apply changes:
```bash
docker-compose restart bot
```

### Check Container Status

```bash
# View running containers
docker-compose ps

# View recent logs
docker-compose logs --tail=100 bot

# Follow logs in real-time
docker-compose logs -f bot
```

### Inspect Running Container

```bash
# Access container shell
docker exec -it simple-telegram-chatbot sh

# Check files inside container
ls -la /app/workplace/
```

### Test Components Individually

```bash
# Run specific tests
go test -v ./internal/channel/
go test -v ./internal/llm/
go test -v ./internal/memory/

# Run all tests
go test ./...
```

### Verify Environment Variables

```bash
# Inside container
docker exec simple-telegram-chatbot env | grep -E "TELEGRAM|OPENROUTER|MODEL"

# From docker-compose
docker-compose config
```

## Common Error Messages

### "Failed to load identity files"

**Cause**: Identity files are missing or corrupted

**Solution**: 
1. Check if files exist in `workplace/agent/`
2. Recreate default files by removing the directory and restarting:
   ```bash
   rm -rf workplace/agent/
   docker-compose restart bot
   ```

### "OpenRouter API error: 401 Unauthorized"

**Cause**: Invalid or expired API key

**Solution**: 
1. Verify your API key at [OpenRouter Dashboard](https://openrouter.ai/)
2. Update `.env` file with correct key
3. Restart the bot

### "Telegram API error: 404 Not Found"

**Cause**: Invalid bot token

**Solution**: 
1. Verify your bot token with @BotFather
2. Update `.env` file with correct token
3. Restart the bot

### "Memory token threshold exceeded"

**Cause**: Conversation is too long and needs summarization

**Solution**: This is normal behavior. The bot will automatically summarize and continue. No action needed.

## Getting Help

If you continue to experience issues:

1. Check the logs for detailed error messages
2. Review the configuration files
3. Ensure all prerequisites are met
4. Open an issue on the project repository with:
   - Error messages from logs
   - Steps to reproduce the issue
   - Environment details (OS, Docker version, etc.)
