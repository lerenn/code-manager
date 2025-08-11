# Feature 004: Validate Project Structure and Git Configuration

## Overview
Implement comprehensive validation for both single repository and workspace modes to ensure all Git repositories are properly configured and accessible before proceeding with worktree operations.

## Background
The Cursor Git WorkTree Manager (cgwt) needs to validate that the project structure and Git configuration are in a working state before creating or managing worktrees. This validation serves as a prerequisite for all worktree operations and ensures that the tool can safely interact with Git repositories.

## Requirements

### Functional Requirements
1. **Single Repository Validation**: Validate that the current directory is a working Git repository
2. **Workspace Repository Validation**: Validate that all repositories in a workspace are working Git repositories
3. **Git Configuration Validation**: Ensure Git is properly configured and working (basic functionality check)
4. **Error Handling**: Provide clear error messages when validation fails
5. **Integration Ready**: Be called automatically by the existing `Run()` method
6. **Fail-Fast Behavior**: Stop program execution with clear error messages on validation failure

### Non-Functional Requirements
1. **Performance**: Validation should be fast (< 200ms for typical projects)
2. **Reliability**: Handle edge cases (broken Git repos, permission issues, etc.)
3. **Cross-Platform**: Work on Windows, macOS, and Linux
4. **Testability**: Use Uber gomock for mocking file system operations
5. **Minimal Dependencies**: Use only Go standard library + gomock for file system operations

## Technical Specification

### Interface Design

#### FS Package Extension (File System Adapter)
**New Interface Methods:**
- `GetCurrentDir() (string, error)`: Get current working directory

**Key Characteristics:**
- Extends existing FS interface with current directory operation
- Pure function-based operations
- No state management required
- Cross-platform compatibility
- **Single source of truth** for all file system operations

#### Git Package (Git Operations Adapter)
**New Package:** `pkg/git/`

**Interface Design:**
- `Status(workDir string) (string, error)`: Execute `git status` in specified directory
- `ConfigGet(workDir, key string) (string, error)`: Execute `git config --get <key>` in specified directory

**Key Characteristics:**
- **ALL Git operations** must be in this adapter
- **ONLY this package** should have integration tests with real Git commands
- Pure function-based operations
- No state management required
- Cross-platform compatibility
- **Single source of truth** for all Git operations

#### CGWT Package Extension (Business Logic)
**New Interface Methods:**
- `validateProjectStructure() error`: Main validation method
- `validateSingleRepository() error`: Validate single repository mode
- `validateWorkspaceRepositories() error`: Validate workspace repositories
- `validateGitConfiguration() error`: Validate Git is working properly

**New Dependencies:**
- Inject Git adapter alongside FS adapter
- Use Git adapter for all Git operations
- Use FS adapter for all file system operations
- Use Logger interface for verbose output
- Import logger package for logging functionality

**New Package:** `pkg/logger/`

**Interface Design:**
```go
type Logger interface {
    Logf(format string, args ...interface{})
}

type noopLogger struct{}

func (n *noopLogger) Logf(format string, args ...interface{}) {}

type defaultLogger struct {
    mu sync.Mutex
}

func (d *defaultLogger) Logf(format string, args ...interface{}) {
    d.mu.Lock()
    defer d.mu.Unlock()
    fmt.Printf(format+"\n", args...)
}
```

**Key Characteristics:**
- **ALL logging code** must be in this package
- **Thread-safe** implementation with mutex
- Pure function-based operations
- No state management required
- Cross-platform compatibility
- **Single source of truth** for all logging operations

**Implementation Structure:**
- Extends existing CGWT package with validation logic
- Private helper methods for different validation types
- Error handling with wrapped errors
- Clean separation of concerns
- Called automatically by existing `Run()` method

**Key Characteristics:**
- **NO direct file system access** - all operations go through FS adapter
- **ONLY unit tests** using mocked FS adapter
- Business logic focused on validation
- Testable through dependency injection
- **Pure business logic** with no file system dependencies
- **Fail-fast** - stops execution on first validation failure

### Implementation Details

#### 1. FS Package Extension
The FS package extends with current directory operation:

**Key Components:**
- New method: `GetCurrentDir()`
- Concrete implementation using Go standard library (`os`)
- Updated `//go:generate` directive for Uber mockgen
- No state - pure function-based operations

**Implementation Notes:**
- Use `os.Getwd()` for current directory
- Handle cross-platform path operations properly
- Update mock generation to include new method
- Generate mock files as `mockfs.gen.go` in same directory
- **ALL file system operations** must be implemented here
- **Integration tests only** - no unit tests for this adapter

