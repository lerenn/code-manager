# Feature 013: Implement Load PR Branches

## Overview
Implement functionality to load branches from remote sources using the `wtm load` command. This feature will allow users to load branches from other users/organizations by specifying a remote source and branch name, automatically adding the remote if it doesn't exist.

## Background
The Git WorkTree Manager (wtm) needs to provide functionality to load branches from external sources (like pull requests from other users). This feature will enable developers to easily work on branches from other contributors by automatically managing remote sources and creating worktrees for the loaded branches.

## Command Syntax

### Load Command
```bash
wtm load [remote-source:]<branch-name>
```

### Examples

```bash
# Load branch from origin (default remote)
wtm load feature/new-feature

# Load branch from specific user/organization
wtm load user:branch-name

# Load branch and open in IDE
wtm load user:fix-bug-123 -i cursor

# Load with verbose output
wtm load feature/improvement -v
```

## Requirements

### Functional Requirements
1. **Remote Source Loading**: Load branches from specified remote sources
2. **Origin Fallback**: Use origin remote when no remote source is specified
3. **Remote Management**: Automatically add remote sources when they don't exist
4. **Branch Validation**: Ensure the branch exists on the remote before loading
5. **Worktree Creation**: Create worktree for the loaded branch
6. **Repository Detection**: Work in single repository mode (current behavior)
7. **Error Handling**: Handle various error conditions gracefully
8. **User Feedback**: Provide clear feedback about the loading operation
9. **Configuration Integration**: Use existing configuration for worktree management
10. **IDE Integration**: Support opening loaded worktree in IDE with `-i` flag

### Non-Functional Requirements
1. **Performance**: Loading should complete within reasonable time (< 10 seconds)
2. **Reliability**: Handle Git command failures, network errors, and concurrent access
3. **Cross-Platform**: Work on Windows, macOS, and Linux
4. **Testability**: Use Uber gomock for mocking Git and file system operations
5. **Minimal Dependencies**: Use only existing adapters and Go standard library
6. **Atomic Operations**: Ensure status file updates are atomic
7. **Safe Cleanup**: Provide proper cleanup on failure

## Technical Specification

### Interface Design

#### Git Package Extension
**New Interface Methods**:
- `AddRemote(repoPath, remoteName, remoteURL string) error`: Add a new remote to the repository
- `FetchRemote(repoPath, remoteName string) error`: Fetch from a specific remote
- `BranchExistsOnRemote(repoPath, remoteName, branch string) (bool, error)`: Check if branch exists on remote
- `GetRemoteURL(repoPath, remoteName string) (string, error)`: Get the URL of a remote
- `RemoteExists(repoPath, remoteName string) (bool, error)`: Check if a remote exists

**Key Characteristics**:
- Extends existing Git package with remote management operations
- Pure function-based operations
- No state management required
- Cross-platform compatibility
- Error handling with wrapped errors

#### WTM Package Extension
**New Interface Methods**:
- `LoadBranch(remoteSource, branchName string, ideName *string) error`: Main entry point for loading branches
- `loadBranchForSingleRepo(remoteSource, branchName string, ideName *string) error`: Load branch for single repository

**Implementation Structure**:
- Extends existing WTM package with branch loading functionality
- Mode detection in main `LoadBranch()` method
- Private helper methods for remote management
- Integration with existing worktree creation
- Error handling with wrapped errors
- User feedback through logger

**Key Characteristics**:
- Builds upon existing detection and validation logic
- Integrates with worktree creation system
- Provides comprehensive user feedback
- Handles remote management automatically
- Uses existing adapters (FS, Git, Status) for all operations

### Implementation Details

#### 1. Git Package Implementation
The Git package will be extended with remote management operations:

**Key Components**:
- `AddRemote()`: Executes `git remote add` command
- `FetchRemote()`: Executes `git fetch` for specific remote
- `BranchExistsOnRemote()`: Checks if branch exists on remote using `git ls-remote`
- `GetRemoteURL()`: Gets remote URL using `git remote get-url`
- `RemoteExists()`: Checks if remote exists using `git remote`

**Implementation Notes**:
- Use `exec.Command()` for Git operations
- Handle Git command failures with proper error wrapping
- Support both HTTPS and SSH URL formats
- Provide detailed error messages for debugging
- Validate remote URLs before adding them

