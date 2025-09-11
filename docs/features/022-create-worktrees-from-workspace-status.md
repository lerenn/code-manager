# Feature 022: Create Worktrees from Workspace Status

## Overview
Implement functionality to create worktrees from workspace definitions stored in the status.yaml file. When using `cm wt create --workspace <workspace> <worktree_name>`, the system should create a worktree in each repository listed in the specified workspace from the status.yaml file.

## Background
The Code Manager (cm) currently supports creating worktrees from .code-workspace files by detecting them in the current directory. However, this approach has limitations:
1. It requires the .code-workspace file to be present in the current directory
2. It doesn't leverage the centralized workspace management through status.yaml
3. It creates a tight coupling between file system structure and workspace operations

This feature will enable users to create worktrees from workspace definitions stored in the status.yaml file, providing a more flexible and centralized approach to workspace management.

## Requirements

### Functional Requirements
1. **Workspace Flag Support**: Add `--workspace <workspace_name>` flag to the `cm wt create` command
2. **Status-Based Workspace Resolution**: Resolve workspace repositories from status.yaml instead of .code-workspace files
3. **Multi-Repository Worktree Creation**: Create worktrees in all repositories listed in the workspace
4. **Workspace Validation**: Validate that the specified workspace exists in status.yaml
5. **Repository Validation**: Validate that all repositories in the workspace exist and are accessible
6. **Error Handling**: Provide clear error messages for missing workspaces or invalid repositories
7. **Backward Compatibility**: Maintain existing functionality for single repository mode
8. **Consistent Worktree Naming**: Use consistent worktree naming across all repositories in the workspace

### Non-Functional Requirements
1. **Performance**: Workspace-based worktree creation should complete within 5 seconds for typical workspaces
2. **Reliability**: Handle file system errors and permission issues gracefully
3. **Cross-Platform**: Work on Windows, macOS, and Linux
4. **Testability**: Use existing mocking infrastructure for testing

## Technical Specification

### CLI Changes

#### Worktree Create Command
**New Flag:**
- `--workspace <workspace_name>`: Specify workspace name from status.yaml

**Updated Usage:**
```bash
cm wt create <worktree_name> --workspace <workspace_name> [--ide <ide-name>] [--force]
```

**Argument Validation:**
- Branch name is mandatory (same as current behavior)
- Exception: Branch name is optional when using `--from-issue` flag
- Workspace name is mandatory when `--workspace` flag is provided

**Examples:**
```bash
cm wt create feature-branch --workspace my-workspace
cm wt create feature-branch --workspace my-workspace --ide cursor
cm wt create feature-branch --workspace my-workspace --force
cm wt create --from-issue 123 --workspace my-workspace
```

### CM Package Changes

#### New Method: CreateWorkTreeFromWorkspace
```go
func (c *realCM) CreateWorkTreeFromWorkspace(workspaceName, branch string, opts ...CreateWorkTreeOpts) error
```

**Parameters:**
- `workspaceName`: Name of the workspace from status.yaml
- `branch`: Branch name for worktree creation
- `opts`: Optional parameters (IDE, force, issue reference)

**Behavior:**
1. Validate workspace exists in status.yaml
2. Get repository list from workspace definition
3. Create worktrees in all repositories with the specified branch using existing repository logic
4. Create `.code-workspace` file in `workspaces_dir` with all repositories from workspace definition
5. Update status.yaml workspace section with worktree name only after ALL worktrees are successfully created
6. Handle IDE opening with the newly created `.code-workspace` file if specified
7. Rollback all changes silently (delete worktrees, remove status entries, clean up workspace file) if any repository fails

#### Updated CreateWorkTree Method
The existing `CreateWorkTree` method will be updated to:
1. Check for `--workspace` flag presence
2. Pass workspace name to `detectProjectMode` method
3. Route to workspace-based creation if workspace flag is provided
4. Fall back to existing single repository mode detection logic if no workspace flag

### Status Package Integration

#### Workspace Retrieval
Use existing `GetWorkspace` method to retrieve workspace definition:
```go
workspace, err := c.statusManager.GetWorkspace(workspaceName)
```

#### Repository Resolution
Extract repository list from workspace definition:
```go
repositories := workspace.Repositories
```

#### Status.yaml Workspace Update
Add worktree name to existing worktree array in workspace definition:
```go
// Add to existing worktree array
workspace.Worktree = append(workspace.Worktree, worktreeName)
```

### Mode Detection Changes

#### Update Project Mode Detection Logic
The following changes will be made to update mode detection:

1. **Update `detectProjectMode` method** in `pkg/cm/cm.go`:
   - Add `workspaceName` parameter to the method signature
   - If `workspaceName` is provided: return `ModeWorkspace`
   - If `workspaceName` is empty: check for Git repository and return `ModeSingleRepo` or `ModeNone`
   - Remove .code-workspace file detection logic (lines 376-383)

2. **Update `IsWorkspaceFile` method** in `pkg/mode/repository/repository.go`:
   - Remove this method as it's no longer needed

3. **Update `DetectWorkspaceFiles` method** in `pkg/mode/workspace/detect_workspace_files.go`:
   - Remove this method as it's no longer needed

4. **Update workspace mode handling**:
   - Mode is determined by presence of `--workspace` flag, not file system detection
   - Require explicit `--workspace` flag for workspace operations

### Workspace Package Changes

#### Add Repository Provider to Workspace Struct
```go
type realWorkspace struct {
    // ... existing fields ...
    repositoryProvider RepositoryProvider
}
```