#### 2. Logger Package Implementation
The Logger package provides logging capabilities:

**Key Components:**
- Interface with methods: `Logf()`
- Concrete implementations: `noopLogger`, `defaultLogger`
- Thread-safe implementation with mutex
- No state - pure function-based operations

**Implementation Notes:**
- Use `sync.Mutex` for thread safety in `defaultLogger`
- Use `fmt.Printf()` for output in `defaultLogger`
- Provide constructor functions: `NewNoopLogger()`, `NewDefaultLogger()`
- **ALL logging code** must be implemented here
- **Integration tests only** - no unit tests for this adapter
- **File structure**: `logger.go` (interface and implementations), `logger_test.go` (integration tests)

#### 3. Git Package Implementation
The Git package provides Git command execution capabilities:

**Key Components:**
- Interface with methods: `Status()`, `ConfigGet()`
- Concrete implementation using Go standard library (`os/exec`)
- `//go:generate` directive for Uber mockgen
- No state - pure function-based operations

**Implementation Notes:**
- Use `exec.Command()` for Git command execution
- Use `exec.Command().Output()` for command output and convert to string
- Allow specifying different working directory for each Git command
- Handle cross-platform command execution properly
- Include both command and output in error context for debugging using `fmt.Errorf("git command failed: %w (command: %s, output: %s)", err, cmd, output)`
- Return empty string for Git commands with no output (like missing config keys)
- Add `//go:generate go run go.uber.org/mock/mockgen@v0.5.2 -source=git.go -destination=mockgit.gen.go -package=git` directive
- Generate mock files as `mockgit.gen.go` in same directory
- **ALL Git operations** must be implemented here
- **Integration tests only** - no unit tests for this adapter

#### 4. CGWT Package Extension
The CGWT package extends with comprehensive validation logic:

**Key Components:**
- `validateProjectStructure()`: Main validation orchestrator
- `validateSingleRepository()`: Single repository validation
- `validateWorkspaceRepositories()`: Workspace repository validation
- `validateGitConfiguration()`: Git configuration validation
- Integration with existing `Run()` method
- Dependency injection of both FS and Git adapters

**Implementation Notes:**
- `validateProjectStructure()` determines mode and calls appropriate validation
- `validateSingleRepository()` checks current directory is a working Git repo
- `validateWorkspaceRepositories()` validates all repositories in workspace
- `validateGitConfiguration()` ensures Git is properly configured and working
- Use `fmt.Errorf()` with `%w` for error wrapping
- Provide clear user messages for different validation states
- **Fail-fast** - stop execution on first validation failure
- Support verbose mode for detailed validation steps
- **NO direct file system access** - all operations go through FS adapter
- **NO direct Git access** - all operations go through Git adapter
- **Verbose mode shows**: which repositories are being validated, exact Git commands being executed, resolved paths for workspace repositories
- **Verbose mode integration**: Field in CGWT struct and separate logging interface
- **Validation methods use struct fields**: No verbose parameter needed in method signatures
- **Verbose output examples**: "Validating repository: /path/to/repo", "Executing git status in: /path/to/repo", "Resolved workspace path: /path/to/repo"
- **Error logging**: Log error details before returning them using same format as verbose messages
- **Independent verbose checks**: Each validation method checks verbose flag independently
- **Thread-safe logging**: Logger interface must be thread-safe
- **Default logger**: Provides default verbose logger that writes to stdout
- **Run method support**: Existing `Run()` method also supports verbose mode
- **Consistent message format**: All verbose messages follow consistent format
- **All methods support verbose**: Existing detection methods (Features 001, 002, 003) also support verbose mode
- **Update existing methods**: Existing detection methods should be updated to include verbose logging

#### 4. Validation Logic

**Single Repository Validation:**
- Check current directory contains `.git` folder (using FS adapter)
- Verify `.git` is a directory (using FS adapter)
- Execute `git status` to ensure repository is working (using Git adapter)
- Validate Git configuration is functional (using Git adapter)

**Workspace Repository Validation:**
- Parse workspace configuration (from Feature 002)
- For each repository in workspace:
  - Resolve relative path using `filepath.Join()` from workspace file directory (using `filepath.Dir()`)
  - Check repository path exists (using FS adapter)
  - Verify path contains `.git` folder (using FS adapter)
  - Execute `git status` in repository directory to ensure repository is working (using Git adapter)
  - Validate Git configuration is functional (using Git adapter)
