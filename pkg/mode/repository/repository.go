// Package repository provides Git repository management functionality for CM.
package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/dependencies"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/mode/repository/interfaces"
	"github.com/lerenn/code-manager/pkg/status"
)

// DefaultRemote is the default remote name used for Git operations.
const DefaultRemote = "origin"

// Repository interface provides repository management capabilities.
// This interface is now defined in pkg/mode/repository/interfaces to avoid circular imports.
type Repository = interfaces.Repository

// CreateWorktreeOpts contains optional parameters for worktree creation in repository mode.
type CreateWorktreeOpts = interfaces.CreateWorktreeOpts

// LoadWorktreeOpts contains optional parameters for LoadWorktree.
type LoadWorktreeOpts = interfaces.LoadWorktreeOpts

// ValidationParams contains parameters for repository validation.
type ValidationParams = interfaces.ValidationParams

// ValidationResult contains the result of repository validation.
type ValidationResult = interfaces.ValidationResult

// StatusParams contains parameters for status operations.
type StatusParams = interfaces.StatusParams

// NewRepositoryParams contains parameters for creating a new Repository instance.
type NewRepositoryParams = interfaces.NewRepositoryParams

// realRepository represents a single Git repository and provides methods for repository operations.
type realRepository struct {
	deps           *dependencies.Dependencies
	repositoryPath string
}

// NewRepository creates a new Repository instance.
func NewRepository(params NewRepositoryParams) Repository {
	// Cast interface{} to concrete type
	deps := params.Dependencies.(*dependencies.Dependencies)
	if deps == nil {
		deps = dependencies.New()
	}

	// Resolve repository path from repository name/path
	repoPath, err := resolveRepositoryPath(params.RepositoryName, deps.StatusManager, deps.Logger)
	if err != nil {
		deps.Logger.Logf("Warning: failed to resolve repository path '%s': %v", params.RepositoryName, err)
		repoPath = "." // Fallback to current directory
	}

	return &realRepository{
		deps:           deps,
		repositoryPath: repoPath,
	}
}

// CreateWorktreeOpts contains optional parameters for CreateWorktree.
// This is now defined in the mode package for consistency.

// resolveRepositoryPath resolves a repository name/path to an actual path, checking status file first.
func resolveRepositoryPath(repoName string, statusManager status.Manager, logger logger.Logger) (string, error) {
	// If empty, use current directory
	if repoName == "" {
		return ".", nil
	}

	// First, check if it's a repository name from status.yaml
	if existingRepo, err := statusManager.GetRepository(repoName); err == nil && existingRepo != nil {
		logger.Logf("Resolved repository '%s' from status.yaml: %s", repoName, existingRepo.Path)
		return existingRepo.Path, nil
	}

	// Check if it's an absolute path
	if filepath.IsAbs(repoName) {
		return repoName, nil
	}

	// Resolve relative path from current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	resolvedPath := filepath.Join(currentDir, repoName)
	return resolvedPath, nil
}
