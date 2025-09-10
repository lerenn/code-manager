# Feature 020: Implement Git-Crypt Hook for Worktree Operations

## Overview

Implement a git-crypt hook that automatically handles encrypted repositories during worktree creation and loading operations. This hook will detect when a repository uses git-crypt, locate the encryption key, and ensure proper decryption of files in newly created worktrees.

## Background

Currently, when creating worktrees from repositories that use git-crypt for file encryption, the operation fails with the following error:

```
git-crypt: Error: Unable to open key file - have you unlocked/initialized this repository yet?
error: external filter '"/opt/homebrew/bin/git-crypt" smudge' failed 1
error: external filter '"/opt/homebrew/bin/git-crypt" smudge' failed
fatal: host_vars/home_assistant.yaml: smudge filter git-crypt failed
```

This occurs because git-crypt's smudge filter cannot find the encryption key during the worktree checkout process. The worktree's `.git` directory structure doesn't include the `git-crypt` directory by default, causing the decryption to fail.

## Requirements

### Functional Requirements

1. **Git-Crypt Detection**
   - Automatically detect if a repository uses git-crypt by checking for `.gitattributes` entries with `filter=git-crypt`
   - Only enable git-crypt handling when needed (no overhead for non-encrypted repositories)

2. **Key Location Strategy**
   - Check for git-crypt key in the default branch at `$default_branch_path/.git/git-crypt/keys/default`
   - If key not found in default branch, prompt user for key file path
   - Never store key paths in configuration files for security

3. **Worktree Creation Support**
   - Handle git-crypt during `CreateWorkTree` operations
   - Handle git-crypt during `LoadWorktree` operations
   - Ensure encrypted files are properly decrypted in new worktrees

4. **Error Handling**
   - Fail the entire worktree creation operation if git-crypt unlock fails
   - Provide clear error messages to users
   - Handle cases where git-crypt is not installed

5. **User Interaction**
   - Use simple text prompt when asking for key file path
   - Allow user to provide absolute or relative paths to key files
   - Validate that provided key files exist and are readable

### Non-Functional Requirements

1. **Security**
   - Never store encryption keys or key paths in configuration files
   - Use keys only from their original locations
   - Ensure proper file permissions on copied git-crypt directories

2. **Performance**
   - Minimal overhead when git-crypt is not used
   - Efficient key detection and copying operations
   - Avoid unnecessary file system operations

3. **Reliability**
   - Graceful handling of missing git-crypt installation
   - Proper cleanup on operation failures
   - Consistent behavior across different repository structures

## Technical Specification

### Git-Crypt + Worktree Solution

Based on testing, the solution involves:

1. **Pre-hook execution**: Run before worktree creation to prepare git-crypt environment
2. **Key preparation**: Ensure git-crypt key is available in the worktree's git directory
3. **Worktree creation with --no-checkout**: Prevent smudge filter from running during creation
4. **Post-creation checkout**: Complete the checkout after git-crypt is properly configured

### Hook Implementation

