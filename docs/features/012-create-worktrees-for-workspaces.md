# Feature 012: Create Worktrees for Workspaces

## Overview
Implement functionality to create Git worktrees for all repositories within a workspace. This feature will extend the existing workspace detection and validation capabilities to provide complete worktree management for multi-repository workspaces, allowing users to create worktrees across all repositories in a workspace simultaneously.

## Background
The Git WorkTree Manager (wtm) currently supports worktree creation for single repositories and has comprehensive workspace detection and validation. However, the workspace mode currently only validates repositories without creating worktrees. This feature will complete the workspace functionality by implementing worktree creation across all repositories in a workspace, enabling developers to work on multiple branches simultaneously across all repositories in their workspace.

## Requirements

### Functional Requirements
1. **Workspace Worktree Creation**: Create Git worktrees for all repositories in a workspace
2. **Branch Management**: Use user-provided branch name from command line argument across all repositories
3. **Directory Management**: Create worktree directories in the configured base path using full repository path structure for each repository
4. **Status Tracking**: Update the status file to track created worktrees for all repositories
5. **Collision Detection**: Prevent creation of worktrees with conflicting names/paths across all repositories
6. **Repository Validation**: Ensure all repositories in the workspace are valid Git repositories
7. **Git State Validation**: Ensure nothing prevents worktree creation in any repository (clean state)
8. **Error Recovery**: Handle failures during worktree creation and provide cleanup for all affected repositories
9. **User Feedback**: Provide clear feedback about worktree creation progress and results for each repository
10. **Configuration Integration**: Use existing configuration for base paths and status file location
11. **Atomic Operations**: Ensure workspace-wide operations are atomic where possible
12. **Complete Rollback Handling**: Stop on first failure and rollback all successful worktree creations

## Technical Specification

### Interface Design

#### Git Package Extension
**Existing Interface Methods** (already implemented):
- `CreateWorktree(repoPath, worktreePath, branch string) error`: Create a new worktree
- `GetCurrentBranch(repoPath string) (string, error)`: Get the current branch name
- `GetRepositoryName(repoPath string) (string, error)`: Get the repository name (remote origin URL)
- `IsClean(repoPath string) (bool, error)`: Check if the repository is in a clean state
- `BranchExists(repoPath, branch string) (bool, error)`: Check if a branch exists
- `CreateBranch(repoPath, branch string) error`: Create a new branch
- `WorktreeExists(repoPath, branch string) (bool, error)`: Check if a worktree exists for the branch

**Key Characteristics**:
- Extends existing Git package with worktree-specific operations
- Pure function-based operations
- No state management required
- Cross-platform compatibility
- Error handling with wrapped errors

#### WTM Package Extension
**New Interface Methods**:
- `createWorktreesForWorkspace(branch string) error`: Create worktrees for all repositories in workspace
- `validateWorkspaceForWorktreeCreation(branch string) error`: Validate workspace state before worktree creation

**Implementation Structure**:
- Extends existing `handleWorkspaceMode()` method in `wtm.go`
- Private helper methods for workspace worktree creation logic
- Integration with status management for multiple repositories
- Error handling with wrapped errors
- User feedback through logger for each repository

**Key Characteristics**:
- Builds upon existing workspace detection and validation logic
- Integrates with status management system for multiple repositories
- Provides comprehensive user feedback for each repository
- Handles both new and existing branches across all repositories
- Supports complete rollback scenarios (no partial success)

### Implementation Details

#### 1. Workspace Worktree Creation Algorithm
The workspace worktree creation will follow this algorithm:

**Key Components**:
- `createWorktreesForWorkspace()`: Main entry point for workspace worktree creation
- `validateWorkspaceForWorktreeCreation()`: Pre-creation validation for all repositories
- `createWorktreeForRepository()`: Individual repository worktree creation
- `cleanupFailedWorktrees()`: Cleanup mechanism for failed operations

**Implementation Flow**:
1. Load and validate workspace configuration (already implemented)
2. Validate all repositories in workspace (already implemented)
3. Pre-validate worktree creation for all repositories:
   - Check for existing worktrees in status file
   - Check for branch existence and create if needed
4. Update status file with worktree entries (using file locking)
5. Create worktree-specific workspace file in `$BASE_PATH/workspaces/`
6. Create worktree directories for all repositories:
   - Use repository name from remote origin URL (fallback to local path)
   - Create directory structure: `$BASE_PATH/{repo-name}/{branch-name}/`
7. Execute Git worktree creation commands for all repositories
8. Handle cleanup on failure (remove directories, status entries, and workspace file for all repositories)
9. Provide comprehensive user feedback for each repository

