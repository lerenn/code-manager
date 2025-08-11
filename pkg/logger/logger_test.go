//go:build integration

package logger

import (
	"testing"
)

func TestNoopLogger_Logf(t *testing.T) {
	logger := NewNoopLogger()

	// This should not panic or produce any output
	logger.Logf("test message")
	logger.Logf("test message with args: %s", "value")
}

func TestDefaultLogger_Logf(t *testing.T) {
	logger := NewDefaultLogger()

	// These should write to stdout
	logger.Logf("test message")
	logger.Logf("test message with args: %s", "value")
}

func TestDefaultLogger_ThreadSafety(t *testing.T) {
	logger := NewDefaultLogger()

	// Test concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Logf("concurrent message from goroutine %d", id)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
