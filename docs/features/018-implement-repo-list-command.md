# Feature 018: Implement Repository List Command

## Overview
Implement functionality to list all repositories present in the status.yaml file. This feature will allow users to see all tracked repositories with visual indicators for those that are not within the configured base_path, providing visibility into the repository management setup.

## Background
The Code Manager (cm) needs to provide visibility into all tracked repositories from the status file. This feature will help users understand what repositories are being managed by CM and identify any repositories that may be outside the expected base path. This builds upon the existing status management capabilities to provide a complete repository management solution.

## Requirements

### Functional Requirements
1. **Repository Listing**: List all repositories from the status.yaml file
2. **Base Path Validation**: Check if each repository's path is within the configured base_path
3. **Visual Indicators**: Display an asterisk (*) for repositories not in the base_path
4. **Numbered Output**: Display repositories in a numbered list format
5. **Repository Name Display**: Show full repository URLs as stored in status.yaml
6. **Empty State Handling**: Display appropriate message when no repositories are found
7. **Error Handling**: Handle status file errors and corruption gracefully
8. *Configuration Integration**: Use existing configuration for status file and base_path location

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
- `ListRepositories() ([]RepositoryInfo, error)`: Main entry point for listing repositories

**Implementation Structure**:
- Extends existing CM package with repository listing functionality
- Integration with status management using existing `ListRepositories()` method
- Base path validation logic using existing FS adapter
- Error handling with wrapped errors
- User feedback through logger

**Key Characteristics**:
- Builds upon existing status management system
- Provides comprehensive user feedback
- Standalone operation (no integration with other commands)
- Uses existing adapters (FS, Status) for all operations
- Simple and focused functionality

### Implementation Details

#### 1. CM Package Implementation
The CM package will implement the repository listing logic:

**Key Components**:
- Add `ListRepositories()` method as the main entry point
- Integration with status management using existing `ListRepositories()` method
- Base path validation using new FS adapter method `IsPathWithinBase()`
- Repository information formatting using dedicated `RepositoryInfo` struct
- User feedback and output formatting

**Implementation Flow**:
1. Load all repositories from status file using status manager
2. For each repository:
   a. Extract repository name and path
   b. Check if path is within configured base_path using FS adapter
   c. Create RepositoryInfo with appropriate indicator
3. Return formatted repository list
4. Handle empty state and error conditions

**Implementation Notes**:
- Use existing Status adapter for repository listing
- Use existing Config for base_path validation
- Add new `IsPathWithinBase()` method to FS adapter using `filepath.HasPrefix()` for cross-platform path validation
- Define `RepositoryInfo` struct in `repo_list.go` file within CM package
- Provide clear user messages for different scenarios
- Support quiet mode (only errors to stderr), verbose mode (detailed steps), and normal mode (user interaction only)
- Verbose mode prints significant steps (e.g., "Loading repositories from status file", "Validating base path", "Formatting repository list")
- Error messages can include additional context in verbose mode before the actual error
- Base path validation uses filepath operations for cross-platform compatibility

#### 2. Output Format
Repositories will be displayed in a numbered list format starting from 1:

**Normal Mode Output**:
```
1. github.com/lerenn/example
2. *github.com/lerenn/out-of-base-path
3. github.com/lerenn/another-repo
```

**Empty State Output**:
```
No repositories found in status.yaml
```

**Error State Output**:
```
Error: failed to load repositories from status file: file not found
```

#### 3. Data Structures

**RepositoryInfo Structure**:
```go
type RepositoryInfo struct {
    Name        string
    Path        string
    InRepositoriesDir  bool
}
```

**Implementation Notes**:
- RepositoryInfo contains all necessary information for display
- InRepositoriesDir field determines whether to show asterisk
- Name field contains repository path (e.g., `github.com/lerenn/example`) from status.yaml
- Path field contains the repository's absolute local path
- Defined in `repo_list.go` file within CM package for reusability in CLI

### CLI Integration

#### Repository Command Extension
**New Subcommand**:
- `list` command under the existing `repository` command group
- Aliases: `ls`, `l` for convenience
- Simple command with no additional flags
- Follows existing command patterns for verbose mode and error handling

