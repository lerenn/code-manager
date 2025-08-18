# Feature 009: Create Worktrees for Single Repositories

## Overview
Implement functionality to create Git worktrees for single repositories. This feature will allow users to create worktrees from their current Git repository, managing branch creation, worktree directory creation, and status tracking.

## Background
The Code Manager (cm) needs to provide the core worktree creation functionality for single repositories. This is a foundational feature that enables developers to work on multiple branches simultaneously by creating separate worktree directories. This feature builds upon the existing detection and validation capabilities to provide a complete worktree management solution.

## Requirements

### Functional Requirements
1. **Worktree Creation**: Create Git worktrees for the current repository
2. **Branch Management**: Use user-provided branch name from command line argument
3. **Directory Management**: Create worktree directories in the configured base path using full repository path structure
4. **Status Tracking**: Update the status file to track created worktrees
5. **Collision Detection**: Prevent creation of worktrees with conflicting names/paths
6. **Repository Validation**: Ensure the current directory is a valid Git repository
7. **Git State Validation**: Ensure nothing prevents worktree creation (clean state)
8. **Error Recovery**: Handle failures during worktree creation and provide cleanup
9. **User Feedback**: Provide clear feedback about worktree creation progress and results
10. **Configuration Integration**: Use existing configuration for base paths and status file location

### Non-Functional Requirements
1. **Performance**: Worktree creation should complete within reasonable time (< 5 seconds)
2. **Reliability**: Handle Git command failures, file system errors, and concurrent access
3. **Cross-Platform**: Work on Windows, macOS, and Linux
4. **Testability**: Use Uber gomock for mocking Git and file system operations
5. **Minimal Dependencies**: Use only Go standard library + gomock for operations
6. **Atomic Operations**: Ensure status file updates are atomic
7. **Safe Cleanup**: Provide proper cleanup on failure

## Technical Specification

### Interface Design

#### Git Package Extension
**New Interface Methods**:
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

#### CM Package Extension
**New Interface Methods**:
- `CreateWorkTree(branch string) error`: Main entry point (updated to accept branch parameter)
- `createWorktreeForSingleRepo(branch string) error`: Create worktree for single repository

**Implementation Structure**:
- Extends existing `CreateWorkTree()` method
- Private helper methods for worktree creation logic
- Integration with status management
- Error handling with wrapped errors
- User feedback through logger

**Key Characteristics**:
- Builds upon existing detection and validation logic
- Integrates with status management system
- Provides comprehensive user feedback
- Handles both new and existing branches

### Implementation Details

#### 1. Git Package Implementation
The Git package will be extended with worktree-specific operations:

**Key Components**:
- `CreateWorktree()`: Executes `git worktree add` command
- `GetCurrentBranch()`: Executes `git branch --show-current`
- `GetRepositoryName()`: Extracts repository name from remote origin URL with fallback to local path
- `IsClean()`: Checks if working directory is clean (placeholder for future validation)
- `BranchExists()`: Checks if branch exists locally or remotely
- `CreateBranch()`: Creates new branch from current branch using worktree
- `WorktreeExists()`: Check if a worktree exists for the branch

**Implementation Notes**:
- Use `exec.Command()` for Git operations
- Handle Git command failures with proper error wrapping
- Support both local and remote branch checking
- Provide detailed error messages for debugging
- Create branches from current branch when they don't exist
- Fallback to local repository path when no remote origin is configured

#### 2. CM Package Implementation
The CM package will implement the worktree creation logic:

**Key Components**:
- Extend existing `CreateWorkTree()` method to accept branch parameter
- Add `createWorktreeForSingleRepo(branch string)` helper method
- Integration with status management
- Collision detection and prevention
- User feedback and progress reporting

