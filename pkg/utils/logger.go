package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	level         LogLevel
	logger        *log.Logger
	logFile       *os.File
	logToFile     bool
	logDir        string
	currentDate   string
	maxAgeDays    int
	mu            sync.Mutex
}

// NewLogger creates a new logger with the specified log level
func NewLogger(levelStr string) *Logger {
	level := parseLogLevel(levelStr)
	return &Logger{
		level:     level,
		logger:    log.New(os.Stdout, "", 0),
		logToFile: false,
	}
}

// NewLoggerWithFile creates a new logger that writes to both stdout and a file
// with automatic daily rotation and cleanup of old log files
func NewLoggerWithFile(levelStr string, logDir string, maxAgeDays int) (*Logger, error) {
	level := parseLogLevel(levelStr)
	
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	
	l := &Logger{
		level:      level,
		logToFile:  true,
		logDir:     logDir,
		maxAgeDays: maxAgeDays,
	}
	
	// Open initial log file
	if err := l.rotateLogFile(); err != nil {
		return nil, err
	}
	
	// Start background goroutine for daily rotation and cleanup
	go l.dailyRotationWorker()
	
	return l, nil
}

// rotateLogFile closes the current log file and opens a new one for today
func (l *Logger) rotateLogFile() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Get current date
	currentDate := time.Now().Format("2006-01-02")
	
	// If already on the correct date, do nothing
	if l.currentDate == currentDate && l.logFile != nil {
		return nil
	}
	
	// Close old log file if open
	if l.logFile != nil {
		l.logFile.Close()
	}
	
	// Open new log file
	logFileName := filepath.Join(l.logDir, fmt.Sprintf("bot-%s.log", currentDate))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	
	l.logFile = logFile
	l.currentDate = currentDate
	
	// Create multi-writer for both stdout and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	l.logger = log.New(multiWriter, "", 0)
	
	return nil
}

// dailyRotationWorker runs in background and rotates log file at midnight
func (l *Logger) dailyRotationWorker() {
	for {
		// Calculate time until next midnight
		now := time.Now()
		tomorrow := now.Add(24 * time.Hour)
		nextMidnight := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, now.Location())
		timeUntilMidnight := nextMidnight.Sub(now)
		
		// Wait until midnight
		time.Sleep(timeUntilMidnight)
		
		// Rotate log file
		if err := l.rotateLogFile(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to rotate log file: %v\n", err)
		}
		
		// Clean up old log files
		if err := l.cleanupOldLogs(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to cleanup old logs: %v\n", err)
		}
	}
}

// cleanupOldLogs removes log files older than maxAgeDays
func (l *Logger) cleanupOldLogs() error {
	if l.maxAgeDays <= 0 {
		return nil // Cleanup disabled
	}
	
	entries, err := os.ReadDir(l.logDir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}
	
	cutoffDate := time.Now().AddDate(0, 0, -l.maxAgeDays)
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		// Check if it's a bot log file
		if !strings.HasPrefix(entry.Name(), "bot-") || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}
		
		// Get file info
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		// Check if file is older than cutoff date
		if info.ModTime().Before(cutoffDate) {
			logPath := filepath.Join(l.logDir, entry.Name())
			if err := os.Remove(logPath); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to remove old log file %s: %v\n", entry.Name(), err)
			} else {
				fmt.Fprintf(os.Stdout, "Removed old log file: %s\n", entry.Name())
			}
		}
	}
	
	return nil
}

// Close closes the log file if it's open
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.logToFile && l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
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
