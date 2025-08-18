# Feature 001: Detect Single Repository Mode

## Overview
Implement functionality to detect when the current working directory is a single Git repository by checking for the presence of a `.git` folder.

## Background
The Code Manager (cm) needs to distinguish between different project types to provide appropriate worktree management. The first step is detecting single repository mode, which is the foundation for all other features.

## Requirements

### Functional Requirements
1. **Git Repository Detection**: Detect if the current working directory contains a `.git` folder
2. **Path Validation**: Ensure the `.git` folder is a valid Git repository (contains necessary Git files)
3. **Error Handling**: Provide clear error messages when no Git repository is found
4. **Integration Ready**: Return detection results in a format suitable for other features

### Non-Functional Requirements
1. **Performance**: Detection should be fast (< 100ms)
2. **Reliability**: Handle edge cases (symlinks, broken Git repos, etc.)
3. **Cross-Platform**: Work on Windows, macOS, and Linux
4. **Testability**: Use Uber gomock for mocking file system operations
5. **Minimal Dependencies**: Use only Go standard library + gomock for file system operations

## Technical Specification

### Interface Design

#### FS Package (File System Adapter)
**Interface Design:**
- `Exists(path string) (bool, error)`: Check if file/directory exists
- `IsDir(path string) (bool, error)`: Check if path is a directory

**Key Characteristics:**
- Minimal file system abstraction
- Pure function-based operations
- No state management required
- Cross-platform compatibility

#### CM Package (Business Logic)
**Interface Design:**
- `Run() error`: Main entry point for application logic

**Implementation Structure:**
- Dependency injection of FS interface
- Private helper method: `detectSingleRepoMode()`
- Error handling with wrapped errors
- Clean separation of concerns

**Key Characteristics:**
- Single public method for simplicity
- Business logic focused on Git repository detection
- Testable through dependency injection

### Implementation Details

#### 1. FS Package Implementation
The FS package provides simple file system operations that can be used by the CM package:

**Key Components:**
- Interface with methods: `Exists()`, `IsDir()`
- Concrete implementation using Go standard library (`os`)
- `//go:generate` directive for Uber mockgen
- No state - pure function-based operations

**Implementation Notes:**
- Use `os.Stat()` for both existence and directory checks
- Handle cross-platform path separators properly
- Add `//go:generate go run go.uber.org/mock/mockgen@v0.5.2 -source=fs.go -destination=mockfs.gen.go -package=fs` directive
- Generate mock files as `mockfs.gen.go` in same directory

#### 2. CM Package Implementation
The CM package uses the FS adapter to implement Git repository detection:

**Key Components:**
- Public `Run()` method as the main entry point
- Private helper methods: `detectSingleRepoMode()`
- Dependency injection of FS interface
- Error handling with wrapped errors

**Implementation Notes:**
- `Run()` orchestrates the detection flow and provides user feedback
- `detectSingleRepoMode()` implements the core detection algorithm
- Use `fmt.Errorf()` with `%w` for error wrapping
- Provide clear user messages for different detection states
- Support quiet mode (only errors to stderr), verbose mode (detailed steps), and normal mode (user interaction only)
- Verbose mode prints significant steps (e.g., "Checking for .git directory", "Verifying .git is a directory")
- Error messages can include additional context in verbose mode before the actual error

#### 2. Git Repository Detection
- Check current working directory for `.git` folder
- Verify that `.git` is a directory
- Return detection result (true/false)



### Error Handling

#### Error Types
1. **NoGitRepositoryError**: When no `.git` folder is found
2. **PermissionError**: When unable to access directories due to permissions
3. **SymlinkError**: When encountering problematic symlinks

#### Error Messages
- `"no git repository found in current directory"`
- `"permission denied: cannot access .git directory"`
- `"broken symlink detected in git repository path"`

### Integration Points

#### 1. Main Application Flow
**Key Components:**
- Cobra root command setup
- Dependency injection: FS adapter → CM manager
- Error handling with `log.Fatal()`
- Clean separation of concerns

