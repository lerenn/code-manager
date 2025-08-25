package cm

import (
	"fmt"
	"sort"
)

// RepositoryInfo contains information about a repository for display purposes.
type RepositoryInfo struct {
	Name       string
	Path       string
	InBasePath bool
}

// ListRepositories lists all repositories from the status file with base path validation.
func (c *realCM) ListRepositories() ([]RepositoryInfo, error) {
	if c.Logger != nil {
		c.Logger.Logf("Loading repositories from status file")
	}

	// Get all repositories from status manager
	repositories, err := c.StatusManager.ListRepositories()
	if err != nil {
		return nil, fmt.Errorf("failed to load repositories from status file: %w", err)
	}

	if c.Logger != nil {
		c.Logger.Logf("Validating base path for repositories")
	}

	// Convert to RepositoryInfo slice with base path validation
	var repoInfos []RepositoryInfo
	for repoName, repo := range repositories {
		// Check if repository path is within configured base path
		inBasePath, err := c.FS.IsPathWithinBase(c.Config.BasePath, repo.Path)
		if err != nil {
			// Log warning but continue processing other repositories
			if c.Logger != nil {
				c.Logger.Logf("Failed to validate base path for repository %s: %v", repoName, err)
			}
			// Default to false if validation fails
			inBasePath = false
		}

		repoInfo := RepositoryInfo{
			Name:       repoName,
			Path:       repo.Path,
			InBasePath: inBasePath,
		}
		repoInfos = append(repoInfos, repoInfo)
	}

	// Sort repositories by name for consistent ordering
	sort.Slice(repoInfos, func(i, j int) bool {
		return repoInfos[i].Name < repoInfos[j].Name
	})

	if c.Logger != nil {
		c.Logger.Logf("Formatting repository list")
	}

	return repoInfos, nil
}
