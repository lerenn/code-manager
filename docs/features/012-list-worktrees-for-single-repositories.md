# Feature 012: List Worktrees for Single Repositories

## Overview
Implement functionality to list worktrees for the current Git repository. This feature will allow users to see all worktrees associated with their current repository, providing visibility into their worktree management setup.

## Background
The Code Manager (cm) needs to provide visibility into existing worktrees for the current repository. This feature will help users understand what worktrees they have created and their current status. This builds upon the existing detection, validation, and status management capabilities to provide a complete worktree management solution.

## Requirements

### Functional Requirements
1. **Worktree Listing**: List all worktrees for the current repository
2. **Repository Detection**: Ensure the current directory is a valid Git repository
3. **Status Integration**: Read worktree information from the status file
4. **Simple Output**: Display worktrees in a simple text list format
5. **Repository Filtering**: Only show worktrees for the current repository
6. **Error Handling**: Handle various error conditions gracefully
7. **User Feedback**: Provide clear feedback about the listing operation
8. **Configuration Integration**: Use existing configuration for status file location

### Non-Functional Requirements
1. **Performance**: Listing should complete quickly (< 100ms)
2. **Reliability**: Handle status file errors and corruption gracefully
3. **Cross-Platform**: Work on Windows, macOS, and Linux
4. **Testability**: Use Uber gomock for mocking dependencies
5. **Minimal Dependencies**: Use only existing adapters and Go standard library
6. **Standalone Operation**: Function independently of other commands

## Technical Specification

### Interface Design

#### CM Package Extension
**New Interface Methods**:
- `ListWorktrees() ([]status.Repository, error)`: Main entry point for listing worktrees with mode detection
- `listWorktreesForSingleRepo() ([]status.Repository, error)`: List worktrees for current repository
- `listWorktreesForWorkspace() ([]status.Repository, error)`: List worktrees for workspace (placeholder for future)

**Implementation Structure**:
- Extends existing CM package with listing functionality
- Mode detection in main `ListWorktrees()` method
- Private helper methods for different modes
- Integration with status management
- Error handling with wrapped errors
- User feedback through logger

**Key Characteristics**:
- Builds upon existing detection and validation logic
- Integrates with status management system
- Provides comprehensive user feedback
- Standalone operation (no integration with create/open commands)
- Uses existing adapters (FS, Git, Status) for all operations
- Mode detection determines which listing method to call

### Implementation Details

#### 1. CM Package Implementation
The CM package will implement the worktree listing logic:

**Key Components**:
- Add `ListWorktrees()` method as the main entry point with mode detection
- Add `listWorktreesForSingleRepo()` helper method for single repository mode
- Add `listWorktreesForWorkspace()` placeholder method for workspace mode
- Integration with status management
- Repository name extraction and filtering
- User feedback and output formatting

**Implementation Flow**:
1. Detect current mode (single repository vs workspace)
2. For single repository mode:
   a. Validate current directory is a Git repository
   b. Extract repository name from remote origin URL (fallback to local path if no remote)
   c. Load all worktrees from status file
   d. Filter worktrees to only include those for the current repository
   e. Return filtered worktrees list
3. For workspace mode (placeholder):
   a. Return empty list with placeholder message
   b. Future implementation will handle workspace detection and listing

**Implementation Notes**:
- Use existing FS adapter for file system operations
- Use existing Git adapter for repository name extraction and mode detection
- Use existing Status adapter for worktree listing
- Provide clear user messages for different scenarios
- Support quiet mode (only errors to stderr), verbose mode (detailed steps), and normal mode (user interaction only)
- Verbose mode prints significant steps (e.g., "Detecting mode", "Checking for .git directory", "Extracting repository name", "Loading worktrees from status file")
- Error messages can include additional context in verbose mode before the actual error
- Mode detection determines which listing method to call
- Workspace mode is placeholder for future implementation

#### 2. Output Format
Worktrees will be displayed in a simple text list format:

**Normal Mode Output**:
```
Worktrees for github.com/lerenn/example:
  feature-a: /Users/lfradin/.cm/github.com/lerenn/example/feature-a
  feature-b: /Users/lfradin/.cm/github.com/lerenn/example/feature-b
  bugfix-123: /Users/lfradin/.cm/github.com/lerenn/example/bugfix-123
```

**Empty List Output**:
```
No worktrees found for github.com/lerenn/example
```

**Verbose Mode Output**:
```
Checking for .git directory...
Extracting repository name from remote origin...
Repository name: github.com/lerenn/example
Loading worktrees from status file...
Found 3 worktrees for current repository
Worktrees for github.com/lerenn/example:
  feature-a: /Users/lfradin/.cm/github.com/lerenn/example/feature-a
  feature-b: /Users/lfradin/.cm/github.com/lerenn/example/feature-b
  bugfix-123: /Users/lfradin/.cm/github.com/lerenn/example/bugfix-123
```

