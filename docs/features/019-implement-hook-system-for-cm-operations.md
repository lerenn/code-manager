# Feature 019: Implement Hook System for CM Operations

## Overview

Implement a comprehensive hook system for all CM operations that allows for pre-execution and post-execution middleware functionality. This system will enable extensibility, logging, validation, and custom behavior injection without modifying the core CM logic. The hook system will treat hooks as middlewares that can be chained and configured per operation.

## Background

Currently, CM operations execute their business logic directly without any extensibility points. This makes it difficult to add cross-cutting concerns like:
- Enhanced logging and metrics
- Validation and authorization
- Custom business logic injection
- Performance monitoring
- Error handling customization
- Integration with external systems

A hook system will provide a clean, extensible architecture that maintains separation of concerns while allowing powerful customization capabilities.

## Requirements

### Functional Requirements

1. **Hook Registration**
   - Register pre-hooks and post-hooks for each CM operation
   - Support multiple hooks per operation
   - Allow hook ordering and priority
   - Support global hooks that apply to all operations
   - Support operation-specific hooks

2. **Hook Execution**
   - Execute pre-hooks before operation logic
   - Execute post-hooks after operation logic (including error cases)
   - Support hook chaining and early termination
   - Handle hook failures gracefully
   - Maintain operation context throughout hook chain

3. **Hook Context**
   - Provide operation name and parameters to hooks
   - Allow hooks to modify operation parameters
   - Allow hooks to access operation results
   - Support hook-specific metadata and configuration
   - Provide access to CM instance and dependencies

4. **Hook Types**
   - **Pre-hooks**: Execute before operation, can modify parameters or abort execution
   - **Post-hooks**: Execute after operation, can process results or handle errors
   - **Error-hooks**: Execute when operations fail, for error handling and recovery
   - **Global hooks**: Apply to all operations automatically

5. **Hook Management**
   - Add/remove hooks dynamically
   - Enable/disable hooks without removal
   - Configure hook behavior through options
   - Support hook dependencies and ordering

### Non-Functional Requirements

1. **Performance**
   - Minimal overhead for hook execution
   - Efficient hook lookup and chaining
   - Support for async hooks where appropriate

2. **Reliability**
   - Hook failures should not break core operations
   - Graceful degradation when hooks are unavailable
   - Proper error propagation and handling

3. **Extensibility**
   - Easy to add new hook types
   - Support for third-party hook implementations
   - Plugin-like architecture for hook management

4. **Testability**
   - Easy to mock and test hooks
   - Support for hook testing in isolation
   - Clear separation between hook and core logic

## Technical Specification

### Hook System Architecture

#### Core Hook Interfaces
```go
// pkg/hooks/hooks.go

// HookContext provides context for hook execution
type HookContext struct {
    OperationName string
    Parameters    map[string]interface{}
    Results       map[string]interface{}
    Error         error
    CM            CM
    Metadata      map[string]interface{}
}

// Hook defines the interface for all hooks
type Hook interface {
    Name() string
    Priority() int
    Execute(ctx *HookContext) error
}

// PreHook executes before an operation
type PreHook interface {
    Hook
    PreExecute(ctx *HookContext) error
}

// PostHook executes after an operation
type PostHook interface {
    Hook
    PostExecute(ctx *HookContext) error
}

// ErrorHook executes when an operation fails
type ErrorHook interface {
    Hook
    OnError(ctx *HookContext) error
}
```

#### Hook Manager
```go
// pkg/hooks/manager.go

type HookManager struct {
    preHooks   map[string][]PreHook
    postHooks  map[string][]PostHook
    errorHooks map[string][]ErrorHook
    globalHooks []Hook
    mu         sync.RWMutex
}

type HookManagerInterface interface {
    // Hook registration
    RegisterPreHook(operation string, hook PreHook) error
    RegisterPostHook(operation string, hook PostHook) error
    RegisterErrorHook(operation string, hook ErrorHook) error
    RegisterGlobalHook(hook Hook) error
    
    // Hook execution
    ExecutePreHooks(operation string, ctx *HookContext) error
    ExecutePostHooks(operation string, ctx *HookContext) error
    ExecuteErrorHooks(operation string, ctx *HookContext) error
    
    // Hook management
    RemoveHook(operation, hookName string) error
    EnableHook(operation, hookName string) error
    DisableHook(operation, hookName string) error
    ListHooks(operation string) ([]Hook, error)
}
```

