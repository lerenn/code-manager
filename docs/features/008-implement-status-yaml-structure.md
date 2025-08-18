# Feature 008: Implement Status YAML Structure

## Overview
Implement functionality to create and manage a `status.yaml` file that tracks worktrees present in `$HOME/.cm` and links them to their original repositories. This provides a centralized registry of all worktrees and their metadata for future worktree management features.

## Background
The Code Manager (cm) needs a persistent state tracking mechanism to maintain information about created worktrees and their relationships to original repositories. This status file will serve as the authoritative source for worktree locations, branch information, and workspace associations, enabling efficient worktree management operations.

## Requirements

### Functional Requirements
1. **Status File Management**: Create and maintain `status.yaml` at configurable location (default: `$HOME/.cm/status.yaml`)
2. **Worktree Tracking**: Track worktrees with repository name, branch, local path, and optional workspace association
3. **Atomic Operations**: Implement file locking mechanism to prevent concurrent access issues
4. **Automatic Creation**: Create status file automatically if it doesn't exist
5. **Update Triggers**: Update status file right before worktree creation and after worktree deletion
6. **Repository Name Handling**: Use full repository names (e.g., "github.com/lerenn/example") as unique identifiers
7. **Workspace Support**: Track workspace file associations for multi-repo workspaces
8. **Path Validation**: Ensure all tracked paths are valid and accessible
9. **Error Recovery**: Handle corruption and provide recovery mechanisms
10. **Configuration Integration**: Support configurable status file location via existing config system

### Non-Functional Requirements
1. **Performance**: Status file operations should be fast (< 100ms for read/write)
2. **Reliability**: Handle file system errors, corruption, and concurrent access
3. **Cross-Platform**: Work on Windows, macOS, and Linux
4. **Testability**: Use Uber gomock for mocking file system operations
5. **Minimal Dependencies**: Use only Go standard library + gomock for file operations
6. **Atomic Writes**: Ensure status file updates are atomic to prevent corruption
7. **Backup Safety**: Maintain data integrity during write operations

## Technical Specification

### Data Structure

#### Status File Format
```yaml
repositories:
  - name: github.com/lerenn/example
    branch: feature-a
    path: /Users/lfradin/Code/example
  - name: github.com/lerenn/example-for-workspace
    branch: feature-b 
    path: /Users/lfradin/Code/example-for-workspace
    workspace: /Users/lfradin/Code/example.code-workspace
  - name: github.com/lerenn/example2-for-workspace
    branch: feature-b 
    path: /Users/lfradin/Code/example2-for-workspace
    workspace: /Users/lfradin/Code/example.code-workspace
```

#### Go Structures
```go
type Status struct {
    Repositories []Repository `yaml:"repositories"`
}

type Repository struct {
    Name      string `yaml:"name"`
    Branch    string `yaml:"branch"`
    Path      string `yaml:"path"`
    Workspace string `yaml:"workspace,omitempty"`
}
```

### Interface Design

#### Status Package (New Package)
**Package Location**: `pkg/status/`
**Key Components**:
- `Status` and `Repository` structs with YAML tags
- `Manager` interface for status file operations
- `realManager` implementation with file locking
- `AddWorktree(repoName, branch, worktreePath, workspacePath string) error`: Add worktree to status
- `RemoveWorktree(repoName, branch string) error`: Remove worktree from status
- `GetWorktree(repoName, branch string) (*Repository, error)`: Get specific worktree
- `ListAllWorktrees() ([]Repository, error)`: List all worktrees

**Key Characteristics**:
- Pure status management with file system operations
- YAML-based status file format
- Atomic write operations with temporary files
- File locking for concurrent access prevention
- Error handling with wrapped errors
- No business logic - pure data management

#### FS Package Extension (File System Adapter)
**New Interface Methods**:
- `WriteFileAtomic(filename string, data []byte, perm os.FileMode) error`: Atomic file write using temporary files
- `FileLock(filename string) (func(), error)`: Acquire file lock, return unlock function
- `CreateFileIfNotExists(filename string, initialContent []byte, perm os.FileMode) error`: Create file with initial content

**Key Characteristics**:
- Extends existing FS package with atomic file operations
- Pure function-based operations
- No state management required
- Cross-platform compatibility
- File locking implementation using system calls