**Implementation Notes:**
- Use Cobra for command-line argument parsing with root command and `create` subcommand
- Create FS adapter and CM manager in the command's `RunE` function
- Handle errors at the top level with `log.Fatal()` and exit code 1
- Keep main function simple and focused on orchestration
- Add global quiet mode flag for silent operation (only errors to stderr)
- Add global verbose mode flag for detailed step output
- Normal mode shows only user interaction messages
- Implement `create` subcommand now (calls detection logic)
- `create` subcommand takes exactly one argument (branch name) but not used yet
- Help text: "Create worktree(s) for the specified branch"

#### 2. CM Package Integration
**Key Components:**
- Simple interface with single `Run()` method
- Future types for project classification
- Clean interface design for extensibility

**Implementation Notes:**
- Keep interface minimal for this feature
- Define project types for future subcommand implementation
- Use iota for enum-like constants
- Plan for workspace mode detection in future features

## Test Cases

### Unit Tests

#### 1. FS Package Tests
**Test Strategy:**
- Use real file system operations with temporary files/directories
- Test all interface methods with various scenarios
- Clean up resources with `defer` statements
- Test both positive and negative cases
- Integration tests only (no mocking for adapter)

**Test Cases:**
- `TestFS_Exists`: Test existing and non-existing files/directories
- `TestFS_IsDir`: Test files vs directories

**Implementation Notes:**
- Use `os.CreateTemp()` and `os.MkdirTemp()` for test files
- Always clean up with `defer` to prevent test pollution
- Test cross-platform path handling
- Verify error conditions and edge cases
- Use standard `fs_test.go` naming

#### 2. CM Package Tests (with Mocked FS)
**Test Strategy:**
- Use Uber gomock for mocking FS interface
- Test the public `Run()` method with various scenarios
- Mock all FS method calls with expected parameters and return values
- Verify error handling and user feedback
- Unit tests only (not an adapter)

**Test Cases:**
- `TestCM_Run_ValidRepo`: Test successful Git repository detection in current directory
- `TestCM_Run_NoRepo`: Test when no Git repository is found in current directory
- `TestCM_Run_Error`: Test FS error propagation
- `TestCM_Run_QuietMode`: Test quiet mode operation (only errors to stderr)
- `TestCM_Run_VerboseMode`: Test verbose mode operation (detailed steps)
- `TestCM_Run_NormalMode`: Test normal mode operation (user interaction only)

**Implementation Notes:**
- Use `gomock.NewController(t)` for mock setup
- Set up mock expectations with `mockFS.EXPECT()`
- Test both success and failure scenarios
- Verify error messages and user output
- Use testify/assert for cleaner assertions
- Use `//go:build unit` tag for unit tests

#### 3. Private Function Tests (with Mocked FS)
**Test Strategy:**
- Test individual private helper methods in isolation
- Use detailed mock expectations for specific scenarios
- Verify algorithm correctness for different Git repository states
- Test edge cases and error conditions
- Unit tests only (not an adapter)

**Test Cases:**
- `TestCM_detectSingleRepoMode_ValidRepo`: Test successful detection in current directory
- `TestCM_detectSingleRepoMode_NoRepo`: Test when no repository found in current directory

**Implementation Notes:**
- Test private methods by accessing them directly in test package
- Set up mock expectations for current directory operations
- Verify both return values and error conditions
- Ensure proper error message content
- Use `//go:build unit` tag for unit tests

### Integration Tests

#### 1. Real Repository Detection
**Test Strategy:**
- Use real FS adapter with actual Git repositories
- Test from various subdirectories within repositories
- Verify consistent detection across different repository structures
- Integration tests for adapters only

**Test Cases:**
- `TestCM_Run_RealRepository`: Test with cloned repositories in current directory
- `TestCM_Run_DifferentRepoTypes`: Test various Git repository configurations

