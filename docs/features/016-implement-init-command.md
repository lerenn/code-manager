# Feature 016: Implement Init Command

## Overview
Implement an `init` command that allows users to configure CM for first-time use. The command will provide an interactive prompt to choose the base path for code storage, with a non-interactive option for using default values. This ensures users can customize their CM setup while maintaining a smooth onboarding experience.

## Background
Currently, CM uses hardcoded default paths for configuration and data storage. Users need a way to customize these paths based on their preferences and existing code organization. The init command will provide a guided setup process that creates the necessary configuration files and directory structure.

## Requirements

### Functional Requirements
1. **First-Time Detection**: Detect when CM is being used for the first time (status.yaml has `initialized: false`)
2. **Interactive Configuration**: Provide interactive prompts for base path selection
3. **Non-Interactive Mode**: Support `--default` flag for automatic setup with default values
4. **Configuration Creation**: Create config.yaml with user-selected paths
5. **Directory Structure**: Create necessary directories (base_path, worktrees_dir)
6. **Status Initialization**: Initialize status.yaml with `initialized: true`
7. **Path Validation**: Validate user-provided paths for accessibility and permissions
8. **Default Values**: Use sensible defaults for all configuration options
9. **Error Recovery**: Handle configuration errors gracefully with helpful messages
10. **Existing Configuration**: Prevent re-initialization if already configured

### Non-Functional Requirements
1. **User Experience**: Provide clear, helpful prompts and error messages
2. **Cross-Platform**: Work consistently on Windows, macOS, and Linux
3. **Performance**: Complete initialization within 5 seconds
4. **Reliability**: Handle file system errors and permission issues gracefully
5. **Testability**: Support unit testing with mocked dependencies
6. **Minimal Dependencies**: Use only Go standard library for interactive prompts

## Technical Specification

### Configuration Structure Changes

#### Updated Config Structure
```go
type Config struct {
    BasePath   string `yaml:"base_path"`   // User's code directory (default: ~/Code)
    StatusFile string `yaml:"status_file"` // Status file path (default: ~/.cm/status.yaml)
    // WorktreesDir field removed - computed as $base_path/worktrees
}

// New method for computing worktrees directory
func (c *Config) GetWorktreesDir() string {
    return filepath.Join(c.BasePath, "worktrees")
}
```

#### Updated Status Structure
```go
type Status struct {
    Initialized   bool          `yaml:"initialized"`   // New field: indicates if CM is initialized
    Repositories  []Repository  `yaml:"repositories"`
}
```

#### Default Configuration Values
```yaml
# Default configuration after init (configs/default.yaml updated)
base_path: ~/Code                    # User's code directory
status_file: ~/.cm/status.yaml       # CM status tracking
# worktrees_dir computed as $base_path/worktrees (not in config)
```

### Interface Design

#### Init Command (CLI)
**Command Structure**:
```bash
cm init [--force] [--base-path <path>]
```

**Flags**:
- `--force`: Skip interactive confirmation for reset (replaces --default flag)
- `--base-path <path>`: Set base path directly, skipping interactive prompt
- `--reset`: Reset CM configuration and start fresh (with confirmation prompt unless --force is used)

**Interactive Prompts**:
1. **Base Path Selection**: "Choose the location of the repositories (ex: ~/Code, ~/Projects, ~/Development): [default: ~/Code]: "
2. **Reset Confirmation**: "This will reset your CM configuration and remove all existing worktrees. Are you sure? [y/N]: "
3. **Final Confirmation**: "CM will be configured with the following settings: ... [Y/n]: "