#### 3. Mode Detection and Repository Name Handling
- Use existing mode detection logic from CM package
- Use existing `GetRepositoryName()` method from Git adapter
- Support both remote origin URL extraction and local path fallback
- Ensure consistent repository name format for filtering
- Single repository mode: detect Git repository in current directory
- Workspace mode: placeholder for future workspace detection

### Error Handling

#### Error Types
1. **NoGitRepositoryError**: When current directory is not a Git repository
2. **StatusFileCorruptedError**: When status file is corrupted or invalid
3. **StatusFileNotFoundError**: When status file doesn't exist (treat as empty list)
4. **RepositoryNameError**: When unable to extract repository name
5. **PermissionError**: When unable to access status file due to permissions
6. **ModeDetectionError**: When unable to detect current mode

#### Error Recovery
- Handle missing status file gracefully (treat as empty list)
- Provide clear error messages with recovery instructions
- Use existing error handling patterns from other features
- Ensure graceful degradation on non-critical errors

### User Interface

#### Command Line Interface
The feature will be accessible through a new `list` subcommand:
```bash
# List worktrees for current repository
cm list

# List worktrees with verbose output
cm list --verbose

# List worktrees with quiet output (errors only)
cm list --quiet
```

**Command Structure**:
- New `list` subcommand with no arguments
- Uses existing global flags (--verbose, --quiet, --config)
- Non-interactive operation (no prompts or confirmations)

#### User Feedback
- **Verbose mode**: Detailed progress information (validation steps, repository name extraction, status file loading)
- **Normal mode**: Success/error messages and worktree list
- **Quiet mode**: Only error messages to stderr
- **Non-interactive**: No prompts or confirmations, fails immediately on critical errors

### Testing Strategy

#### Unit Tests (Business Logic)
**Test Strategy**:
- Use Uber gomock for mocking all dependencies (FS, Git, Status)
- Test the public `ListWorktrees()` method with various scenarios
- Mock all adapter method calls with expected parameters and return values
- Verify error handling and user feedback
- Unit tests only (not an adapter)

**Test Cases**:
- `TestCM_ListWorktrees_SingleRepoModeWithWorktrees`: Test successful listing in single repository mode with existing worktrees
- `TestCM_ListWorktrees_SingleRepoModeNoWorktrees`: Test listing in single repository mode when no worktrees exist
- `TestCM_ListWorktrees_WorkspaceMode`: Test listing in workspace mode (placeholder, returns empty list)
- `TestCM_ListWorktrees_NoRepo`: Test when current directory is not a Git repository
- `TestCM_ListWorktrees_StatusFileCorrupted`: Test handling of corrupted status file
- `TestCM_ListWorktrees_StatusFileNotFound`: Test handling of missing status file
- `TestCM_ListWorktrees_RepositoryNameError`: Test handling of repository name extraction errors
- `TestCM_ListWorktrees_ModeDetectionError`: Test handling of mode detection errors
- `TestCM_ListWorktrees_QuietMode`: Test quiet mode operation (only errors to stderr)
- `TestCM_ListWorktrees_VerboseMode`: Test verbose mode operation (detailed steps)
- `TestCM_ListWorktrees_NormalMode`: Test normal mode operation (user interaction only)

**Implementation Notes**:
- Use `gomock.NewController(t)` for mock setup
- Set up mock expectations with `mockFS.EXPECT()`, `mockGit.EXPECT()`, `mockStatus.EXPECT()`
- Test both success and failure scenarios
- Verify error messages and user output
- Use testify/assert for cleaner assertions
- Use `//go:build unit` tag for unit tests

#### Private Function Tests (with Mocked Dependencies)
**Test Strategy**:
- Test individual private helper methods in isolation
- Use detailed mock expectations for specific scenarios
- Verify algorithm correctness for different repository states
- Test edge cases and error conditions
- Unit tests only (not an adapter)

**Test Cases**:
- `TestCM_listWorktreesForSingleRepo_ValidRepoWithWorktrees`: Test successful listing logic for single repository
- `TestCM_listWorktreesForSingleRepo_ValidRepoNoWorktrees`: Test empty list logic for single repository
- `TestCM_listWorktreesForSingleRepo_RepositoryNameError`: Test repository name extraction errors
- `TestCM_listWorktreesForWorkspace_Placeholder`: Test workspace mode placeholder implementation

**Implementation Notes**:
- Test private methods by accessing them directly in test package
- Set up mock expectations for repository name extraction and status operations
- Verify both return values and error conditions
- Ensure proper error message content
- Use `//go:build unit` tag for unit tests

#### Integration Tests (Adapters)
**Test Strategy**:
- Use real adapters with actual Git repositories and status files
- Test from various subdirectories within repositories
- Verify consistent listing across different repository structures
- Integration tests for adapters only

**Test Cases**:
- `TestCM_ListWorktrees_RealRepository`: Test with real Git repositories and status files
- `TestCM_ListWorktrees_DifferentRepoTypes`: Test various Git repository configurations

**Implementation Notes**:
- Use real Git repositories for testing (e.g., small test repos)
- Test from current directory only
- Verify listing works with different Git configurations
- Test with repositories that have different structures
- Use `//go:build integration` tag for real file system tests

