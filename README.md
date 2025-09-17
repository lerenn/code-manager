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
- Organized directory structure: `$repositories_dir/<repo_url>/<remote_name>/<branch>`

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

### 🏗️ Repository Management
- Clone, list, and delete repositories with automatic CM initialization
- Organized repository structure with remote tracking
- Default branch detection and management

### 🏢 Workspace Management
- Create, list, and delete multi-repository workspaces
- Automatic repository addition to status tracking
- Workspace-specific worktree management

### 🔧 Extensible Hook System
- Pre/post/error hooks for all operations
- Custom middleware for logging, validation, and business logic
- Plugin-like architecture for extensibility

## Installation

```bash
# Install directly from GitHub
go install github.com/lerenn/code-manager/cmd/cm@latest

# Verify installation
cm --help
```

**Prerequisites:**
- Go 1.19 or later
- `$GOPATH/bin` in your `$PATH` (usually already configured)

**For GitHub Integration:**
- `GITHUB_TOKEN` environment variable (optional, for private repositories or rate limit increases)

## First-Time Setup

Before using CM, you need to initialize it:

```bash
# Interactive initialization
cm init

# Initialize with default settings
cm init --repositories-dir ~/Code/src

# Initialize with custom directories
cm init --repositories-dir ~/Projects/src --workspaces-dir ~/Projects/workspaces

# Reset existing configuration
cm init --reset
```

## Usage

### Basic Commands

```bash
# Initialize CM configuration
cm init

# Clone a repository
cm repository clone <repository-url>

# Create a worktree for a branch
cm worktree create <branch-name>

# Create worktree and open in IDE
cm worktree create <branch-name> -i cursor

# List all worktrees
cm worktree list

# Load a branch from remote
cm worktree load <remote>:<branch-name>

# Open existing worktree in IDE
cm worktree open <branch-name> -i cursor

# Delete a worktree
cm worktree delete <branch-name>
```

### Project Structure

#### Single Repository Mode
Worktrees are created at:
```
$repositories_dir/<repo_url>/<remote_name>/<branch>/
```

#### Workspace Mode
Worktrees are created at:
```
$repositories_dir/<repo_url>/<remote_name>/<branch>/<repo_name>/
```

## Command Reference

### `init [options]`
Initializes CM configuration for first-time use.

**Options:**
- `--repositories-dir <path>, -r`: Set the repositories directory directly
- `--workspaces-dir <path>, -w`: Set the workspaces directory directly
- `--status-file <path>, -s`: Set the status file location directly
- `--reset, -R`: Reset existing CM configuration and start fresh
- `--force, -f`: Skip interactive confirmation when using --reset flag

**Examples:**
```bash
# Interactive initialization
cm init

# Initialize with specific repositories directory
cm init --repositories-dir ~/Projects

# Initialize with custom settings
cm init --repositories-dir ~/Code/src --workspaces-dir ~/Code/workspaces

# Reset existing configuration
cm init --reset --force
```

### `repository clone <repository-url> [options]`
Clones a repository and initializes it in CM.

**Options:**
- `--shallow, -s`: Perform a shallow clone (non-recursive)

**Examples:**
```bash
# Clone repository
cm repository clone https://github.com/octocat/Hello-World.git

# Shallow clone
cm repository clone git@github.com:lerenn/example.git --shallow

# Using aliases
cm repo clone https://github.com/octocat/Hello-World.git
cm r clone git@github.com:lerenn/example.git
```

### `repository list [options]`
Lists all repositories tracked by CM.

**Examples:**
```bash
# List all repositories
cm repository list

# Using aliases
cm repo list
cm r list
```

### `repository delete <repository-name> [options]`
Removes a repository from CM tracking and optionally deletes the local directory.

**Options:**
- `--force`: Force deletion without confirmation

**Examples:**
```bash
# Delete repository with confirmation
cm repository delete my-repo

# Force delete without confirmation
cm repository delete my-repo --force

# Using aliases
cm repo delete my-repo
cm r delete my-repo --force
```

### `worktree create <branch> [options]`
Creates a new worktree for the specified branch.

**Options:**
- `-i, --ide <ide-name>`: Open the worktree in IDE after creation
- `-f, --force`: Force creation without prompts

