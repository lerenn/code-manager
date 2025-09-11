# Feature 017: Reorganize Worktree Structure and Add Clone Command

## Overview

This feature reorganizes the worktree directory structure to be more organized and predictable, and adds a new `clone` command for initializing repositories. The new structure will place all worktrees under `$base_path/<repo_url>/<remote_name>/<branch>` instead of the current flat structure.

## Current State

Currently, worktrees are stored in a flat structure under `$base_path/` with potentially confusing names. The status file structure is also flat, making it difficult to organize repositories by URL.

## Desired State

### New Directory Structure

All worktrees will be organized under `$base_path/<repo_url>/<remote_name>/<branch>`:

```
$base_path/
├── github.com/lerenn/example/
│   ├── origin/
│   │   ├── main/           # Default branch worktree
│   │   ├── feature1/       # Feature branch worktree
│   │   └── bugfix/         # Bugfix branch worktree
│   └── upstream/
│       └── develop/        # Upstream branch worktree
├── github.com/other/repo/
│   └── origin/
│       ├── master/         # Default branch worktree
│       └── develop/        # Development branch worktree
```

### New Status File Structure

The status file will be reorganized to group repositories by URL and separate workspaces, with remotes and default branches separate from worktrees:

```yaml
repositories:
  github.com/lerenn/example:
    path: /path/to/base/github.com/lerenn/example/origin/main
    remotes:
      origin:
        default_branch: main
    worktrees:
      origin:feature1:
        remote: origin
        branch: feature1
        issue:
          number: 123
          title: "Add new feature"
          url: "https://github.com/lerenn/example/issues/123"
workspaces:
  /path/to/workspace.code-workspace:
    worktree: origin:feature1
    repositories:
      github.com/lerenn/example:
      github.com/other/repo:
```

### New Clone Command

Add a new `cm clone` command that:
1. Clones a repository to `$base_path/<repo_url>/<remote_name>/<default_branch>` (recursive by default)
2. Automatically detects the default branch from the remote by querying directly
3. Initializes the repository in CM
4. Creates the initial status entry with remote and default branch in `remotes` section
5. Returns an error if the target path already exists (regardless of protocol differences)
6. Supports non-recursive cloning option with `--shallow` flag
7. Requires full repository URLs (e.g., `https://github.com/lerenn/example.git`)

## Implementation Plan

**Important**: The status file structure changes must be implemented first, before the clone command. This ensures the foundation is in place before adding new functionality.

### Phase 1: Status File Structure Changes (MUST BE FIRST)

1. **Update Status struct**:
   - Change from flat `[]Repository` to nested structure
   - Add new types for the hierarchical organization
   - Update all methods to work with new structure

2. **Clean State Implementation**:
   - Start with new structure from scratch
   - No migration needed - clean slate approach
   - Handle both single repository and workspace modes
   - Treat existing status files as new format (no backward compatibility)

### Phase 2: Worktree Path Changes

1. **Update CM package**:
   - Modify worktree creation to use new path structure
   - Update path generation logic to include remote name
   - Handle path conflicts and validation (return error if exists)
   - Validate repository exists and default branch exists before creating worktrees
   - Validate remote exists in repository's remotes section
   - Automatically detect and add remotes for existing repositories

2. **Update Git package**:
   - Ensure worktree operations work with new paths
   - Update remote management for new structure

### Phase 3: Clone Command Implementation (AFTER STATUS CHANGES)

1. **Add clone command**:
   - Parse repository URL (normalize to github.com/lerenn/example.git format)
   - Detect default branch from remote by querying directly
   - Clone to correct path with remote name
   - Initialize in CM

2. **Default branch detection**:
   - Query remote directly using `git ls-remote --symref origin HEAD`
   - Handle different default branch names (main, master, etc.)
   - Always use remote's default branch name
   - Fail and return error if detection fails

### Phase 4: Workspace Integration

