//go:build integration

package logger

import (
	"testing"
)

func TestNoopLogger_Logf(t *testing.T) {
	logger := NewNoopLogger()
	// This should not panic or produce any output
	logger.Logf("test message")
}

func TestVerboseLogger_Logf(t *testing.T) {
	logger := NewVerboseLogger()
	// This should not panic
	logger.Logf("test message %s", "with args")
}

func TestVerboseLogger_ThreadSafety(t *testing.T) {
	logger := NewVerboseLogger()
	// This should not panic when called concurrently
	for i := 0; i < 100; i++ {
		go logger.Logf("concurrent message %d", i)
	}
}