**Path Handling**:
- Handle path expansion (`~` to user's home directory) in FS package as utility function
- Support both relative and absolute paths
- Validate expanded paths for accessibility and permissions

#### CM Package Extension (Business Logic)
**New Interface Methods**:
- `IsInitialized() (bool, error)`: Check if CM is initialized
- `Init(opts InitOpts) error`: Initialize CM configuration
- `ValidateInitPaths(basePath string) error`: Validate user-provided paths

**Option Struct**:
```go
type InitOpts struct {
    Force    bool
    Reset    bool
    BasePath string
}
```

**Key Characteristics**:
- **NO direct file system access** - all operations go through adapters
- **ONLY unit tests** using mocked dependencies
- Business logic for initialization workflow
- Path validation and configuration management
- Error handling with wrapped errors

#### Status Package Extension
**New Interface Methods**:
- `IsInitialized() (bool, error)`: Check initialization status
- `SetInitialized(initialized bool) error`: Set initialization status
- `CreateInitialStatus() error`: Create status file with initial structure

**Updated Status Structure**:
- Add `Initialized` field to Status struct
- Update default status creation to include initialization flag
- Support checking and setting initialization status

#### Config Package Extension
**New Interface Methods**:
- `SaveConfig(config *Config, configPath string) error`: Save configuration to file
- `CreateConfigDirectory(configPath string) error`: Create config directory structure
- `ValidateBasePath(basePath string) error`: Validate base path accessibility
- `GetWorktreesDir() string`: Compute worktrees directory path

**Updated Default Configuration**:
- Change default `base_path` from `~/.cm` to `~/Code`
- Remove `worktrees_dir` field (computed dynamically via `GetWorktreesDir()`)
- Keep `status_file` configurable (default: `~/.cm/status.yaml`)
- Update `configs/default.yaml` to reflect new defaults

#### FS Package Extension
**New Interface Methods**:
- `CreateDirectory(path string, perm os.FileMode) error`: Create directory with permissions
- `CreateFileWithContent(path string, content []byte, perm os.FileMode) error`: Create file with content
- `IsDirectoryWritable(path string) (bool, error)`: Check directory write permissions
- `ExpandPath(path string) (string, error)`: Expand `~` to user's home directory using `os.UserHomeDir()`

**Key Characteristics**:
- Extends existing FS package with directory creation
- Pure function-based operations
- Cross-platform compatibility
- Proper error handling and validation
- Path expansion utility for handling `~` in paths (returns error if home directory not found)

### Implementation Details

#### 1. CLI Implementation
The init command will be added to the main CLI:

**Command Structure**:
```go
func createInitCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "init [--force] [--base-path <path>] [--reset]",
        Short: "Initialize CM configuration",
        Long: `Initialize CM configuration with interactive prompts or direct path specification.

Flags:
  --force       Skip interactive confirmation when using --reset flag
  --base-path   Set the base path for code storage directly (skips interactive prompt)
  --reset       Reset existing CM configuration and start fresh`,
        RunE: func(_ *cobra.Command, args []string) error {
            cfg := loadConfig()
            cmManager := cm.NewCM(cfg)
            
            opts := cm.InitOpts{
                Force:    force,
                Reset:    reset,
                BasePath: basePath,
            }
            
            return cmManager.Init(opts)
        },
    }
}
```

**Interactive Prompts**:
- Use `bufio.NewReader(os.Stdin)` for user input
- Provide clear prompts with default values and common options
- Handle empty input (use defaults)
- Validate user input before proceeding
- Exit with error on validation failure (require user to run init again)
- Path expansion handled in FS package as utility function

#### 2. CM Package Implementation
Extends business logic with initialization functionality:

**New Methods**:
- `IsInitialized()`: Check status file for initialization flag
- `Init(opts InitOpts)`: Main initialization workflow
- `ValidateInitPaths(basePath)`: Validate user-provided paths

**Initialization Workflow**:
1. Check if already initialized (unless --reset)
2. Handle --reset flag (clear existing configuration with confirmation, skip if --force)
3. Get base path (interactive, --base-path flag, or default)
4. Validate paths and permissions using FS adapter (check if exists, create if doesn't using `os.MkdirAll`)
5. Update configuration file with new base_path
6. Create base_path directory (worktrees created when needed)
7. Initialize status file with `initialized: false`
8. Set initialized flag to true

#### 3. Status Package Implementation
Extends status management with initialization support:

**New Methods**:
- `IsInitialized()`: Check initialization status
- `SetInitialized(initialized bool)`: Update initialization status
- `CreateInitialStatus()`: Create status file with initial structure

**Status File Structure**:
```yaml
initialized: true
repositories: []
```

#### 4. Config Package Implementation
Extends configuration management with save functionality:

**New Methods**:
- `SaveConfig(config, path)`: Save configuration to YAML file
- `CreateConfigDirectory(path)`: Create config directory structure
- `ValidateBasePath(path)`: Validate base path accessibility
- `GetWorktreesDir() string`: Compute worktrees directory path

**Updated Defaults**:
- `base_path`: `~/Code` (instead of `~/.cm`)
- `status_file`: Configurable (default: `~/.cm/status.yaml`)
- `worktrees_dir`: Computed as `$base_path/worktrees` via `GetWorktreesDir()`
- Update `configs/default.yaml` to reflect new defaults (change base_path to ~/Code, remove worktrees_dir, keep status_file)

#### 5. FS Package Implementation
Extends file system operations with directory creation:

**New Methods**:
- `CreateDirectory(path, perm)`: Create directory with permissions
- `CreateFileWithContent(path, content, perm)`: Create file with content
- `IsDirectoryWritable(path)`: Check directory write permissions

**Path Expansion Integration**:
- Create helper function `expandPath(path string) (string, error)` for `~` expansion
- Integrate path expansion into existing FS methods that accept paths
- Use `os.UserHomeDir()` for home directory resolution

**Implementation**:
- Use `os.MkdirAll` for directory creation
- Use atomic file operations for file creation
- Proper error handling and validation
- Automatic path expansion in relevant methods

### File Structure

```
cmd/cm/
├── main.go                    # Extended with init command
├── init.go                    # Init command implementation
└── prompts.go                 # Interactive prompt utilities

pkg/
├── cm/
│   ├── cm.go                  # Extended with init functionality
│   ├── init.go                # Init business logic
│   ├── cm_test.go             # Unit tests for CM
│   ├── init_test.go           # Unit tests for init functionality
│   └── mockcm.gen.go          # Generated mock for testing
├── status/
│   ├── status.go              # Extended with initialization support
│   ├── status_test.go         # Unit tests for status
│   └── mockstatus.gen.go      # Generated mock for testing
├── config/
│   ├── config.go              # Extended with save functionality
│   ├── config_test.go         # Unit tests for config
│   └── mockconfig.gen.go      # Generated mock for testing
└── fs/
    ├── fs.go                  # Extended with directory creation
    ├── fs_test.go             # Integration tests for FS
    └── mockfs.gen.go          # Generated mock for testing
```

### Testing Strategy

#### Unit Tests (Business Logic)
- **CM Package**: Mock Status, Config, and FS adapters for init operations
- **Status Package**: Mock FS adapter for status file operations
- **Config Package**: Mock FS adapter for config file operations
- **Mock Regeneration**: Update interface definitions and regenerate mocks in same implementation phase

#### Integration Tests (Adapters)
- **FS Package**: Real file system operations with temporary directories and cleanup
- **CLI Package**: End-to-end command execution with temporary directories

#### Test Scenarios
1. **First-Time Initialization**: Test interactive initialization flow
2. **Base Path Flag**: Test `--base-path` flag functionality (skips interactive prompt)
3. **Reset Initialization**: Test `--reset` flag functionality with confirmation
4. **Force Reset**: Test `--force` flag with reset (skips confirmation)
5. **Re-Initialization Prevention**: Test prevention of duplicate initialization
6. **Path Validation**: Test invalid path handling
7. **Permission Errors**: Test insufficient permissions handling
8. **Configuration Update**: Test config file update with new base_path
9. **Status Initialization**: Test status file creation and flag setting
10. **Directory Creation**: Test directory structure creation
11. **Error Recovery**: Test various error conditions and recovery
12. **Cross-Platform**: Test path handling on different operating systems
13. **Path Expansion**: Test `~` expansion in paths

### Error Handling

#### Initialization Errors
- **Already Initialized**: Return specific error type `ErrAlreadyInitialized` with clear message about existing configuration (unless --reset): "CM is already initialized. Use --reset to clear existing configuration and start fresh."
- **Invalid Path**: Provide helpful suggestions for valid paths
- **Permission Denied**: Explain permission requirements and solutions
- **File System Errors**: Handle disk space, corruption, and access issues
- **Configuration Errors**: Validate and handle malformed configuration
- **Reset Errors**: Handle errors during configuration reset
- **Path Expansion Errors**: Return specific error type with helpful message suggesting absolute paths if home directory not found

#### User Input Errors
- **Invalid Input**: Handle malformed user input gracefully
- **Empty Input**: Use default values when user provides empty input
- **Cancellation**: Handle Ctrl+C and other interruption signals
- **Timeout**: Handle cases where user doesn't respond
- **Validation Failure**: Exit with error and require user to run init again

### Dependencies

#### Direct Dependencies
- `bufio`: Interactive user input handling
- `os`: File system operations and user input
- `path/filepath`: Cross-platform path handling

#### Indirect Dependencies
- Go standard library: File system operations, path handling
- Existing packages: `pkg/fs`, `pkg/config`, `pkg/status`, `pkg/logger`

### Migration Strategy

#### Backward Compatibility
- **Existing Configurations**: Detect and preserve existing configurations
- **Status File Migration**: Add `initialized` field to existing status files (defaults to false)
- **Default Path Changes**: Handle migration from old default paths

#### Upgrade Path
- **Automatic Detection**: Detect uninitialized CM instances (missing or false initialized field)
- **Graceful Migration**: Migrate existing configurations without data loss
- **User Notification**: Inform users about configuration changes

## Success Criteria

1. **First-Time Setup**: New users can successfully initialize CM with interactive prompts
2. **Default Values**: Users can initialize with default values using `--default` flag
3. **Reset Functionality**: Users can reset configuration with `--reset` flag and confirmation
4. **Configuration Creation**: Config files are created with correct paths and permissions
5. **Directory Structure**: Base path directory is created with proper permissions (worktrees created when needed)
6. **Status Initialization**: Status file is created with `initialized: true`
7. **Path Validation**: Invalid paths are detected and rejected with helpful messages
8. **Re-Initialization Prevention**: Already initialized CM instances are detected (unless reset)
9. **Error Handling**: All error conditions are handled gracefully with clear messages
10. **Cross-Platform**: Works consistently on Windows, macOS, and Linux
11. **Testing**: Comprehensive test coverage for all initialization scenarios

## Blocking Dependencies

- **Feature 4**: Validate project structure and Git configuration
- **Feature 8**: Implement status YAML structure

## Blocked Features

- **Feature 9**: Create worktrees for single repositories (requires initialization)
- **Feature 10**: Integrate IDE opening functionality (requires initialization)
- **Feature 11**: Implement worktree deletion (requires initialization)
- **Feature 12**: List worktrees for single repositories (requires initialization)
- **Feature 13**: Implement load PR branches (requires initialization)
- **Feature 14**: Create worktrees for workspaces (requires initialization)
- **Feature 15**: Create worktrees from GitHub issues (requires initialization)

## Questions for Clarification

1. **Default Base Path**: Should the default base path be `~/Code` or would you prefer a different default?
   - **Answer**: Keep `~/Code` as default, include examples in prompt

2. **Interactive Prompts**: Should we provide suggestions for common code directory locations (e.g., `~/Code`, `~/Projects`, `~/Development`)?
   - **Answer**: Provide examples in prompt, user types their choice or uses default if empty

3. **Configuration Location**: Should the config file always be at `~/.cm/config.yaml` or should users be able to choose a different location?
   - **Answer**: Fixed at `~/.cm/config.yaml`

4. **Worktrees Directory**: You mentioned worktrees_dir should be fixed to `$base_path/worktrees`. Should this be completely non-configurable or should we allow users to override it if needed?
   - **Answer**: Completely non-configurable, computed as `$base_path/worktrees`

5. **Status File Location**: Should the status file always remain at `~/.cm/status.yaml` or should it be configurable relative to the base path?
   - **Answer**: Fixed at `~/.cm/status.yaml` (non-configurable)

6. **Re-Initialization**: Should we provide a way to re-initialize CM (e.g., `cm init --force`) or should initialization be a one-time process?
   - **Answer**: Use `--force` flag to skip interactive confirmation for reset

7. **Migration**: How should we handle existing CM installations that don't have the `initialized` field in their status file?
   - **Answer**: Consider as `false` (uninitialized)

8. **Error Recovery**: Should we provide a way to reset CM configuration if something goes wrong during initialization?
   - **Answer**: Exit with error and require user to run init again

9. **Base Path Flag**: Should we add a `--base-path` flag to skip interactive prompts?
   - **Answer**: Yes, add `--base-path <path>` flag to set base path directly

10. **Path Expansion**: Should `~` expansion be integrated into existing FS methods?
    - **Answer**: Yes, create helper function and integrate into existing methods

11. **Config File Handling**: Should init update existing config files or only create new ones?
    - **Answer**: Update existing config files with new base_path

12. **Directory Creation**: Should init create the base_path directory if it doesn't exist?
    - **Answer**: Yes, create base_path directory if it doesn't exist

13. **Error Types**: Should we create specific error types for initialization errors?
    - **Answer**: Yes, create `ErrAlreadyInitialized` and other specific error types
