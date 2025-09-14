package repository

import "fmt"

// ValidateOriginRemote validates that the origin remote exists and is a valid Git hosting service URL.
func (r *realRepository) ValidateOriginRemote() error {
	return r.validateRemote(DefaultRemote)
}

// validateRemote validates that the specified remote exists and has a valid URL.
func (r *realRepository) validateRemote(remote string) error {
	r.logger.Logf("Validating remote: %s", remote)

	// Check if remote exists
	exists, err := r.git.RemoteExists(r.repositoryPath, remote)
	if err != nil {
		return fmt.Errorf("failed to check remote %s: %w", remote, err)
	}
	if !exists {
		return fmt.Errorf("%w: remote '%s' not found", ErrOriginRemoteNotFound, remote)
	}

	// Get remote URL
	remoteURL, err := r.git.GetRemoteURL(r.repositoryPath, remote)
	if err != nil {
		return fmt.Errorf("failed to get remote %s URL: %w", remote, err)
	}

	// Validate that it's a valid Git hosting service URL
	if r.ExtractHostFromURL(remoteURL) == "" {
		return fmt.Errorf("%w: remote '%s' has invalid URL", ErrOriginRemoteInvalidURL, remote)
	}

	return nil
}
