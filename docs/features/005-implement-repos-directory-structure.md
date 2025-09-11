# Feature 005: Implement Repos Directory Structure

## Overview
Implement functionality to create and manage the directory structure `$HOME/.cm/repos/<repo>/<branch>/` for organizing Git worktrees for individual repositories.

## Background
The Code Manager (cm) needs a standardized directory structure to organize and manage Git worktrees. This feature establishes the foundation for storing worktrees for single repositories, providing a clean separation between different repos and branches.

## Requirements

### Functional Requirements
1. **Directory Structure Creation**: Create the full directory path `$HOME/.cm/repos/<repo>/<branch>/`
2. **Repository Name Handling**: Extract and sanitize repository names from Git remote URLs (full path format: e.g., "github.com/lerenn/code-manager")
3. **Branch Name Handling**: Validate and sanitize branch names for safe directory creation (replace invalid characters with underscores/hyphens)
4. **Path Validation**: Ensure the base directory (`$HOME/.cm/`) can be accessed and created
5. **Error Handling**: Provide clear error messages for directory creation failures
6. **Integration Ready**: Return the created path in a format suitable for other features
7. *Configuration Support**: Support configurable base path via config file
8. **Idempotent Behavior**: Allow parent directories to exist, only fail if final branch directory exists

### Non-Functional Requirements
1. **Performance**: Directory creation should be fast (< 50ms)
2. **Reliability**: Handle edge cases (permissions, existing directories, etc.)
3. **Cross-Platform**: Work on Windows, macOS, and Linux
4. **Testability**: Use Uber gomock for mocking file system operations
5. **Minimal Dependencies**: Use only Go standard library + gomock for file system operations
6. **Default Permissions**: Use default system permissions (0755) for created directories

## Technical Specification

### Interface Design

#### Config Package (New Package)
**Package Location**: `pkg/config/`
**Key Components**:
- `Config` struct with YAML tags
- `LoadConfig(configPath string) (Config, error)`: Load configuration from file
- `DefaultConfig() Config`: Return default configuration
- `Validate() error`: Validate configuration values

*Config Structure**:
```go
type Config struct {
    RepositoriesDir string `yaml:"base_path"`
    // Future configuration fields can be added here
}
```

**Key Characteristics**:
- Pure configuration management
- YAML-based configuration files
- Default values when config file is missing
- Validation of configuration values
- No state management required

#### FS Package Extension (File System Adapter)
**New Interface Methods:**
- `MkdirAll(path string, perm os.FileMode) error`: Create directory and all parent directories
- `GetHomeDir() (string, error)`: Get user's home directory path
- `IsNotExist(err error) bool`: Check if error indicates file/directory doesn't exist

**Key Characteristics:**
- Extends existing FS package with directory creation capabilities
- Pure function-based operations
- No state management required
- Cross-platform compatibility

#### CM Package Extension (Business Logic)
**New Interface Methods:**
- `CreateReposDirectoryStructure(repoName, branchName string) (string, error)`: Create and return the full directory path
- `sanitizeRepositoryName(remoteURL string) (string, error)`: Extract and sanitize repo name from Git remote
- `sanitizeBranchName(branchName string) (string, error)`: Validate and sanitize branch name
- `getRepositoriesDir() (string, error)`: Get the configurable base path for CM

**Updated Constructor:**
- `NewCM(config config.Config) CM`: Accept configuration as parameter

**Implementation Structure:**
- Extends existing CM package with directory structure management
- Private helper methods for sanitization and validation
- Error handling with wrapped errors
- Clean separation of concerns
- Integration with existing `Run()` method
- Configuration dependency injection

**Key Characteristics:**
- **NO direct file system access** - all operations go through FS adapter
- **ONLY unit tests** using mocked FS adapter
- Business logic focused on path construction and validation
- Testable through dependency injection
- Method will be called from `Run()` to keep function small
- Configuration passed through constructor

### Implementation Details

#### 1. Config Package
The Config package manages application configuration:

**Key Components:**
- `Config` struct with JSON tags
- `LoadConfig()` function for file-based configuration
- `DefaultConfig()` function for default values
- `Validate()` method for configuration validation
- Error handling for missing/invalid config files

**Implementation Notes:**
- Use `gopkg.in/yaml.v3` for YAML parsing
- Default base path: `$HOME/.cm/`
- Config file location: `$HOME/.cm/config.yaml`
- Graceful fallback to defaults if config file is missing
- Validation of base path accessibility

#### 2. Example Configs Directory
**Location**: `configs/`
**Purpose**: Provide example configuration files for users

**Example Files**:
- `configs/default.yaml`: Default configuration example
- `configs/custom-path.yaml`: Example with custom base path
- `configs/README.md`: Documentation for configuration options

**Example Config Structure**:
```yaml
base_path: /custom/path/to/cm
```

#### 3. FS Package Extension
The FS package extends with directory creation operations:

**Key Components:**
- New methods: `MkdirAll()`, `GetHomeDir()`, `IsNotExist()`
- Concrete implementation using Go standard library (`os`, `path/filepath`)
- Updated `//go:generate` directive for Uber mockgen
- No state - pure function-based operations

