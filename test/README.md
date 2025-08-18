# End-to-End Tests

This directory contains end-to-end tests for the CM CLI application. These tests verify that the CLI behaves correctly by:

1. **Compiling the binary** in isolation
2. **Running CLI commands** in temporary directories
3. **Verifying file system changes** (status.yaml, worktrees, etc.)
4. **Testing error conditions** and edge cases

## Test Structure

### File Organization
- `common.go` - Shared test utilities and setup functions
- `create_single_repo_test.go` - Tests for single repository mode
- `README.md` - This documentation file

### Test Environment
Each test creates a temporary directory structure:
```
/tmp/cm-e2e-test-*/
├── config.yaml          # Custom test configuration
├── cm                  # Compiled binary
├── repo/                # Test Git repository
└── .cm/                # CM data directory
    └── status.yaml      # Status file (created by tests)
```

### Test Coverage

#### Positive Cases
- `TestCreateWorktreeSingleRepo`: Creates worktree successfully
- `TestCreateWorktreeWithVerboseFlag`: Tests verbose output
- `TestCreateWorktreeWithQuietFlag`: Tests quiet output

#### Error Cases
- `TestCreateWorktreeNonExistentBranch`: Tests invalid branch
- `TestCreateWorktreeAlreadyExists`: Tests duplicate worktree creation
- `TestCreateWorktreeOutsideGitRepo`: Tests non-Git directory

## Running Tests

### Run All E2E Tests
```bash
go test ./test/ -tags=e2e -v
```

### Run Specific Test
```bash
go test ./test/ -tags=e2e -v -run TestCreateWorktreeSingleRepo
```

### Run with Coverage
```bash
go test ./test/ -tags=e2e -v -cover
```

### Run Only Unit/Integration Tests (Exclude E2E)
```bash
go test ./test/ -v
```

## Test Features

### Isolation
- Tests run in temporary directories
- Never touch system config (`~/.cm`)
- Clean up after each test
- Custom config files for each test

### Git Repository Setup
- Creates realistic Git repositories
- Multiple branches with commits
- Proper Git configuration for testing

### Verification
- Checks status.yaml file structure
- Verifies worktree creation in `.cm` directory
- Confirms worktree linking in original repository
- Validates CLI output and error messages

## Future Enhancements

The test structure is designed to support:
- **Workspace mode** testing
- **Multiple repository** scenarios
- **Additional subcommands** (list, remove, etc.)
- **Complex Git workflows** (merges, conflicts, etc.)

## Dependencies

- `github.com/stretchr/testify/assert` - Assertions
- `github.com/stretchr/testify/require` - Test requirements
- `gopkg.in/yaml.v3` - YAML parsing for status files
- Git CLI (for repository operations)