#### Git-Crypt Hook Structure
```go
// pkg/hooks/gitcrypt/gitcrypt.go

package gitcrypt

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    
    "github.com/lerenn/code-manager/pkg/cm/consts"
    "github.com/lerenn/code-manager/pkg/fs"
    "github.com/lerenn/code-manager/pkg/git"
    "github.com/lerenn/code-manager/pkg/hooks"
    "github.com/lerenn/code-manager/pkg/logger"
    "github.com/lerenn/code-manager/pkg/prompt"
)

// GitCryptHook provides git-crypt functionality as a pre-hook.
type GitCryptHook struct {
    fs     fs.FSInterface
    git    git.GitInterface
    prompt prompt.PromptInterface
    logger logger.LoggerInterface
}

// NewGitCryptHook creates a new GitCryptHook instance.
func NewGitCryptHook() *GitCryptHook {
    return &GitCryptHook{
        fs:     fs.NewFS(),
        git:    git.NewGit(),
        prompt: prompt.NewPrompt(),
        logger: logger.NewNoopLogger(),
    }
}

// RegisterForOperations registers this hook for worktree operations.
func (h *GitCryptHook) RegisterForOperations(registerHook func(operation string, hook hooks.PreHook) error) error {
    // Register as pre-hook for operations that create worktrees
    if err := registerHook(consts.CreateWorkTree, h); err != nil {
        return err
    }
    
    if err := registerHook(consts.LoadWorktree, h); err != nil {
        return err
    }
    
    return nil
}

// Name returns the hook name.
func (h *GitCryptHook) Name() string {
    return "git-crypt"
}

// Priority returns the hook priority (lower numbers execute first).
func (h *GitCryptHook) Priority() int {
    return 50 // Execute early to prepare git-crypt before worktree creation
}

// Execute is a no-op for GitCryptHook as it implements specific methods.
func (h *GitCryptHook) Execute(_ *hooks.HookContext) error {
    return nil
}

// PreExecute handles git-crypt preparation before worktree operations.
func (h *GitCryptHook) PreExecute(ctx *hooks.HookContext) error {
    // Get repository path from parameters
    repoPath, err := h.getRepositoryPath(ctx)
    if err != nil {
        return err
    }
    
    // Check if repository uses git-crypt
    usesGitCrypt, err := h.detectGitCryptUsage(repoPath)
    if err != nil {
        return fmt.Errorf("failed to detect git-crypt usage: %w", err)
    }
    
    if !usesGitCrypt {
        // No git-crypt usage detected, nothing to do
        return nil
    }
    
    // Get branch information
    branch, err := h.getBranchFromContext(ctx)
    if err != nil {
        return err
    }
    
    // Prepare git-crypt for worktree creation
    return h.prepareGitCryptForWorktree(repoPath, branch, ctx)
}

// PostExecute is a no-op for GitCryptHook.
func (h *GitCryptHook) PostExecute(_ *hooks.HookContext) error {
    return nil
}

// OnError is a no-op for GitCryptHook.
func (h *GitCryptHook) OnError(_ *hooks.HookContext) error {
    return nil
}
```

#### Git-Crypt Detection
```go
// detectGitCryptUsage checks if the repository uses git-crypt.
func (h *GitCryptHook) detectGitCryptUsage(repoPath string) (bool, error) {
    gitattributesPath := filepath.Join(repoPath, ".gitattributes")
    
    // Check if .gitattributes exists
    exists, err := h.fs.Exists(gitattributesPath)
    if err != nil {
        return false, err
    }
    
    if !exists {
        return false, nil
    }
    
    // Read .gitattributes and check for git-crypt filter
    content, err := h.fs.ReadFile(gitattributesPath)
    if err != nil {
        return false, err
    }
    
    return strings.Contains(string(content), "filter=git-crypt"), nil
}
```

#### Key Location and Preparation
```go
// prepareGitCryptForWorktree prepares git-crypt for worktree creation.
func (h *GitCryptHook) prepareGitCryptForWorktree(repoPath, branch string, ctx *hooks.HookContext) error {
    // Get default branch
    defaultBranch, err := h.git.GetDefaultBranch(repoPath)
    if err != nil {
        return fmt.Errorf("failed to get default branch: %w", err)
    }
    
    // Try to find key in default branch
    keyPath, err := h.findGitCryptKey(repoPath, defaultBranch)
    if err != nil {
        return fmt.Errorf("failed to find git-crypt key: %w", err)
    }
    
    if keyPath == "" {
        // Key not found, prompt user
        keyPath, err = h.promptUserForKeyPath()
        if err != nil {
            return fmt.Errorf("failed to get key path from user: %w", err)
        }
    }
    
    // Store key path in context for post-hook to use
    ctx.Metadata["gitCryptKeyPath"] = keyPath
    ctx.Metadata["usesGitCrypt"] = true
    
    return nil
}

// findGitCryptKey looks for git-crypt key in the default branch.
func (h *GitCryptHook) findGitCryptKey(repoPath, defaultBranch string) (string, error) {
    // Check if default branch has git-crypt key
    keyPath := filepath.Join(repoPath, ".git", "git-crypt", "keys", "default")
    
    exists, err := h.fs.Exists(keyPath)
    if err != nil {
        return "", err
    }
    
    if exists {
        return keyPath, nil
    }
    
    return "", nil // Key not found
}

// promptUserForKeyPath prompts the user for the git-crypt key file path.
func (h *GitCryptHook) promptUserForKeyPath() (string, error) {
    prompt := "Git-crypt key not found in default branch. Please provide the path to the git-crypt key file:"
    
    keyPath, err := h.prompt.Input(prompt)
    if err != nil {
        return "", err
    }
    
    // Validate that the key file exists
    exists, err := h.fs.Exists(keyPath)
    if err != nil {
        return "", fmt.Errorf("failed to check key file existence: %w", err)
    }
    
    if !exists {
        return "", fmt.Errorf("git-crypt key file does not exist: %s", keyPath)
    }
    
    return keyPath, nil
}
```