#

### CM Interface Extension

#### Updated CM Interface
```go
// pkg/cm/cm.go

type CM interface {
    // Existing methods...
    CreateWorkTree(branch string, opts ...CreateWorkTreeOpts) error
    DeleteWorkTree(branch string, force bool) error
    OpenWorktree(worktreeName, ideName string) error
    ListWorktrees(force bool) ([]status.WorktreeInfo, ProjectType, error)
    LoadWorktree(branchArg string, opts ...LoadWorktreeOpts) error
    Init(opts InitOpts) error
    Clone(repoURL string, opts ...CloneOpts) error
    ListRepositories() ([]RepositoryInfo, error)
    SetVerbose(verbose bool)
    
    // New hook management methods
    RegisterHook(operation string, hook hooks.Hook) error
    UnregisterHook(operation, hookName string) error
}
```

#### Updated realCM Implementation
```go
// pkg/cm/cm.go

type realCM struct {
    *basepkg.Base
    ideManager  ide.ManagerInterface
    repository  repository.Repository
    workspace   workspace.Workspace
    hookManager hooks.HookManagerInterface
}

// Wrapper methods with hook execution
func (c *realCM) CreateWorkTree(branch string, opts ...CreateWorkTreeOpts) error {
    ctx := &hooks.HookContext{
        OperationName: "CreateWorkTree",
        Parameters: map[string]interface{}{
            "branch": branch,
            "opts":   opts,
        },
        CM: c,
    }
    
    // Execute pre-hooks
    if err := c.hookManager.ExecutePreHooks("CreateWorkTree", ctx); err != nil {
        return err
    }
    
    // Execute operation
    var resultErr error
    func() {
        defer func() {
            if r := recover(); r != nil {
                resultErr = fmt.Errorf("panic in CreateWorkTree: %v", r)
            }
        }()
        resultErr = c.executeCreateWorkTree(branch, opts...)
    }()
    
    // Update context with results
    ctx.Error = resultErr
    if resultErr == nil {
        ctx.Results = map[string]interface{}{
            "success": true,
        }
    }
    
    // Execute post-hooks or error-hooks
    if resultErr != nil {
        c.hookManager.ExecuteErrorHooks("CreateWorkTree", ctx)
    } else {
        c.hookManager.ExecutePostHooks("CreateWorkTree", ctx)
    }
    
    return resultErr
}

// Similar wrapper methods for all other operations...
```

### Operation Constants

All operation names used in the hook system are centralized in `pkg/cm/consts/operations.go` to ensure consistency and prevent typos:

```go
// pkg/cm/consts/operations.go
package consts

// Operation names for the hook system
const (
    // Worktree operations
    CreateWorkTree = "CreateWorkTree"
    DeleteWorkTree = "DeleteWorkTree"
    LoadWorktree   = "LoadWorktree"
    ListWorktrees  = "ListWorktrees"
    OpenWorktree   = "OpenWorktree"

    // Repository operations
    CloneRepository = "CloneRepository"
    ListRepositories = "ListRepositories"
    Clone = "Clone" // Legacy name for backward compatibility

    // Initialization operations
    Init = "Init"
)
```

This approach provides several benefits:
- **Consistency**: All operation names are defined in one place
- **Type Safety**: Prevents typos in operation names
- **Maintainability**: Easy to update operation names across the codebase
- **IDE Support**: Better autocomplete and refactoring support

### Built-in Hooks

The hook system provides a framework for creating custom hooks. Users can implement their own hooks based on the provided interfaces:

