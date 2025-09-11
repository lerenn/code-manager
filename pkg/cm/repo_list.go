package cm

import (
	"fmt"
	"sort"

	"github.com/lerenn/code-manager/pkg/cm/consts"
)

// RepositoryInfo contains information about a repository for display purposes.
type RepositoryInfo struct {
	Name              string
	Path              string
	InRepositoriesDir bool
}

// ListRepositories lists all repositories from the status file with base path validation.
func (c *realCM) ListRepositories() ([]RepositoryInfo, error) {
	// Prepare parameters for hooks
	params := map[string]interface{}{}

	// Execute with hooks
	return c.executeWithHooksAndReturnRepositories(consts.ListRepositories, params, func() ([]RepositoryInfo, error) {
		if c.logger != nil {
			c.logger.Logf("Loading repositories from status file")
		}

		// Get all repositories from status manager
		repositories, err := c.statusManager.ListRepositories()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedToLoadRepositories, err)
		}

		if c.logger != nil {
			c.logger.Logf("Validating base path for repositories")
		}

		// Convert to RepositoryInfo slice with base path validation
		var repoInfos []RepositoryInfo
		for repoName, repo := range repositories {
			// Check if repository path is within configured repositories directory
			inRepositoriesDir, err := c.fs.IsPathWithinBase(c.config.RepositoriesDir, repo.Path)
			if err != nil {
				// Log warning but continue processing other repositories
				if c.logger != nil {
					c.logger.Logf("Failed to validate base path for repository %s: %v", repoName, err)
				}
				// Default to false if validation fails
				inRepositoriesDir = false
			}

			repoInfo := RepositoryInfo{
				Name:              repoName,
				Path:              repo.Path,
				InRepositoriesDir: inRepositoriesDir,
			}
			repoInfos = append(repoInfos, repoInfo)
		}

		// Sort repositories by name for consistent ordering
		sort.Slice(repoInfos, func(i, j int) bool {
			return repoInfos[i].Name < repoInfos[j].Name
		})

		if c.logger != nil {
			c.logger.Logf("Formatting repository list")
		}

		return repoInfos, nil
	})
}
