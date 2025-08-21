# Feature 015: Create Worktrees from GitHub Issues

## Overview
Implement functionality to create Git worktrees based on GitHub issues. This feature will allow users to create worktrees by specifying a GitHub issue URL or issue number, automatically creating a branch with a descriptive name based on the issue title and creating a worktree for development.

## Background
The Code Manager (cm) currently supports creating worktrees from branch names and loading branches from remote sources. However, developers often want to work on specific GitHub issues and need a convenient way to create worktrees directly from issue references. This feature will streamline the workflow by automatically creating descriptive branch names and worktrees based on GitHub issue information.

## Command Syntax

### Create from Issue Command
```bash
cm create [branch-name] --from-issue <issue-reference> [options]
```

### Issue Reference Formats
```bash
# GitHub issue URL
cm create --from-issue https://github.com/owner/repo/issues/123

# Issue number (requires remote origin to be GitHub)
cm create --from-issue 123

# Owner/repo#issue format
cm create --from-issue owner/repo#123

# With custom branch name
cm create custom-branch-name --from-issue 456
```

### Examples

```bash
# Create worktree from GitHub issue URL (infer branch name)
cm create --from-issue https://github.com/lerenn/code-manager/issues/42

# Create worktree from issue number (current repo, infer branch name)
cm create --from-issue 42

# Create worktree with custom branch name
cm create feature-login-fix --from-issue 123

# Create worktree and open in IDE
cm create --from-issue 123 -i cursor

# Create worktree with verbose output
cm create --from-issue https://github.com/owner/repo/issues/789 -v
```

## Requirements

### Functional Requirements
1. **Forge Issue Parsing**: Parse forge issue URLs, issue numbers, and owner/repo#issue formats
2. **Issue Information Retrieval**: Fetch issue title, description, and metadata from forge API
3. **Branch Name Generation**: Create descriptive branch names based on issue title and number
4. **Worktree Creation**: Create worktree using the generated branch name
5. **Repository Detection**: Work in both single repository and workspace modes
6. **Remote Origin Validation**: Ensure remote origin is a supported forge repository when using issue numbers
7. **Error Handling**: Handle various error conditions gracefully (invalid URLs, API errors, etc.)
8. **User Feedback**: Provide clear feedback about the issue parsing and worktree creation process
9. **Configuration Integration**: Use existing configuration for worktree management
10. **IDE Integration**: Support opening created worktree in IDE with `-i` flag
11. **Branch Name Override**: Allow custom branch name via positional argument or infer from issue title
12. **Issue Description Integration**: Create initial commit with issue description (but do not push)
13. **Forge Agnostic Design**: Support multiple forges through a common interface

### Non-Functional Requirements
1. **Performance**: Issue fetching and worktree creation should complete within reasonable time (< 15 seconds)
2. **Reliability**: Handle forge API failures, network errors, and concurrent access
3. **Cross-Platform**: Work on Windows, macOS, and Linux
4. **Testability**: Use Uber gomock for mocking forge API and file system operations
5. **Minimal Dependencies**: Use only existing adapters and Go standard library
6. **Atomic Operations**: Ensure status file updates are atomic
7. **Safe Cleanup**: Provide proper cleanup on failure with rollback
8. **Rate Limiting**: Respect forge API rate limits
9. **Authentication**: Support forge personal access tokens via environment variables
10. **Forge Extensibility**: Design should allow easy addition of new forge implementations

## Technical Specification

### Interface Design

#### Forge Package (New)
**New Interface Methods**:
- `GetIssueInfo(issueRef string) (*IssueInfo, error)`: Fetch issue information from forge
- `ValidateForgeRepository(repoPath string) error`: Validate that repository has supported forge remote origin
- `ParseIssueReference(issueRef string) (*IssueReference, error)`: Parse various issue reference formats
- `GenerateBranchName(issueInfo *IssueInfo) string`: Generate branch name from issue information

**Data Structures**:
```go
type IssueInfo struct {
    Number      int
    Title       string
    Description string
    State       string
    URL         string
    Repository  string
    Owner       string
}

type IssueReference struct {
    Owner      string
    Repository string
    IssueNumber int
    URL        string
}

type Forge interface {
    GetIssueInfo(issueRef string) (*IssueInfo, error)
    ValidateForgeRepository(repoPath string) error
    ParseIssueReference(issueRef string) (*IssueReference, error)
    GenerateBranchName(issueInfo *IssueInfo) string
    Name() string
}
```

**Key Characteristics**:
- New package for forge API interactions (forge-agnostic)
- Pure function-based operations
- No state management required
- Cross-platform compatibility
- Error handling with wrapped errors
- Support for forge API authentication via environment variables
- Extensible design for multiple forge implementations

#### CM Package Extension
**New Interface Methods**:
- `CreateWorkTreeFromIssue(branchName *string, issueRef string, ideName *string) error`: Main entry point for creating worktrees from issues
- `createWorkTreeFromIssueForSingleRepo(branchName *string, issueRef string, ideName *string) error`: Create worktree from issue for single repository
- `createWorkTreeFromIssueForWorkspace(branchName *string, issueRef string, ideName *string) error`: Create worktree from issue for workspace

**Implementation Structure**:
- Extends existing CM package with GitHub issue functionality
- Mode detection in main `CreateWorkTreeFromIssue()` method
- Private helper methods for issue parsing and branch name generation
- Integration with existing worktree creation
- Error handling with wrapped errors
- User feedback through logger