- **PreHook**: Execute before operations, can modify parameters or abort execution
- **PostHook**: Execute after operations, can process results or handle errors  
- **ErrorHook**: Execute when operations fail, for error handling and recovery
- **GlobalHook**: Apply to all operations automatically

Example hook implementations can be created for:
- Logging and metrics collection
- Parameter validation
- User confirmation for destructive operations
- Performance monitoring
- Error handling and recovery

### IDE Opening Hook Implementation

The IDE opening functionality has been implemented as a post-hook that validates IDE opening parameters and stores the information for the CM to handle. The hook focuses on validation and information storage, while the CM handles the actual IDE opening:

```go
// pkg/hooks/ide/opening.go

type OpeningHook struct{}

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

func (h *OpeningHook) PostExecute(ctx *hooks.HookContext) error {
    // Only proceed if operation was successful
    if ctx.Error != nil {
        return nil
    }

    // Check if IDE name is provided in parameters
    ideName, hasIDEName := ctx.Parameters["ideName"]
    if !hasIDEName {
        return nil
    }

    // Get worktree path from parameters
    var worktreePath string
    if branch, hasBranch := ctx.Parameters["branch"]; hasBranch {
        if branchStr, ok := branch.(string); ok && branchStr != "" {
            worktreePath = branchStr
        }
    } else if worktreeName, hasWorktreeName := ctx.Parameters["worktreeName"]; hasWorktreeName {
        if worktreeNameStr, ok := worktreeName.(string); ok && worktreeNameStr != "" {
            worktreePath = worktreeNameStr
        }
    }

    if worktreePath == "" {
        return fmt.Errorf("cannot open IDE: worktree path is empty")
    }

    // Store the IDE opening information in the context for the CM to handle
    ctx.Results["ideName"] = ideNameStr
    ctx.Results["worktreePath"] = worktreePath
    ctx.Results["shouldOpenIDE"] = true

    return nil
}
```

This hook automatically registers itself for `CreateWorkTree`, `LoadWorktree`, and `OpenWorktree` operations and replaces the previous direct IDE opening calls in the CM operations. The CM then handles the actual IDE opening after hook execution.

The CM package handles IDE opening after hook execution:

```go
// pkg/cm/cm.go

// Execute post-hooks or error-hooks (if hook manager is available)
if c.hookManager != nil {
    if resultErr != nil {
        _ = c.hookManager.ExecuteErrorHooks(operationName, ctx)
    } else {
        _ = c.hookManager.ExecutePostHooks(operationName, ctx)
        
        // Handle IDE opening if requested by hooks
        if shouldOpenIDE, exists := ctx.Results["shouldOpenIDE"]; exists && shouldOpenIDE == true {
            if ideName, hasIDE := ctx.Results["ideName"]; hasIDE {
                if worktreePath, hasPath := ctx.Results["worktreePath"]; hasPath {
                    if ideNameStr, ok := ideName.(string); ok {
                        if worktreePathStr, ok := worktreePath.(string); ok {
                            _ = c.ideManager.OpenIDE(ideNameStr, worktreePathStr, c.IsVerbose())
                        }
                    }
                }
            }
        }
    }
}
```

### Package Organization

The codebase has been organized into focused packages for better modularity:

#### IDE Opening Package Organization

The IDE opening functionality has been integrated into a coherent package that combines the hook system with IDE management:

```
pkg/hooks/ide/
├── opening.go    # IDE opening hook implementation
├── opening_test.go
├── ide.go           # IDE manager and interfaces
├── vscode.go        # VS Code IDE implementation
├── cursor.go        # Cursor IDE implementation
├── dummy.go         # Dummy IDE for testing
├── errors.go        # IDE-related errors
└── ide_test.go      # IDE manager tests
```

#### Branch Package Organization

Branch-related functionality has been moved to its own package:

```
pkg/branch/
├── sanitize.go       # Branch name sanitization
└── sanitize_test.go  # Branch sanitization tests
```

