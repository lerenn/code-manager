package cm

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/status"
)

// CloneOpts contains optional parameters for Clone.
type CloneOpts struct {
	Recursive bool // defaults to true
}

// Clone clones a repository and initializes it in CM.
func (c *realCM) Clone(repoURL string, opts ...CloneOpts) error {
	// Extract and validate options
	recursive := true // default to true
	if len(opts) > 0 {
		recursive = opts[0].Recursive
	}

	// Prepare parameters for hooks
	params := map[string]interface{}{
		"repoURL":   repoURL,
		"recursive": recursive,
	}

	// Execute with hooks
	return c.executeWithHooks(consts.Clone, params, func() error {
		c.VerbosePrint("Starting repository clone: %s", repoURL)

		// 1. Validate repository URL
		normalizedURL, err := c.normalizeRepositoryURL(repoURL)
		if err != nil {
			return err
		}

		c.VerbosePrint("Normalized URL: %s", normalizedURL)

		// 2. Check if repository already exists
		if err := c.checkRepositoryExists(normalizedURL); err != nil {
			return err
		}

		// 3. Detect default branch from remote
		defaultBranch, err := c.git.GetDefaultBranch(repoURL)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToDetectDefaultBranch, err)
		}

		c.VerbosePrint("Detected default branch: %s", defaultBranch)

		// 4. Generate target path
		targetPath := c.generateClonePath(normalizedURL, defaultBranch)

		c.VerbosePrint("Target path: %s", targetPath)

		// 5. Create parent directories for the target path
		parentDir := filepath.Dir(targetPath)
		if err := c.fs.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create parent directories: %w", err)
		}

		// 6. Clone repository
		if err := c.git.Clone(git.CloneParams{
			RepoURL:    repoURL,
			TargetPath: targetPath,
			Recursive:  recursive,
		}); err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToCloneRepository, err)
		}

		// 7. Initialize repository in CM
		if err := c.initializeRepositoryInCM(normalizedURL, targetPath, defaultBranch); err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToInitializeRepository, err)
		}

		c.VerbosePrint("Repository cloned and initialized successfully")
		return nil
	})
}

// normalizeRepositoryURL normalizes a repository URL to a consistent format.
func (c *realCM) normalizeRepositoryURL(repoURL string) (string, error) {
	if repoURL == "" {
		return "", ErrRepositoryURLEmpty
	}

	// Remove .git suffix if present
	normalized := strings.TrimSuffix(repoURL, ".git")

	// Handle SSH URLs (git@host:user/repo) first
	if strings.Contains(normalized, "@") && strings.Contains(normalized, ":") && !strings.HasPrefix(normalized, "http") {
		parts := strings.Split(normalized, ":")
		if len(parts) == 2 {
			hostParts := strings.Split(parts[0], "@")
			if len(hostParts) == 2 {
				host := hostParts[1]
				path := parts[1]
				return host + "/" + path, nil
			}
		}
	}

	// Handle HTTPS URLs
	if strings.HasPrefix(normalized, "http") {
		parsedURL, err := url.Parse(normalized)
		if err != nil {
			return "", fmt.Errorf("invalid repository URL: %w", err)
		}

		host := parsedURL.Host
		path := strings.TrimPrefix(parsedURL.Path, "/")
		return host + "/" + path, nil
	}

	// If we get here, the URL format is not supported
	return "", fmt.Errorf("%w: %s", ErrUnsupportedRepositoryURLFormat, repoURL)
}

// checkRepositoryExists checks if a repository already exists in the status file.
func (c *realCM) checkRepositoryExists(normalizedURL string) error {
	repos, err := c.statusManager.ListRepositories()
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	if _, exists := repos[normalizedURL]; exists {
		return fmt.Errorf("%w: %s", ErrRepositoryExists, normalizedURL)
	}

	return nil
}

// generateClonePath generates the target path for cloning a repository.
func (c *realCM) generateClonePath(normalizedURL, defaultBranch string) string {
	// Use the new path structure: $repositories_dir/<repo_url>/<remote_name>/<default_branch>
	remoteName := "origin" // Default remote name
	return filepath.Join(c.config.RepositoriesDir, normalizedURL, remoteName, defaultBranch)
}

// initializeRepositoryInCM initializes a cloned repository in CM.
func (c *realCM) initializeRepositoryInCM(normalizedURL, targetPath, defaultBranch string) error {
	// Create repository entry in status file
	remotes := map[string]status.Remote{
		"origin": {
			DefaultBranch: defaultBranch,
		},
	}

	err := c.statusManager.AddRepository(normalizedURL, status.AddRepositoryParams{
		Path:    targetPath,
		Remotes: remotes,
	})
	if err != nil {
		return fmt.Errorf("failed to add repository to status: %w", err)
	}

	return nil
}
