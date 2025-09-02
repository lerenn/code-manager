// Package logger provides logging functionality for the CM application.
package logger

import (
	"fmt"
	"sync"
)

//go:generate mockgen -source=logger.go -destination=mocklogger.gen.go -package=logger

// Logger interface provides logging capabilities.
type Logger interface {
	// Logf logs a formatted message.
	Logf(format string, args ...interface{})
}

// noopLogger is a logger that does nothing.
type noopLogger struct{}

// NewNoopLogger creates a new noop logger.
func NewNoopLogger() Logger {
	return &noopLogger{}
}

// Logf does nothing for noop logger.
func (n *noopLogger) Logf(_ string, _ ...interface{}) {}

// verboseLogger is a thread-safe logger that writes to stdout.
type verboseLogger struct {
	mu sync.Mutex
}

// NewVerboseLogger creates a new verbose logger.
func NewVerboseLogger() Logger {
	return &verboseLogger{}
}

// Logf writes a formatted message to stdout with thread safety.
func (v *verboseLogger) Logf(format string, args ...interface{}) {
	v.mu.Lock()
	defer v.mu.Unlock()
	fmt.Printf(format+"\n", args...)
}
