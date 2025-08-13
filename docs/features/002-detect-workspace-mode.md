# Feature 002: Detect Workspace Mode

## Overview
Implement functionality to detect when the current working directory contains a `.code-workspace` file, indicating a multi-repository workspace configuration for VS Code like IDEs.

## Background
The Git WorkTree Manager (wtm) needs to distinguish between single repository mode and workspace mode. Workspace mode is indicated by the presence of a `.code-workspace` file, which contains configuration for managing multiple repositories within a single workspace. This feature builds upon the single repository detection from Feature 001.

## Requirements

### Functional Requirements
1. **Workspace File Detection**: Detect if the current working directory contains a `.code-workspace` file
2. **Workspace File Validation**: Ensure the `.code-workspace` file is a valid JSON file
3. **Workspace Configuration Parsing**: Parse the workspace configuration to identify repository paths
4. **Repository Path Validation**: Validate that all repository paths in the workspace exist and are Git repositories
5. **Error Handling**: Provide clear error messages for invalid workspace configurations
6. **Integration Ready**: Return detection results in a format suitable for other features

### Non-Functional Requirements
1. **Performance**: Detection and parsing should be fast (< 200ms for typical workspaces)
2. **Reliability**: Handle edge cases (malformed JSON, missing repositories, etc.)
3. **Cross-Platform**: Work on Windows, macOS, and Linux
4. **Testability**: Use Uber gomock for mocking file system operations
5. **Minimal Dependencies**: Use only Go standard library + gomock for file system operations

## Technical Specification

### Interface Design

#### FS Package Extension (File System Adapter)
**New Interface Methods:**
- `ReadFile(path string) ([]byte, error)`: Read file contents
- `ReadDir(path string) ([]os.DirEntry, error)`: Read directory contents
- `Glob(pattern string) ([]string, error)`: Find files matching pattern (for `*.code-workspace`)

**Key Characteristics:**
- **ALL file system access** (list, read, write, exists, etc.) must be in this adapter
- **ONLY this package** should have integration tests with real file system
- Pure function-based operations
- No state management required
- Cross-platform compatibility
- **Single source of truth** for all file system operations

#### WTM Package Extension (Business Logic)
**New Interface Methods:**
- `detectWorkspaceMode() ([]string, error)`: Detect workspace files (returns list of files, no error for multiple files)
- `getWorkspaceInfo(workspaceFile string) (*WorkspaceConfig, error)`: Parse and validate workspace configuration

**New Data Structures:**
```go
type WorkspaceConfig struct {
    Name string `json:"name,omitempty"`
    Folders []WorkspaceFolder `json:"folders"`
    Settings map[string]interface{} `json:"settings,omitempty"`
    Extensions map[string]interface{} `json:"extensions,omitempty"`
}

type WorkspaceFolder struct {
    Name string `json:"name,omitempty"`
    Path string `json:"path"`
}
```

**Implementation Structure:**
- Extends existing WTM package with workspace detection
- Private helper method: `detectWorkspaceMode()`
- Private helper method: `parseWorkspaceFile()`
- Private helper method: `validateWorkspaceRepositories()`
- Error handling with wrapped errors
- Clean separation of concerns

**Key Characteristics:**
- **NO direct file system access** - all operations go through FS adapter
- **ONLY unit tests** using mocked FS adapter
- Business logic focused on workspace detection and validation
- Testable through dependency injection
- **Pure business logic** with no file system dependencies

### Implementation Details

#### 1. FS Package Extension
The FS package extends with file reading operations:

**Key Components:**
- New methods: `ReadFile()`, `ReadDir()`, `Glob()`
- Concrete implementation using Go standard library (`os`, `io`, `path/filepath`)
- Updated `//go:generate` directive for Uber mockgen
- No state - pure function-based operations

**Implementation Notes:**
- Use `os.ReadFile()` for file reading operations
- Use `os.ReadDir()` for directory listing
- Use `filepath.Glob()` for pattern matching (`*.code-workspace`)
- Handle cross-platform path separators properly
- Update mock generation to include new methods
- Generate mock files as `mockfs.gen.go` in same directory
- **ALL file system operations** must be implemented here
- **Integration tests only** - no unit tests for this adapter

#### 2. WTM Package Extension
The WTM package extends with workspace detection capabilities:

**Key Components:**
- Extended `Run()` method to detect both single repo and workspace modes
- Private helper methods: `detectWorkspaceMode()`, `parseWorkspaceFile()`, `validateWorkspaceRepositories()`
- New data structures for workspace configuration
- Enhanced error handling with wrapped errors