#### New Hook Type: WorktreeCheckoutHook

A new hook type will be introduced to handle worktree checkout operations, keeping all git-crypt logic within the hook system:

```go
// pkg/hooks/hooks.go (addition to existing interfaces)

// WorktreeCheckoutHook executes between worktree creation and checkout.
type WorktreeCheckoutHook interface {
    Hook
    OnWorktreeCheckout(ctx *HookContext) error
}
```

#### Updated Hook Manager

```go
// pkg/hooks/manager.go (addition to existing methods)

type HookManagerInterface interface {
    // ... existing methods ...
    
    // Worktree checkout hook management
    RegisterWorktreeCheckoutHook(operation string, hook WorktreeCheckoutHook) error
    ExecuteWorktreeCheckoutHooks(operation string, ctx *HookContext) error
}

// RegisterWorktreeCheckoutHook registers a worktree checkout hook.
func (hm *HookManager) RegisterWorktreeCheckoutHook(operation string, hook WorktreeCheckoutHook) error {
    hm.mu.Lock()
    defer hm.mu.Unlock()

    if hook == nil {
        return fmt.Errorf("hook cannot be nil")
    }

    if hm.worktreeCheckoutHooks[operation] == nil {
        hm.worktreeCheckoutHooks[operation] = make([]WorktreeCheckoutHook, 0)
    }

    hm.worktreeCheckoutHooks[operation] = append(hm.worktreeCheckoutHooks[operation], hook)
    hm.sortWorktreeCheckoutHooksByPriority(operation)
    return nil
}

// ExecuteWorktreeCheckoutHooks executes all worktree checkout hooks.
func (hm *HookManager) ExecuteWorktreeCheckoutHooks(operation string, ctx *HookContext) error {
    hm.mu.RLock()
    defer hm.mu.RUnlock()

    for _, hook := range hm.worktreeCheckoutHooks[operation] {
        if err := hook.OnWorktreeCheckout(ctx); err != nil {
            return fmt.Errorf("worktree checkout hook %s failed: %w", hook.Name(), err)
        }
    }

    return nil
}
```

#### Updated CM Worktree Creation Logic

```go
// pkg/cm/worktrees_create.go (simplified approach):

func (c *realCM) executeCreateWorkTree(branch string, opts ...CreateWorkTreeOpts) error {
    // ... existing parameter processing ...
    
    // Always create worktree with --no-checkout to allow hooks to prepare
    worktreePath, err := c.git.CreateWorktreeWithNoCheckout(repoPath, branch, worktreeName)
    if err != nil {
        return fmt.Errorf("failed to create worktree: %w", err)
    }
    
    // Update context with worktree information
    ctx.Parameters["worktreePath"] = worktreePath
    ctx.Parameters["branch"] = branch
    ctx.Parameters["worktreeName"] = worktreeName
    
    // Execute worktree checkout hooks (for git-crypt setup, etc.)
    if c.hookManager != nil {
        if err := c.hookManager.ExecuteWorktreeCheckoutHooks("CreateWorkTree", ctx); err != nil {
            // Cleanup failed worktree
            _ = c.git.DeleteWorktree(worktreePath)
            return fmt.Errorf("worktree checkout hooks failed: %w", err)
        }
    }
    
    // Now checkout the branch
    if err := c.git.CheckoutBranch(worktreePath, branch); err != nil {
        // Cleanup failed worktree
        _ = c.git.DeleteWorktree(worktreePath)
        return fmt.Errorf("failed to checkout branch in worktree: %w", err)
    }
    
    // ... rest of the function ...
}
```

#### Git-Crypt WorktreeCheckoutHook Implementation