**RepositoryProvider Type:**
```go
type RepositoryProvider func(params repository.NewRepositoryParams) repository.Repository
```

**Usage:**
- Use repository provider to create repository instances for worktree operations
- Leverage existing `repoInstance.CreateWorktree(branch)` logic
- Status.yaml updates are handled automatically by the repository worktree creation

### Config Package Changes

#### Rename BasePath to RepositoriesDir and Add WorkspacesDir Field
Update config structure:
```go
type Config struct {
    RepositoriesDir string `yaml:"repositories_dir"`  // Renamed from BasePath
    WorkspacesDir   string `yaml:"workspaces_dir"`    // New field
    // ... existing fields ...
}
```

**Changes:**
- Rename `BasePath` field to `RepositoriesDir` for clarity
- Add new `WorkspacesDir` field
- Update YAML tag from `base_path` to `repositories_dir`
- No migration needed - this is a breaking change that users will need to update their config

**Default Values:**
- `RepositoriesDir`: `~/Code/repos` (existing default)
- `WorkspacesDir`: `~/Code/workspaces`

**Environment Variables:**
- `CM_REPOSITORIES_DIR` (renamed from existing)
- `CM_WORKSPACES_DIR` (new)

### Error Handling

#### New Error Types
```go
var (
    ErrWorkspaceNotFound = errors.New("workspace not found in status.yaml")
    ErrWorkspaceRepositoryNotFound = errors.New("repository in workspace not found")
    ErrWorkspaceRepositoryInvalid = errors.New("repository in workspace is invalid")
)
```

#### Error Messages
- "Workspace 'workspace-name' not found in status.yaml"
- "Failed to create worktree in repository 'github.com/user/repo': branch 'feature-branch' already exists"
- "Repository 'repo-name' in workspace 'workspace-name' is not a valid Git repository"

### Implementation Details

#### Workflow
1. **CLI Validation**: Validate that `--workspace` flag is provided
2. **Mode Detection**: Pass workspace name to `detectProjectMode` to determine workspace mode
3. **Workspace Resolution**: Get workspace definition from status.yaml
4. **Worktree Creation**: Create worktrees in all repositories using existing repository logic
5. **Workspace File Creation**: Create `.code-workspace` file in `workspaces_dir` with all repositories from workspace definition
6. **Status Update**: Update status.yaml workspace section with worktree name only after ALL worktrees are successfully created
7. **IDE Opening**: Open IDE with the newly created `.code-workspace` file if specified
8. **Rollback**: If any repository fails, silently rollback all changes (delete worktrees, remove status entries, clean up workspace file)

#### Repository Processing
For each repository in the workspace:
1. Create repository instance using repository provider
2. Call `repoInstance.CreateWorktree(branch)` (handles validation, creation, hooks, and status.yaml updates automatically)
3. Track created worktrees for potential rollback

#### Workspace File Creation
Create `.code-workspace` file in `workspaces_dir`:
- Path: `~/Code/workspaces/{workspace-name}-{branch-name}.code-workspace`
- Include all repositories from workspace definition in status.yaml (regardless of worktree creation success)
- Use existing workspace file creation logic

#### Worktree Naming
Use consistent naming pattern across all repositories:
- Pattern: `{repositories-dir}/{repo-name}/worktrees/{remote}/{branch}`
- Same branch name used across all repositories in workspace

### Testing Strategy

#### Unit Tests
- Test workspace resolution from status.yaml
- Test repository validation for workspace repositories
- Test worktree creation for multiple repositories
- Test error handling for missing workspaces and invalid repositories

#### Integration Tests
- Test end-to-end workflow with real file system
- Test status.yaml updates
- Test IDE opening functionality

#### E2E Tests
- Test complete workflow with real CM struct and Git operations
- Test with multiple repositories in workspace
- Test error scenarios

### Migration Strategy

#### Backward Compatibility
- Existing single repository mode remains unchanged
- Existing workspace mode detection is removed
- Users must explicitly specify `--workspace` flag for workspace operations

#### Documentation Updates
- Update CLI documentation to reflect new `--workspace` flag requirement
- Update examples to show workspace-based usage
- Document migration path from .code-workspace to status.yaml based workspaces

### Success Criteria

1. **Functional**: Users can create worktrees from workspace definitions in status.yaml
2. **Performance**: Workspace-based worktree creation completes within 5 seconds
3. **Reliability**: All error scenarios are handled gracefully
4. **Usability**: Clear error messages guide users to correct usage
5. **Compatibility**: Existing single repository functionality remains unchanged

### Future Enhancements

1. **Workspace Templates**: Support for workspace templates
2. **Workspace Cloning**: Clone workspaces with all repositories
3. **Workspace Synchronization**: Sync workspace definitions across team members
4. **Workspace Validation**: Validate workspace consistency across repositories

## Implementation Plan

### Phase 1: Core Implementation
1. Add `--workspace`/`-w` flag to CLI
2. Implement `CreateWorkTreeFromWorkspace` method
3. Update `CreateWorkTree` to handle workspace flag
4. Add workspace validation logic
5. Rename `BasePath` to `RepositoriesDir` in config structure
6. Add `WorkspacesDir` field to config structure

### Phase 2: Mode Detection Cleanup
1. Remove .code-workspace detection from `detectProjectMode`
2. Update workspace mode handling
3. Remove deprecated methods
4. Update error handling

### Phase 3: Testing and Documentation
1. Add comprehensive unit tests
2. Add integration tests
3. Add E2E tests
4. Update documentation

