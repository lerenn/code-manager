# Feature 021: Implement Workspace Creation Command

## Overview
Implement a `workspace create` command that allows users to create new workspace definitions in the status.yaml file. The command will support selecting repositories from both local paths and repositories tracked in the status.yaml file, automatically add new repositories to the status file for consistency, and create workspace entries that can be used later for worktree creation.

## Background
The Code Manager (cm) currently supports workspace detection, validation, and worktree creation for existing workspaces. However, there's no capability to create new workspace definitions from scratch. This feature will complete the workspace management functionality by allowing users to create new workspace entries in the status.yaml file that can be used later for worktree creation and management.

## Requirements

### Functional Requirements
1. **Workspace Creation Command**: Create a new `workspace create` command with workspace name and repository selection
2. **Repository Selection**: Support multiple repository input formats (repository names from status.yaml, absolute paths, relative paths)
3. **Status File Integration**: Automatically add created workspaces to status.yaml for tracking
4. **Repository Auto-Addition**: Automatically add new repositories to status.yaml if they don't exist
5. **Repository Validation**: Validate all provided repositories exist and are accessible
6. **Path Resolution**: Resolve relative paths from command execution directory using FS adapter
7. **Error Handling**: Fail entire operation if any repository is invalid (no partial success)
8. **Non-Interactive Operation**: All parameters provided via command line (no prompts)
9. **Workspace Structure**: Update Workspace struct to contain worktree and repositories lists

### Non-Functional Requirements
1. **Performance**: Workspace creation should complete within 2 seconds for typical workspaces
2. **Reliability**: Handle file system errors and permission issues gracefully
3. **Cross-Platform**: Work consistently on Windows, macOS, and Linux
4. **Testability**: Support unit testing with mocked dependencies
5. **Minimal Dependencies**: Use only Go standard library and existing CM packages

## Technical Specification

### Interface Design

#### Config Package Extension
**Updated Config Structure**:
```go
type Config struct {
    RepositoriesDir string `yaml:"repositories_dir"` // User's repositories directory (default: ~/Code/src)
    StatusFile      string `yaml:"status_file"`      // Status file path (default: ~/.cm/status.yaml)
    WorkspacesDir   string `yaml:"workspaces_dir"`   // Workspaces directory (default: ~/Code/workspaces)
}
```

**Default Configuration Values**:
```yaml
# Default configuration (configs/default.yaml updated)
repositories_dir: ~/Code/src         # User's repositories directory
status_file: ~/.cm/status.yaml       # CM status tracking
workspaces_dir: ~/Code/workspaces    # Workspaces directory
```

#### Status Package Extension
**Updated Workspace Structure**:
```go
// Updated Workspace struct - contains worktree and repositories lists
type Workspace struct {
    Worktree     []string `yaml:"worktree"`     // List of worktree references
    Repositories []string `yaml:"repositories"` // List of repository URLs/names
}
```

**New Interface Methods**:
- `AddWorkspace(workspaceName string, params AddWorkspaceParams) error`: Add workspace to status file
- `GetWorkspace(workspaceName string) (*Workspace, error)`: Retrieve workspace from status file
- `ListWorkspaces() (map[string]Workspace, error)`: List all tracked workspaces

**Updated AddWorkspaceParams**:
```go
type AddWorkspaceParams struct {
    Repositories []string // List of repository URLs/names
}
```

#### CM Package Extension
**New Interface Methods**:
- `CreateWorkspace(params CreateWorkspaceParams) error`: Create new workspace with repository selection
- `ValidateAndResolveRepositories(repositories []string) ([]string, error)`: Validate and resolve repository paths/names
- `AddRepositoryToStatus(repoPath string) (string, error)`: Add new repository to status file and return repository URL

**New Data Structures**:
```go
type CreateWorkspaceParams struct {
    WorkspaceName string   // Name of the workspace
    Repositories  []string // Repository identifiers (names, paths, URLs)
}
```

**Key Characteristics**:
- **NO direct file system access** - all operations go through adapters
- **ONLY unit tests** using mocked dependencies
- Business logic for workspace creation workflow
- Repository validation and path resolution
- Error handling with wrapped errors


#### FS Package Extension
**New Interface Methods**:
- `ResolvePath(repositoriesDir, relativePath string) (string, error)`: Resolve relative paths from base directory
- `ValidateRepositoryPath(path string) (bool, error)`: Validate that path contains a Git repository

**Key Characteristics**:
- **ALL file system access** must be in this adapter
- **ONLY this package** should have integration tests with real file system
- Pure function-based operations
- Cross-platform compatibility
- **Single source of truth** for all file system operations

### Implementation Details

#### 1. Workspace Creation Algorithm
The workspace creation will follow this algorithm:

**Key Components**:
- `CreateWorkspace()`: Main entry point for workspace creation
- `ValidateAndResolveRepositories()`: Validate and resolve repository paths/names
- `AddRepositoryToStatus()`: Add new repositories to status file if needed
- `AddWorkspaceToStatus()`: Add workspace to status file for tracking

**Implementation Flow**:
1. Validate workspace name (no special characters, not empty)
2. Validate and resolve all repository identifiers:
   - Check if repository name exists in status.yaml
   - Resolve relative paths from current working directory
   - Validate absolute paths exist and contain Git repositories
   - Extract repository URL from Git remote origin for new repositories
3. Add new repositories to status.yaml if they don't exist
4. Add workspace entry to status file with resolved repository URLs

**Implementation Notes**:
- Use existing FS adapter for all file system operations
- Use existing Status adapter for status file operations
- Use existing Git adapter for repository URL extraction
- Provide comprehensive error handling and cleanup
- Non-interactive operation (fail immediately on any issue)
- Support multiple repository input formats
- Fail entire operation if any repository is invalid
- Automatically add new repositories to status.yaml for consistency

#### 2. Repository Path Resolution
Repository paths will be resolved in the following order:

**Path Resolution Priority**:
1. **Repository Name from Status**: Check if identifier matches a repository name in status.yaml
2. **Absolute Path**: Use path as-is if it's absolute
3. **Relative Path**: Resolve relative to current working directory using FS adapter

**Path Validation**:
- Ensure path exists and is accessible
- Validate path contains a `.git` directory (Git repository)
- Convert all paths to absolute paths for workspace file
- Check for duplicate repositories in the same workspace

#### 3. Status File Integration
Workspace creation will update the status file to track the new workspace:

**Status File Structure**:
```yaml
workspaces:
  my-workspace:
    worktree: []  # Empty initially, populated when worktrees are created
    repositories:
      - github.com/user/repo1
      - github.com/user/repo2
      - /absolute/path/to/local/repo3
```

**Status Manager Enhancement**:
- Add workspace to workspaces map in status file
- Use workspace name as the key (e.g., "my-workspace")
- Store worktree and repositories lists
- Update workspaces map computation for efficient operations

### Error Handling

#### Error Types
1. **InvalidWorkspaceNameError**: When workspace name contains invalid characters
2. **RepositoryNotFoundError**: When a repository path doesn't exist
3. **InvalidRepositoryError**: When a repository path doesn't contain a Git repository
4. **DuplicateRepositoryError**: When the same repository is specified multiple times
5. **WorkspaceAlreadyExistsError**: When a workspace with the same name already exists
6. **StatusUpdateError**: When status file update fails
7. **RepositoryAdditionError**: When adding new repository to status file fails
8. **PathResolutionError**: When relative path resolution fails

#### Error Recovery
- Clean up any partially added repositories from status file on failure
- Remove workspace entry from status file if workspace creation fails
- Provide clear error messages with recovery instructions
- Use existing file locking mechanism for status file operations
- Ensure atomic operations where possible

### User Interface

#### Command Line Interface
The feature will be accessible through a new workspace command:
```bash
# Create workspace with repository names from status.yaml
cm workspace create my-workspace repo1 repo2

# Create workspace with absolute paths
cm workspace create my-workspace /path/to/repo1 /path/to/repo2

# Create workspace with relative paths
cm workspace create my-workspace ./repo1 ../repo2

# Create workspace with mixed repository sources
cm workspace create my-workspace repo1 /path/to/repo2 ./repo3
```

**Command Structure**:
```bash
cm workspace create <workspace-name> [repositories...]
```

**Parameters**:
- `<workspace-name>`: Name of the workspace
- `[repositories...]`: Repository identifiers (names from status.yaml, absolute paths, or relative paths)

#### User Feedback
- **Verbose mode**: Detailed progress information (validation steps, repository addition, status updates)
- **Normal mode**: Success/error messages with summary
- **Non-interactive**: No prompts or confirmations, fails immediately on any issue
- **Progress reporting**: Show progress for each repository validation
- **Summary reporting**: Provide summary of successful workspace creation

**Feedback Examples**:
```
Creating workspace: my-workspace
Validating repositories:
  ✓ repo1 (from status.yaml): github.com/user/repo1
  ✓ /path/to/repo2: github.com/user/repo2 (added to status)
  ✓ ./repo3: /current/dir/repo3 (added to status)
Adding workspace to status file
✓ Workspace created successfully
```

### Testing Strategy

#### Unit Tests
- Mock FS operations for testing workspace creation logic
- Mock Status operations for testing status file updates
- Mock Git operations for testing repository URL extraction
- Test error scenarios and recovery mechanisms
- Test repository path resolution with various input formats
- Test automatic repository addition to status file

#### Integration Tests
- Test actual repository validation with real file system
- Test status file updates with real YAML files
- Test Git repository URL extraction with real repositories
- Test cross-platform path handling

