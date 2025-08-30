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
    HookManager() hooks.HookManagerInterface
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

### Built-in Hooks

#### Logging Hook
```go
// pkg/hooks/logging.go

type LoggingHook struct {
    logger logger.Logger
    level  string
}

func (h *LoggingHook) PreExecute(ctx *HookContext) error {
    h.logger.Info("Starting operation", "operation", ctx.OperationName, "params", ctx.Parameters)
    return nil
}

func (h *LoggingHook) PostExecute(ctx *HookContext) error {
    if ctx.Error != nil {
        h.logger.Error("Operation failed", "operation", ctx.OperationName, "error", ctx.Error)
    } else {
        h.logger.Info("Operation completed", "operation", ctx.OperationName, "results", ctx.Results)
    }
    return nil
}
```

#### Metrics Hook
```go
// pkg/hooks/metrics.go

type MetricsHook struct {
    metrics MetricsCollector
}

func (h *MetricsHook) PreExecute(ctx *HookContext) error {
    h.metrics.IncCounter("cm_operations_started", map[string]string{
        "operation": ctx.OperationName,
    })
    ctx.Metadata["start_time"] = time.Now()
    return nil
}

func (h *MetricsHook) PostExecute(ctx *HookContext) error {
    if startTime, ok := ctx.Metadata["start_time"].(time.Time); ok {
        duration := time.Since(startTime)
        h.metrics.RecordHistogram("cm_operation_duration", duration, map[string]string{
            "operation": ctx.OperationName,
        })
    }
    
    if ctx.Error != nil {
        h.metrics.IncCounter("cm_operations_failed", map[string]string{
            "operation": ctx.OperationName,
        })
    } else {
        h.metrics.IncCounter("cm_operations_succeeded", map[string]string{
            "operation": ctx.OperationName,
        })
    }
    return nil
}
```

#### Validation Hook
```go
// pkg/hooks/validation.go

type ValidationHook struct {
    validators map[string]Validator
}

func (h *ValidationHook) PreExecute(ctx *HookContext) error {
    if validator, exists := h.validators[ctx.OperationName]; exists {
        return validator.Validate(ctx.Parameters)
    }
    return nil
}
```

### Programmatic Hook Registration

#### Hook Setup in main.go
```go
// cmd/cm/main.go

func setupHooks(cmInstance cm.CM) {
    hookManager := cmInstance.HookManager()
    
    // Register global hooks
    hookManager.RegisterGlobalHook(&hooks.LoggingHook{
        Logger: logger.NewLogger(),
        Level:  "info",
    })
    
    hookManager.RegisterGlobalHook(&hooks.MetricsHook{
        Metrics: metrics.NewCollector(),
    })
    
    // Register operation-specific hooks
    hookManager.RegisterPreHook("CreateWorkTree", &hooks.ValidationHook{
        Validators: map[string]hooks.Validator{
            "CreateWorkTree": &CreateWorkTreeValidator{},
        },
    })
    
    hookManager.RegisterPreHook("DeleteWorkTree", &hooks.ConfirmationHook{
        RequireConfirmation: true,
    })
    
    hookManager.RegisterPreHook("Clone", &hooks.ValidationHook{
        Validators: map[string]hooks.Validator{
            "Clone": &CloneValidator{},
        },
    })
    
    // Add more operation-specific hooks as needed...
}

func main() {
    // ... existing main.go code ...
    
    // Initialize CM with hooks
    cmInstance := cm.NewCM(config)
    setupHooks(cmInstance)
    
    // ... rest of main.go ...
}
```



## Implementation Plan

### Phase 1: Core Hook System
1. Create `pkg/hooks` package with core interfaces
2. Implement `HookManager` with basic registration and execution
3. Update CM interface to include hook management methods
4. Add hook execution wrappers to all CM operations
5. Create basic logging and metrics hooks

### Phase 2: Built-in Hooks
1. Implement validation hooks for all operations
2. Add confirmation hooks for destructive operations
3. Create performance monitoring hooks
4. Implement error handling and recovery hooks

### Phase 3: Programmatic Setup
1. Implement hook setup function in main.go
2. Add hook registration for all operations
3. Create hook testing utilities

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
