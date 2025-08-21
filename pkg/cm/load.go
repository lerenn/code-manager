package cm

import (
	"fmt"
	"strings"

	repo "github.com/lerenn/code-manager/pkg/repository"
)

// LoadWorktreeOpts contains optional parameters for LoadWorktree.
type LoadWorktreeOpts struct {
	IDEName string
}

// LoadWorktree loads a branch from a remote source and creates a worktree.
func (c *realCM) LoadWorktree(branchArg string, opts ...LoadWorktreeOpts) error {
	c.VerbosePrint("Starting branch loading: %s", branchArg)

	// 1. Parse the branch argument to extract remote and branch name
	remoteSource, branchName, err := c.parseBranchArg(branchArg)
	if err != nil {
		return err
	}

	c.VerbosePrint("Parsed: remote=%s, branch=%s", remoteSource, branchName)

	// 2. Detect project mode (repository or workspace)
	projectType, err := c.detectProjectMode()
	if err != nil {
		c.VerbosePrint("Error: %v", err)
		return fmt.Errorf("failed to detect project mode: %w", err)
	}

	// 3. Handle based on project type
	var loadErr error
	switch projectType {
	case ProjectTypeSingleRepo:
		loadErr = c.loadWorktreeForSingleRepo(remoteSource, branchName)
	case ProjectTypeWorkspace:
		return ErrWorkspaceModeNotSupported
	case ProjectTypeNone:
		return ErrNoGitRepositoryOrWorkspaceFound
	default:
		return fmt.Errorf("unknown project type")
	}

	// 4. Open IDE if specified and branch loading was successful
	var ideName *string
	if len(opts) > 0 && opts[0].IDEName != "" {
		ideName = &opts[0].IDEName
	}
	if err := c.handleIDEOpening(loadErr, branchName, ideName); err != nil {
		return err
	}

	return loadErr
}

// loadWorktreeForSingleRepo loads a worktree for single repository mode.
func (c *realCM) loadWorktreeForSingleRepo(remoteSource, branchName string) error {
	c.VerbosePrint("Loading worktree for single repository mode")

	repoInstance := repo.NewRepository(repo.NewRepositoryParams{
		FS:            c.FS,
		Git:           c.Git,
		Config:        c.Config,
		StatusManager: c.StatusManager,
		Logger:        c.Logger,
		Prompt:        c.Prompt,
		Verbose:       c.IsVerbose(),
	})
	return repoInstance.LoadWorktree(remoteSource, branchName)
}

// parseBranchArg parses the remote:branch argument format.
func (c *realCM) parseBranchArg(arg string) (remoteSource, branchName string, err error) {
	// Check for edge cases
	if arg == "" {
		return "", "", fmt.Errorf("argument cannot be empty")
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
			return "", "", fmt.Errorf("branch name contains invalid character ':'")
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
			return "", "", fmt.Errorf("branch name contains invalid character ':'")
		}

		return remoteSource, branchName, nil
	}

	return "", "", fmt.Errorf("invalid argument format")
}
