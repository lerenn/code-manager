package hooks

import (
	"github.com/lerenn/code-manager/pkg/logger"
)

// LoggingHook provides logging functionality for all operations.
type LoggingHook struct {
	logger logger.Logger
	level  string
}

// NewLoggingHook creates a new LoggingHook instance.
func NewLoggingHook(logger logger.Logger, level string) *LoggingHook {
	return &LoggingHook{
		logger: logger,
		level:  level,
	}
}

// Name returns the hook name.
func (h *LoggingHook) Name() string {
	return "logging"
}

// Priority returns the hook priority (lower numbers execute first).
func (h *LoggingHook) Priority() int {
	return 100
}

// Execute is a no-op for LoggingHook as it implements specific methods.
func (h *LoggingHook) Execute(_ *HookContext) error {
	return nil
}

// PreExecute logs the start of an operation.
func (h *LoggingHook) PreExecute(ctx *HookContext) error {
	h.logger.Logf("Starting operation: %s with params: %v", ctx.OperationName, ctx.Parameters)
	return nil
}

// PostExecute logs the completion of an operation.
func (h *LoggingHook) PostExecute(ctx *HookContext) error {
	if ctx.Error != nil {
		h.logger.Logf("Operation failed: %s, error: %v", ctx.OperationName, ctx.Error)
	} else {
		h.logger.Logf("Operation completed: %s with results: %v", ctx.OperationName, ctx.Results)
	}
	return nil
}

// OnError logs when an operation fails.
func (h *LoggingHook) OnError(ctx *HookContext) error {
	h.logger.Logf("Operation error: %s, error: %v", ctx.OperationName, ctx.Error)
	return nil
}