#### Config Package Extension
**New Config Fields**:
```go
type Config struct {
    BasePath   string `yaml:"base_path"`
    StatusFile string `yaml:"status_file"` // New field
}
```

**Updated Methods**:
- `DefaultConfig()`: Include default status file path (`$HOME/.cm/status.yaml`)

#### CM Package Extension (Business Logic)
**New Interface Methods**:
- `AddWorktreeToStatus(repoName, branch, worktreePath, workspacePath string) error`: Add worktree to status
- `RemoveWorktreeFromStatus(repoName, branch string) error`: Remove worktree from status
- `GetWorktreeStatus(repoName, branch string) (*status.Repository, error)`: Get worktree status
- `ListAllWorktrees() ([]status.Repository, error)`: List all tracked worktrees

**Updated Constructor**:
- `NewCM(cfg *config.Config) CM`: Creates Status Manager internally

**Implementation Structure**:
- Extends existing CM package with status management
- Integration with existing `Run()` method for status updates
- Business logic for worktree status tracking
- Clean separation of concerns
- Error handling with wrapped errors

**Key Characteristics**:
- **NO direct file system access** - all operations go through Status Manager
- **ONLY unit tests** using mocked Status Manager
- Business logic focused on worktree status tracking
- Testable through dependency injection
- Methods will be called from future worktree creation/deletion features
- Status Manager created internally with shared FS instance

### Implementation Details

#### 1. Status Package
The Status package manages the status.yaml file:

**Key Components**:
- `Status` and `Repository` structs with YAML tags
- `Manager` interface for status operations
- `realManager` implementation with file locking
- Atomic file operations using temporary files
- File locking for concurrent access prevention

**File Operations**:
- Atomic writes using temporary files and rename
- File locking using system-specific mechanisms
- Automatic file creation with initial structure
- Error recovery and validation

#### 2. FS Package Extension
Extends the existing FS package with atomic file operations:

**New Methods**:
- `WriteFileAtomic`: Write file atomically using temporary file and rename
- `FileLock`: Acquire file lock for concurrent access prevention
- `CreateFileIfNotExists`: Create file with initial content if not exists

**Implementation**:
- Cross-platform file locking (flock on Unix, LockFileEx on Windows)
- Temporary file creation and atomic rename
- Proper error handling and cleanup

#### 3. Config Package Extension
Extends configuration to support status file path:

**New Fields**:
- `StatusFile` field for configurable status file location
- Default value: `$HOME/.cm/status.yaml`

**Updated Methods**:
- `DefaultConfig`: Include default status file path

#### 4. CM Package Extension
Extends business logic with status management:

**New Methods**:
- `AddWorktreeToStatus`: Add worktree entry to status file
- `RemoveWorktreeFromStatus`: Remove worktree entry from status file
- `GetWorktreeStatus`: Retrieve specific worktree status
- `ListAllWorktrees`: List all tracked worktrees

**Integration**:
- Status Manager dependency injection
- Integration with future worktree creation/deletion features
- Error handling and validation

### File Structure

```
pkg/
├── status/
│   ├── status.go          # Status structs and Manager interface
│   ├── status_test.go     # Unit tests with mocked dependencies
│   └── mockstatus.gen.go  # Generated mock for testing
├── fs/
│   ├── fs.go                    # Extended with atomic file operations
│   ├── fs_test.go               # Integration tests for file operations
│   ├── fs_integration_test.go   # Integration tests for new methods
│   └── mockfs.gen.go            # Generated mock for testing
├── config/
│   ├── config.go                # Extended with StatusFile field
│   ├── config_test.go           # Unit tests for config
│   └── mockconfig.gen.go        # Generated mock for testing
└── cm/
    ├── cm.go                  # Extended with status management
    ├── cm_test.go             # Unit tests with mocked dependencies
    ├── status_test.go           # Unit tests for status functionality
    └── mockcm.gen.go          # Generated mock for testing
```

### Testing Strategy

#### Unit Tests (Business Logic)
- **Status Package**: Mock FS adapter for file operations
- **CM Package**: Mock Status Manager for status operations
- **Config Package**: Mock file system for config operations