**Key Characteristics**:
- Builds upon existing detection and validation logic
- Integrates with worktree creation system
- Provides comprehensive user feedback
- Handles both single repository and workspace modes
- Uses existing adapters (FS, Git, Status) for all operations

### Implementation Details

#### 1. Forge Package Implementation
The Forge package will handle all forge API interactions with a forge-agnostic design:

**Key Components**:
- `GetIssueInfo()`: Fetches issue information using forge REST API
- `ValidateForgeRepository()`: Validates that repository remote origin is supported forge
- `ParseIssueReference()`: Parses various issue reference formats
- `GenerateBranchName()`: Generates descriptive branch names from issue title

**Implementation Notes**:
- Use forge REST API for issue fetching
- Support forge personal access token authentication via environment variables
- Handle rate limiting with exponential backoff
- Parse issue references: URLs, numbers, and owner/repo#issue format
- Generate branch names: `<issue-nb>-<sanitized-issue-title>` format
- Handle API errors with proper error wrapping
- Support both public and private repositories
- Forge-agnostic design allowing multiple forge implementations

#### 2. CM Package Implementation
The CM package will implement the issue-based worktree creation logic:

**Key Components**:
- `CreateWorkTreeFromIssue()`: Main entry point with mode detection
- `createWorkTreeFromIssueForSingleRepo()`: Single repository implementation
- `createWorkTreeFromIssueForWorkspace()`: Workspace implementation
- `generateBranchNameFromIssue()`: Branch name generation logic

**Implementation Flow**:
1. Parse issue reference to extract repository and issue number
2. Validate that current repository has supported forge remote origin (for issue numbers)
3. Fetch issue information from forge API
4. Validate that issue is open/active
5. Generate descriptive branch name from issue title and number (`<issue-nb>-<sanitized-issue-title>`)
6. Use provided branch name or generate from issue title
7. Create worktree using existing worktree creation logic
8. Create initial commit with issue description (but do not push)
9. Update status file and provide user feedback

**Implementation Notes**:
- Use existing worktree creation methods after branch name generation
- Handle both single repository and workspace modes
- Provide comprehensive error handling and cleanup with rollback
- Support custom branch name via positional argument
- Integrate with existing IDE opening functionality
- Use existing adapters for all file system and Git operations
- Only allow worktrees for open issues

#### 3. CLI Implementation
The CLI will extend the existing create command with a new flag:

**New Flag**:
- `--from-issue`: Create worktree from forge issue (optional)
- When provided, the positional branch name argument becomes optional

**Implementation Notes**:
- Extend existing `cm create` command with `--from-issue` flag
- Follow existing CLI patterns and architecture
- Use existing flag parsing and validation
- Integrate with existing configuration loading
- Provide clear help text and examples
- Handle authentication token from environment variables (e.g., `GITHUB_TOKEN`)
- Always create initial commit with issue description (but do not push)
- Branch name is optional when `--from-issue` is used (inferred from issue title)

## Implementation Notes

### Forge Package Structure
Following the IDE package pattern, the forge package will have the following structure:
- `pkg/forge/forge.go`: Main interface and manager (similar to `pkg/ide/ide.go`)
- `pkg/forge/github.go`: GitHub implementation (similar to `pkg/ide/cursor.go`)
- `pkg/forge/errors.go`: Forge-specific errors
- `pkg/forge/mockforge.gen.go`: Generated mocks for testing

### Branch Name Generation
The branch name will follow the format: `<issue-nb>-<sanitized-issue-title>`
- Issue number is preserved as-is
- Issue title is sanitized (lowercase, replace spaces with hyphens, remove all non-alphanumeric characters except hyphens)
- Limit the length of the sanitized title to 80 characters
- Ensure there are no two or more consecutive hyphens
- Example: Issue #123 "Fix Login Bug" → `123-fix-login-bug`

### Authentication
- Use environment variables for authentication tokens (e.g., `GITHUB_TOKEN`)
- No command-line flags for tokens to avoid security issues
- Support for private repositories when token is provided
- Return error on rate limiting (no exponential backoff)

### Error Handling
- Rollback all changes on any failure
- Return descriptive error messages
- Return error on rate limiting (no exponential backoff)
- Validate issue state (only open issues allowed)

### Initial Commit
- Always create an initial commit with the issue description
- Do not push the commit automatically
- Use issue title as git commit title
- Use issue description as git commit description

### Forge Extensibility
- Design the forge interface to be platform-agnostic
- Allow easy addition of new forge implementations (GitLab, Bitbucket, etc.)
- Use dependency injection pattern similar to IDE manager

### Workspace Mode Behavior
- Use the same issue reference for all repositories in the workspace
- Create worktrees for all repositories with the same branch name (from the issue)
- For issue numbers: Test all repositories in order until finding the one where the issue exists
- For owner/repo#issue format or issue URLs: Use the specified repository directly
- Do not require the issue to exist in each repository's forge

### Error Handling and Validation
- Provide specific error messages for common scenarios:
  - "Issue #123 not found in repository"
  - "Issue #123 is closed, only open issues are supported"
  - "Repository does not have a supported forge as remote origin"
- Validate issue exists and is open before checking if branch name already exists