**Implementation Notes:**
- Clone real repositories for testing (e.g., small test repos)
- Test from current directory only
- Verify detection works with different Git configurations
- Test with repositories that have different structures
- Use `//go:build integration` tag for real file system tests

#### 2. Edge Cases
**Test Strategy:**
- Test with problematic file system scenarios
- Verify error handling for edge cases
- Ensure graceful degradation
- Integration tests for adapters only

**Test Cases:**
- `TestCM_Run_Symlinks`: Test with symbolic links
- `TestCM_Run_BrokenPermissions`: Test with permission issues
- `TestCM_Run_NetworkDrives`: Test with network-mounted directories

**Implementation Notes:**
- Create controlled edge case scenarios
- Verify appropriate error messages
- Test cross-platform compatibility
- Ensure no panics or crashes
- Use `//go:build integration` tag for real file system tests

## Implementation Plan

### Phase 1: FS Package Implementation (Priority: High)
1. Implement basic file system operations (Exists, IsDir)
2. Add `//go:generate go run go.uber.org/mock/mockgen@v0.5.2` directive
3. Create comprehensive integration tests for FS package
4. Ensure cross-platform compatibility
5. Generate mock files as `mockfs.gen.go` using `go generate` and commit them

### Phase 2: CM Package Implementation (Priority: High)
1. Implement `Run()` function with basic Git repository detection
2. Add private helper functions (`detectSingleRepoMode()`)
3. Create error types and messages
4. Write unit tests using mocked FS from gomock with build tags
5. Add Cobra integration in main.go with root command, `create` subcommand, and mode flags

### Phase 3: Integration (Priority: Medium)
1. Integrate FS and CM packages in main application
2. Add integration tests with real file system using build tags
3. Performance optimization
4. Documentation updates

## Success Criteria

### Functional
- [ ] Successfully detects Git repositories in current directory
- [ ] Returns appropriate errors for invalid scenarios
- [ ] Handles edge cases gracefully

### Non-Functional
- [ ] Detection completes in < 100ms for typical repositories
- [ ] Works on all supported platforms (Windows, macOS, Linux)
- [ ] No external dependencies required
- [ ] Comprehensive test coverage (> 90%)

### Integration
- [ ] Integrates cleanly with main application
- [ ] Provides clear interface for other features
- [ ] Follows project coding standards
- [ ] Includes proper documentation

## Dependencies
- **Blocked by**: None
- **Blocks**: Features 2, 3, 4, 11, 12, 17, 18, 21, 22, 23, 24

## Dependencies

### Required Go Modules
- `github.com/spf13/cobra v1.7.0`: Command-line argument parsing
- `go.uber.org/mock v0.5.2`: Mocking framework for testing
- `github.com/stretchr/testify v1.8.4`: Testing utilities and assertions

### Development Dependencies
- `go generate ./pkg/fs`: Command to generate mock files using Uber mockgen
- Build tags: `unit` for unit tests, `integration` for real file system tests (adapters only)
- `.cursorrules` file for testing conventions (only adapters should have integration tests)
- Separate test files: integration tests for adapters, unit tests for business logic

## Future Considerations
- Consider caching detection results for performance
- Plan for workspace mode detection (Feature 002)
- Consider supporting bare repositories
- Plan for Git submodule detection
- Extend mocking strategy for other packages as needed

## Notes
- This feature is foundational and must be implemented first
- Focus on reliability over performance initially
- Ensure error messages are user-friendly
- Consider adding debug logging for troubleshooting
- Use build tags to organize tests: `unit` for unit tests, `integration` for real file system tests (adapters only)
- Mock files should be committed to the repository as `mockfs.gen.go`
- Exit immediately with exit code 1 on errors
- FS package uses integration tests only (adapter)
- CM package uses unit tests only (business logic)
- Support three output modes: quiet (errors to stderr only), verbose (detailed steps), normal (user interaction only)
- Only adapters should have integration tests
- Separate test files: integration tests for adapters, unit tests for business logic
- Keep mockgen parameters simple