#### Integration Tests (Adapters)
- **FS Package**: Real file system operations with cleanup
- **New FS Methods**: Integration tests for atomic operations and file locking

#### Test Scenarios
1. **Status File Creation**: Test automatic creation of status file
2. **Atomic Operations**: Test atomic file writes and error recovery
3. **File Locking**: Test concurrent access prevention
4. **Data Integrity**: Test corruption handling and recovery
5. **Repository Operations**: Test add/remove/list operations
6. **Error Handling**: Test various error conditions and recovery

### Error Handling

#### Status File Errors
- **File Not Found**: Create with initial structure
- **Permission Denied**: Return clear error with path information
- **Corruption**: Attempt recovery or return error
- **Concurrent Access**: Use file locking to prevent conflicts

#### Data Validation
- **Invalid Paths**: Validate all paths are accessible
- **Duplicate Entries**: Prevent duplicate repository/branch combinations
- **Missing Fields**: Validate required fields are present

#### Recovery Mechanisms
- **Backup Creation**: Create backup before major operations
- **Rollback Support**: Support rolling back failed operations
- **Data Validation**: Validate data integrity after operations

### Performance Considerations

#### File Operations
- **Atomic Writes**: Use temporary files for atomic operations
- **Minimal I/O**: Read/write only when necessary
- **Efficient Locking**: Use appropriate lock granularity

#### Memory Usage
- **Streaming Parsing**: Use streaming YAML parsing for large files
- **Lazy Loading**: Load status only when needed
- **Efficient Data Structures**: Use appropriate data structures for lookups

### Security Considerations

#### File Permissions
- **Restrictive Permissions**: Use 0600 for status file
- **Directory Permissions**: Use 0700 for status directory
- **Path Validation**: Validate all paths to prevent directory traversal

#### Data Validation
- **Input Sanitization**: Sanitize all input data
- **Path Validation**: Validate all file paths
- **YAML Safety**: Use safe YAML parsing

### Future Integration

#### Worktree Management Features
- **Feature 11**: Create worktrees for single repositories
- **Feature 12**: Create worktrees for multi-repo workspaces
- **Feature 17**: List worktrees for current project
- **Feature 18**: List all worktrees across projects

#### Status File Evolution
- **Versioning**: Add version field for future compatibility
- **Migration Support**: Support migrating between status file versions
- **Backup Strategy**: Implement backup and restore functionality

### Dependencies

#### Direct Dependencies
- `gopkg.in/yaml.v3`: YAML parsing and marshaling
- `go.uber.org/mock`: Mock generation for testing

#### Indirect Dependencies
- Go standard library: File system operations, path handling, syscall
- Existing packages: `pkg/fs`, `pkg/config`, `pkg/logger`

### Migration Strategy

#### Backward Compatibility
- **Default Values**: Provide sensible defaults for new fields
- **Optional Fields**: Make new fields optional in YAML
- **Validation**: Validate existing status files during loading

#### Upgrade Path
- **Automatic Migration**: Migrate existing status files automatically
- **Backup Creation**: Create backups before migration
- **Rollback Support**: Support rolling back failed migrations

## Success Criteria

1. **Status File Creation**: Status file is created automatically at first use
2. **Atomic Operations**: All status file operations are atomic and safe
3. **Concurrent Access**: File locking prevents concurrent access issues
4. **Data Integrity**: Status file maintains data integrity under all conditions
5. **Performance**: Status operations complete within 100ms
6. **Error Handling**: All error conditions are handled gracefully
7. **Testing**: Comprehensive test coverage for all components
8. **Integration**: Seamless integration with existing CM functionality
9. **Future Ready**: Status file structure supports future worktree management features
10. **Cross-Platform**: Works consistently across Windows, macOS, and Linux

## Blocking Dependencies

- **Feature 4**: Validate project structure and Git configuration
- **Feature 5**: Implement repos directory structure

## Blocked Features

- **Feature 9**: Workspace name extraction from `.code-workspace` files
- **Feature 10**: Multi-repo workspace support
- **Feature 11**: Create worktrees for single repositories
- **Feature 12**: Create worktrees for multi-repo workspaces
- **Feature 17**: List worktrees for current project
- **Feature 18**: List all worktrees across projects
- **Feature 21**: Safe deletion with confirmation
