# CGWT Configuration

This directory contains example configuration files for the Cursor Git WorkTree Manager (CGWT).

## Configuration Files

### default.yaml
Default configuration example showing the standard configuration structure.

### Configuration Options

#### base_path
- **Type**: string
- **Default**: `$HOME/.cursor/cgwt`
- **Description**: The base directory where CGWT will store its data, including repository worktrees.

## Usage

1. Copy the example configuration file to your home directory:
   ```bash
   cp configs/default.yaml ~/.cursor/cgwt/config.yaml
   ```

2. Edit the configuration file to customize your settings:
   ```bash
   nano ~/.cursor/cgwt/config.yaml
   ```

3. CGWT will automatically load the configuration from `~/.cursor/cgwt/config.yaml` when it starts.

## File Format

CGWT uses YAML format for configuration files. The configuration file should be located at:
`$HOME/.cursor/cgwt/config.yaml`

## Example Configuration

```yaml
# Base path for CGWT data storage
base_path: /custom/path/to/cgwt
```

## Validation

CGWT validates the configuration on startup:
- The `base_path` must not be empty
- The parent directory of `base_path` must be accessible and writable
- If validation fails, CGWT will fall back to the default configuration