### Implementation Plan

#### Phase 1: CM Package Extension (Priority: High)
1. Add `ListWorktrees()` method to CM package with mode detection
2. Add `listWorktreesForSingleRepo()` helper method for single repository mode
3. Add `listWorktreesForWorkspace()` placeholder method for workspace mode
4. Implement repository name extraction and filtering logic
5. Add output formatting for simple text list
6. Create comprehensive unit tests using mocked dependencies

#### Phase 2: Command Integration (Priority: High)
1. Add `list` subcommand to main application
2. Integrate with existing global flags (--verbose, --quiet, --config)
3. Add proper error handling and user feedback
4. Test command-line interface

#### Phase 3: Integration Testing (Priority: Medium)
1. Add integration tests with real file system and Git operations
2. Performance optimization
3. Documentation updates

### File Structure

```
cmd/cm/
├── main.go                    # Extended with list subcommand
pkg/cm/
├── cm.go                    # Extended with ListWorktrees method
├── cm_test.go               # Unit tests with mocked dependencies
├── list_test.go              # Unit tests for listing functionality
└── mockcm.gen.go            # Generated mock for testing
```

### Dependencies

#### Direct Dependencies
- **Blocked by**: Features 1, 4, 8 (detection, validation, status management)
- **Dependencies**: Git package, FS package, Status package, Config package
- **External**: Git command-line tool
- **Adapters**: Use existing FS adapter for file system operations, Git adapter for Git operations, Status adapter for status file operations

#### Required Go Modules
- `github.com/spf13/cobra v1.7.0`: Command-line argument parsing
- `go.uber.org/mock v0.5.2`: Mocking framework for testing
- `github.com/stretchr/testify v1.8.4`: Testing utilities and assertions

### Success Criteria

#### Functional
- [ ] Successfully detects mode and lists worktrees accordingly
- [ ] Lists worktrees for current Git repository in single repository mode
- [ ] Returns empty list for workspace mode (placeholder)
- [ ] Displays worktrees in simple text format
- [ ] Handles empty list case gracefully
- [ ] Filters worktrees to current repository only
- [ ] Provides appropriate error messages for various scenarios

#### Non-Functional
- [ ] Listing completes in < 100ms for typical repositories
- [ ] Works on all supported platforms (Windows, macOS, Linux)
- [ ] Uses only existing adapters and dependencies
- [ ] Comprehensive test coverage (> 90%)

#### Integration
- [ ] Integrates cleanly with main application
- [ ] Follows project coding standards
- [ ] Includes proper documentation
- [ ] Standalone operation (no integration with create/open commands)

### Error Scenarios

#### 1. No Worktrees Exist
**Behavior**: Display "No worktrees found for {repository-name}"
**Implementation**: Check if filtered worktrees list is empty
**User Experience**: Clear indication that no worktrees exist for current repository

#### 2. Status File is Corrupted
**Behavior**: Return error with clear message
**Implementation**: Handle YAML parsing errors from status file
**User Experience**: Clear error message with recovery instructions

#### 3. Worktree Directories No Longer Exist on Disk
**Behavior**: Not checked (as per requirements)
**Implementation**: Only read from status file, no disk validation
**User Experience**: List shows worktrees as they exist in status file

#### 4. Current Directory is Not a Git Repository
**Behavior**: Return error with clear message
**Implementation**: Use existing Git repository detection logic in mode detection
**User Experience**: Clear error message indicating no Git repository found

#### 5. Mode Detection Fails
**Behavior**: Return error with clear message
**Implementation**: Handle mode detection errors in main ListWorktrees method
**User Experience**: Clear error message indicating mode detection failure

#### 6. Status File Not Found
**Behavior**: Treat as empty list
**Implementation**: Handle file not found errors gracefully
**User Experience**: Display "No worktrees found for {repository-name}"

### Future Considerations

#### All Repositories Flag
- Future feature will add `-a` or `--all` flag to list worktrees across all repositories
- Current implementation focuses only on current repository
- Design should be extensible for future all-repositories functionality

#### Output Format Options
- Future consideration for JSON/YAML output formats
- Current implementation uses simple text format
- Design should be extensible for multiple output formats

#### Worktree Status Validation
- Future consideration for validating worktree directories exist on disk
- Current implementation only reads from status file
- Could add optional validation flag in future

#### Performance Optimization
- Consider caching for large status files
- Current implementation always reads from status file
- Could add caching mechanism for future performance improvements

### Notes
- This feature is standalone and does not integrate with create/open commands
- Uses existing adapters (FS, Git, Status) for all external operations
- Follows existing testing conventions (unit tests for business logic, integration tests for adapters)
- Supports three output modes: quiet (errors to stderr only), verbose (detailed steps), normal (user interaction only)
- Only adapters should have integration tests
- Business logic uses unit tests only with mocked dependencies
- Exit immediately with exit code 1 on critical errors
- Non-interactive operation (no prompts or confirmations)