**Examples:**
```bash
# Create persistent worktree
cm worktree create feature/new-feature

# Create worktree and open in Cursor IDE
cm worktree create hotfix/bug-fix -i cursor

# Force creation
cm worktree create feature-branch --force

# Using aliases
cm wt create feature-branch
cm w create feature-branch -i vscode
```

### `worktree load [remote:]<branch-name> [options]`
Loads a branch from a remote source and creates a worktree.

**Options:**
- `-i, --ide <ide-name>`: Open in specified IDE after loading

**Examples:**
```bash
# Load branch from origin
cm worktree load origin:feature-branch

# Load branch from another user's fork
cm worktree load otheruser:feature-branch

# Load and open in IDE
cm worktree load feature-branch -i cursor

# Using aliases
cm wt load upstream:main
cm w load feature-branch --ide vscode
```

### `worktree list [options]`
Lists all active worktrees for the current project.

**Options:**
- `-f, --force`: Force listing without prompts

**Examples:**
```bash
# List worktrees
cm worktree list

# Force listing
cm worktree list --force

# Using aliases
cm wt list
cm w list
```

**Output Format:**
```
Worktrees:
  [origin] main
  [origin] feature/new-feature
  [upstream] develop
```

### `worktree open <branch> [options]`
Opens a worktree in the specified IDE.

**Options:**
- `-i, --ide <ide-name>`: Open in specified IDE (defaults to cursor)

**Examples:**
```bash
# Open worktree in default IDE (cursor)
cm worktree open feature-branch

# Open in specific IDE
cm worktree open main -i vscode

# Using aliases
cm wt open feature-branch
cm w open main --ide goland
```

### `worktree delete <branch> [options]`
Safely removes a worktree and cleans up Git state.

**Options:**
- `--force`: Force deletion without confirmation

**Examples:**
```bash
# Delete with confirmation
cm worktree delete feature/new-feature

# Force delete without confirmation
cm worktree delete bugfix/issue-123 --force

# Using aliases
cm wt delete feature-branch
cm w delete hotfix/critical-fix --force
```

### `workspace create <workspace-name> [repositories...] [options]`
Creates a new workspace definition with the specified repositories.

**Examples:**
```bash
# Create workspace with repository names from status
cm workspace create my-workspace repo1 repo2

# Create workspace with absolute paths
cm workspace create my-workspace /path/to/repo1 /path/to/repo2

# Create workspace with relative paths
cm workspace create my-workspace ./repo1 ../repo2

# Using aliases
cm ws create my-workspace repo1 repo2
```

### `workspace list [options]`
Lists all workspaces tracked by CM.

**Examples:**
```bash
# List all workspaces
cm workspace list

# Using aliases
cm ws list
```

### `workspace delete <workspace-name> [options]`
Deletes a workspace and all associated worktrees and files.

**Options:**
- `--force`: Skip confirmation prompts

**Examples:**
```bash
# Delete workspace with confirmation
cm workspace delete my-workspace

# Force delete without confirmation
cm workspace delete my-workspace --force

# Using aliases
cm ws delete my-workspace
```

## Global Options

All commands support these global options:

- `-v, --verbose`: Enable verbose output
- `-q, --quiet`: Suppress all output except errors
- `-c, --config <path>`: Specify a custom config file path

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
- **Repository Validation**: Validates repository structure and Git configuration

## Configuration

Configuration files are stored in `$HOME/.cm/`:

- `config.yaml`: Main configuration file with repositories directory and status file location
- `status.yaml`: Status file tracking repositories, worktrees, and workspaces

### Default Configuration
```yaml
# Repositories directory
repositories_dir: ~/Code/src

# Status file path
status_file: ~/.cm/status.yaml

# Worktrees directory (computed as $repositories_dir/worktrees)
worktrees_dir: ~/Code/src/worktrees
```

## Extension Integration

The `--json` flag enables structured output for extension development:

```bash
# Get worktree list in JSON format
cm worktree list --json

# Create worktree with JSON response
cm worktree create feature-branch --json
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

- [x] Hook system for extensibility
- [x] Workspace creation and management
- [x] Repository management (clone, list, delete)
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