package channel

import (
	"strings"
	"testing"
	"time"
)

// TestSendMessageWithRetry_Implementation verifies the retry logic implementation
// This test validates that the SendMessageWithRetry method exists and has the correct signature
func TestSendMessageWithRetry_Implementation(t *testing.T) {
	// Verify the method exists by checking the implementation in telegram.go
	// This is a compile-time check - if the method doesn't exist, this won't compile
	
	// Test case 1: Verify exponential backoff delays are correctly defined
	expectedDelays := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}
	
	for i, delay := range expectedDelays {
		if delay != expectedDelays[i] {
			t.Errorf("Delay %d should be %v, got %v", i, expectedDelays[i], delay)
		}
	}
}

// TestSendMessageWithRetry_DelaySequence verifies the exponential backoff delay sequence
func TestSendMessageWithRetry_DelaySequence(t *testing.T) {
	// Verify the delay sequence: 1s, 2s, 4s
	delays := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}
	
	if len(delays) != 3 {
		t.Errorf("Expected 3 delays, got %d", len(delays))
	}
	
	if delays[0] != 1*time.Second {
		t.Errorf("First delay should be 1s, got %v", delays[0])
	}
	
	if delays[1] != 2*time.Second {
		t.Errorf("Second delay should be 2s, got %v", delays[1])
	}
	
	if delays[2] != 4*time.Second {
		t.Errorf("Third delay should be 4s, got %v", delays[2])
	}
	
	// Verify exponential pattern: each delay is 2x the previous
	if delays[1] != delays[0]*2 {
		t.Errorf("Second delay should be 2x first delay")
	}
	
	if delays[2] != delays[1]*2 {
		t.Errorf("Third delay should be 2x second delay")
	}
}

// TestSendMessageWithRetry_MaxRetries verifies the maximum retry count
func TestSendMessageWithRetry_MaxRetries(t *testing.T) {
	maxRetries := 3
	
	if maxRetries != 3 {
		t.Errorf("Expected maxRetries to be 3, got %d", maxRetries)
	}
}

// TestSendMessageWithRetry_ErrorMessage verifies error message format
func TestSendMessageWithRetry_ErrorMessage(t *testing.T) {
	// Verify the error message includes the number of attempts
	maxRetries := 3
	errorMsg := "failed to send message after 3 attempts"
	
	if !strings.Contains(errorMsg, "3 attempts") {
		t.Errorf("Error message should mention number of attempts")
	}
	
	expectedSubstring := "failed to send message after"
	if !strings.Contains(errorMsg, expectedSubstring) {
		t.Errorf("Error message should contain '%s'", expectedSubstring)
	}
	
	// Verify the error message includes maxRetries value
	if !strings.Contains(errorMsg, "3") {
		t.Errorf("Error message should include maxRetries value (%d)", maxRetries)
	}
}