This organization provides several benefits:
- **Coherent Integration**: IDE opening hook and IDE management are in the same package
- **Separation of Concerns**: Hook validates and stores information, CM handles actual IDE opening
- **Modularity**: Each package can be developed and tested independently
- **Clear Dependencies**: Explicit imports show which packages are being used
- **Scalability**: Easy to add new functionality in focused packages
- **Maintainability**: Package-specific code is isolated and focused

### IDE Opening Hook Coverage

The IDE opening hook is registered for the following operations that support IDE opening:

| Operation | IDE Opening Support | Parameters |
|-----------|-------------------|------------|
| `CreateWorkTree` | ✅ Yes | `ideName`, `branch` |
| `LoadWorktree` | ✅ Yes | `ideName`, `branchArg` |
| `OpenWorktree` | ✅ Yes | `ideName`, `worktreeName` |
| `DeleteWorkTree` | ❌ No | No IDE parameters |
| `ListWorktrees` | ❌ No | No IDE parameters |
| `Clone` | ❌ No | No IDE parameters |
| `ListRepositories` | ❌ No | No IDE parameters |
| `Init` | ❌ No | No IDE parameters |

The hook validates that `ideName` and branch/worktree information are provided and stores the IDE opening information in the hook context for the IDE manager to use.

### Programmatic Hook Registration

#### Hook Setup in CM Creation
```go
// pkg/cm/cm.go

func NewCM(cfg *config.Config) (CM, error) {
    // ... create dependencies ...
    
    cmInstance := &realCM{
        // ... initialize fields ...
        hookManager: hooks.NewHookManager(),
    }

    // Setup hooks for the CM instance
    if err := setupHooks(cmInstance); err != nil {
        return nil, err
    }

    return cmInstance, nil
}

// setupHooks configures and registers all hooks for the CM instance.
func setupHooks(cmInstance *realCM) error {
    // Register IDE opening hook for operations that create worktrees
    if err := ide.NewOpeningHook().RegisterForOperations(cmInstance); err != nil {
        return err
    }

    return nil
}

// Usage in CLI commands (simplified):
func createCreateCmdRunE(ideName *string, force *bool, fromIssue *string) func(*cobra.Command, []string) error {
    return func(_ *cobra.Command, args []string) error {
        // ... existing code ...
        
        cmManager := cm.NewCM(cfg) // Hooks are automatically set up
        cmManager.SetVerbose(config.Verbose)
        
        // ... rest of the function ...
    }
}
```



## Implementation Plan

### Phase 1: Core Hook System
1. Create `pkg/hooks` package with core interfaces
2. Implement `HookManager` with basic registration and execution
3. Update CM interface to include hook management methods
4. Add hook execution wrappers to all CM operations
5. Create basic logging and metrics hooks

### Phase 2: Hook Framework ✅ COMPLETED
1. ✅ Provide hook interfaces and framework
2. ✅ Implement IDE opening hook to replace direct IDE calls in existing functions
3. ✅ Implement hook setup function in main.go
4. ✅ Add hook registration for all operations
5. ✅ Create hook testing utilities

## Testing Strategy

### Unit Tests
- Test hook registration and execution
- Test hook ordering and priority
- Test hook error handling
- Test hook context management

### Integration Tests
- Test hook integration with CM operations
- Test hook registration and setup

### E2E Tests
- Test complete hook workflows
- Test hook performance impact
- Test hook failure scenarios

## Migration Strategy

1. **Backward Compatibility**: All existing CM operations will continue to work without hooks
2. **Gradual Rollout**: Hooks can be enabled/disabled programmatically
3. **Default Setup**: Only logging hooks enabled by default in main.go
4. **Documentation**: Provide comprehensive examples and migration guides

## Success Criteria

1. **Functionality**: All CM operations support pre/post/error hooks
2. **Performance**: Hook overhead < 5ms per operation
3. **Reliability**: Hook failures don't break core operations
4. **Usability**: Easy to add and configure custom hooks
5. **Testability**: Comprehensive test coverage for hook system
