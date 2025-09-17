# Contributing to Code Manager (CM)

Thank you for your interest in contributing to Code Manager! This document provides guidelines and information for contributors.

## Table of Contents

- [Development Environment Setup](#development-environment-setup)
- [Coding Standards and Style Guidelines](#coding-standards-and-style-guidelines)
- [Testing Strategy](#testing-strategy)
- [Submitting Pull Requests](#submitting-pull-requests)
- [Code of Conduct](#code-of-conduct)
- [Contributors](#contributors)

## Development Environment Setup

### Prerequisites

- Go 1.24.4 or later
- Git
- Docker (for Dagger-based testing and building)

### Setup Steps

1. **Clone the repository:**
   ```bash
   git clone https://github.com/lerenn/code-manager.git
   cd code-manager
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Install development tools:**
   ```bash
   # Install Dagger
   curl -L https://dl.dagger.io/dagger/install.sh | sh

   # Install golangci-lint (optional, for local linting)
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

4. **Verify setup:**
   ```bash
   go version
   dagger version
   ```

### Building and Running

```bash
# Build the project
go build -o cm ./cmd/cm

# Run tests
dagger call unit-tests --source-dir=.
dagger call integration-tests --source-dir=.
dagger call end-to-end-tests --source-dir=.

# Run linter
dagger call lint --source-dir=.
```

## Coding Standards and Style Guidelines

### Go Code Standards

- Follow standard Go formatting (`go fmt`)
- Use `golangci-lint` for code quality checks (configuration in `.golangci.yml`)
- Write clear, concise, and well-documented code
- Use meaningful variable and function names
- Follow Go naming conventions (exported/unexported identifiers)

### Project-Specific Rules

#### Testing Strategy
- **Adapters** (`pkg/fs`): Integration tests only using real file system
- **Business Logic** (`pkg/cm`): Unit tests only with mocked dependencies
- **E2E Tests** (`test/`): Real CM struct with Git operations
- Use build tags: `//go:build unit|integration|e2e`
- File naming: `*_test.go`, `*_integration_test.go`, `*_test.go` in `test/` for e2e
- Mocking: Use Uber gomock with `//go:generate` directives
- Assertions: Use `testify/assert` with `assert.ErrorIs()` for specific errors
- **NEVER** touch `~/.cm/` files in tests; use `os.MkdirTemp()` and `-c` flag

#### Function Design
- Use `XXXOpts` structs for optional parameters with variadic `...XXXOpts`
- Use `XXXParams` structs for functions with >3 arguments
- Check `len(opts) > 0` for optional parameters
- No proxy functions - accept parameter structs directly

#### CLI Architecture
- CLI layer: Only boilerplate (flags, config, basic validation, CM calls, display)
- **NO business logic in CLI** - all parsing/validation in CM package
- Keep CLI thin: Pass args to CM methods, display results

### Commit Messages

Use conventional commit format:
```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New features
- `fix`: Bug fixes
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Testing changes
- `chore`: Maintenance tasks

## Testing Strategy

### Test Categories

1. **Unit Tests** (`//go:build unit`)
   - Test business logic with mocked dependencies
   - Focus on individual functions and methods
   - Use Uber gomock for mocking

2. **Integration Tests** (`//go:build integration`)
   - Test adapters with real file system operations
   - Test external service integrations
   - Use temporary directories and cleanup

3. **End-to-End Tests** (`//go:build e2e`)
   - Test complete workflows with real CM binary
   - Test Git operations and CLI interactions
   - Located in `test/` directory

### Running Tests

```bash
# Generate mocks
go generate ./...

# Run all tests
dagger call unit-tests --source-dir=.
dagger call integration-tests --source-dir=.
dagger call end-to-end-tests --source-dir=.

# Run specific test types
go test -tags=unit ./...
go test -tags=integration ./...
go test -tags=e2e ./test/...
```

### Test Guidelines

- Write tests for all new functionality
- Use descriptive test names: `TestFunctionName_Scenario_ExpectedResult`
- Test both success and failure scenarios
- Clean up resources in tests (use `defer`)
- Mock immediate dependencies only, not transitive ones

## Submitting Pull Requests

### Before Submitting

1. **Ensure all tests pass:**
   ```bash
   go generate ./...
   dagger call lint --source-dir=. stdout
   dagger call unit-tests --source-dir=. stdout
   dagger call integration-tests --source-dir=. stdout
   dagger call end-to-end-tests --source-dir=. stdout
   ```

2. **Follow coding standards:**
   - Run `go fmt` on all files
   - Ensure golangci-lint passes
   - Update documentation as needed

3. **Test your changes:**
   - Add unit tests for new functionality
   - Update existing tests if behavior changes
   - Test manually with the CM binary

### Pull Request Process

1. **Create a feature branch:**
   ```bash
   git checkout -b feat/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Make your changes:**
   - Write clear, focused commits
   - Update documentation and tests
   - Follow the coding standards

3. **Submit the PR:**
   - Use a clear, descriptive title
   - Provide detailed description of changes
   - Reference any related issues
   - Request review from maintainers

4. **Address feedback:**
   - Respond to review comments
   - Make requested changes
   - Update tests and documentation as needed

### CI/CD Pipeline

All PRs must pass:
- **Linting**: Code quality checks
- **Unit Tests**: Business logic testing
- **Integration Tests**: Adapter testing
- **E2E Tests**: Full workflow testing

The pipeline uses Dagger for containerized testing and building.

## Code of Conduct

### Our Pledge

We pledge to make participation in our project and community a harassment-free experience for everyone, regardless of age, body size, disability, ethnicity, gender identity and expression, level of experience, nationality, personal appearance, race, religion, or sexual identity and orientation.

### Our Standards

- Be respectful and inclusive
- Focus on constructive feedback
- Accept responsibility for mistakes
- Show empathy towards other community members
- Help create a positive environment

### Unacceptable Behavior

- Harassment, discrimination, or offensive comments
- Personal attacks or trolling
- Publishing others' private information
- Any other conduct that could reasonably be considered inappropriate

### Enforcement

Violations of the code of conduct may result in temporary or permanent bans from the community, depending on the severity of the violation.

## Contributors

### Core Contributors

- **lerenn**: Project maintainer and lead developer

### How to Become a Contributor

1. Fork the repository
2. Create a feature branch
3. Make your changes following the guidelines above
4. Submit a pull request
5. Get your changes reviewed and merged

### Recognition

Contributors are recognized in:
- Git commit history
- This CONTRIBUTORS.md file
- GitHub's contributor insights

Thank you for contributing to Code Manager! 🎉