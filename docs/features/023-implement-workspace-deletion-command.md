# Feature 023: Implement Workspace Deletion Command

## Overview
Implement a `workspace delete` command that allows users to completely remove workspace definitions from the status.yaml file. The command will delete all associated worktrees, remove workspace files, and clean up the workspace entry from the status file. This feature provides complete workspace lifecycle management by complementing the existing workspace creation functionality.

## Background
The Code Manager (cm) currently supports workspace creation, worktree management within workspaces, and individual worktree deletion. However, there's no capability to completely delete an entire workspace and all its associated resources. This feature will complete the workspace management functionality by allowing users to remove workspace entries and clean up all associated worktrees and files.

## Requirements

### Functional Requirements
1. **Workspace Deletion Command**: Create a new `workspace delete` command with workspace name and optional force flag
2. **Complete Worktree Cleanup**: Delete all worktrees associated with the workspace from all repositories
3. **Workspace File Removal**: Delete both main workspace file and all worktree-specific workspace files
4. **Status File Cleanup**: Remove the workspace entry completely from status.yaml
5. **Repository Preservation**: Keep individual repository entries in status.yaml (they may be used by other workspaces)
6. **Confirmation System**: Require user confirmation before deletion (unless --force flag is used)
7. **Error Handling**: Stop and report error if any worktree deletion fails
8. **Worktree Listing**: Provide reusable worktree listing functionality for workspace mode
9. **Non-Interactive Operation**: All parameters provided via command line (no prompts except confirmation)

### Non-Functional Requirements
1. **Performance**: Workspace deletion should complete within 10 seconds for typical workspaces
2. **Reliability**: Handle file system errors and permission issues gracefully
3. **Cross-Platform**: Work consistently on Windows, macOS, and Linux
4. **Testability**: Support unit testing with mocked dependencies
5. **Minimal Dependencies**: Use only Go standard library and existing CM packages

## Technical Specification

### Interface Design

#### Status Package Extension
**New Interface Methods**:
- `RemoveWorkspace(workspaceName string) error`: Remove workspace entry from status file
- `GetWorkspaceByName(workspaceName string) (*Workspace, error)`: Retrieve workspace by name (not path)

**Key Characteristics**:
- **NO direct file system access** - all operations go through adapters
- **ONLY unit tests** using mocked dependencies
- Business logic for workspace removal workflow
- Error handling with wrapped errors

#### CM Package Extension
**New Interface Methods**:
- `DeleteWorkspace(params DeleteWorkspaceParams) error`: Delete workspace and all associated resources
- `ListWorkspaceWorktrees(workspaceName string) ([]status.WorktreeInfo, error)`: List all worktrees for a workspace

**New Data Structures**:
```go
type DeleteWorkspaceParams struct {
    WorkspaceName string // Name of the workspace to delete
    Force         bool   // Skip confirmation prompts
}
```

**Key Characteristics**:
- **NO direct file system access** - all operations go through adapters
- **ONLY unit tests** using mocked dependencies
- Business logic for workspace deletion workflow
- Worktree deletion orchestration
- Error handling with wrapped errors

#### Workspace Mode Package Extension
**New Interface Methods**:
- `DeleteWorkspace(workspaceName string, force bool) error`: Delete entire workspace and all resources
- `ListWorkspaceWorktrees(workspaceName string) ([]status.WorktreeInfo, error)`: List worktrees for workspace

**Updated list_worktrees.go**:
- Replace existing code with reusable worktree listing functionality
- Support both workspace mode and general workspace worktree listing
- Return comprehensive worktree information for deletion operations

**Key Characteristics**:
- **NO direct file system access** - all operations go through adapters
- **ONLY unit tests** using mocked dependencies
- Business logic for workspace deletion workflow
- Worktree deletion orchestration
- Error handling with wrapped errors

### Implementation Details

#### 1. Workspace Deletion Algorithm
The workspace deletion will follow this algorithm:

**Key Components**:
- `DeleteWorkspace()`: Main entry point for workspace deletion
- `ListWorkspaceWorktrees()`: List all worktrees associated with workspace
- `DeleteWorktreeFromWorkspace()`: Delete individual worktree from workspace
- `RemoveWorkspaceFromStatus()`: Remove workspace entry from status file
- `DeleteWorkspaceFiles()`: Delete workspace and worktree-specific files

**Implementation Flow**:
1. Validate workspace name exists in status.yaml
2. List all worktrees associated with the workspace
3. Show confirmation prompt with deletion summary (unless --force)
4. Delete all worktrees associated with the workspace:
   - For each worktree, delete the worktree directory
   - Remove worktree entries from status file
   - Delete worktree-specific workspace files
5. Delete main workspace file
6. Remove workspace entry from status file
7. Report success or failure

