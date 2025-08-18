# Feature 011: Implement Worktree Deletion

## Overview

Add the ability to safely delete Git worktrees using the `delete` command. This feature will allow users to remove worktrees and clean up both the file system and Git state, with proper validation and safety measures.

## Background

The Code Manager (cm) currently supports creating worktrees but lacks the ability to delete them. Users need a safe way to remove worktrees when they're no longer needed, ensuring proper cleanup of both the worktree directory and Git's internal worktree tracking.

## Command Syntax

### Delete Command
```bash
cm delete <branch-name> [options]
```

### Examples

```bash
# Delete a worktree with confirmation
cm delete feature/new-feature

# Force delete without confirmation
cm delete bugfix/issue-123 --force

# Force delete without confirmation (short flag)
cm delete hotfix/critical-fix -f

# Delete with verbose output
cm delete feature/old-feature -v
```

## Requirements

### Functional Requirements

1. **Worktree Deletion**: Remove worktree directories from the file system
2. **Git State Cleanup**: Remove worktree references from Git's internal tracking
3. **Status Management**: Remove worktree entries from the status.yaml file
4. **Safety Validation**: Ensure worktree exists before deletion
5. **Confirmation**: Require user confirmation with detailed information unless --force/-f flag is used
6. **Error Recovery**: Handle failures during deletion and provide cleanup
7. **User Feedback**: Provide clear feedback about deletion progress and results
8. **Branch Name Resolution**: Support deletion by branch name (not full path)
9. **Mode Detection**: Automatically detect single repository mode and handle accordingly
10. **Workspace Placeholder**: Prepare structure for future workspace deletion support

### Non-Functional Requirements

1. **Performance**: Worktree deletion should complete within reasonable time (< 3 seconds)
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
- `RemoveWorktree(repoPath, worktreePath string) error`: Remove worktree from Git's tracking
- `GetWorktreePath(repoPath, branch string) (string, error)`: Get the path of a worktree for a branch

**Key Characteristics**:
- Extends existing Git package with worktree deletion operations
- Pure function-based operations
- No state management required
- Cross-platform compatibility
- Error handling with wrapped errors

#### CM Package Extension
**New Interface Methods**:
- `DeleteWorkTree(branch string, force bool) error`: Main entry point for worktree deletion

#### Repository Package Extension
**New Interface Methods**:
- `DeleteWorktree(branch string, force bool) error`: Delete worktree for single repository

#### Workspace Package Extension
**New Interface Methods**:
- `DeleteWorktree(branch string, force bool) error`: Placeholder for future workspace deletion support

**Implementation Structure**:
- CM package: Main entry point that detects mode and delegates to appropriate handler
- Repository package: Implements deletion logic for single repository mode
- Workspace package: Placeholder for future workspace deletion support
- Integration with status management
- Error handling with wrapped errors
- User feedback through logger

**Key Characteristics**:
- Builds upon existing detection and validation logic
- Follows existing architectural patterns (repository.go, workspace.go)
- Integrates with status management system
- Provides comprehensive user feedback
- Handles both single repos and workspaces

### Implementation Details

#### 1. Git Package Implementation
The Git package will be extended with worktree deletion operations:

**Key Components**:
- `RemoveWorktree()`: Executes `git worktree remove` command
- `GetWorktreePath()`: Parses `git worktree list` output to find worktree path

**Implementation Notes**:
- Use `exec.Command()` for Git operations
- Handle Git command failures with proper error wrapping
- Parse `git worktree list` output to find worktree paths
- Provide detailed error messages for debugging

#### 2. CM Package Implementation
The CM package will implement the main entry point for worktree deletion:

**Key Components**:
- `DeleteWorkTree()`: Main entry point that detects mode and delegates to appropriate handler
- Mode detection (single repo vs workspace)
- Delegation to repository or workspace deletion methods
- Integration with status management
- Safety validation and confirmation

**Deletion Process**:
1. **CM Package**: Detect project mode (single repo vs workspace)
2. **CM Package**: Delegate to appropriate handler (repository or workspace)
3. **Repository/Workspace Package**: Validate worktree exists in status.yaml
4. **Repository/Workspace Package**: Get worktree path from Git
5. **Repository/Workspace Package**: Prompt for confirmation with detailed information (unless --force/-f)
6. **Repository/Workspace Package**: Remove worktree from Git tracking
7. **Repository/Workspace Package**: Delete worktree directory
8. **Repository/Workspace Package**: Remove entry from status.yaml
9. **Repository/Workspace Package**: Provide user feedback

