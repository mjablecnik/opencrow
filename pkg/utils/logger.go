package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// Logger provides structured logging with configurable log levels
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new logger with the specified log level
func NewLogger(levelStr string) *Logger {
	level := parseLogLevel(levelStr)
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
	}
}

// parseLogLevel converts a string log level to LogLevel
func parseLogLevel(levelStr string) LogLevel {
	switch strings.ToLower(levelStr) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// formatMessage creates a formatted log message with timestamp, level, and component
func (l *Logger) formatMessage(level string, component string, message string, keyvals ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	
	// Build the base message
	msg := fmt.Sprintf("[%s] [%s]", timestamp, level)
	
	if component != "" {
		msg += fmt.Sprintf(" [%s]", component)
	}
	
	msg += " " + message
	
	// Add key-value pairs if provided
	if len(keyvals) > 0 {
		for i := 0; i < len(keyvals); i += 2 {
			if i+1 < len(keyvals) {
				msg += fmt.Sprintf(" %v=%v", keyvals[i], keyvals[i+1])
			}
		}
	}
	
	return msg
}

// Debug logs a debug-level message
func (l *Logger) Debug(message string, keyvals ...interface{}) {
	if l.level <= DebugLevel {
		l.logger.Println(l.formatMessage("DEBUG", "", message, keyvals...))
	}
}

// DebugWithComponent logs a debug-level message with component name
func (l *Logger) DebugWithComponent(component string, message string, keyvals ...interface{}) {
	if l.level <= DebugLevel {
		l.logger.Println(l.formatMessage("DEBUG", component, message, keyvals...))
	}
}

// Info logs an info-level message
func (l *Logger) Info(message string, keyvals ...interface{}) {
	if l.level <= InfoLevel {
		l.logger.Println(l.formatMessage("INFO", "", message, keyvals...))
	}
}

// InfoWithComponent logs an info-level message with component name
func (l *Logger) InfoWithComponent(component string, message string, keyvals ...interface{}) {
	if l.level <= InfoLevel {
		l.logger.Println(l.formatMessage("INFO", component, message, keyvals...))
	}
}

// Warn logs a warning-level message
func (l *Logger) Warn(message string, keyvals ...interface{}) {
	if l.level <= WarnLevel {
		l.logger.Println(l.formatMessage("WARN", "", message, keyvals...))
	}
}

// WarnWithComponent logs a warning-level message with component name
func (l *Logger) WarnWithComponent(component string, message string, keyvals ...interface{}) {
	if l.level <= WarnLevel {
		l.logger.Println(l.formatMessage("WARN", component, message, keyvals...))
	}
}

// Error logs an error-level message
func (l *Logger) Error(message string, keyvals ...interface{}) {
	if l.level <= ErrorLevel {
		l.logger.Println(l.formatMessage("ERROR", "", message, keyvals...))
	}
}

// ErrorWithComponent logs an error-level message with component name
func (l *Logger) ErrorWithComponent(component string, message string, keyvals ...interface{}) {
	if l.level <= ErrorLevel {
		l.logger.Println(l.formatMessage("ERROR", component, message, keyvals...))
	}
}

// ErrorWithDetails logs an error with additional details and context
func (l *Logger) ErrorWithDetails(component string, message string, err error, context ...interface{}) {
	if l.level <= ErrorLevel {
		msg := l.formatMessage("ERROR", component, message)
		msg += fmt.Sprintf("\nDetails: %v", err)
		
		if len(context) > 0 {
			msg += "\nContext:"
			for i := 0; i < len(context); i += 2 {
				if i+1 < len(context) {
					msg += fmt.Sprintf(" %v=%v", context[i], context[i+1])
				}
			}
		}
		
		l.logger.Println(msg)
	}
}