#### End-to-End Tests
- Test complete workspace creation workflow with real repositories
- Test integration with existing workspace features
- Test workspace listing and management

#### Test Cases
- `TestWorkspace_Create_Success`: Test successful workspace creation
- `TestWorkspace_Create_RepositoryValidation`: Test repository validation
- `TestWorkspace_Create_PathResolution`: Test path resolution with various formats
- `TestWorkspace_Create_DuplicateRepositories`: Test duplicate repository detection
- `TestWorkspace_Create_InvalidRepositories`: Test invalid repository handling
- `TestWorkspace_Create_StatusIntegration`: Test status file integration
- `TestWorkspace_Create_RepositoryAddition`: Test automatic repository addition to status
- `TestWorkspace_Create_ErrorRecovery`: Test error handling and cleanup
- `TestWorkspace_Create_CrossPlatform`: Test cross-platform compatibility

### Dependencies

#### Direct Dependencies
- **Blocked by**: Features 2, 4, 8, 16 (workspace detection, validation, status management, init command)
- **Dependencies**: FS package, Status package, Git package, Config package
- **External**: Git command-line tool (for repository validation and URL extraction)
- **Adapters**: Use existing FS adapter for file system operations, Status adapter for status management, Git adapter for repository operations

#### Required Go Modules
- `github.com/spf13/cobra v1.7.0`: Command-line argument parsing
- `go.uber.org/mock v0.5.2`: Mocking framework for testing
- `github.com/stretchr/testify v1.8.4`: Testing utilities and assertions
- `encoding/json`: Go standard library for JSON parsing

### Success Criteria
1. Successfully create workspace entries in status.yaml with user-specified repositories
2. Support multiple repository input formats (status.yaml names, absolute paths, relative paths)
3. Automatically add new repositories to status.yaml for consistency
4. Automatically add created workspaces to status.yaml for tracking
5. Validate all repositories exist and are accessible before workspace creation
6. Extract repository URLs from Git remote origin for new repositories
7. Provide comprehensive error handling with clear recovery instructions
8. Support non-interactive operation with command-line parameters
9. Pass all unit, integration, and end-to-end tests
10. Use existing FS, Status, and Git adapters for all external operations
11. Maintain atomic operations where possible for workspace creation
12. Support cross-platform operation on Windows, macOS, and Linux

### Implementation Plan

#### Phase 1: Status Package Updates (Priority: High)
1. Update Workspace struct to contain worktree and repositories lists
2. Update AddWorkspaceParams to match new structure
3. Update status file operations for new workspace structure
4. Write unit tests for updated status operations

#### Phase 2: FS Package Extension (Priority: High)
1. Add path resolution methods to FS interface
2. Add repository validation methods
3. Update concrete implementation using Go standard library
4. Create comprehensive integration tests for new FS methods
5. Generate mock files using `go generate` and commit them

#### Phase 3: CM Package Extension (Priority: High)
1. Add CreateWorkspace method to CM interface
2. Implement workspace creation business logic
3. Add repository validation and path resolution
4. Add automatic repository addition to status file
5. Integrate with Status, FS, and Git adapters
6. Write unit tests using mocked dependencies

#### Phase 4: CLI Integration (Priority: Medium)
1. Create workspace command structure
2. Add workspace create subcommand
3. Integrate with CM package workspace creation
4. Add command-line argument parsing and validation
5. Test CLI integration with real workspaces

#### Phase 5: Testing and Validation (Priority: Medium)
1. Create comprehensive unit tests for all components
2. Create integration tests with real file system operations
3. Create end-to-end tests with real workspace creation

### Future Considerations
- Consider workspace templates for common repository combinations
- Plan for workspace cloning from existing workspaces
- Consider workspace sharing and collaboration features
- Plan for workspace-specific configuration and settings
- Consider workspace versioning and history tracking
- Plan for workspace import/export functionality

## Notes
- This feature builds upon existing workspace detection, validation, and status management capabilities
- Focus on reliability and proper error handling for workspace creation operations
- Ensure error messages are user-friendly and provide clear recovery instructions
- Consider adding debug logging for troubleshooting workspace creation
- Use build tags to organize tests: `unit` for unit tests, `integration` for real file system tests (adapters only), `e2e` for end-to-end tests
- Mock files should be committed to the repository as `*_gen.go`
- Exit immediately with exit code 1 on critical errors
- Support three output modes: quiet (errors to stderr only), verbose (detailed steps), normal (user interaction only)
- Maintain separation of concerns: FS adapter handles file system operations, Status adapter handles status management, Git adapter handles repository operations, CM handles business logic
- Ensure workspace operations are atomic where possible
- No partial success scenarios - either workspace creation succeeds completely or fails with cleanup
- Workspace entries are created in status.yaml with worktree and repositories lists
- Repository path resolution supports multiple input formats with clear priority order
- Automatic repository addition to status.yaml ensures consistency across CM operations