**Implementation Flow**:
1. Validate current directory is a Git repository
2. Extract repository name from remote origin URL (fallback to local path if no remote)
3. Check if worktree already exists (status file only, fail on discrepancy)
4. Validate repository state (placeholder for future validation)
5. Create worktree directory structure in `.cm/{repo-name}/{branch-name}/`
6. Update status file with worktree entry (using file locking)
7. Execute Git worktree creation command
8. Handle cleanup on failure (remove directory and status entry)

**Implementation Notes**:
- Use existing FS adapter for all file system operations
- Use existing Git adapter for all Git operations
- Update status file before worktree creation for proper cleanup
- Use existing file locking mechanism for concurrent access prevention
- Provide comprehensive error handling and cleanup
- Non-interactive operation (fail immediately on any issue)
- Create branches from current branch when they don't exist
- Handle repository name fallback when no remote origin is configured

#### 3. Worktree Directory Structure
Worktrees will be created in the configured base path with the following structure:
```
$HOME/.cm/
├── status.yaml
└── {repository-name}/
    └── {branch-name}/
        └── (worktree contents)
```

**Directory Naming**:
- Repository name: Extracted from remote origin URL (e.g., `github.com/lerenn/example`) with fallback to local path
- Branch name: User-provided branch name from command line argument
- Full path example: `$HOME/.cm/github.com/lerenn/example/feature-branch/`
- Fallback example: `$HOME/.cm/local-repo-name/feature-branch/` (when no remote origin)

#### 4. Status File Integration
Worktree creation will update the status file to track:
- Repository name
- Branch name
- Worktree path
- Creation timestamp (optional)

### Error Handling

#### Error Types
1. **RepositoryNotCleanError**: When repository has uncommitted changes or other issues preventing worktree creation (placeholder for future validation)
2. **WorktreeExistsError**: When a worktree already exists for the specified branch (checked in status file only)
3. **GitCommandError**: When Git operations fail
4. **StatusUpdateError**: When status file update fails
5. **DirectoryExistsError**: When worktree directory already exists
6. **DiscrepancyError**: When there's a discrepancy between Git worktrees and status file entries

#### Error Recovery
- Remove worktree directory (recursively until parent directory is not empty)
- Remove status file entry if worktree creation fails (since status is updated before worktree creation)
- Use existing file locking mechanism for status file operations
- Provide clear error messages with recovery instructions
- Ensure complete rollback to initial state on any failure

### User Interface

#### Command Line Interface
The feature will be accessible through the existing `cm` command:
```bash
# Create worktree with specified branch
cm create feature-name

# Create worktree with existing branch
cm create existing-branch
```

**Command Structure**:
- Branch name is provided as the first argument to the `create` command
- No optional flags for branch specification
- Non-interactive operation (fails on any issue)

#### User Feedback
- Verbose mode: Detailed progress information (validation steps, Git operations, status updates)
- Normal mode: Success/error messages only
- Non-interactive: No prompts or confirmations, fails immediately on any issue

### Testing Strategy

#### Unit Tests
- Mock Git operations for testing worktree creation logic
- Mock file system operations for directory creation
- Mock status management for status file updates
- Test error scenarios and recovery mechanisms

#### Integration Tests
- Test actual Git worktree creation with real repositories
- Test file system operations with real directories
- Test status file updates with real YAML files

### Dependencies
- **Blocked by**: Features 4, 5, 6, 7 (detection, validation, status management)
- **Dependencies**: Git package, FS package, Status package, Config package
- **External**: Git command-line tool
- **Adapters**: Use existing FS adapter for file system operations, Git adapter for Git operations

### Success Criteria
1. Successfully create worktrees for single repositories using user-provided branch names
2. Use full repository path structure for worktree directories (e.g., `.cm/github.com/lerenn/example/branch-name/`)
3. Prevent collisions by checking both Git worktrees and status file entries
4. Update status file correctly with repository name from remote origin URL
5. Provide comprehensive user feedback in verbose and normal modes
6. Handle errors gracefully with complete cleanup (remove status entry if worktree creation fails)
7. Pass all unit and integration tests
8. Use existing FS and Git adapters for all external operations