**Implementation Notes**:
- Use existing FS adapter for all file system operations
- Use existing Status adapter for status file operations
- Use existing Git adapter for worktree operations
- Provide comprehensive error handling and cleanup
- Stop on first error (no partial success)
- Support confirmation prompts with detailed summary
- Preserve repository entries in status file

#### 2. Worktree Listing Enhancement
The worktree listing functionality will be enhanced to support workspace deletion:

**Enhanced list_worktrees.go**:
- Replace existing implementation with comprehensive worktree listing
- Support listing worktrees by workspace name
- Return detailed worktree information including repository, branch, and file paths
- Support both workspace mode and general workspace operations
- Provide reusable functionality for future commands

**Worktree Information Structure**:
```go
type WorkspaceWorktreeInfo struct {
    Repository string // Repository URL
    Branch     string // Branch name
    WorktreePath string // Path to worktree directory
    WorkspaceFile string // Path to worktree-specific workspace file
}
```

#### 3. Confirmation System
The confirmation system will provide detailed information about what will be deleted:

**Confirmation Prompt Structure**:
```
Are you sure you want to delete workspace 'my-workspace'?

This will delete:
- 3 worktrees across 2 repositories:
  * repo1: feature-branch-1, feature-branch-2
  * repo2: feature-branch-3
- Workspace file: /path/to/my-workspace.code-workspace
- 3 worktree-specific workspace files

Type 'yes' to confirm deletion: 
```

**Confirmation Behavior**:
- Show detailed summary of what will be deleted
- Require typing "yes" to confirm
- Default to "no" if user presses Enter
- Skip confirmation with --force flag

#### 4. Error Handling and Recovery
The error handling will ensure safe operation and clear error reporting:

**Error Types**:
1. **WorkspaceNotFoundError**: When workspace name doesn't exist in status.yaml
2. **WorktreeDeletionError**: When worktree deletion fails
3. **FileDeletionError**: When workspace file deletion fails
4. **StatusUpdateError**: When status file update fails
5. **ConfirmationCancelledError**: When user cancels deletion

**Error Recovery**:
- Stop on first error and report failure
- Leave status file in current state (no rollback)
- Provide clear error messages with recovery instructions
- Use existing file locking mechanism for status file operations

### User Interface

#### Command Line Interface
The feature will be accessible through a new workspace delete command:
```bash
# Delete workspace with confirmation
cm workspace delete my-workspace

# Delete workspace without confirmation
cm workspace delete my-workspace --force
```

**Command Structure**:
```bash
cm workspace delete <workspace-name> [--force]
```

**Parameters**:
- `<workspace-name>`: Name of the workspace to delete
- `--force`: Skip confirmation prompts

#### User Feedback
- **Confirmation mode**: Detailed deletion summary with confirmation prompt
- **Force mode**: Direct deletion with progress reporting
- **Verbose mode**: Detailed progress information for each deletion step
- **Normal mode**: Success/error messages with summary
- **Error mode**: Clear error messages with recovery instructions

**Feedback Examples**:
```
Deleting workspace: my-workspace
Found 3 worktrees to delete:
  - repo1: feature-branch-1
  - repo1: feature-branch-2  
  - repo2: feature-branch-3
Deleting worktrees...
  ✓ Deleted worktree: repo1/feature-branch-1
  ✓ Deleted worktree: repo1/feature-branch-2
  ✓ Deleted worktree: repo2/feature-branch-3
Deleting workspace files...
  ✓ Deleted: /path/to/my-workspace.code-workspace
  ✓ Deleted: /path/to/my-workspace-feature-branch-1.code-workspace
  ✓ Deleted: /path/to/my-workspace-feature-branch-2.code-workspace
  ✓ Deleted: /path/to/my-workspace-feature-branch-3.code-workspace
Removing workspace from status file...
✓ Workspace 'my-workspace' deleted successfully
```

### Testing Strategy

#### Unit Tests
- Mock FS operations for testing workspace deletion logic
- Mock Status operations for testing status file updates
- Mock Git operations for testing worktree deletion
- Test error scenarios and recovery mechanisms
- Test confirmation system with various inputs
- Test worktree listing functionality

#### Integration Tests
- Test actual worktree deletion with real file system
- Test status file updates with real YAML files
- Test workspace file deletion with real files
- Test cross-platform file handling

#### End-to-End Tests
- Test complete workspace deletion workflow with real workspaces
- Test integration with existing workspace features
- Test confirmation system with real user input