**Implementation Notes:**
- Use `os.MkdirAll()` for recursive directory creation
- Use `os.UserHomeDir()` for home directory detection
- Use `os.IsNotExist()` for existence checking
- Handle cross-platform path separators properly
- Update mock generation to include new methods
- Generate mock files as `mockfs.gen.go` in same directory

#### 4. CM Package Extension
The CM package implements directory structure management:

**Key Components:**
- Updated constructor: `NewCM(config config.Config)`
- Public method: `CreateReposDirectoryStructure()`
- Private helper methods: `sanitizeRepositoryName()`, `sanitizeBranchName()`, `getRepositoriesDir()`
- Dependency injection of FS interface and Config
- Error handling with wrapped errors
- Integration with existing `Run()` method

**Implementation Notes:**
- `CreateReposDirectoryStructure()` orchestrates the full directory creation process
- `sanitizeRepositoryName()` extracts full repo path from Git remote URL and sanitizes it
- `sanitizeBranchName()` validates and sanitizes branch names, replacing invalid characters with underscores/hyphens
- `getRepositoriesDir()` retrieves base path from injected Config, with fallback to default
- Use `fmt.Errorf()` with `%w` for error wrapping
- Provide clear user messages for different creation states
- Support verbose mode for detailed directory creation steps
- Idempotent behavior: allow parent directories to exist, only fail if final branch directory exists

### Configuration

#### Config Package Structure
- **Package**: `pkg/config/`
- **Main File**: `pkg/config/config.go`
- **Test File**: `pkg/config/config_test.go`
- **Mock File**: `pkg/config/mockconfig.gen.go`

#### Config File Support
- Support configurable base path via config file
- Default to `$HOME/.cm/` if no config file or base path specified
- Config file location: `$HOME/.cm/config.yaml`
- Config structure:
```yaml
base_path: /custom/path/to/cm
```

#### Example Configs
- **Location**: `configs/`
- **Files**: 
  - `configs/default.yaml`: Default configuration
  - `configs/README.md`: Configuration documentation

### Error Handling

#### Error Types
1. **HomeDirectoryError**: When unable to access or determine home directory
2. **PermissionError**: When unable to create directories due to permissions
3. **InvalidRepositoryNameError**: When repository name cannot be extracted or sanitized
4. **InvalidBranchNameError**: When branch name is invalid or cannot be sanitized
5. **DirectoryCreationError**: When directory creation fails for other reasons
6. *ConfigError**: When unable to read or parse configuration file
7. *ConfigValidationError**: When configuration values are invalid

#### Error Messages
- Clear, user-friendly error messages
- Include context about what operation failed
- Provide suggestions for resolution when possible

### Testing Strategy

#### Unit Tests (CM Package)
- Mock FS adapter for all file system operations
- Mock Config for configuration testing
- Test repository name sanitization with various remote URL formats
- Test branch name sanitization with various branch name formats
- Test error handling for different failure scenarios
- Test path construction logic
- Test configuration integration
- Test idempotent behavior with existing directories

#### Unit Tests (Config Package)
- Test configuration loading from file
- Test default configuration generation
- Test configuration validation
- Test error handling for invalid config files
- Mock file system for config file operations

#### Integration Tests (FS Package)
- Test actual directory creation on real file system
- Test home directory detection across platforms
- Test permission handling
- Test cross-platform path handling

### Dependencies
- **Blocked by:** Feature 004 (Validate Project Structure and Git Configuration)
- **Enables:** Features 006, 007, 011 (Repository/Branch handling and worktree creation)

### Integration with Run() Method
The `CreateReposDirectoryStructure()` method will be called from the existing `Run()` method to keep the function small and maintain separation of concerns. The integration will occur after project validation (Feature 004) and before any worktree operations.

### Integration with NewCM()
The `NewCM()` function will be updated to accept a `config.Config` parameter, allowing configuration to be injected at construction time. This maintains dependency injection principles and makes the code more testable.

### Repository Name Format Decision
- **Format**: Full repository path (e.g., "github.com/lerenn/code-manager")
- **Rationale**: Provides better organization and avoids naming conflicts
- **Sanitization**: Replace invalid characters with underscores/hyphens

### Branch Name Sanitization Decision
- **Allowed Characters**: Alphanumeric, hyphens, underscores
- **Invalid Characters**: Replace `/`, `\`, `:`, `*`, `?`, `"`, `<`, `>`, `|` with underscores
- **Maximum Length**: 255 characters (filesystem limit)
- **Examples**: 
  - `feature/new-branch` → `feature_new-branch`
  - `bugfix/issue#123` → `bugfix_issue_123`

### Directory Permissions Decision
- **Default**: Use system default permissions (0755)
- **Rationale**: Standard practice, allows group/other read access
- **Override**: Configurable via config file if needed

### Existing Directory Handling Decision
- **Behavior**: Idempotent - allow parent directories to exist
- **Failure Condition**: Only fail if final branch directory exists
- **Rationale**: Supports incremental directory creation and avoids unnecessary errors