1. **Update workspace handling**:
   - Ensure workspaces work with new structure
   - Update workspace path references to point to new worktree paths
   - Maintain workspace functionality with new schema
   - Validate workspace references only point to worktrees (create them if they don't exist)
   - Workspace repositories reference repository entries
   - Each repository in workspace must have a matching worktree
   - Automatically create worktrees from repository's default branch when missing

## Technical Details

### New Status Types

```go
type Status struct {
    Repositories  map[string]Repository   `yaml:"repositories"`
    Workspaces    map[string]Workspace    `yaml:"workspaces"`
}

type Repository struct {
    Path     string                    `yaml:"path"`
    Remotes  map[string]Remote         `yaml:"remotes"`
    Worktrees map[string]WorktreeInfo  `yaml:"worktrees"`
}

type Remote struct {
    DefaultBranch string `yaml:"default_branch"`
}

type Workspace struct {
    Worktree    string   `yaml:"worktree"`
    Repositories []string `yaml:"repositories"`
}

type WorktreeInfo struct {
    Remote string      `yaml:"remote"`
    Issue  *issue.Info `yaml:"issue,omitempty"`
    Branch string      `yaml:"branch"`
}
```

### Error Types

```go
var (
    ErrRepositoryNotFound = errors.New("repository not found in status")
    ErrRemoteNotFound     = errors.New("remote not found in repository")
    ErrWorktreeNotFound   = errors.New("worktree not found")
    ErrWorktreeExists     = errors.New("worktree already exists")
    ErrInvalidWorkspace   = errors.New("invalid workspace reference")
    ErrRepositoryExists   = errors.New("repository already exists")
)
```

### Path Generation

```go
func generateWorktreePath(repositoriesDir, repoURL, remoteName, branch string) string {
    return filepath.Join(repositoriesDir, repoURL, remoteName, branch)
}
```

### Clone Command Interface

```go
type CloneOpts struct {
    Recursive bool // defaults to true
}

func (c *CM) Clone(repoURL string, opts ...CloneOpts) error
```

## Implementation Strategy

### Clean State Approach

Since we're starting from a clean state, no migration is needed. The new structure will be implemented from scratch:

1. **New status file format**: Implement the new hierarchical structure directly
2. **New directory structure**: All new worktrees will use the new path format
3. **New workspace handling**: Implement the separated workspace structure
4. **No backward compatibility**: Treat existing status files as new format

### Repository Management Rules

1. **Repository path**: Points to default branch worktree path
2. **URL normalization**: Use github.com/lerenn/example.git format regardless of original protocol
3. **Path conflicts**: Return error if repository already exists (even with different protocol)

## Implementation Tasks

**Note**: All tasks must be implemented sequentially. Testing should be done incrementally after each task to ensure stability.

### Task 1: Status File Structure Foundation
- Update Status struct to use new hierarchical organization
- Add new types (Repository, Remote, Workspace, WorktreeInfo)
- Update all status management methods
- Add specific error types for validation
- Update status file loading and saving logic
- **Testing**: Unit tests for new status structure

### Task 2: Path Generation and Validation
- Implement new path generation logic with remote names
- Add path conflict detection and validation
- Update worktree path handling throughout the system
- Implement remote existence validation
- **Testing**: Unit tests for path generation and validation

### Task 3: Repository Management Updates
- Update repository creation and management logic
- Add automatic remote detection for existing repositories
- Update repository validation logic
- **Testing**: Unit tests for repository management

### Task 4: Worktree Creation Updates
- Update worktree creation to use new path structure
- Implement validation for repository and remote existence
- Add worktree conflict detection
- Update worktree listing and management
- **Testing**: Unit tests for worktree operations

### Task 5: Workspace Integration
- Update workspace handling for new structure
- Implement workspace validation logic
- Update workspace path references
- Ensure workspace functionality with new schema
- **Testing**: Unit tests for workspace integration

### Task 6: Clone Command Implementation
- Implement repository URL parsing and normalization (require full URLs)
- Add default branch detection from remote (fail on error)
- Implement recursive cloning logic with shallow option
- Add clone command to CLI interface with `--shallow` flag
- Implement status entry creation for cloned repositories
- **Testing**: Unit and integration tests for clone functionality

### Task 7: Final Testing and Validation
- Comprehensive integration tests for complete workflows
- End-to-end tests for new features
- Performance and stability validation
- **Testing**: Full test suite validation

## Testing Strategy

### Unit Tests

1. **Status structure tests**:
   - Test new hierarchical structure
   - Test workspace separation
   - Test error handling with specific error types

2. **Path generation tests**:
   - Test new path structure with remote names
   - Test conflict resolution (error if exists)
   - Test validation

3. **Clone command tests**:
   - Test default branch detection (success and failure cases)
   - Test cloning process (recursive and shallow)
   - Test error scenarios
   - Test URL validation (require full URLs)

4. **Remote detection tests**:
   - Test automatic remote detection for existing repositories
   - Test remote validation

### Integration Tests

1. **End-to-end structure tests**:
   - Test complete new structure implementation
   - Test with real repositories
   - Test workspace integration

2. **Clone command integration**:
   - Test with real Git repositories
   - Test default branch detection
   - Test CM initialization

### End-to-End Tests

1. **Full workflow tests**:
   - Clone → Create worktree → List → Delete
   - Workspace integration
   - New structure validation

## Breaking Changes

1. **Status file format**: New YAML structure
2. **Worktree paths**: All worktrees will be in new locations with remote names
3. **CLI commands**: New `clone` command added
4. **Workspace structure**: Separated workspace handling
5. **No backward compatibility**: Existing status files treated as new format

## Implementation Notes

- **Clean state**: No migration needed - start fresh with new structure
- **Path conflicts**: Return error if worktree path already exists
- **Default branch**: Always use remote's default branch for clone
- **Validation**: Validate workspace references to existing worktrees and repositories
- **Error handling**: Return error immediately for missing references with specific error messages
- **Repository path**: Points to default branch worktree path
- **Clone behavior**: Recursive by default, with shallow option (`--shallow` flag)
- **Remote detection**: Automatically detect and add remotes for existing repositories
- **URL normalization**: Use consistent format regardless of original protocol
- **Special characters**: Use URLs as-is without special handling
- **Workspace references**: Only reference worktrees, create them if they don't exist
- **Workspace structure**: Each repository must have a matching worktree
- **URL parsing**: Require full repository URLs
- **Remote validation**: Return error if remote exists but has no default branch information
- **Error messages**: Include specific details (repository URL, remote name, etc.)
- **Workspace worktree creation**: Create worktrees from repository's default branch when missing

## Future Considerations

1. **Multiple remotes**: Support for multiple remotes per repository
2. **Branch naming conflicts**: Handle branches with same name from different remotes
3. **Repository aliases**: Support for custom repository names
4. **Bulk operations**: Support for cloning multiple repositories
5. **Special character handling**: Future support for special characters in URLs

## Success Criteria

1. **All existing functionality preserved**: No regression in current features
2. **Improved organization**: Clear, predictable worktree structure with remote names
3. **Clone command works**: Successfully clones and initializes repositories
4. **New structure implemented**: Clean implementation of new status and path structure
5. **All tests pass**: Unit, integration, and end-to-end tests
6. **Documentation updated**: User guides and examples updated
