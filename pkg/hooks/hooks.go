// Package hooks provides a middleware system for CM operations.
package hooks

// HookContext provides context for hook execution.
type HookContext struct {
	OperationName string
	Parameters    map[string]interface{}
	Results       map[string]interface{}
	Error         error
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

// PostWorktreeCheckoutHook executes between worktree creation and checkout.
type PostWorktreeCheckoutHook interface {
	Hook
	OnPostWorktreeCheckout(ctx *HookContext) error
}

// PreWorktreeCreationHook executes before worktree creation for detection/configuration.
type PreWorktreeCreationHook interface {
	Hook
	OnPreWorktreeCreation(ctx *HookContext) error
}
