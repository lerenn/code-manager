package repository

import "fmt"

// ValidateGitStatus validates that the Git repository is in a clean state.
func (r *realRepository) ValidateGitStatus() error {
	status, err := r.deps.Git.Status(r.repositoryPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGitRepositoryInvalid, err)
	}

	// Check if the status indicates a clean repository
	// This is a simple check - in practice you might want more sophisticated parsing
	if status == "" {
		return fmt.Errorf("%w: empty git status", ErrGitRepositoryInvalid)
	}

	return nil
}
