package repository

import (
	"fmt"
	"strings"

	"github.com/lerenn/code-manager/pkg/git"
)

// HandleRemoteManagement handles remote addition if the remote doesn't exist.
func (r *realRepository) HandleRemoteManagement(remoteSource string) error {
	// If remote source is "origin", no need to add it
	if remoteSource == "origin" {
		r.deps.Logger.Logf("Using existing origin remote")
		return nil
	}

	// Check if remote already exists and handle existing remote
	if err := r.handleExistingRemote(remoteSource); err != nil {
		// Remote doesn't exist, add it
		return r.addNewRemote(remoteSource)
	}

	// Remote exists, no need to add it
	return nil
}

// handleExistingRemote checks if remote exists and handles it appropriately.
func (r *realRepository) handleExistingRemote(remoteSource string) error {
	exists, err := r.deps.Git.RemoteExists(r.repositoryPath, remoteSource)
	if err != nil {
		return fmt.Errorf("failed to check if remote '%s' exists: %w", remoteSource, err)
	}

	if exists {
		remoteURL, err := r.deps.Git.GetRemoteURL(r.repositoryPath, remoteSource)
		if err != nil {
			return fmt.Errorf("failed to get remote URL: %w", err)
		}

		r.deps.Logger.Logf("Using existing remote '%s' with URL: %s", remoteSource, remoteURL)
		return nil
	}

	// Remote doesn't exist, return a specific error
	return fmt.Errorf("remote '%s' does not exist", remoteSource)
}

// addNewRemote adds a new remote for the given remote source.
func (r *realRepository) addNewRemote(remoteSource string) error {
	r.deps.Logger.Logf("Adding new remote '%s'", remoteSource)

	// Get repository information
	repoName, err := r.deps.Git.GetRepositoryName(r.repositoryPath)
	if err != nil {
		return fmt.Errorf("failed to get repository name: %w", err)
	}

	originURL, err := r.deps.Git.GetRemoteURL(r.repositoryPath, "origin")
	if err != nil {
		return fmt.Errorf("failed to get origin remote URL: %w", err)
	}

	// Construct remote URL
	remoteURL, err := r.ConstructRemoteURL(originURL, remoteSource, repoName)
	if err != nil {
		return err
	}

	r.deps.Logger.Logf("Constructed remote URL: %s", remoteURL)

	// Add the remote
	if err := r.deps.Git.AddRemote(r.repositoryPath, remoteSource, remoteURL); err != nil {
		return fmt.Errorf("%w: %w", git.ErrRemoteAddFailed, err)
	}

	return nil
}

// ConstructRemoteURL constructs the remote URL based on origin URL and remote source.
func (r *realRepository) ConstructRemoteURL(originURL, remoteSource, repoName string) (string, error) {
	protocol := r.DetermineProtocol(originURL)
	host := r.ExtractHostFromURL(originURL)

	if host == "" {
		return "", fmt.Errorf("failed to extract host from origin URL: %s", originURL)
	}

	repoNameShort := r.ExtractRepoNameFromFullPath(repoName)

	if protocol == "ssh" {
		return fmt.Sprintf("git@%s:%s/%s.git", host, remoteSource, repoNameShort), nil
	}
	return fmt.Sprintf("https://%s/%s/%s.git", host, remoteSource, repoNameShort), nil
}

// ExtractHostFromURL extracts the host from a Git remote URL.
func (r *realRepository) ExtractHostFromURL(url string) string {
	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Handle SSH format: git@host:user/repo
	if strings.Contains(url, "@") && strings.Contains(url, ":") {
		parts := strings.Split(url, ":")
		if len(parts) == 2 {
			hostParts := strings.Split(parts[0], "@")
			if len(hostParts) == 2 {
				return hostParts[1] // host
			}
		}
	}

	// Handle HTTPS format: https://host/user/repo
	if strings.HasPrefix(url, "http") {
		parts := strings.Split(url, "/")
		if len(parts) >= 3 {
			return parts[2] // host
		}
	}

	return ""
}

// DetermineProtocol determines the protocol (https or ssh) from the origin URL.
func (r *realRepository) DetermineProtocol(originURL string) string {
	if strings.HasPrefix(originURL, "git@") || strings.HasPrefix(originURL, "ssh://") {
		return "ssh"
	}
	return "https"
}

// ExtractRepoNameFromFullPath extracts just the repository name from the full path.
func (r *realRepository) ExtractRepoNameFromFullPath(fullPath string) string {
	parts := strings.Split(fullPath, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1] // Return the last part (repo name)
	}
	return fullPath
}
