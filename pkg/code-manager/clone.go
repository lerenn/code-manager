package codemanager

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/status"
)

// CloneOpts contains optional parameters for Clone.
type CloneOpts struct {
	Recursive bool // defaults to true
}

// Clone clones a repository and initializes it in CM.
func (c *realCodeManager) Clone(repoURL string, opts ...CloneOpts) error {
	// Extract and validate options
	options := c.extractCloneOptions(opts)

	// Prepare parameters for hooks
	params := map[string]interface{}{
		"repoURL":   repoURL,
		"recursive": options.Recursive,
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
			Recursive:  options.Recursive,
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
func (c *realCodeManager) normalizeRepositoryURL(repoURL string) (string, error) {
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
func (c *realCodeManager) checkRepositoryExists(normalizedURL string) error {
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
func (c *realCodeManager) generateClonePath(normalizedURL, defaultBranch string) string {
	// Get config from ConfigManager
	cfg, err := c.configManager.GetConfigWithFallback()
	if err != nil {
		// Fallback to a default path if config cannot be loaded
		return filepath.Join("~/Code/repos", normalizedURL, "origin", defaultBranch)
	}

	// Use the new path structure: $repositories_dir/<repo_url>/<remote_name>/<default_branch>
	remoteName := "origin" // Default remote name
	return filepath.Join(cfg.RepositoriesDir, normalizedURL, remoteName, defaultBranch)
}

// initializeRepositoryInCM initializes a cloned repository in CM.
func (c *realCodeManager) initializeRepositoryInCM(normalizedURL, targetPath, defaultBranch string) error {
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

// extractCloneOptions extracts and merges options from the variadic parameter.
func (c *realCodeManager) extractCloneOptions(opts []CloneOpts) CloneOpts {
	result := CloneOpts{
		Recursive: true, // default to true
	}

	// Merge all provided options, with later options overriding earlier ones
	for _, opt := range opts {
		result.Recursive = opt.Recursive
	}

	return result
}