**Implementation Notes:**
- `Run()` orchestrates detection flow: first check for single repo (`.git`), then workspace (`*.code-workspace`)
- `Run()` handles multiple workspace files error (fails if more than 1)
- `detectWorkspaceMode()` detects workspace files and returns list (no error for multiple files)
- `getWorkspaceInfo()` parses and validates workspace configuration from a specific file
- `parseWorkspaceFile()` handles JSON parsing of workspace configuration
- `validateWorkspaceRepositories()` validates all repository paths in workspace (strict validation)
- Use `encoding/json` for JSON parsing with strict validation
- Use `fmt.Errorf()` with `%w` for error wrapping
- Provide clear user messages for different detection states
- Support quiet mode (only errors to stderr), verbose mode (detailed steps), and normal mode (user interaction only)
- **Normal mode**: Display workspace name from JSON config if available, otherwise use filename without `.code-workspace` extension (e.g., "Found workspace: MyProject")
- **Verbose mode**: Show workspace configuration details, repository names, resolved paths, and significant steps (e.g., "Checking for .code-workspace files", "Parsing workspace configuration", "Validating repository paths")
- Error messages can include additional context in verbose mode before the actual error
- **Always detect before any actions** to avoid unstable state
- **NO direct file system calls** - all operations through FS adapter interface
- **Unit tests only** - use mocked FS adapter for all testing
- **Fail on first error** encountered during validation
- **Resolve relative paths** relative to current working directory using `filepath.Join()`
- **Resolve symlinks** in all paths using `filepath.EvalSymlinks()` (fail immediately on broken symlinks)
- **Validate unique paths** after resolution using `filepath.Clean()` for normalized separators (fail if duplicates found)
- **No security validation** for resolved paths (allow paths outside current directory)
- **Multiple workspace files**: Throw fatal error with count (selection is for task 3)
- **Allow additional fields** in workspace configuration (don't validate unknown fields)
- **Validate folders array** is an array (not null, not object)
- **Validate folder objects** are objects (not primitive values)
- **Skip null values** in folders array during validation
- **Validate array not empty** after skipping null values
- **Keep Glob simple**: Only search for `*.code-workspace` in current directory, return full paths
- **Extract workspace name**: Use workspace name from JSON config if available, otherwise use `strings.TrimSuffix(filepath.Base(filename), ".code-workspace")`

#### 3. Workspace Detection Algorithm
1. **First**: Check current working directory for `.git` folder (single repository mode)
2. **If no `.git` found**: Check for files matching `*.code-workspace` pattern in current directory only (workspace mode)
3. **If multiple workspace files found**: Throw fatal error with count (multiple workspace selection is for task 3)
4. **If single workspace file found**: Parse and validate the JSON configuration
5. **Extract repository paths** from the `folders` array (skip null values)
6. **Validate folders array** is not empty after skipping null values
7. **Resolve relative paths** relative to current working directory using `filepath.Join()`
8. **Resolve symlinks** in all paths using `filepath.EvalSymlinks()` (fail immediately on broken symlinks)
9. **Validate unique paths** after resolution using `filepath.Clean()` for normalized separators (fail if duplicates found)
10. **Validate that each repository path** exists and contains a `.git` directory
11. **Return workspace configuration** if all repositories are valid (strict validation)
12. **Fail entire workspace detection** if any repository is invalid (fail on first error)

### Error Handling

#### Error Types
1. **NoWorkspaceFileError**: When no `.code-workspace` files are found
2. **MultipleWorkspaceFilesError**: When multiple `.code-workspace` files are found in the same directory
3. **InvalidWorkspaceFileError**: When the workspace file is not valid JSON
4. **EmptyFoldersError**: When the `folders` array is empty
5. **InvalidFolderStructureError**: When `folders` array contains objects without `path` field
6. **MissingRepositoryError**: When a repository path in the workspace doesn't exist
7. **InvalidRepositoryError**: When a repository path doesn't contain a Git repository
8. **PermissionError**: When unable to access files due to permissions

#### Error Messages
- `"no .code-workspace files found in current directory"`
- `"%d .code-workspace files found in current directory"`
- `"invalid .code-workspace file: malformed JSON"`
- `"workspace file must contain non-empty folders array"`
- `"workspace folder must contain path field"`
- `"workspace folder path field must be non-empty string"`
- `"workspace folder name field must be string if present"`
- `"workspace repository not found: %s"`
- `"workspace repository is not a git repository: %s"`
- `"permission denied: cannot access workspace file"`
- `"duplicate repository paths found after resolution: %s"`
- `"broken symlink detected in workspace repository: %s"`

### Integration Points

#### 1. Main Application Flow
**Key Components:**
- Enhanced Cobra root command setup
- Dependency injection: FS adapter → WTM manager
- Error handling with `log.Fatal()`
- Clean separation of concerns

**Implementation Notes:**
- Extend existing Cobra command structure
- Create FS adapter and WTM manager in the command's `RunE` function
- Handle errors at the top level with `log.Fatal()` and exit code 1
- Keep main function simple and focused on orchestration
- Maintain existing global quiet mode and verbose mode flags
- Normal mode shows only user interaction messages
- Extend `create` subcommand to handle workspace mode (branch name argument not used yet)
- Help text: "Create worktree(s) for the specified branch"

#### 2. WTM Package Integration
**Key Components:**
- Extended interface with workspace detection capabilities
- Project type classification (SingleRepo, Workspace, Unknown)
- Clean interface design for extensibility

**Implementation Notes:**
- Extend existing interface to support workspace detection
- Define project types for classification: `ProjectType` enum with `Unknown`, `SingleRepo`, `Workspace`
- Use iota for enum-like constants
- Plan for multi-repo workspace support in future features
- **Detection priority**: Single repo first, then workspace if no single repo found
- **Return type unchanged**: `Run()` method still returns only `error`
- **Multiple workspace files**: Throw fatal error (selection functionality is for task 3)

## Test Cases

### Unit Tests

#### 1. FS Package Extension Tests
**Test Strategy:**
- Use real file system operations with temporary files/directories
- Test all new interface methods with various scenarios
- Clean up resources with `defer` statements
- Test both positive and negative cases
- **Integration tests only** - this is an adapter, no mocking

**Test Cases:**
- `TestFS_ReadFile`: Test reading existing and non-existing files
- `TestFS_ReadDir`: Test reading directory contents
- `TestFS_Glob`: Test pattern matching for `*.code-workspace` files

**Implementation Notes:**
- Use `os.CreateTemp()` and `os.MkdirTemp()` for test files
- Always clean up with `defer` to prevent test pollution
- Test cross-platform path handling
- Verify error conditions and edge cases
- Use standard `fs_test.go` naming
- **Use `//go:build integration` tag** for all tests
- **NO unit tests** - adapters should only have integration tests

#### 2. WTM Package Extension Tests (with Mocked FS)
**Test Strategy:**
- Use Uber gomock for mocking FS interface
- Test the extended `Run()` method with various scenarios
- Mock all FS method calls with expected parameters and return values
- Verify error handling and user feedback
- **Unit tests only** - this is business logic, no real file system access

**Test Cases:**
- `TestWTM_Run_ValidWorkspace`: Test successful workspace detection with valid configuration
- `TestWTM_Run_ValidWorkspaceWithName`: Test workspace detection when workspace has a name field
- `TestWTM_Run_ValidWorkspaceWithoutName`: Test workspace detection when workspace has no name field (uses filename)
- `TestWTM_Run_InvalidWorkspaceJSON`: Test when workspace file contains invalid JSON
- `TestWTM_Run_MissingRepository`: Test when workspace references non-existent repository
- `TestWTM_Run_InvalidRepository`: Test when workspace references non-Git repository
- `TestWTM_Run_NoWorkspaceFile`: Test when no workspace file exists (falls back to single repo detection)
- `TestWTM_Run_MultipleWorkspaceFiles`: Test when multiple workspace files are found (fatal error with count)
- `TestWTM_Run_EmptyFolders`: Test when workspace has empty folders array
- `TestWTM_Run_InvalidFolderStructure`: Test when folders array contains objects without path field
- `TestWTM_Run_InvalidPathField`: Test when path field is empty or not a string
- `TestWTM_Run_InvalidNameField`: Test when name field is present but not a string
- `TestWTM_Run_DuplicatePaths`: Test when multiple paths resolve to same location
- `TestWTM_Run_BrokenSymlink`: Test when symlink points to non-existent location
- `TestWTM_Run_NullValuesInFolders`: Test when folders array contains null values (should skip)
- `TestWTM_Run_QuietMode`: Test quiet mode operation (only errors to stderr)
- `TestWTM_Run_VerboseMode`: Test verbose mode operation (detailed steps)
- `TestWTM_Run_NormalMode`: Test normal mode operation (user interaction only)

**Implementation Notes:**
- Use `gomock.NewController(t)` for mock setup
- Set up mock expectations with `mockFS.EXPECT()`
- Test both success and failure scenarios
- Verify error messages and user output
- Use testify/assert for cleaner assertions
- **Use `//go:build unit` tag** for all tests
- **NO integration tests** - business logic should only have unit tests
- **Mock ALL file system operations** - no real file system access in tests
- **Create realistic workspace JSON examples** in test data (keep it simple, no multiple examples)

#### 3. Private Function Tests (with Mocked FS)
**Test Strategy:**
- Test individual private helper methods in isolation
- Use detailed mock expectations for specific scenarios
- Verify algorithm correctness for different workspace states
- Test edge cases and error conditions
- **Unit tests only** - this is business logic, no real file system access

**Test Cases:**
- `TestWTM_detectWorkspaceMode_ValidWorkspace`: Test successful workspace detection
- `TestWTM_detectWorkspaceMode_NoWorkspaceFile`: Test when no workspace file found
- `TestWTM_detectWorkspaceMode_MultipleFiles`: Test when multiple workspace files found
- `TestWTM_parseWorkspaceFile_ValidJSON`: Test successful JSON parsing
- `TestWTM_parseWorkspaceFile_InvalidJSON`: Test JSON parsing errors
- `TestWTM_parseWorkspaceFile_EmptyFolders`: Test when folders array is empty
- `TestWTM_parseWorkspaceFile_InvalidFolderStructure`: Test when folder objects lack required fields
- `TestWTM_parseWorkspaceFile_InvalidFoldersType`: Test when folders is not an array
- `TestWTM_parseWorkspaceFile_InvalidFolderType`: Test when folder is not an object
- `TestWTM_parseWorkspaceFile_NullValues`: Test when folders array contains null values (should skip)
- `TestWTM_validateWorkspaceRepositories_ValidRepos`: Test successful repository validation
- `TestWTM_validateWorkspaceRepositories_InvalidRepos`: Test repository validation errors
- `TestWTM_validateWorkspaceRepositories_DuplicatePaths`: Test when multiple paths resolve to same location
- `TestWTM_validateWorkspaceRepositories_SymlinkResolution`: Test symlink resolution in paths
- `TestWTM_validateWorkspaceRepositories_BrokenSymlinks`: Test broken symlink detection

**Implementation Notes:**
- Test private methods by accessing them directly in test package
- Set up mock expectations for file operations
- Verify both return values and error conditions
- Ensure proper error message content
- **Use `//go:build unit` tag** for all tests
- **Mock ALL file system operations** - no real file system access in tests

### Integration Tests

#### 1. Real Workspace Detection
**Test Strategy:**
- Use real FS adapter with actual workspace files
- Test with various workspace configurations
- Verify consistent detection across different workspace structures
- **Integration tests for FS adapter only** - not for WTM business logic

**Test Cases:**
- `TestFS_ReadFile_RealWorkspace`: Test reading actual `.code-workspace` files
- `TestFS_Glob_RealWorkspace`: Test pattern matching with real workspace files
- `TestFS_ReadDir_RealWorkspace`: Test directory listing with real workspace structures

**Implementation Notes:**
- Create real workspace files for testing
- Test with different repository structures
- Verify file system operations work with various Git configurations
- Test with workspaces that have different folder structures
- **Use `//go:build integration` tag** for real file system tests
- **Only test FS adapter** - WTM business logic uses unit tests with mocked FS

#### 2. Edge Cases
**Test Strategy:**
- Test with problematic file system scenarios
- Verify error handling for edge cases
- Ensure graceful degradation
- **Integration tests for FS adapter only** - not for WTM business logic

**Test Cases:**
- `TestFS_ReadFile_MalformedWorkspace`: Test reading corrupted workspace files
- `TestFS_Glob_EdgeCases`: Test pattern matching with edge cases
- `TestFS_ReadDir_EdgeCases`: Test directory listing with problematic structures
- `TestFS_ReadFile_PermissionErrors`: Test permission-related file system errors

**Implementation Notes:**
- Create controlled edge case scenarios
- Verify appropriate error messages from file system operations
- Test cross-platform compatibility
- Ensure no panics or crashes
- **Use `//go:build integration` tag** for real file system tests
- **Only test FS adapter** - WTM business logic uses unit tests with mocked FS

## Implementation Plan

### Phase 1: FS Package Extension (Priority: High)
1. Extend FS interface with `ReadFile()`, `ReadDir()`, and `Glob()` methods
2. Update concrete implementation using Go standard library (`os`, `io`, `path/filepath`)
3. Update `//go:generate` directive for Uber mockgen
4. Create comprehensive **integration tests only** for new FS methods
5. Ensure cross-platform compatibility
6. Generate mock files as `mockfs.gen.go` using `go generate` and commit them
7. **Ensure ALL file system operations** are implemented in this adapter

### Phase 2: WTM Package Extension (Priority: High)
1. Extend `Run()` function with workspace detection logic
2. Add private helper functions (`detectWorkspaceMode()`, `parseWorkspaceFile()`, `validateWorkspaceRepositories()`)
3. Define workspace configuration data structures
4. Create error types and messages
5. Write **unit tests only** using mocked FS from gomock with build tags
6. Extend Cobra integration in main.go to handle workspace mode
7. **Ensure NO direct file system calls** - all operations through FS adapter interface

### Phase 3: Integration (Priority: Medium)
1. Integrate extended FS and WTM packages in main application
2. **Verify separation of concerns**: FS adapter (integration tests) vs WTM (unit tests)
3. Performance optimization
4. Documentation updates
5. **Ensure proper test organization** with build tags

## Success Criteria

### Functional
- [ ] Successfully detects `.code-workspace` files in current directory
- [ ] Correctly parses valid workspace JSON configurations
- [ ] Validates all repository paths in workspace
- [ ] Returns appropriate errors for invalid scenarios
- [ ] Handles edge cases gracefully
- [ ] Falls back to single repository detection when no workspace file exists

### Non-Functional
- [ ] Detection and parsing completes in < 200ms for typical workspaces
- [ ] Works on all supported platforms (Windows, macOS, Linux)
- [ ] No external dependencies required beyond Go standard library
- [ ] Comprehensive test coverage (> 90%)

### Integration
- [ ] Integrates cleanly with existing single repository detection
- [ ] Provides clear interface for other features
- [ ] Follows project coding standards
- [ ] Includes proper documentation

## Dependencies
- **Blocked by**: Feature 001 (Detect Single Repository Mode)
- **Blocks**: Features 3, 4, 8, 9, 10, 12, 17, 18, 21, 22, 23, 24

## Dependencies

### Required Go Modules
- `github.com/spf13/cobra v1.7.0`: Command-line argument parsing
- `go.uber.org/mock v0.5.2`: Mocking framework for testing
- `github.com/stretchr/testify v1.8.4`: Testing utilities and assertions
- `encoding/json`: Go standard library for JSON parsing

### Development Dependencies
- `go generate ./pkg/fs`: Command to generate mock files using Uber mockgen
- Build tags: `unit` for unit tests, `integration` for real file system tests (adapters only)
- `.cursorrules` file for testing conventions (only adapters should have integration tests)
- Separate test files: integration tests for adapters, unit tests for business logic

## Future Considerations
- Consider caching workspace configuration for performance
- Plan for multi-repo workspace support (Feature 003)
- Consider supporting workspace-specific settings
- Plan for workspace extensions configuration
- Extend mocking strategy for other packages as needed
- Consider supporting workspace templates

## Notes
- This feature builds upon Feature 001 and must be implemented after it
- Focus on reliability over performance initially
- Ensure error messages are user-friendly
- Consider adding debug logging for troubleshooting
- Use build tags to organize tests: `unit` for unit tests, `integration` for real file system tests (adapters only)
- Mock files should be committed to the repository as `mockfs.gen.go`
- Exit immediately with exit code 1 on errors
- **FS package uses integration tests only (adapter)** - ALL file system operations here
- **WTM package uses unit tests only (business logic)** - NO direct file system access
- Support three output modes: quiet (errors to stderr only), verbose (detailed steps), normal (user interaction only)
- **Only adapters should have integration tests**
- **Separate test files**: integration tests for adapters, unit tests for business logic
- Keep mockgen parameters simple
- **Detection priority**: Single repo first (`.git`), then workspace (`*.code-workspace`) if no single repo
- Handle both relative and absolute paths in workspace configuration
- Validate that workspace repositories are accessible and contain valid Git repositories
- **Clear separation**: FS adapter handles ALL file system operations, WTM handles business logic only