#### 2. WTM Package Implementation
The WTM package will implement the branch loading logic:

**Key Components**:
- Add `LoadBranch()` method as the main entry point with mode detection
- Add `loadBranchForSingleRepo()` helper method for single repository mode
- Remote source parsing and validation
- Remote management (add if doesn't exist)
- Branch validation and loading
- Integration with existing worktree creation
- User feedback and progress reporting

**Implementation Flow**:
1. **Parse remote source and branch name** from input
2. **Detect current mode** (single repository vs workspace)
3. **For single repository mode**:
   a. Validate current directory is a Git repository
   b. Validate origin remote exists and is a valid GitHub URL
   c. Parse remote source (default to "origin" if not specified)
   d. Check if remote exists, add if it doesn't (using existing repository name logic)
   e. Fetch from the remote
   f. Validate branch exists on remote
   g. Create worktree for the branch (using existing worktree creation logic directly)
   h. Open in IDE if specified
4. **For workspace mode** (placeholder):
   a. Return error with placeholder message
   b. Future implementation will handle workspace loading

**Remote Management Logic**:
- If remote source is "origin" or not specified: use existing origin remote
- If remote source is specified but doesn't exist:
  - Use existing repository name extraction logic from origin remote
  - Determine protocol (HTTPS/SSH) from origin remote, default to HTTPS
  - Support both `git@github.com:` and `ssh://git@github.com/` SSH formats
  - Construct remote URL: `{protocol}://github.com/{remote-source}/{repo-name}.git`
  - Add remote with name matching the remote source
  - Fetch from the new remote
- If remote source exists: validate URL matches expected format, fail if different

**URL Construction**:
- For GitHub: `{protocol}://github.com/{remote-source}/{repo-name}.git`
- Protocol matching: Use the same protocol (HTTPS/SSH) as the origin remote, default to HTTPS
- Support both `git@github.com:` and `ssh://git@github.com/` SSH formats
- Repository name: Use existing repository name extraction logic from origin remote
- Future support: Other platforms (GitLab, Bitbucket) will be added in future versions

#### 3. Command Line Integration
The main command will be added to the CLI:

**Command Structure**:
```go
func createLoadCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "load [remote-source:branch-name]",
        Short: "Load branch from remote source",
        Long:  `Load a branch from a remote source and create a worktree. Supports loading from origin or other users/organizations.`,
        Args:  cobra.ExactArgs(1),
        RunE: func(_ *cobra.Command, args []string) error {
            // Parse remote:branch format
            // Call WTM.LoadBranch()
        },
    }
}
```

**Argument Parsing**:
- Parse `remote-source:branch-name` format by splitting on first `:` character
- Support `branch-name` format (defaults to origin)
- Branch names cannot contain `:` characters
- Handle edge cases: `:branch-name` and `remote:` return errors
- Provide helpful error messages with usage examples

### Error Handling

#### Error Types
1. **InvalidArgumentError**: When remote:branch format is invalid (empty remote, empty branch, or branch contains ':')
2. **RemoteNotFoundError**: When remote source doesn't exist and can't be added
3. **BranchNotFoundOnRemoteError**: When branch doesn't exist on remote (includes remote name and branch name)
4. **RemoteAddError**: When adding remote fails
5. **FetchError**: When fetching from remote fails
6. **OriginRemoteError**: When origin remote doesn't exist or is invalid
7. **WorktreeCreationError**: When worktree creation fails (reuses existing errors)

#### Error Recovery
- Clean up any partially added remotes on failure
- Provide clear error messages with suggested solutions
- Support verbose mode for detailed error information

### User Experience

#### Success Scenarios
```
$ wtm load user:branch-name
✓ Added remote 'user'
✓ Fetched from remote 'user'
✓ Branch 'branch-name' found on remote
✓ Created worktree for 'branch-name'
✓ Worktree ready at: /Users/user/.wtm/github.com/lerenn/wtm/branch-name
```

#### Error Scenarios
```
$ wtm load invalid-user:branch
✗ Failed to add remote 'invalid-user': repository not found
  Suggestion: Check the username and repository name

$ wtm load user:non-existent-branch
✗ Branch 'non-existent-branch' not found on remote 'user'
  Suggestion: Check the branch name or contact the repository owner

$ wtm load invalid:format
✗ Invalid argument format. Expected: [remote-source:]<branch-name> or <branch-name>
  Usage: wtm load [remote-source:]<branch-name>
  Examples:
    wtm load feature/new-feature          # Load from origin
    wtm load user:branch-name             # Load from specific remote

$ wtm load :branch-name
✗ Invalid argument format: empty remote source

$ wtm load user:
✗ Invalid argument format: empty branch name
```

### Testing Strategy

#### Unit Tests
- Test remote source parsing logic
- Test URL construction for different platforms
- Test error handling scenarios
- Mock Git operations for isolated testing

#### Integration Tests
- Test actual Git operations with temporary repositories
- Test remote addition and fetching
- Test worktree creation integration

#### End-to-End Tests
- Test complete workflow with real Git repositories
- Test error scenarios with invalid remotes/branches
- Test IDE integration

### Future Enhancements

#### Platform Support
- Support for GitLab (`gitlab.com/{user}/{repo}`)
- Support for Bitbucket (`bitbucket.org/{user}/{repo}`)
- Support for custom Git hosting platforms

#### Advanced Features
- Support for loading specific commits
- Support for loading pull request branches by PR number
- Support for workspace mode
- Support for SSH URL detection and usage

#### Configuration
- Configurable remote URL templates
- Default remote source preferences
- SSH vs HTTPS URL preferences

## Design Decisions

### Remote Source Format
- **Username only**: Remote sources will be just usernames (e.g., `justenstall`)
- **GitHub only**: Initial implementation will support GitHub only
- **Future expansion**: Support for other platforms (GitLab, Bitbucket) will be added in future versions

### Branch Existence
- **Fail if not exists**: Command will fail with an error if the branch doesn't exist on the remote
- **No local creation**: Will not create branches locally if they don't exist remotely
- **Clear error messages**: Provide specific error messages when branches don't exist

### Remote Management
- **Username as remote name**: Use the remote source name as the remote name (e.g., `git remote add justenstall ...`)
- **Protocol matching**: Support the same protocol (HTTPS/SSH) as the origin remote
- **URL validation**: Validate remote URLs before adding them

### Repository Detection
- **Single repository mode only**: Command will only work in single repository mode initially
- **Future workspace support**: Workspace mode support will be added in future versions
- **Origin requirement**: Repository must have an origin remote configured

### Worktree Creation
- **Automatic creation**: Automatically create a worktree for the loaded branch
- **Integration**: Use existing worktree creation logic from the `create` command
- **No user choice**: No option to just checkout without creating worktree

### Error Handling
- **Fail on invalid remote**: Return error if remote source doesn't exist as a user/organization
- **Fail on URL mismatch**: Return error if remote already exists but points to a different URL
- **Helpful suggestions**: Provide suggestions for common error scenarios

### IDE Integration
- **Support -i flag**: Load command will support the `-i` flag like other commands
- **Automatic opening**: Open the worktree in IDE after successful loading when flag is used

### Status Tracking
- **Track loaded worktrees**: Loaded branches will be tracked in the status.yaml file like other worktrees
- **No distinction**: No special distinction between locally created and loaded worktrees in the status

## Implementation Summary

### Key Features
- **Dual format support**: `wtm load branch-name` (origin) and `wtm load user:branch-name` (specific remote)
- **Automatic remote management**: Add remotes automatically when they don't exist
- **Protocol matching**: Use same protocol (HTTPS/SSH) as origin remote, default to HTTPS
- **GitHub-only**: Initial implementation supports GitHub only
- **Single repository mode**: Only works in single repository mode initially
- **Direct integration**: Uses existing worktree creation logic directly
- **IDE support**: Supports `-i` flag for opening worktree in IDE
- **Comprehensive error handling**: Specific error types with helpful messages

### Technical Approach
- Extends existing Git package with remote management methods
- Extends WTM package with `LoadBranch()` method
- Reuses existing repository name extraction and worktree creation logic
- Follows existing architectural patterns and testing conventions
- Integrates seamlessly with current command structure
