// Package hooks provides a middleware system for CM operations.
package hooks

// CMInterface defines the interface that hooks need from CM.
type CMInterface interface {
	// Add any methods that hooks might need from CM.
	// For now, we'll keep it minimal to avoid circular dependencies.
}

// HookContext provides context for hook execution.
type HookContext struct {
	OperationName string
	Parameters    map[string]interface{}
	Results       map[string]interface{}
	Error         error
	CM            CMInterface
	Metadata      map[string]interface{}
}

// Hook defines the interface for all hooks.
type Hook interface {
	Name() string
	Priority() int
	Execute(ctx *HookContext) error
}

// PreHook executes before an operation.
type PreHook interface {
	Hook
	PreExecute(ctx *HookContext) error
}

// PostHook executes after an operation.
type PostHook interface {
	Hook
	PostExecute(ctx *HookContext) error
}

// ErrorHook executes when an operation fails.
type ErrorHook interface {
	Hook
	OnError(ctx *HookContext) error
}