**Implementation Notes**:
- Use existing FS adapter for all file system operations
- Use existing Git adapter for all Git operations
- Update status file before worktree creation for proper cleanup
- Use existing file locking mechanism for concurrent access prevention
- Provide comprehensive error handling and cleanup
- Non-interactive operation (fail immediately on any issue)
- Create branches from current branch when they don't exist
- Handle repository name fallback when no remote origin is configured
- Support complete rollback scenarios (no partial success)

#### 2. Workspace Worktree Directory Structure
Worktrees will be created in the configured base path with the following structure for each repository:
```
$BASE_PATH/
├── status.yaml
├── workspaces/
│   └── {original-workspace-name}-{branch-name}.code-workspace
├── {repository-name-1}/
│   └── {branch-name}/
│       └── (worktree contents)
├── {repository-name-2}/
│   └── {branch-name}/
│       └── (worktree contents)
└── {repository-name-n}/
    └── {branch-name}/
        └── (worktree contents)
```

**Directory Naming**:
- Repository name: Extracted from remote origin URL (e.g., `github.com/lerenn/example`) with fallback to local path
- Branch name: User-provided branch name from command line argument (same across all repositories)
- Full path example: `$BASE_PATH/github.com/lerenn/example/feature-branch/`
- Fallback example: `$BASE_PATH/local-repo-name/feature-branch/` (when no remote origin)

#### 2.1. Worktree-Specific Workspace File Creation
A worktree-specific workspace file will be created to enable IDE opening with all worktrees:

**File Location**: `$BASE_PATH/workspaces/{original-workspace-name}-{branch-name}.code-workspace`

**File Naming Convention**: 
- Use the original workspace file name (without `.code-workspace` extension) as the base name
- Append the branch name with a hyphen separator
- Example: `MyProject-feat1.code-workspace`

**File Structure**: The worktree workspace file will contain all repositories from the original workspace but with paths pointing to the worktree directories instead of the original repository paths.

