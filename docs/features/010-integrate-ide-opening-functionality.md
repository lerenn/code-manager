# Feature 010: Integrate IDE Opening Functionality

## Overview

Add the ability to open repositories directly in supported IDEs using the `-o` command flag. This feature will allow users to quickly open worktrees in their preferred development environment.

## Command Syntax

### Create Command with IDE Opening
```bash
wtm create [branch-name] -i <ide-name>
```

### Open Existing Worktree
```bash
wtm open <worktree-name> -i <ide-name>
```

### Examples

```bash
# Create a new branch and open in VS Code
wtm create feature/new-feature -i vscode

# Create a new branch and open in GoLand
wtm create bugfix/issue-123 -i goland

# Create a new branch and open in IntelliJ IDEA
wtm create hotfix/critical-fix -i intellij

# Open existing worktree in Cursor
wtm open feature/new-feature -i cursor

# Open existing worktree in VS Code
wtm open bugfix/issue-123 -i vscode
```



## Requirements

### Functional Requirements

1. **IDE Detection**
   - Detect if the specified IDE is installed on the system
   - Return error if IDE is not supported or not installed
   - Support multiple installation paths per IDE

2. **Repository Resolution**
   - For create command: Use the current worktree directory after branch creation
   - For open command: Look up worktree path in status.yaml file
   - Support both relative and absolute paths
   - Do not support repositories that don't exist yet

3. **Command Execution**
   - Execute the appropriate command to open the IDE with the repository
   - Handle different IDE command-line interfaces
   - Support both GUI and terminal-based IDEs

4. **Error Handling**
   - Return error when IDE is not supported or not installed
   - Return error when worktree is not found in status.yaml
   - Clear error messages for unsupported IDEs
   - Validation of repository paths

### Non-Functional Requirements

1. **Performance**
   - Quick command execution
   - No caching or detection overhead

2. **User Experience**
   - Intuitive command syntax
   - Clear error messages for unsupported IDEs

3. **Cross-Platform Support**
   - macOS: Support for Applications folder and Homebrew installations
   - Linux: Support for system packages and snap/flatpak installations
   - Windows: Support for Program Files and user installations

## Implementation Details

### New Package: `pkg/ide`

```go
// pkg/ide/ide.go
type IDE interface {
    Name() string
    IsInstalled() bool
    OpenRepository(path string) error
}

type Manager struct {
    ides map[string]IDE
}

func (m *Manager) GetIDE(name string) (IDE, error)
func (m *Manager) OpenIDE(name, path string, verbose bool) error
```

### IDE Implementations

```go
// pkg/ide/cursor.go
type Cursor struct {
    fs fs.FS
}

func (c *Cursor) Name() string { return "cursor" }
func (c *Cursor) IsInstalled() bool { /* use fs.Which("cursor") */ }
func (c *Cursor) OpenRepository(path string) error { /* use fs.ExecuteCommand("cursor", path) */ }

// Similar implementations for other IDEs
```

### FS Adapter Extensions

```go
// pkg/fs/fs.go
// Add to existing FS interface:
Which(command string) (string, error)
ExecuteCommand(command string, args ...string) error
```

### CLI Integration

```go
// cmd/wtm/main.go
// Add -o flag to existing create command
var createCmd = &cobra.Command{
    Use:   "create [branch-name]",
    Short: "Create a new worktree",
    Long:  "Create a new worktree with optional IDE opening",
    Args:  cobra.ExactArgs(1),
    RunE:  runCreate,
}

// Add IDE flag
createCmd.Flags().StringP("ide", "i", "", "Open in specified IDE after creation")

// New open command
var openCmd = &cobra.Command{
    Use:   "open [worktree-name]",
    Short: "Open existing worktree in IDE",
    Long:  "Open an existing worktree in the specified IDE",
    Args:  cobra.ExactArgs(1),
    RunE:  runOpen,
}

// Add IDE flag
openCmd.Flags().StringP("ide", "i", "", "IDE to open worktree in")
```

## Configuration

### IDE Configuration in `configs/default.yaml`

```yaml
ide:
  # Custom IDE paths (optional)
  custom_paths:
    cursor: "/usr/local/bin/cursor"
    vscode: "/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code"
    goland: "/Applications/GoLand.app/Contents/MacOS/goland"
```

## Error Handling

### Error Types

```go
// pkg/ide/errors.go
var (
    ErrIDENotInstalled = errors.New("IDE not installed")
    ErrUnsupportedIDE = errors.New("unsupported IDE")
    ErrIDEExecutionFailed = errors.New("failed to execute IDE command")
)
```

### Error Messages

- `"IDE 'cursor' is not installed"`
- `"Unsupported IDE 'unknown-ide'"`
- `"Failed to open cursor: command not found"`
- `"Worktree 'branch-name' not found"`

## Testing Strategy

### Unit Tests
- IDE detection logic using mocked FS adapter
- Error handling with specific error types
- IDE registry and manager functionality

## Dependencies

### New Dependencies
- None (uses standard library for process execution)

### Modified Dependencies
- `pkg/wtm`: Add IDE opening functionality
- `cmd/wtm`: Add new CLI command

## Migration and Backward Compatibility

- This is a new feature, no migration required
- Backward compatible with existing commands
- Optional feature that doesn't affect core functionality



## Implementation Decisions

1. **IDE Interface**: `Name()`, `IsInstalled()`, `OpenRepository(path string) error`
2. **Detection**: Use `which` command through FS adapter to check if IDE is available in PATH
3. **Execution**: Launch in background using `fs.ExecuteCommand()`
4. **Error Handling**: 
   - `ErrIDENotInstalled` when `IsInstalled()` fails
   - `ErrUnsupportedIDE` when implementation doesn't exist
   - `ErrWorktreeNotFound` when worktree not found in status.yaml
5. **Integration**: 
   - Create command: Open IDE after worktree creation, don't reverse on failure
   - Open command: Look up worktree in status.yaml, validate existence, then open IDE
6. **Logging**: Log failures always, log successes only when verbose is active
7. **First Implementation**: Cursor IDE with command `cursor <repo-path>`
8. **Testing**: Unit tests only using mocked FS adapter
9. **CLI Structure**: 
   - Create command: `wtm create [branch-name] -i <ide-name>`
   - Open command: `wtm open <worktree-name> -i <ide-name>`
