# Code Manager (CM)

A powerful Go CLI tool for managing code development workflows, Git worktrees, and IDE integration. Enables parallel development across different branches and repositories with seamless IDE integration and forge connectivity.

## Overview

`cm` is a command-line interface that simplifies code development workflows for both single repositories and multi-repo workspaces. It automatically detects your project type and provides intelligent worktree creation, management, IDE integration, and forge connectivity for enhanced development productivity.

## Features

### 🔍 Smart Project Detection
- **Single Repository Mode**: Automatically detected when `.git` folder is present
- **Workspace Mode**: Detected when `.code-workspace` files exist (prompts for selection if multiple)

### 🌳 Worktree Management
- Create ephemeral or persistent worktrees for any branch
- Safe creation with collision detection
- Automatic cleanup for ephemeral worktrees
- Support for both single repos and multi-repo workspaces

### 🚀 IDE Integration
- Direct IDE launch with `-i` flag
- Seamless workspace duplication
- Optimized for modern IDE workflows (VSCode, Cursor, etc.)

### 🔗 Forge Integration
- Create worktrees directly from GitHub issues
- Automatic branch name generation from issue titles
- Support for multiple issue reference formats
- Issue information stored in status file for tracking
- Enhanced development workflow with forge connectivity

### 📊 Flexible Output
- Human-readable output for terminal usage
- JSON output for extension integration (`--json` flag)

### 🔄 Remote Branch Management
- Load branches from remote sources
- Support for multiple remote configurations
- Automatic remote management and validation

## Installation

```bash
# Install directly from GitHub
go install github.com/lerenn/code-manager@latest

# Verify installation
cm --help
```

**Prerequisites:**
- Go 1.19 or later
- `$GOPATH/bin` in your `$PATH` (usually already configured)

**For GitHub Integration:**
- `GITHUB_TOKEN` environment variable (optional, for private repositories or rate limit increases)

## Usage

### Basic Commands

```bash
# Create a worktree for a branch
cm create <branch-name>

# Create an ephemeral worktree
cm create <branch-name> -e

# Open worktree in IDE
cm create <branch-name> -i cursor

# List all worktrees
cm list

# List worktrees in JSON format
cm list --json

# Delete a worktree
cm delete <branch-name>

# Load a branch from remote
cm load <remote>:<branch-name>
```

### Project Structure

#### Single Repository Mode
Worktrees are created at:
```
$HOME/.cm/repos/<repo-name>/<branch-name>/
```

#### Workspace Mode
Worktrees are created at:
```
$HOME/.cm/workspaces/<workspace-name>/<branch-name>/<repo-name>/
```

## Command Reference

### `create <branch> [options]`
Creates a new worktree for the specified branch.

**Options:**
- `-i, --ide`: Open the worktree in IDE after creation
- `--from-issue`: Create worktree from a forge issue (GitHub issue URL, issue number, or owner/repo#issue format)
- `--json`: Output creation details in JSON format

**Examples:**
```bash
# Create persistent worktree
cm create feature/new-feature

# Create worktree and open in Cursor IDE
cm create hotfix/bug-fix -i cursor

# Create worktree from GitHub issue (auto-generates branch name)
cm create --from-issue https://github.com/owner/repo/issues/123

# Create worktree from GitHub issue with custom branch name
cm create custom-branch-name --from-issue owner/repo#456

# Create worktree from issue and open in IDE
cm create --from-issue 789 -i cursor
```

### Issue Reference Formats

The `--from-issue` flag supports multiple formats for referencing GitHub issues:

- **GitHub URL**: `https://github.com/owner/repo/issues/123`
- **Owner/Repo format**: `owner/repo#456`
- **Issue number only**: `789` (requires current repository to be GitHub)

**Branch Name Generation:**
When using `--from-issue` without specifying a branch name, CM automatically generates a branch name in the format:
```
<issue-number>-<sanitized-issue-title>
```

The title is sanitized by:
- Converting to lowercase
- Replacing spaces with hyphens
- Removing non-alphanumeric characters (except hyphens)
- Limiting to 80 characters
- Ensuring no consecutive hyphens

### `load [remote-source:]<branch-name>`
Loads a branch from a remote source and creates a worktree.

**Examples:**
```bash
# Load branch from origin
cm load origin:feature-branch

# Load branch from another user's fork
cm load otheruser:feature-branch

# Load branch using default remote (origin)
cm load feature-branch
```

### `list [options]`
Lists all active worktrees for the current project.

**Options:**
- `--json`: Output in JSON format for extension parsing
- `--all`: List worktrees for all projects

**JSON Output Format:**
```json
{
  "worktrees": [
    {
      "repo": "my-project",
      "branch": "feature/new-feature",
      "path": "/home/user/.cm/repos/my-project/feature/new-feature",
      "type": "persistent",
      "workspace": "my-workspace"
    }
  ]
}
```

### `delete <branch> [options]`
Safely removes a worktree and cleans up Git state.

**Options:**
- `--force`: Force deletion without confirmation
- `--json`: Output deletion details in JSON format

## Worktree Types

### Persistent Worktrees
- Survive IDE restarts
- Manual cleanup required
- Ideal for long-term feature development

## Safety Features

- **Collision Detection**: Prevents accidental overwrites of existing worktrees
- **Safe Deletion**: Confirms before removing worktrees
- **Git State Cleanup**: Properly removes worktree references from Git
- **Path Validation**: Ensures valid worktree paths

## Extension Integration

The `--json` flag enables structured output for extension development:

```bash
# Get worktree list in JSON format
cm list --json

# Create worktree with JSON response
cm create feature-branch --json
```

## Configuration

Configuration files are stored in `$HOME/.cm/config/`:

- `settings.json`: Global settings
- `workspaces.json`: Workspace-specific configurations

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

- [ ] Workspace template support
- [ ] Branch naming conventions
- [ ] Integration with Git hooks
- [ ] Advanced filtering options
- [ ] Performance optimizations
- [ ] Plugin system for custom workflows
- [ ] Enhanced forge integrations (GitLab, Bitbucket)
- [ ] Code review workflow integration
- [ ] Automated testing workflow support
- [ ] Multi-language project support