#### Test Cases
- `TestWorkspace_Delete_Success`: Test successful workspace deletion
- `TestWorkspace_Delete_Confirmation`: Test confirmation system
- `TestWorkspace_Delete_Force`: Test force flag behavior
- `TestWorkspace_Delete_WorktreeListing`: Test worktree listing functionality
- `TestWorkspace_Delete_FileCleanup`: Test workspace file deletion
- `TestWorkspace_Delete_StatusCleanup`: Test status file cleanup
- `TestWorkspace_Delete_ErrorHandling`: Test error handling and recovery
- `TestWorkspace_Delete_NotFound`: Test workspace not found handling
- `TestWorkspace_Delete_CrossPlatform`: Test cross-platform compatibility

### Dependencies

#### Direct Dependencies
- **Blocked by**: Features 2, 4, 8, 16, 21 (workspace detection, validation, status management, init command, workspace creation)
- **Dependencies**: FS package, Status package, Git package, Config package, Workspace mode package
- **External**: Git command-line tool (for worktree operations)
- **Adapters**: Use existing FS adapter for file system operations, Status adapter for status management, Git adapter for worktree operations

#### Required Go Modules
- `github.com/spf13/cobra v1.7.0`: Command-line argument parsing
- `go.uber.org/mock v0.5.2`: Mocking framework for testing
- `github.com/stretchr/testify v1.8.4`: Testing utilities and assertions
- `encoding/json`: Go standard library for JSON parsing

### Success Criteria
1. Successfully delete workspace entries from status.yaml with all associated resources
2. Delete all worktrees associated with the workspace from all repositories
3. Delete both main workspace file and all worktree-specific workspace files
4. Remove workspace entry completely from status.yaml while preserving repository entries
5. Provide comprehensive confirmation system with detailed deletion summary
6. Support force flag to skip confirmation prompts
7. Stop and report error if any worktree deletion fails
8. Provide reusable worktree listing functionality for workspace mode
9. Pass all unit, integration, and end-to-end tests
10. Use existing FS, Status, and Git adapters for all external operations
11. Maintain atomic operations where possible for workspace deletion
12. Support cross-platform operation on Windows, macOS, and Linux

### Implementation Plan

#### Phase 1: Status Package Updates (Priority: High)
1. Add RemoveWorkspace method to Status Manager interface
2. Add GetWorkspaceByName method for workspace lookup by name
3. Update concrete implementation using existing status file operations
4. Write unit tests for new status operations

#### Phase 2: Workspace Mode Package Updates (Priority: High)
1. Replace list_worktrees.go with comprehensive worktree listing functionality
2. Add DeleteWorkspace method to Workspace interface
3. Add ListWorkspaceWorktrees method for workspace worktree listing
4. Implement workspace deletion business logic
5. Write unit tests using mocked dependencies

#### Phase 3: CM Package Extension (Priority: High)
1. Add DeleteWorkspace method to CM interface
2. Add ListWorkspaceWorktrees method for workspace worktree listing
3. Implement workspace deletion orchestration
4. Integrate with Status, FS, and Git adapters
5. Write unit tests using mocked dependencies

#### Phase 4: CLI Integration (Priority: Medium)
1. Add workspace delete subcommand to workspace command
2. Integrate with CM package workspace deletion
3. Add command-line argument parsing and validation
4. Add confirmation system with detailed summary
5. Test CLI integration with real workspaces

#### Phase 5: Testing and Validation (Priority: Medium)
1. Create comprehensive unit tests for all components
2. Create integration tests with real file system operations
3. Create end-to-end tests with real workspace deletion

### Future Considerations
- Consider workspace deletion with selective worktree preservation
- Plan for workspace deletion with backup/archive functionality
- Consider workspace deletion with dependency checking (other workspaces using same repositories)
- Plan for workspace deletion with undo/restore functionality
- Consider workspace deletion with batch operations (multiple workspaces)
- Plan for workspace deletion with dry-run mode

## Notes
- This feature builds upon existing workspace creation, worktree management, and status management capabilities
- Focus on reliability and proper error handling for workspace deletion operations
- Ensure error messages are user-friendly and provide clear recovery instructions
- Consider adding debug logging for troubleshooting workspace deletion
- Use build tags to organize tests: `unit` for unit tests, `integration` for real file system tests (adapters only), `e2e` for end-to-end tests
- Mock files should be committed to the repository as `*_gen.go`
- Exit immediately with exit code 1 on critical errors
- Support three output modes: quiet (errors to stderr only), verbose (detailed steps), normal (user interaction only)
- Maintain separation of concerns: FS adapter handles file system operations, Status adapter handles status management, Git adapter handles worktree operations, CM handles business logic
- Ensure workspace operations are atomic where possible
- No partial success scenarios - either workspace deletion succeeds completely or fails with cleanup
- Workspace entries are completely removed from status.yaml
- Worktree deletion covers all worktrees associated with the workspace
- Confirmation system provides detailed summary of what will be deleted
- Force flag skips confirmation prompts for automated operations
