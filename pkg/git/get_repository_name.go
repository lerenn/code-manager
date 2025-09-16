package git

import (
	"fmt"
	"path/filepath"
	"strings"
)

// GetRepositoryName gets the repository name from remote origin URL with fallback to local path.
func (g *realGit) GetRepositoryName(repoPath string) (string, error) {
	// Try to get remote origin URL first
	originURL, err := g.ConfigGet(repoPath, "remote.origin.url")
	if err != nil {
		return "", fmt.Errorf("failed to get remote origin URL: %w", err)
	}

	// Trim whitespace and newlines from the URL
	originURL = strings.TrimSpace(originURL)

	if originURL != "" {
		// Extract repository name from URL
		// Handle different URL formats: https://host/user/repo.git, git@host:user/repo.git
		repoName := g.extractRepoNameFromURL(originURL)
		if repoName != "" {
			return repoName, nil
		}
	}

	// Fallback to local repository path
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Use the directory name as repository name, removing .git suffix if present
	dirName := filepath.Base(absPath)
	return strings.TrimSuffix(dirName, ".git"), nil
}