#### 3. Repository Package Implementation
The Repository package will implement the worktree deletion logic for single repositories:

**Key Components**:
- `DeleteWorktree()`: Main deletion method for single repository mode
- Worktree validation and path resolution
- Git worktree removal
- File system cleanup
- Status file updates
- User confirmation handling

#### 4. Workspace Package Implementation
The Workspace package will provide a placeholder for future workspace deletion support:

**Key Components**:
- `DeleteWorktree()`: Placeholder method that returns an error indicating workspace deletion is not yet supported
- Future implementation will handle multiple repositories in workspace mode

#### 5. CLI Integration
Add new delete command to the CLI:

**Command Structure**:
```go
var deleteCmd = &cobra.Command{
    Use:   "delete [branch-name]",
    Short: "Delete a worktree",
    Long:  "Delete a worktree and clean up Git state",
    Args:  cobra.ExactArgs(1),
    RunE:  runDelete,
}

// Add flags
deleteCmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")
```

### Error Handling

#### Error Types

```go
// pkg/git/errors.go
var (
    ErrWorktreeNotFound = errors.New("worktree not found")
    ErrWorktreeDeletionFailed = errors.New("failed to delete worktree")
    ErrWorktreePathNotFound = errors.New("worktree path not found")
)
```

// pkg/cm/errors.go
var (
    ErrWorktreeNotInStatus = errors.New("worktree not found in status file")
    ErrDeletionCancelled = errors.New("deletion cancelled by user")
    ErrWorktreeValidationFailed = errors.New("worktree validation failed")
)
```

#### Error Messages

- `"Worktree 'feature/branch' not found"`
- `"Failed to delete worktree: git worktree remove failed"`
- `"Deletion cancelled by user"`
- `"Worktree 'feature/branch' not found in status file"`

### Safety Features

1. **Existence Validation**: Verify worktree exists before deletion
2. **Confirmation**: Require user confirmation with detailed information unless --force/-f flag
3. **Atomic Operations**: Ensure status file updates are atomic
4. **Error Recovery**: Handle partial failures gracefully
5. **Path Validation**: Validate worktree paths before deletion
6. **Git-First Deletion**: Delete worktree with Git before updating status file

## Configuration

No new configuration required. Uses existing configuration for:
- Base paths for worktree locations
- Status file location
- Logging configuration

## Testing Strategy

### Unit Tests
- Worktree deletion logic using mocked Git and FS adapters
- Error handling with specific error types
- Status management integration
- CLI command validation

### Integration Tests
- Real worktree deletion with actual Git operations
- File system cleanup verification
- Status file updates

### End-to-End Tests
- Complete deletion workflow from CLI
- Cross-platform compatibility
- Error scenarios and recovery

## Dependencies

### New Dependencies
- None (uses standard library for process execution)

### Modified Dependencies
- `pkg/git`: Add worktree deletion methods
- `pkg/cm`: Add main deletion entry point
- `pkg/cm/repository.go`: Add repository deletion method
- `pkg/cm/workspace.go`: Add workspace deletion placeholder
- `cmd/cm`: Add new CLI command

## Migration and Backward Compatibility

- This is a new feature, no migration required
- Backward compatible with existing commands
- Optional feature that doesn't affect core functionality

## Implementation Decisions

1. **Git Interface**: Add `RemoveWorktree()`, `GetWorktreePath()` methods
2. **Architecture**: Follow existing patterns with CM as main entry point, delegating to repository/workspace packages
3. **Deletion Process**: 
   - Validate existence in status.yaml
   - Get worktree path from Git
   - Remove from Git tracking first
   - Delete directory
   - Update status.yaml
4. **Safety**: Require confirmation with detailed information unless --force/-f flag
5. **Error Handling**: 
   - `ErrWorktreeNotFound` when worktree doesn't exist
   - `ErrDeletionCancelled` when user cancels
6. **CLI Structure**: `cm delete <branch-name> [--force/-f]` (following create/open pattern)
7. **Testing**: Unit tests with mocked dependencies, integration tests with real Git
8. **Cross-Platform**: Use standard library for file operations and process execution
9. **Logging**: Log all operations with appropriate verbosity levels
10. **Status Management**: Atomic updates to status.yaml file
11. **Validation**: Check worktree existence in both Git and status file
12. **Mode Support**: Single repository mode only, with placeholder for workspace mode
13. **Confirmation**: Simple yes/no prompt with detailed information about what will be deleted
14. **Package Structure**: Repository deletion logic in `repository.go`, workspace placeholder in `workspace.go`