```go
// pkg/hooks/gitcrypt/worktree_checkout.go

type GitCryptWorktreeCheckoutHook struct {
    fs     fs.FSInterface
    git    git.GitInterface
    prompt prompt.PromptInterface
    logger logger.LoggerInterface
}

// NewGitCryptWorktreeCheckoutHook creates a new GitCryptWorktreeCheckoutHook instance.
func NewGitCryptWorktreeCheckoutHook() *GitCryptWorktreeCheckoutHook {
    return &GitCryptWorktreeCheckoutHook{
        fs:     fs.NewFS(),
        git:    git.NewGit(),
        prompt: prompt.NewPrompt(),
        logger: logger.NewNoopLogger(),
    }
}

// RegisterForOperations registers this hook for worktree operations.
func (h *GitCryptWorktreeCheckoutHook) RegisterForOperations(registerHook func(operation string, hook hooks.WorktreeCheckoutHook) error) error {
    // Register for operations that create worktrees
    if err := registerHook(consts.CreateWorkTree, h); err != nil {
        return err
    }
    
    if err := registerHook(consts.LoadWorktree, h); err != nil {
        return err
    }
    
    return nil
}

// Name returns the hook name.
func (h *GitCryptWorktreeCheckoutHook) Name() string {
    return "git-crypt-worktree-checkout"
}

// Priority returns the hook priority.
func (h *GitCryptWorktreeCheckoutHook) Priority() int {
    return 50
}

// Execute is a no-op for GitCryptWorktreeCheckoutHook.
func (h *GitCryptWorktreeCheckoutHook) Execute(_ *hooks.HookContext) error {
    return nil
}

// OnWorktreeCheckout handles git-crypt setup before worktree checkout.
func (h *GitCryptWorktreeCheckoutHook) OnWorktreeCheckout(ctx *hooks.HookContext) error {
    // Get worktree path from context
    worktreePath, ok := ctx.Parameters["worktreePath"].(string)
    if !ok || worktreePath == "" {
        return fmt.Errorf("worktree path not found in context")
    }
    
    // Get repository path
    repoPath, err := h.getRepositoryPath(ctx)
    if err != nil {
        return err
    }
    
    // Check if repository uses git-crypt
    usesGitCrypt, err := h.detectGitCryptUsage(repoPath)
    if err != nil {
        return fmt.Errorf("failed to detect git-crypt usage: %w", err)
    }
    
    if !usesGitCrypt {
        // No git-crypt usage detected, nothing to do
        return nil
    }
    
    // Get branch information
    branch, ok := ctx.Parameters["branch"].(string)
    if !ok || branch == "" {
        return fmt.Errorf("branch not found in context")
    }
    
    // Setup git-crypt for the worktree
    return h.setupGitCryptForWorktree(repoPath, worktreePath, branch)
}

// setupGitCryptForWorktree sets up git-crypt in the worktree.
func (h *GitCryptWorktreeCheckoutHook) setupGitCryptForWorktree(repoPath, worktreePath, branch string) error {
    // Get default branch
    defaultBranch, err := h.git.GetDefaultBranch(repoPath)
    if err != nil {
        return fmt.Errorf("failed to get default branch: %w", err)
    }
    
    // Try to find key in default branch
    keyPath, err := h.findGitCryptKey(repoPath, defaultBranch)
    if err != nil {
        return fmt.Errorf("failed to find git-crypt key: %w", err)
    }
    
    if keyPath == "" {
        // Key not found, prompt user
        keyPath, err = h.promptUserForKeyPath()
        if err != nil {
            return fmt.Errorf("failed to get key path from user: %w", err)
        }
    }
    
    // Get the worktree's git directory
    worktreeGitDir := filepath.Join(repoPath, ".git", "worktrees", filepath.Base(worktreePath))
    
    // Create git-crypt directory in worktree's git directory
    gitCryptDir := filepath.Join(worktreeGitDir, "git-crypt")
    if err := h.fs.MkdirAll(gitCryptDir, 0755); err != nil {
        return fmt.Errorf("failed to create git-crypt directory: %w", err)
    }
    
    // Copy the key to the worktree's git-crypt directory
    destKeyPath := filepath.Join(gitCryptDir, "keys", "default")
    if err := h.fs.MkdirAll(filepath.Dir(destKeyPath), 0755); err != nil {
        return fmt.Errorf("failed to create keys directory: %w", err)
    }
    
    if err := h.fs.CopyFile(keyPath, destKeyPath); err != nil {
        return fmt.Errorf("failed to copy git-crypt key: %w", err)
    }
    
    return nil
}
```

