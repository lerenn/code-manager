package cm

import (
	"fmt"
	"strings"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/mode"
	repo "github.com/lerenn/code-manager/pkg/mode/repository"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// LoadWorktreeOpts contains optional parameters for LoadWorktree.
type LoadWorktreeOpts struct {
	IDEName        string
	RepositoryName string
}

// LoadWorktree loads a branch from a remote source and creates a worktree.
func (c *realCM) LoadWorktree(branchArg string, opts ...LoadWorktreeOpts) error {
	// Parse options
	options := c.extractLoadWorktreeOptions(opts)

	// Prepare parameters for hooks
	params := c.prepareLoadWorktreeParams(branchArg, options)

	// Execute with hooks
	return c.executeWithHooks(consts.LoadWorktree, params, func() error {
		return c.executeLoadWorktreeLogic(branchArg, options, params)
	})
}

func (c *realCM) prepareLoadWorktreeParams(branchArg string, options LoadWorktreeOpts) map[string]interface{} {
	params := map[string]interface{}{
		"branchArg":       branchArg,
		"repository_name": options.RepositoryName,
	}
	if options.IDEName != "" {
		params["ideName"] = options.IDEName
	}
	return params
}

func (c *realCM) executeLoadWorktreeLogic(
	branchArg string,
	options LoadWorktreeOpts,
	params map[string]interface{},
) error {
	c.VerbosePrint("Starting branch loading: %s", branchArg)

	// 1. Parse the branch argument to extract remote and branch name
	remoteSource, branchName, err := c.parseBranchArg(branchArg)
	if err != nil {
		return err
	}

	c.VerbosePrint("Parsed: remote=%s, branch=%s", remoteSource, branchName)

	// 2. Handle repository-specific loading if repository name is provided
	if options.RepositoryName != "" {
		return c.handleRepositorySpecificLoading(options.RepositoryName, remoteSource, branchName, params)
	}

	// 3. Handle general loading based on project mode
	return c.handleGeneralLoading(remoteSource, branchName, params)
}

func (c *realCM) handleRepositorySpecificLoading(
	repositoryName, remoteSource, branchName string,
	params map[string]interface{},
) error {
	worktreePath, err := c.loadWorktreeForRepository(repositoryName, remoteSource, branchName)
	if err != nil {
		return err
	}
	params["worktreePath"] = worktreePath
	return nil
}

func (c *realCM) handleGeneralLoading(remoteSource, branchName string, params map[string]interface{}) error {
	// Detect project mode (repository or workspace)
	projectType, err := c.detectProjectMode("", "")
	if err != nil {
		c.VerbosePrint("Error: %v", err)
		return fmt.Errorf("failed to detect project type: %w", err)
	}

	// Handle based on project type
	worktreePath, err := c.loadWorktreeByProjectType(projectType, remoteSource, branchName)
	if err != nil {
		return err
	}

	// Set worktreePath in params for the IDE opening hook
	params["worktreePath"] = worktreePath
	return nil
}

func (c *realCM) loadWorktreeByProjectType(projectType mode.Mode, remoteSource, branchName string) (string, error) {
	switch projectType {
	case mode.ModeSingleRepo:
		return c.loadWorktreeForSingleRepo(remoteSource, branchName)
	case mode.ModeWorkspace:
		return "", ErrWorkspaceModeNotSupported
	case mode.ModeNone:
		return "", ErrNoGitRepositoryOrWorkspaceFound
	default:
		return "", fmt.Errorf("unknown project type")
	}
}

// loadWorktreeForSingleRepo loads a worktree for single repository mode.
func (c *realCM) loadWorktreeForSingleRepo(remoteSource, branchName string) (string, error) {
	c.VerbosePrint("Loading worktree for single repository mode")

	// Create repository instance
	repoInstance := c.repositoryProvider(repo.NewRepositoryParams{
		FS:               c.fs,
		Git:              c.git,
		Config:           c.config,
		StatusManager:    c.statusManager,
		Logger:           c.logger,
		Prompt:           c.prompt,
		WorktreeProvider: worktree.NewWorktree,
		HookManager:      c.hookManager,
		RepositoryName:   ".",
	})

	worktreePath, err := repoInstance.LoadWorktree(remoteSource, branchName)
	if err != nil {
		return "", err
	}

	return worktreePath, nil
}

// parseBranchArg parses the remote:branch argument format.
func (c *realCM) parseBranchArg(arg string) (remoteSource, branchName string, err error) {
	// Validate branch argument
	if arg == "" {
		return "", "", ErrArgumentEmpty
	}

	// Split on first colon
	parts := strings.SplitN(arg, ":", 2)

	if len(parts) == 1 {
		// No colon found, treat as branch name only (default to origin)
		branchName = parts[0]
		if branchName == "" {
			return "", "", fmt.Errorf("branch name cannot be empty")
		}
		if strings.Contains(branchName, ":") {
			return "", "", ErrBranchNameContainsColon
		}
		return "", branchName, nil // empty remoteSource defaults to origin
	}

	if len(parts) == 2 {
		remoteSource = parts[0]
		branchName = parts[1]

		// Validate parts
		if remoteSource == "" {
			return "", "", fmt.Errorf("remote source cannot be empty")
		}
		if branchName == "" {
			return "", "", fmt.Errorf("branch name cannot be empty")
		}
		if strings.Contains(branchName, ":") {
			return "", "", ErrBranchNameContainsColon
		}

		return remoteSource, branchName, nil
	}

	return "", "", fmt.Errorf("invalid argument format")
}

// loadWorktreeForRepository loads a worktree for a specific repository.
func (c *realCM) loadWorktreeForRepository(repositoryName, remoteSource, branchName string) (string, error) {
	c.VerbosePrint("Loading worktree for repository: %s", repositoryName)

	// Create repository instance - let repositoryProvider handle repository name resolution
	repoInstance := c.repositoryProvider(repo.NewRepositoryParams{
		FS:               c.fs,
		Git:              c.git,
		Config:           c.config,
		StatusManager:    c.statusManager,
		Logger:           c.logger,
		Prompt:           c.prompt,
		WorktreeProvider: worktree.NewWorktree,
		HookManager:      c.hookManager,
		RepositoryName:   repositoryName, // Pass repository name directly, let provider handle resolution
	})

	// Load the worktree
	worktreePath, err := repoInstance.LoadWorktree(remoteSource, branchName)
	if err != nil {
		return "", c.translateRepositoryError(err)
	}

	return worktreePath, nil
}

// extractLoadWorktreeOptions extracts and merges options from the variadic parameter.
func (c *realCM) extractLoadWorktreeOptions(opts []LoadWorktreeOpts) LoadWorktreeOpts {
	var result LoadWorktreeOpts

	// Merge all provided options, with later options overriding earlier ones
	for _, opt := range opts {
		if opt.IDEName != "" {
			result.IDEName = opt.IDEName
		}
		if opt.RepositoryName != "" {
			result.RepositoryName = opt.RepositoryName
		}
	}

	return result
}
