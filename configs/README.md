# WTM Configuration

This directory contains example configuration files for the Git WorkTree Manager (WTM).

## Configuration Files

### default.yaml
Default configuration example showing the standard configuration structure.

### Configuration Options

#### base_path
- **Type**: string
- **Default**: `$HOME/.wtm`
- **Description**: The base directory where WTM will store its data, including repository worktrees.

## Usage

1. Copy the example configuration file to your home directory:
   ```bash
   cp configs/default.yaml ~/.wtm/config.yaml
   ```

2. Edit the configuration file to customize your settings:
   ```bash
   nano ~/.wtm/config.yaml
   ```

3. WTM will automatically load the configuration from `~/.wtm/config.yaml` when it starts.

## File Format

WTM uses YAML format for configuration files. The configuration file should be located at:
`$HOME/.wtm/config.yaml`

## Example Configuration

```yaml
# Base path for WTM data storage
base_path: /custom/path/to/wtm
```

## Validation

WTM validates the configuration on startup:
- The `base_path` must not be empty
- The parent directory of `base_path` must be accessible and writable
- If validation fails, WTM will fall back to the default configuration
