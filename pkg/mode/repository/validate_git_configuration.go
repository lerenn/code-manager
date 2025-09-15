package repository

import "fmt"

// ValidateGitConfiguration validates that Git configuration is functional.
func (r *realRepository) ValidateGitConfiguration(workDir string) error {
	// Check if Git is available and functional by running a simple command
	_, err := r.deps.Git.GetCurrentBranch(workDir)
	if err != nil {
		return fmt.Errorf("git configuration validation failed: %w", err)
	}
	return nil
}