**Creation Timing**: The worktree-specific workspace file will be created:
1. After status file update (to track the workspace file path)
2. Before worktree creation (to ensure it's available for IDE opening)

**Deletion Timing**: The worktree-specific workspace file will be deleted:
1. After worktree deletion (when using `wtm delete` command)
2. Before status file update (to maintain consistency)

**IDE Opening**: When IDE opening is triggered, it will open the worktree-specific workspace file directly.

**Error Handling**: If workspace file creation fails, it will be treated as a critical failure that triggers complete rollback (remove from status file as worktrees are not created yet).

**Example Worktree Workspace File**:
```json
{
  "folders": [
    {
      "name": "repo1",
      "path": "/Users/lfradin/.wtm/github.com/lerenn/example/feat1"
    },
    {
      "name": "repo2", 
      "path": "/Users/lfradin/.wtm/github.com/lerenn/toto/feat1"
    }
  ]
}
```

#### 3. Status File Integration
Worktree creation will update the status file to track worktrees for all repositories with workspace information:
- Repository URL
- Repository Path
- Branch name (same across all repositories)
- Original workspace file path

**Status File Structure**:
```yaml
repositories:
  - url: github.com/lerenn/example
    branch: feat1
    path: /Users/lfradin/Code/src/github.com/lerenn/example
    workspace: /Users/lfradin/Code/src/github.com/lerenn/myworkspace.code-workspace
  - url: github.com/lerenn/toto
    branch: feat1
    path: /Users/lfradin/Code/src/github.com/lerenn/toto
    workspace: /Users/lfradin/Code/src/github.com/lerenn/myworkspace.code-workspace
```

**Status Manager Enhancement**:
The Status Manager will include a private computed map for efficient workspace operations:
```go
workspaces map[string]map[string][]Repository
```
Where:
- First string: Original workspace name (from workspace file name or JSON name field)
- Second string: Branch name
- `[]Repository`: Array of repositories for that workspace and branch

This map will be:
- Computed when Status is loaded
- Used internally for efficient workspace operations (listing, deletion)
- Not persisted to the status file (only the flat array is persisted)
- Thread-safe for concurrent access

#### 4. Workspace Validation for Worktree Creation
Before creating worktrees, the system will validate:

**Pre-Creation Validation**:
- All repositories are valid Git repositories (already implemented)
- No existing worktrees for the specified branch in any repository (status file check)
- Branch exists or can be created in all repositories
- Sufficient disk space for all worktrees
- No path conflicts between repositories

**Validation Flow**:
1. Check status file for existing worktrees with the specified branch name
2. Check branch existence in each repository
3. Validate directory creation permissions for all worktree paths
4. Check for path conflicts between different repositories

**Worktree Path Derivation**:
Worktree paths will be derived from repository URL and branch name:
- Pattern: `$BASE_PATH/{repository-url}/{branch-name}/`
- Example: `$BASE_PATH/github.com/lerenn/example/feat1/`

### Error Handling

#### Error Types
1. **WorkspaceWorktreeExistsError**: When worktrees already exist for the specified branch in any repository
2. **WorkspaceGitCommandError**: When Git operations fail in any repository
3. **WorkspaceStatusUpdateError**: When status file update fails
4. **WorkspaceDirectoryExistsError**: When worktree directories already exist
5. **WorkspacePartialFailureError**: When some repositories succeed and others fail
6. **WorkspacePathConflictError**: When there are path conflicts between repositories
7. **WorkspaceInsufficientSpaceError**: When there's insufficient disk space for all worktrees
8. **WorkspaceFileCreationError**: When worktree-specific workspace file creation fails

#### Error Recovery
- Remove worktree directories for failed repositories (recursively until parent directory is not empty)
- Remove status file entries for failed repositories
- Use existing file locking mechanism for status file operations
- Provide clear error messages with recovery instructions
- Ensure complete rollback to initial state for failed repositories
- Support partial success scenarios with proper cleanup

#### Partial Success Handling
- Stop on first failure and rollback all successful worktree creations
- Remove worktree directories for all repositories (successful and failed)
- Remove status file entries for all repositories
- Remove worktree-specific workspace file if created
- Provide clear error messages with recovery instructions
- Maintain atomic status file updates
- No partial success scenarios - either all worktrees succeed or all fail

### User Interface

#### Command Line Interface
The feature will be accessible through the existing `wtm` command:
```bash
# Create worktrees for all repositories in workspace with specified branch
wtm create feature-name

# Create worktrees for all repositories in workspace with existing branch
wtm create existing-branch
```

**Command Structure**:
- Branch name is provided as the first argument to the `create` command
- No optional flags for branch specification
- Non-interactive operation (fails on any issue)
- Works with both single repository and workspace modes

#### User Feedback
- **Verbose mode**: Detailed progress information for each repository (validation steps, Git operations, status updates)
- **Normal mode**: Success/error messages for each repository with summary
- **Non-interactive**: No prompts or confirmations, fails immediately on any issue
- **Progress reporting**: Show progress for each repository in the workspace
- **Summary reporting**: Provide summary of successful and failed worktree creations

#### Workspace Listing
When listing worktrees for workspace mode, the output will be compact and organized by workspace:

**Display Format**:
```
Workspace {workspace-name}:
  - {branch-name-1}
  - {branch-name-2}
  - {branch-name-n}
```

**Example Output**:
```
Workspace toto:
  - feat1
  - feat2
```

**Implementation Notes**:
- Group worktrees by original workspace name
- Show unique branch names for each workspace
- Display only repositories mentioned in the original `.code-workspace` file
- Compact format to avoid repetition since all repositories use the same branch name
- Show worktrees of the repositories mentioned in the original `.code-workspace` file

**Feedback Examples**:
```
Found workspace: MyProject
Creating worktrees for branch: feature-branch
  [1/3] github.com/lerenn/repo1: Creating worktree...
  [1/3] github.com/lerenn/repo1: ✓ Worktree created successfully
  [2/3] github.com/lerenn/repo2: Creating worktree...
  [2/3] github.com/lerenn/repo2: ✓ Worktree created successfully
  [3/3] github.com/lerenn/repo3: Creating worktree...
  [3/3] github.com/lerenn/repo3: ✗ Failed to create worktree: branch already exists
Rolling back all worktrees due to failure...
Workspace worktree creation failed: all worktrees rolled back
```

### Testing Strategy

#### Unit Tests
- Mock Git operations for testing workspace worktree creation logic
- Mock file system operations for directory creation across multiple repositories
- Mock status management for status file updates with multiple repositories
- Test error scenarios and recovery mechanisms for workspace-wide operations
- Test complete rollback scenarios (no partial success)

#### Integration Tests
- Test actual Git worktree creation with real workspace repositories
- Test file system operations with real directories for multiple repositories
- Test status file updates with real YAML files containing multiple repositories
- Test workspace-wide error handling and cleanup
- Test worktree-specific workspace file creation and deletion

#### End-to-End Tests
- Test complete workspace worktree creation workflow with real repositories
- Test workspace listing functionality
- Test workspace deletion with proper cleanup
- Test IDE opening with worktree-specific workspace files

#### Test Cases
- `TestWorkspace_CreateWorktrees_Success`: Test successful worktree creation for all repositories
- `TestWorkspace_CreateWorktrees_CompleteRollback`: Test complete rollback on first failure
- `TestWorkspace_CreateWorktrees_ExistingWorktrees`: Test when worktrees already exist
- `TestWorkspace_CreateWorktrees_InvalidRepositories`: Test when some repositories are invalid
- `TestWorkspace_CreateWorktrees_GitErrors`: Test Git command failures in repositories
- `TestWorkspace_CreateWorktrees_StatusFileErrors`: Test status file update failures
- `TestWorkspace_CreateWorktrees_PathConflicts`: Test path conflicts between repositories
- `TestWorkspace_CreateWorktrees_CleanupOnFailure`: Test cleanup mechanisms on failure
- `TestWorkspace_CreateWorktrees_WorkspaceFileCreation`: Test worktree-specific workspace file creation
- `TestWorkspace_CreateWorktrees_WorkspaceFileDeletion`: Test workspace file deletion during cleanup
- `TestWorkspace_ListWorktrees_CompactFormat`: Test compact workspace listing format
- `TestWorkspace_DeleteWorktrees_CompleteCleanup`: Test workspace-wide deletion with cleanup

### Dependencies
- **Blocked by**: Features 2, 4, 5, 6, 7, 8, 9 (workspace detection, validation, status management, single repository worktree creation)
- **Dependencies**: Git package, FS package, Status package, Config package, Workspace package
- **External**: Git command-line tool
- **Adapters**: Use existing FS adapter for file system operations, Git adapter for Git operations

### Success Criteria
1. Successfully create worktrees for all repositories in a workspace using user-provided branch names
2. Use full repository path structure for worktree directories (e.g., `$BASE_PATH/github.com/lerenn/example/branch-name/`)
3. Prevent collisions by checking both Git worktrees and status file entries across all repositories
4. Update status file correctly with repository names from remote origin URLs and workspace information
5. Create worktree-specific workspace files in `$BASE_PATH/workspaces/` with proper naming convention
6. Provide comprehensive user feedback for each repository in verbose and normal modes
7. Handle errors gracefully with complete cleanup for all repositories (stop on first failure)
8. Support workspace listing with compact format organized by workspace name
9. Pass all unit, integration, and end-to-end tests
10. Use existing FS and Git adapters for all external operations
11. Maintain atomic operations where possible for workspace-wide changes
12. Support workspace-wide deletion with proper cleanup of workspace files
13. Open worktree-specific workspace files directly in IDEs
14. Handle workspace file creation failures as critical errors with complete rollback

### Implementation Plan

#### Phase 1: Workspace Worktree Creation Logic (Priority: High)
1. Extend `workspace.go` with worktree creation methods
2. Implement `createWorktreesForWorkspace()` method
3. Implement `validateWorkspaceForWorktreeCreation()` method
4. Implement individual repository worktree creation logic
5. Add error handling and cleanup mechanisms
6. Write unit tests using mocked dependencies

#### Phase 2: WTM Integration (Priority: High)
1. Update `handleWorkspaceMode()` method in `wtm.go`
2. Integrate workspace worktree creation with existing workflow
3. Add user feedback and progress reporting
4. Handle complete rollback scenarios
5. Update error handling and recovery mechanisms

#### Phase 3: Status Management Integration (Priority: Medium)
1. Extend status management for multiple repositories
2. Implement atomic status file updates for workspace operations
3. Add cleanup mechanisms for failed worktree creations
4. Test status file consistency across workspace operations

#### Phase 4: Testing and Validation (Priority: Medium)
1. Create comprehensive unit tests for workspace worktree creation
2. Create integration tests with real workspace repositories
3. Test error scenarios and recovery mechanisms
4. Test complete rollback scenarios (no partial success)
5. Test worktree-specific workspace file creation and deletion
6. Test workspace listing functionality
7. Test workspace deletion with proper cleanup
8. Performance testing with large workspaces

### Future Considerations
- Consider parallel worktree creation for improved performance
- Plan for workspace-specific worktree management features
- Consider workspace templates for common worktree configurations
- Plan for workspace-wide branch synchronization
- Consider workspace-specific IDE opening configurations
- Plan for workspace worktree cleanup and maintenance features

## Notes
- This feature builds upon existing workspace detection and validation capabilities
- Focus on reliability and proper error handling for workspace-wide operations
- Ensure error messages are user-friendly and provide clear recovery instructions
- Consider adding debug logging for troubleshooting workspace operations
- Use build tags to organize tests: `unit` for unit tests, `integration` for real file system tests (adapters only), `e2e` for end-to-end tests
- Mock files should be committed to the repository as `*_gen.go`
- Exit immediately with exit code 1 on critical errors
- Support three output modes: quiet (errors to stderr only), verbose (detailed steps), normal (user interaction only)
- Maintain separation of concerns: FS adapter handles file system operations, Git adapter handles Git operations, WTM handles business logic
- Ensure workspace operations are atomic where possible
- No partial success scenarios - either all worktrees succeed or all fail with complete rollback
- Worktree-specific workspace files are created in `$BASE_PATH/workspaces/` with naming convention `{original-name}-{branch-name}.code-workspace`
- IDE opening uses worktree-specific workspace files directly
- Workspace listing shows compact format organized by workspace name with unique branch names
