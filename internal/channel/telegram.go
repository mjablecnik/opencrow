package channel

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"simple-telegram-chatbot/internal/llm"
	"simple-telegram-chatbot/internal/memory"
	"simple-telegram-chatbot/internal/session"
	"simple-telegram-chatbot/pkg/utils"
)

// TelegramChannel handles Telegram bot communication
type TelegramChannel struct {
	bot                  *tgbotapi.BotAPI
	llmClient            *llm.OpenRouterClient
	sessionManager       *session.SessionManager
	memorySessionManager *memory.SessionManager
	logger               *utils.Logger
	updatesChan          tgbotapi.UpdatesChannel
	stopChan             chan struct{}
}

// NewTelegramChannel creates a new Telegram channel instance
func NewTelegramChannel(
	botToken string,
	llmClient *llm.OpenRouterClient,
	sessionManager *session.SessionManager,
	memorySessionManager *memory.SessionManager,
	logger *utils.Logger,
) (*TelegramChannel, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	logger.InfoWithComponent("TelegramChannel", "Authorized on account", "username", bot.Self.UserName)

	return &TelegramChannel{
		bot:                  bot,
		llmClient:            llmClient,
		sessionManager:       sessionManager,
		memorySessionManager: memorySessionManager,
		logger:               logger,
		stopChan:             make(chan struct{}),
	}, nil
}

// Start initializes bot connection and starts polling for messages
func (tc *TelegramChannel) Start(ctx context.Context) error {
	tc.logger.InfoWithComponent("TelegramChannel", "Starting Telegram bot polling")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tc.bot.GetUpdatesChan(u)
	tc.updatesChan = updates

	go tc.pollUpdates(ctx)

	tc.logger.InfoWithComponent("TelegramChannel", "Telegram bot started successfully")
	return nil
}

// pollUpdates processes incoming updates from Telegram
func (tc *TelegramChannel) pollUpdates(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			tc.logger.InfoWithComponent("TelegramChannel", "Context cancelled, stopping polling")
			return
		case <-tc.stopChan:
			tc.logger.InfoWithComponent("TelegramChannel", "Stop signal received, stopping polling")
			return
		case update := <-tc.updatesChan:
			if update.Message != nil {
				if err := tc.HandleMessage(update); err != nil {
					tc.logger.ErrorWithComponent("TelegramChannel", "Error handling message", "error", err)
				}
			}
		}
	}
}

// HandleMessage routes incoming messages to appropriate handlers
func (tc *TelegramChannel) HandleMessage(update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	chatID := update.Message.Chat.ID
	userMessage := update.Message.Text

	tc.logger.InfoWithComponent("TelegramChannel", "Received message", "chatID", chatID, "message", userMessage)

	// Check for /reset command
	if userMessage == "/reset" {
		tc.logger.InfoWithComponent("TelegramChannel", "Processing /reset command", "chatID", chatID)
		return tc.handleResetCommand(chatID)
	}

	// Log user message to memory session log
	if err := tc.memorySessionManager.AppendToSessionLog("User", userMessage); err != nil {
		tc.logger.WarnWithComponent("TelegramChannel", "Failed to log user message to memory", "error", err)
		// Continue processing even if logging fails
	}

	// Store user message in session
	if err := tc.sessionManager.AppendMessage(chatID, "user", userMessage); err != nil {
		tc.logger.ErrorWithComponent("TelegramChannel", "Failed to append user message", "error", err)
		return fmt.Errorf("failed to append user message: %w", err)
	}

	// Send request to LLM
	ctx := context.Background()
	response, err := tc.llmClient.SendRequest(ctx, chatID, userMessage)
	if err != nil {
		tc.logger.ErrorWithComponent("TelegramChannel", "Failed to get LLM response", "error", err)
		errorMsg := "Sorry, I encountered an error processing your message. Please try again."
		if sendErr := tc.SendMessageWithRetry(chatID, errorMsg, 3); sendErr != nil {
			tc.logger.ErrorWithComponent("TelegramChannel", "Failed to send error message", "error", sendErr)
		}
		return fmt.Errorf("failed to get LLM response: %w", err)
	}

	// Store assistant response in session
	if err := tc.sessionManager.AppendMessage(chatID, "assistant", response); err != nil {
		tc.logger.ErrorWithComponent("TelegramChannel", "Failed to append assistant message", "error", err)
		return fmt.Errorf("failed to append assistant message: %w", err)
	}

	// Log assistant response to memory session log
	if err := tc.memorySessionManager.AppendToSessionLog("Assistant", response); err != nil {
		tc.logger.WarnWithComponent("TelegramChannel", "Failed to log assistant message to memory", "error", err)
		// Continue processing even if logging fails
	}

	// Send response to user with retry logic
	if err := tc.SendMessageWithRetry(chatID, response, 3); err != nil {
		tc.logger.ErrorWithComponent("TelegramChannel", "Failed to send message after retries", "error", err)
		return fmt.Errorf("failed to send message after retries: %w", err)
	}

	return nil
}