- **Fail if ANY repository validation fails** - all listed repositories must be valid
- **Check ALL folders** listed in `.code-workspace` file, even if some don't exist
- **Resolve relative paths from workspace file location** (not current directory)
- **Workspace file always exists** - no need to handle missing workspace file case
- Provide specific error messages for each invalid/missing repository

**Git Configuration Validation:**
- Execute `git status` to ensure basic Git functionality (using Git adapter)
- Provide clear error messages for missing configuration
- Include both command and output in error context for debugging

### Error Handling

#### Error Types
1. **ValidationError**: When project structure validation fails
2. **GitConfigurationError**: When Git configuration is invalid or missing
3. **RepositoryError**: When a repository is not accessible or corrupted
4. **WorkspaceError**: When workspace configuration is invalid
5. **PermissionError**: When unable to access directories due to permissions

#### Error Messages
- **Single Repository**: "Not a valid Git repository: <reason>"
- **Workspace Repository**: "Invalid repository in workspace: <path> - <reason>"
- **Missing Repository**: "Repository not found in workspace: <path>"
- **Git Configuration**: "Git configuration error: <reason> (command: <cmd>, output: <output>)"
- **General**: "Project structure validation failed: <reason>"
- **Git Command Error**: "git command failed: <error> (command: <cmd>, output: <output>)"

### Integration with Existing Features

#### Run() Method Integration
The validation should be integrated into the existing `Run()` method:

```go
type CGWT struct {
    fs      fs.FS
    git     git.Git
    verbose bool
    logger  logger.Logger
    // ... other fields
}

func NewCGWT() *CGWT {
    return &CGWT{
        fs:      fs.NewFS(),
        git:     git.NewGit(),
        verbose: false,
        logger:  logger.NewNoopLogger(),
    }
}



func (c *CGWT) SetVerbose(verbose bool) {
    c.verbose = verbose
    if verbose && c.logger == logger.NewNoopLogger() {
        c.logger = logger.NewDefaultLogger()
    } else if !verbose {
        c.logger = logger.NewNoopLogger()
    }
}

func (c *CGWT) SetLogger(logger logger.Logger) {
    c.logger = logger
}

func (c *CGWT) Run() error {
    if c.verbose {
        c.logger.Logf("Starting CGWT execution")
    }
    
    // Existing detection logic (Features 001, 002, 003)
    if err := c.detectProjectMode(); err != nil {
        if c.verbose {
            c.logger.Logf("Error: %v", err)
        }
        return err
    }
    
    // NEW: Validation logic (Feature 004)
    if err := c.validateProjectStructure(); err != nil {
        if c.verbose {
            c.logger.Logf("Error: %v", err)
        }
        return err
    }
    
    if c.verbose {
        c.logger.Logf("CGWT execution completed successfully")
    }
    
    // Continue with existing logic...
}
```

#### Mode Detection Integration
Validation should work with both detection modes:
- **Single Repository Mode**: Validate current directory
- **Workspace Mode**: Validate all repositories in workspace
- **Multiple Workspace Files**: Validate repositories in selected workspace

### Testing Strategy

#### Unit Tests (CGWT Package)
- Mock FS adapter for all file system operations
- Mock Git adapter for all Git operations
- Override adapters after `NewCGWT()` call in tests
- Test verbose mode and logging functionality
- Test validation logic with various scenarios:
  - Valid single repository
  - Valid workspace with multiple repositories
  - Invalid Git repository
  - Missing repository in workspace
  - Missing Git configuration
  - Corrupted workspace file
  - Permission errors
- Use `//go:build unit` tag

#### Integration Tests (FS Package)
- Real file system operations
- Test with actual file system
- Use `//go:build integration` tag

#### Integration Tests (Git Package)
- Real Git command execution
- Test with actual Git repositories
- Test with actual workspace files
- Use `//go:build integration` tag

### Success Criteria
1. **Single Repository Mode**: Successfully validates working Git repository
2. **Workspace Mode**: Successfully validates all repositories in workspace
3. **Error Handling**: Provides clear error messages for validation failures
4. **Integration**: Seamlessly integrates with existing `Run()` method
5. **Performance**: Validation completes within 200ms for typical projects
6. **Reliability**: Handles edge cases gracefully with appropriate error messages

### Dependencies
- **Blocked by**: Features 001, 002, 003 (project mode detection)
- **Blocks**: Feature 005 (directory structure implementation)
- **Uses**: FS package for all file system operations
- **Uses**: Git package for all Git operations
- **Uses**: Logger package for all logging operations
- **Extends**: CGWT package with validation capabilities
