package utils

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name      string
		levelStr  string
		wantLevel LogLevel
	}{
		{"debug level", "debug", DebugLevel},
		{"info level", "info", InfoLevel},
		{"warn level", "warn", WarnLevel},
		{"warning level", "warning", WarnLevel},
		{"error level", "error", ErrorLevel},
		{"default to info", "invalid", InfoLevel},
		{"empty defaults to info", "", InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.levelStr)
			if logger.level != tt.wantLevel {
				t.Errorf("NewLogger(%q) level = %v, want %v", tt.levelStr, logger.level, tt.wantLevel)
			}
		})
	}
}

func TestLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name       string
		logLevel   string
		logFunc    func(*Logger, *bytes.Buffer)
		wantOutput bool
	}{
		{
			name:     "debug logs at debug level",
			logLevel: "debug",
			logFunc: func(l *Logger, buf *bytes.Buffer) {
				l.logger = log.New(buf, "", 0)
				l.Debug("test message")
			},
			wantOutput: true,
		},
		{
			name:     "debug does not log at info level",
			logLevel: "info",
			logFunc: func(l *Logger, buf *bytes.Buffer) {
				l.logger = log.New(buf, "", 0)
				l.Debug("test message")
			},
			wantOutput: false,
		},
		{
			name:     "info logs at info level",
			logLevel: "info",
			logFunc: func(l *Logger, buf *bytes.Buffer) {
				l.logger = log.New(buf, "", 0)
				l.Info("test message")
			},
			wantOutput: true,
		},
		{
			name:     "info does not log at warn level",
			logLevel: "warn",
			logFunc: func(l *Logger, buf *bytes.Buffer) {
				l.logger = log.New(buf, "", 0)
				l.Info("test message")
			},
			wantOutput: false,
		},
		{
			name:     "warn logs at warn level",
			logLevel: "warn",
			logFunc: func(l *Logger, buf *bytes.Buffer) {
				l.logger = log.New(buf, "", 0)
				l.Warn("test message")
			},
			wantOutput: true,
		},
		{
			name:     "error logs at all levels",
			logLevel: "debug",
			logFunc: func(l *Logger, buf *bytes.Buffer) {
				l.logger = log.New(buf, "", 0)
				l.Error("test message")
			},
			wantOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.logLevel)
			buf := &bytes.Buffer{}
			tt.logFunc(logger, buf)

			output := buf.String()
			hasOutput := len(output) > 0

			if hasOutput != tt.wantOutput {
				t.Errorf("Log output presence = %v, want %v. Output: %q", hasOutput, tt.wantOutput, output)
			}
		})
	}
}

func TestLogger_MessageFormat(t *testing.T) {
	logger := NewLogger("debug")
	buf := &bytes.Buffer{}
	logger.logger = log.New(buf, "", 0)

	logger.Info("test message", "key1", "value1", "key2", "value2")
	output := buf.String()

	// Check for required components
	if !strings.Contains(output, "[INFO]") {
		t.Error("Output missing [INFO] level indicator")
	}
	if !strings.Contains(output, "test message") {
		t.Error("Output missing message text")
	}
	if !strings.Contains(output, "key1=value1") {
		t.Error("Output missing key-value pair key1=value1")
	}
	if !strings.Contains(output, "key2=value2") {
		t.Error("Output missing key-value pair key2=value2")
	}

	// Check timestamp format (YYYY-MM-DD HH:MM:SS)
	if !strings.Contains(output, "[20") {
		t.Error("Output missing timestamp")
	}
}

func TestLogger_ComponentLogging(t *testing.T) {
	logger := NewLogger("info")
	buf := &bytes.Buffer{}
	logger.logger = log.New(buf, "", 0)

	logger.InfoWithComponent("TestComponent", "test message")
	output := buf.String()

	if !strings.Contains(output, "[TestComponent]") {
		t.Errorf("Output missing component name. Output: %q", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("Output missing message. Output: %q", output)
	}
}

func TestLogger_ErrorWithDetails(t *testing.T) {
	logger := NewLogger("error")
	buf := &bytes.Buffer{}
	logger.logger = log.New(buf, "", 0)

	testErr := &testError{msg: "test error"}
	logger.ErrorWithDetails("TestComponent", "operation failed", testErr, "context1", "value1")
	output := buf.String()

	if !strings.Contains(output, "[ERROR]") {
		t.Error("Output missing [ERROR] level indicator")
	}
	if !strings.Contains(output, "[TestComponent]") {
		t.Error("Output missing component name")
	}
	if !strings.Contains(output, "operation failed") {
		t.Error("Output missing message")
	}
	if !strings.Contains(output, "Details:") {
		t.Error("Output missing Details section")
	}
	if !strings.Contains(output, "test error") {
		t.Error("Output missing error details")
	}
	if !strings.Contains(output, "Context:") {
		t.Error("Output missing Context section")
	}
	if !strings.Contains(output, "context1=value1") {
		t.Error("Output missing context key-value pair")
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