// escapeMarkdownV2 intelligently escapes special characters for Telegram MarkdownV2
// while preserving markdown formatting syntax
func (tc *TelegramChannel) escapeMarkdownV2(text string) string {
	var result strings.Builder
	result.Grow(len(text) * 2)
	
	inCodeBlock := false
	inInlineCode := false
	
	i := 0
	for i < len(text) {
		// Check for code blocks (```...```)
		if i+2 < len(text) && text[i:i+3] == "```" {
			if !inCodeBlock {
				// Opening code block
				result.WriteString("```")
				i += 3
				inCodeBlock = true
				continue
			} else {
				// Closing code block
				result.WriteString("```")
				i += 3
				inCodeBlock = false
				continue
			}
		}
		
		// Inside code block, write everything as-is
		if inCodeBlock {
			result.WriteByte(text[i])
			i++
			continue
		}
		
		// Check for inline code (`...`)
		if text[i] == '`' && !inInlineCode {
			// Opening inline code
			result.WriteByte('`')
			i++
			inInlineCode = true
			continue
		} else if text[i] == '`' && inInlineCode {
			// Closing inline code
			result.WriteByte('`')
			i++
			inInlineCode = false
			continue
		}
		
		// Inside inline code, write everything as-is
		if inInlineCode {
			result.WriteByte(text[i])
			i++
			continue
		}
		
		// Check for links ([text](url))
		if text[i] == '[' {
			// Find the closing ] and check for (url)
			closeBracket := strings.IndexByte(text[i+1:], ']')
			if closeBracket != -1 && closeBracket > 0 { // Ensure there's content
				closeBracket += i + 1
				if closeBracket+1 < len(text) && text[closeBracket+1] == '(' {
					closeParen := strings.IndexByte(text[closeBracket+2:], ')')
					if closeParen != -1 && closeParen > 0 { // Ensure there's content
						closeParen += closeBracket + 2 + 1
						// Write the entire link as-is
						result.WriteString(text[i:closeParen])
						i = closeParen
						continue
					}
				}
			}
		}
		
		// Check for bold (**text**)
		if i+1 < len(text) && text[i:i+2] == "**" {
			// Find closing **
			end := strings.Index(text[i+2:], "**")
			if end != -1 && end > 0 { // Ensure there's content between
				end += i + 2 + 2
				// Write the bold text
				result.WriteString(text[i:end])
				i = end
				continue
			}
		}
		
		// Check for italic (*text*)
		if text[i] == '*' {
			// Find closing *
			end := strings.IndexByte(text[i+1:], '*')
			if end != -1 && end > 0 { // Ensure there's content between
				end += i + 1 + 1
				result.WriteString(text[i:end])
				i = end
				continue
			}
		}
		
		// Check for underline (__text__)
		if i+1 < len(text) && text[i:i+2] == "__" {
			end := strings.Index(text[i+2:], "__")
			if end != -1 && end > 0 { // Ensure there's content between
				end += i + 2 + 2
				result.WriteString(text[i:end])
				i = end
				continue
			}
		}
		
		// Escape special characters
		char := text[i]
		switch char {
		case '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!':
			result.WriteByte('\\')
			result.WriteByte(char)
		default:
			result.WriteByte(char)
		}
		i++
	}
	
	return result.String()
}

