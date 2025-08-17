# Git WorkTree Manager (wtm)

A powerful Go CLI tool for managing Git worktrees specifically designed for IDE. Enables parallel development across different branches and repositories with seamless IDE integration.

## Overview

`wtm` is a command-line interface that simplifies Git worktree management for both single repositories and multi-repo workspaces. It automatically detects your project type and provides intelligent worktree creation, management, and IDE integration.

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
- Optimized for VSCode Based IDE's workflow

### 🔗 Forge Integration
- Create worktrees directly from GitHub issues
- Automatic branch name generation from issue titles
- Support for multiple issue reference formats
- Issue information stored in status file for tracking

### 📊 Flexible Output
- Human-readable output for terminal usage
- JSON output for extension integration (`--json` flag)

## Installation

```bash
# Install directly from GitHub
go install github.com/lerenn/wtm@latest

# Verify installation
wtm --help
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
wtm create <branch-name>

# Create an ephemeral worktree
wtm create <branch-name> -e

# Open worktree in IDE
wtm create <branch-name> -i cursor

# List all worktrees
wtm list

# List worktrees in JSON format
wtm list --json

# Delete a worktree
wtm delete <branch-name>
```

### Project Structure

#### Single Repository Mode
Worktrees are created at:
```
$HOME/.wtm/repos/<repo-name>/<branch-name>/
```

#### Workspace Mode
Worktrees are created at:
```
$HOME/.wtm/workspaces/<workspace-name>/<branch-name>/<repo-name>/
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
wtm create feature/new-feature

# Create worktree and open in Cursor IDE
wtm create hotfix/bug-fix -i cursor

# Create worktree from GitHub issue (auto-generates branch name)
wtm create --from-issue https://github.com/owner/repo/issues/123

# Create worktree from GitHub issue with custom branch name
wtm create custom-branch-name --from-issue owner/repo#456

# Create worktree from issue and open in IDE
wtm create --from-issue 789 -i cursor
```

### Issue Reference Formats

The `--from-issue` flag supports multiple formats for referencing GitHub issues:

- **GitHub URL**: `https://github.com/owner/repo/issues/123`
- **Owner/Repo format**: `owner/repo#456`
- **Issue number only**: `789` (requires current repository to be GitHub)

**Branch Name Generation:**
When using `--from-issue` without specifying a branch name, WTM automatically generates a branch name in the format:
```
<issue-number>-<sanitized-issue-title>
```

The title is sanitized by:
- Converting to lowercase
- Replacing spaces with hyphens
- Removing non-alphanumeric characters (except hyphens)
- Limiting to 80 characters
- Ensuring no consecutive hyphens

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
      "path": "/home/user/.wtm/repos/my-project/feature/new-feature",
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
wtm list --json

# Create worktree with JSON response
wtm create feature-branch --json
```

## Configuration

Configuration files are stored in `$HOME/.wtm/config/`:

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