### Package Organization

The git-crypt hook will be organized in its own package:

```
pkg/hooks/gitcrypt/
├── gitcrypt.go           # Main git-crypt hook implementation
├── gitcrypt_test.go      # Unit tests for git-crypt hook
├── detection.go          # Git-crypt detection logic
├── key_management.go     # Key location and validation logic
├── worktree_setup.go     # Worktree-specific git-crypt setup
└── errors.go             # Git-crypt specific errors
```

### Error Handling

```go
// pkg/hooks/gitcrypt/errors.go

package gitcrypt

import "errors"

var (
    ErrGitCryptNotInstalled = errors.New("git-crypt is not installed")
    ErrKeyFileNotFound      = errors.New("git-crypt key file not found")
    ErrKeyFileInvalid       = errors.New("git-crypt key file is invalid or corrupted")
    ErrWorktreeSetupFailed  = errors.New("failed to setup git-crypt in worktree")
    ErrDecryptionFailed     = errors.New("failed to decrypt files in worktree")
)
```

### Integration with Default Hooks

Update the default hooks manager to include git-crypt hook:

```go
// pkg/hooks/default/default.go

import (
    "github.com/lerenn/code-manager/pkg/hooks"
    "github.com/lerenn/code-manager/pkg/hooks/gitcrypt"
    "github.com/lerenn/code-manager/pkg/hooks/ide"
)

// NewDefaultHooksManager creates a new default hooks manager with all default hooks.
func NewDefaultHooksManager() (hooks.HookManagerInterface, error) {
    hm := hooks.NewHookManager()

    // Register IDE opening hook
    if err := ide.NewOpeningHook().RegisterForOperations(hm.RegisterPostHook); err != nil {
        return nil, err
    }

    // Register git-crypt worktree checkout hook
    if err := gitcrypt.NewGitCryptWorktreeCheckoutHook().RegisterForOperations(hm.RegisterWorktreeCheckoutHook); err != nil {
        return nil, err
    }

    return hm, nil
}
```

## Implementation Plan

### Phase 1: New Hook Type Infrastructure
1. Add `WorktreeCheckoutHook` interface to `pkg/hooks/hooks.go`
2. Extend `HookManager` to support worktree checkout hooks
3. Add registration and execution methods for worktree checkout hooks
4. Update hook manager tests

### Phase 2: CM Integration
1. Modify worktree creation logic to always use `--no-checkout`
2. Add worktree checkout hook execution between creation and checkout
3. Update CM to pass worktree context to hooks
4. Add proper error handling and cleanup

### Phase 3: Git-Crypt Hook Implementation
1. Create `pkg/hooks/gitcrypt` package with basic structure
2. Implement `GitCryptWorktreeCheckoutHook` with detection logic
3. Implement key location and validation
4. Implement git-crypt setup in worktree directories

## Testing Strategy

### Unit Tests
- Test git-crypt detection logic
- Test key location and validation
- Test hook registration and execution
- Test error handling scenarios

### E2E Tests
- Test complete git-crypt workflow
- Test with missing git-crypt installation
- Test with invalid or missing key files
- Test user interaction scenarios

## Migration Strategy

1. **Backward Compatibility**: All existing worktree operations will continue to work without git-crypt
2. **Automatic Detection**: Git-crypt handling is only enabled when needed
3. **Graceful Degradation**: Operations fail gracefully with clear error messages
4. **No Configuration Required**: No additional configuration needed for basic usage

## Success Criteria

1. **Functionality**: Worktree creation succeeds with git-crypt repositories
2. **Security**: No keys or sensitive information stored in configuration
3. **Usability**: Clear error messages and user guidance
4. **Performance**: Minimal overhead for non-encrypted repositories
5. **Reliability**: Proper cleanup and error handling in all scenarios

## Future Enhancements

1. **Multiple Key Support**: Support for repositories with multiple git-crypt keys
2. **Key Rotation**: Support for key rotation scenarios
3. **Team Key Sharing**: Integration with team key sharing mechanisms
4. **Performance Optimization**: Caching of git-crypt detection results
5. **Advanced Configuration**: Optional configuration for advanced use cases