**Command Structure**:
```bash
cm repository list    # Full command
cm repo list         # Alias
cm r list           # Short alias
```

**Implementation Notes**:
- Extends existing repository command structure
- Uses existing configuration loading
- Calls CM.ListRepositories() method
- Formats output for user display
- Handles errors and displays appropriate messages
- Creates new file `cmd/cm/repository/list.go` following existing patterns

### Error Handling

#### Error Types
1. **StatusFileError**: When status.yaml file cannot be read or parsed (reuse existing status package errors)
2. *ConfigurationError**: When base_path configuration is invalid (reuse existing config package errors)
3. **PermissionError**: When unable to access status file due to permissions (reuse existing FS package errors)

#### Error Handling Strategy
- **Status File Missing**: Return error with clear message
- **Status File Corrupted**: Return error with parsing details
- *Configuration Issues**: Return error with configuration details
- **Permission Issues**: Return error with permission details
- **Base Path Issues**: Continue with warning in verbose mode

### Testing Strategy

#### Unit Tests
- **Test ListRepositories Success**: Test successful repository listing
- **Test ListRepositories Empty**: Test empty repository list
- **Test Base Path Validation**: Test repositories inside and outside base_path
- **Test Error Handling**: Test various error conditions
- **Test Output Formatting**: Test numbered list format and asterisk indicators
- **Mock Strategy**: Mock Status Manager (first level dependency) for CM package tests

#### Integration Tests
- **Test Status File Integration**: Test with real status.yaml file
- **Test Configuration Integration**: Test with real configuration
- **Test File System Operations**: Test with real file system
- **Test FS IsPathWithinBase Method**: Test positive cases (path within base), negative cases (path outside base), edge cases (relative paths, absolute paths, different separators)

#### End-to-End Tests
- **Test Complete Workflow**: Test from CLI command to output
- **Test Error Scenarios**: Test various error conditions in real environment
- **Test Cross-Platform**: Test on different operating systems

### Implementation Plan

#### Phase 1: Core Implementation
1. **Rename Existing Files**: Rename all worktree-related files in `pkg/cm/` to have `worktrees_` prefix
2. **Add RepositoryInfo Structure**: Define data structure for repository information in `repo_list.go`
3. **Add FS Adapter Method**: Add `IsPathWithinBase()` method to FS interface and implementation with integration tests
4. **Implement ListRepositories Method**: Add main listing functionality to CM package in `repo_list.go`
5. **Add Base Path Validation**: Implement path checking logic using FS adapter
6. **Add Error Handling**: Implement comprehensive error handling reusing existing errors

#### Phase 2: CLI Integration
1. **Create List Command File**: Create new file `cmd/cm/repository/list.go` following existing patterns
2. **Add Output Formatting**: Implement numbered list format with asterisk indicators
3. **Add Configuration Integration**: Integrate with existing configuration system
4. **Add Command Registration**: Register list command with existing repository command group

#### Phase 3: Testing and Validation
1. **Add Unit Tests**: Implement comprehensive unit test coverage
2. **Add Integration Tests**: Test with real file system and status file
3. **Add End-to-End Tests**: Test complete CLI workflow
4. **Add Documentation**: Update documentation and examples

### Success Criteria
1. **Functional**: Command successfully lists all repositories from status.yaml
2. **Visual**: Asterisk correctly indicates repositories outside base_path
3. **Performance**: Command completes within 100ms for typical repository counts
4. **Reliability**: Handles all error conditions gracefully
5. **Usability**: Clear and intuitive output format
6. **Testability**: Comprehensive test coverage with mocked dependencies
7. **Integration**: Seamless integration with existing CM architecture

### Future Enhancements
1. **Filtering Options**: Add flags to filter repositories by various criteria
2. **Sorting Options**: Add flags to sort repositories by name, path, or date
3. **Detailed Output**: Add verbose mode with additional repository information
4. **Export Options**: Add flags to export repository list in different formats
5. **Interactive Mode**: Add interactive selection for repository operations
