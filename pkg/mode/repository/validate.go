package repository

// Validate validates that the current directory is a working Git repository.
func (r *realRepository) Validate() error {
	r.deps.Logger.Logf("Validating repository: %s", r.repositoryPath)

	// Check if we're in a Git repository
	exists, err := r.IsGitRepository()
	if err != nil {
		return err
	}
	if !exists {
		return ErrGitRepositoryNotFound
	}

	if err := r.ValidateGitStatus(); err != nil {
		return err
	}

	// Validate Git configuration is functional
	return r.ValidateGitConfiguration(r.repositoryPath)
}