// SendMessage sends a basic message to a user without formatting
func (tc *TelegramChannel) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	// Disable markdown parsing - send as plain text
	msg.ParseMode = ""
	_, err := tc.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}


// SendMessageWithRetry implements exponential backoff retry logic (3 attempts)
func (tc *TelegramChannel) SendMessageWithRetry(chatID int64, text string, maxRetries int) error {
	delays := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	for attempt := 0; attempt < maxRetries; attempt++ {
		tc.logger.DebugWithComponent("TelegramChannel", "Attempting to send message", "attempt", attempt+1, "chatID", chatID)

		err := tc.SendMessage(chatID, text)
		if err == nil {
			tc.logger.InfoWithComponent("TelegramChannel", "Message sent successfully", "chatID", chatID, "attempt", attempt+1)
			return nil
		}

		tc.logger.WarnWithComponent("TelegramChannel", "Failed to send message", "attempt", attempt+1, "error", err)

		// If this was the last attempt, return the error
		if attempt == maxRetries-1 {
			return fmt.Errorf("failed to send message after %d attempts: %w", maxRetries, err)
		}

		// Wait before retrying with exponential backoff
		delay := delays[attempt]
		tc.logger.DebugWithComponent("TelegramChannel", "Waiting before retry", "delay", delay)
		time.Sleep(delay)
	}

	return fmt.Errorf("failed to send message after %d attempts", maxRetries)
}

// Stop gracefully shuts down the bot
func (tc *TelegramChannel) Stop() error {
	tc.logger.InfoWithComponent("TelegramChannel", "Stopping Telegram bot")
	
	close(tc.stopChan)
	tc.bot.StopReceivingUpdates()
	
	tc.logger.InfoWithComponent("TelegramChannel", "Telegram bot stopped successfully")
	return nil
}

// handleResetCommand handles the /reset command to clear conversation history
func (tc *TelegramChannel) handleResetCommand(chatID int64) error {
	tc.logger.InfoWithComponent("TelegramChannel", "Executing session reset", "chatID", chatID)

	// Step 1: Clear in-memory session
	if err := tc.sessionManager.ClearSession(chatID); err != nil {
		tc.logger.WarnWithComponent("TelegramChannel", "Failed to clear in-memory session (may not exist)", "chatID", chatID, "error", err)
		// Continue even if session doesn't exist
	}

	// Step 2: Perform manual session reset (archives session-latest.log, NO summarization)
	if err := tc.memorySessionManager.PerformManualSessionReset(); err != nil {
		tc.logger.ErrorWithComponent("TelegramChannel", "Failed to perform manual session reset", "chatID", chatID, "error", err)
		// Send error message to user
		errorMsg := "Failed to reset session. Please try again."
		if sendErr := tc.SendMessageWithRetry(chatID, errorMsg, 3); sendErr != nil {
			tc.logger.ErrorWithComponent("TelegramChannel", "Failed to send error message", "error", sendErr)
		}
		return fmt.Errorf("failed to perform manual session reset: %w", err)
	}

	// Step 3: Send confirmation to user
	confirmMsg := "History has been reset. How can I help you today?"
	if err := tc.SendMessageWithRetry(chatID, confirmMsg, 3); err != nil {
		tc.logger.ErrorWithComponent("TelegramChannel", "Failed to send reset confirmation", "error", err)
		return fmt.Errorf("failed to send reset confirmation: %w", err)
	}

	tc.logger.InfoWithComponent("TelegramChannel", "Session reset completed successfully", "chatID", chatID)
	return nil
}
