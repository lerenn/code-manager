# Feature 024: Support Devcontainer Worktrees

## Problem

Git worktrees contain a `.git` file that references the original repository's `.git` directory. When opening a worktree in a devcontainer, this reference breaks because the original repo isn't mounted, making the worktree unusable.

## Solution

CM automatically detects devcontainer configurations and creates "detached worktrees" (standalone clones) instead of Git worktrees. These behave like worktrees in cm's tracking system but are fully independent repositories that work perfectly inside containers.

## How It Works

### Automatic Detection

CM automatically detects devcontainer configurations by checking for:
- `.devcontainer/devcontainer.json`
- `.devcontainer.json` (in repository root)

### Detached Worktrees

When a devcontainer is detected, CM creates a detached worktree instead of a regular Git worktree:

- **Regular worktree**: Contains `.git` file pointing to original repo
- **Detached worktree**: Contains `.git` directory (standalone clone)

### Status Tracking

Detached worktrees are tracked in `status.yaml` with a `detached: true` field:

```yaml
repositories:
  github.com/myorg/myrepo:
    path: /home/user/Code/repos/github.com/myorg/myrepo/origin/main
    worktrees:
      origin:feature-branch:
        remote: origin
        branch: feature-branch
      origin:devcontainer-feature:
        remote: origin
        branch: devcontainer-feature
        detached: true  # Only present when true
```

### Deletion Handling

CM handles deletion differently based on worktree type:

- **Regular worktrees**: Uses `git worktree remove` + directory cleanup
- **Detached worktrees**: Uses directory cleanup only (no Git worktree removal)

## Architecture

### PreWorktreeCreationHook

The solution uses a new hook type `PreWorktreeCreationHook` that runs before worktree creation for detection/configuration:

```go
type PreWorktreeCreationHook interface {
    Hook
    OnPreWorktreeCreation(ctx *HookContext) error
}
```

### Hook Execution Flow

1. **PreWorktreeCreationHooks** (detection) - runs first
2. **Worktree creation** (detached or regular based on detection)
3. **WorktreeCheckoutHooks** (setup like git-crypt) - runs after creation

### Priority System

- Devcontainer hook: Priority 10 (high priority, runs first)
- Git-crypt hook: Priority 50 (runs after devcontainer detection)

## Usage

### Automatic Behavior

No user configuration needed! CM automatically:

1. Detects devcontainer configurations
2. Creates detached worktrees for devcontainer repos
3. Creates regular worktrees for non-devcontainer repos
4. Handles deletion appropriately for each type

### Example Workflow

```bash
# Clone a repository with devcontainer config
cm clone https://github.com/myorg/myproject

# Create worktree - automatically becomes detached
cm worktree create feature-branch

# Worktree is now a standalone clone that works in devcontainers
# Open in devcontainer - works perfectly!

# Delete works normally
cm worktree delete feature-branch
```

## Benefits

### For Users

- **Seamless experience**: No configuration needed
- **Devcontainer compatibility**: Worktrees work inside containers
- **Transparent**: Behaves exactly like regular worktrees
- **Backward compatible**: Existing worktrees unaffected

### For Teams

- **Shared devcontainer configs**: Works with team devcontainer setups
- **No devcontainer changes needed**: Works with existing configurations
- **Consistent behavior**: All team members get same experience

## Technical Details

### Disk Usage

Detached worktrees use slightly more disk space than regular worktrees because they're full clones rather than references. This is the trade-off for devcontainer compatibility.

### Performance

- **Creation**: Slightly slower (full clone vs worktree reference)
- **Operations**: Same performance as regular worktrees
- **Deletion**: Same performance (both clean up directories)

### Status File Changes

The `detached: true` field is only written when true (using `omitempty` YAML tag), so existing status files remain unchanged.

## Implementation Files

### Core Implementation

- `pkg/hooks/hooks.go` - PreWorktreeCreationHook interface
- `pkg/hooks/manager.go` - Hook manager for new hook type
- `pkg/hooks/devcontainer/` - Devcontainer detection and hook
- `pkg/status/status.go` - Detached field in WorktreeInfo
- `pkg/git/clone_to_path.go` - Local repository cloning
- `pkg/worktree/` - Detached mode support

### Integration

- `pkg/mode/repository/create_worktree.go` - PreWorktreeCreationHook integration
- `pkg/mode/repository/delete_worktree.go` - Dual-path deletion logic
- `pkg/hooks/default/default.go` - Hook registration

### Testing

- `test/worktree_create_devcontainer_test.go` - E2E creation tests
- `test/worktree_delete_devcontainer_test.go` - E2E deletion tests
- Unit tests for all components

## Future Enhancements

### Potential Improvements

1. **Configurable detection**: Allow users to override detection
2. **Performance optimization**: Cache devcontainer detection results
3. **Advanced devcontainer support**: Handle more devcontainer configurations
4. **Metrics**: Track usage of detached vs regular worktrees

### Extension Points

The hook system allows for additional pre-worktree creation hooks:

- **Environment detection**: Detect other container environments
- **Project-specific rules**: Custom worktree creation logic
- **Integration hooks**: Connect with other development tools

## Troubleshooting

### Common Issues

1. **Worktree not detached**: Check that devcontainer config exists and is valid JSON
2. **Deletion fails**: Ensure worktree is properly tracked in status.yaml
3. **Performance concerns**: Consider if detached worktrees are needed for your use case

### Debug Information

CM logs when devcontainer detection occurs:

```
Devcontainer detected, enabling detached mode for container compatibility
```

Check status.yaml to verify worktrees are marked as detached:

```yaml
detached: true
```

## Conclusion

This feature provides seamless devcontainer support for CM worktrees without requiring any user configuration or devcontainer changes. It automatically detects devcontainer configurations and creates compatible worktrees that work perfectly inside containers.
