package ide_opening

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/hooks"
)

// IDEOpeningHook provides IDE opening functionality as a post-hook.
type IDEOpeningHook struct{}

// NewIDEOpeningHook creates a new IDEOpeningHook instance and registers it for appropriate operations.
func NewIDEOpeningHook() *IDEOpeningHook {
	return &IDEOpeningHook{}
}

// RegisterForOperations registers this hook for the operations that create worktrees.
func (h *IDEOpeningHook) RegisterForOperations(cmInstance interface {
	RegisterHook(operation string, hook hooks.Hook) error
}) error {
	// Register as post-hook for operations that create worktrees
	if err := cmInstance.RegisterHook(consts.CreateWorkTree, h); err != nil {
		return err
	}

	if err := cmInstance.RegisterHook(consts.LoadWorktree, h); err != nil {
		return err
	}

	// Register as post-hook for operations that open worktrees
	if err := cmInstance.RegisterHook(consts.OpenWorktree, h); err != nil {
		return err
	}

	return nil
}

// Name returns the hook name.
func (h *IDEOpeningHook) Name() string {
	return "ide-opening"
}

// Priority returns the hook priority (lower numbers execute first).
func (h *IDEOpeningHook) Priority() int {
	return 150
}

// Execute is a no-op for IDEOpeningHook as it implements specific methods.
func (h *IDEOpeningHook) Execute(_ *hooks.HookContext) error {
	return nil
}

// PreExecute is a no-op for IDEOpeningHook.
func (h *IDEOpeningHook) PreExecute(_ *hooks.HookContext) error {
	return nil
}

// PostExecute validates IDE opening parameters and opens the IDE after successful operations.
func (h *IDEOpeningHook) PostExecute(ctx *hooks.HookContext) error {
	// Only proceed if operation was successful
	if ctx.Error != nil {
		// Operation failed, don't process IDE opening
		return nil
	}

	// Check if IDE name is provided in parameters
	ideName, hasIDEName := ctx.Parameters["ideName"]
	if !hasIDEName {
		return nil
	}

	ideNameStr, ok := ideName.(string)
	if !ok || ideNameStr == "" {
		return nil
	}

	// Get worktree path from parameters
	worktreePath := h.extractWorktreePath(ctx.Parameters)
	if worktreePath == "" {
		return fmt.Errorf("cannot open IDE: worktree path is empty")
	}

	// Store the IDE opening information in the context for the CM to handle
	ctx.Results["ideName"] = ideNameStr
	ctx.Results["worktreePath"] = worktreePath
	ctx.Results["shouldOpenIDE"] = true

	return nil
}

// extractWorktreePath extracts the worktree path from parameters.
func (h *IDEOpeningHook) extractWorktreePath(params map[string]interface{}) string {
	if branch, hasBranch := params["branch"]; hasBranch {
		if branchStr, ok := branch.(string); ok && branchStr != "" {
			return branchStr
		}
	}
	if worktreeName, hasWorktreeName := params["worktreeName"]; hasWorktreeName {
		if worktreeNameStr, ok := worktreeName.(string); ok && worktreeNameStr != "" {
			return worktreeNameStr
		}
	}
	return ""
}

// OnError is a no-op for IDEOpeningHook.
func (h *IDEOpeningHook) OnError(_ *hooks.HookContext) error {
	return nil
}
