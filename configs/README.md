# CM Configuration

This directory contains example configuration files for the Code Manager (CM).

## Configuration Structure

The CM configuration file uses YAML format and contains the following settings:

### Base Path
- **Key**: `base_path`
- **Default**: `$HOME/.cm`
- **Description**: The base directory where CM will store its data, including repository worktrees.

### Status File
- **Key**: `status_file`
- **Default**: `$HOME/.cm/status.yaml`
- **Description**: The path to the status file that tracks CM worktrees and their metadata.

### Worktrees Directory
- **Key**: `worktrees_dir`
- **Default**: `$HOME/.cm/worktrees`
- **Description**: The directory where CM will store all repository worktrees. If not specified, worktrees will be stored directly under the base_path.

## Installation

1. Copy the default configuration to your home directory:
```bash
cp configs/default.yaml ~/.cm/config.yaml
```

2. Edit the configuration file to customize settings:
```bash
nano ~/.cm/config.yaml
```

3. CM will automatically load the configuration from `~/.cm/config.yaml` when it starts.

## Configuration Format

CM uses YAML format for configuration files. The configuration file should be located at:
`$HOME/.cm/config.yaml`

## Example Configuration

```yaml
# Base path for CM data storage
base_path: /custom/path/to/cm

# Status file path
status_file: /custom/path/to/cm/status.yaml

# Worktrees directory path
worktrees_dir: /custom/path/to/cm/worktrees
```

## Validation

CM validates the configuration on startup:
- Checks if all required fields are present
- Validates file paths and permissions
- If validation fails, CM will fall back to the default configuration
