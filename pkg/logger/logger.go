// Package logger provides logging functionality for the WTM application.
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

// defaultLogger is a thread-safe logger that writes to stdout.
type defaultLogger struct {
	mu sync.Mutex
}

// NewDefaultLogger creates a new default logger.
func NewDefaultLogger() Logger {
	return &defaultLogger{}
}

// Logf writes a formatted message to stdout with thread safety.
func (d *defaultLogger) Logf(format string, args ...interface{}) {
	d.mu.Lock()
	defer d.mu.Unlock()
	fmt.Printf(format+"\n", args...)
}
