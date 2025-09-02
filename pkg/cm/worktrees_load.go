package cm

import (
	"fmt"
	"strings"

	"github.com/lerenn/code-manager/pkg/cm/consts"
)

// LoadWorktreeOpts contains optional parameters for LoadWorktree.
type LoadWorktreeOpts struct {
	IDEName string
}

// LoadWorktree loads a branch from a remote source and creates a worktree.
func (c *realCM) LoadWorktree(branchArg string, opts ...LoadWorktreeOpts) error {
	// Prepare parameters for hooks
	params := map[string]interface{}{
		"branchArg": branchArg,
	}
	if len(opts) > 0 && opts[0].IDEName != "" {
		params["ideName"] = opts[0].IDEName
	}

	// Execute with hooks
	return c.executeWithHooks(consts.LoadWorktree, params, func() error {
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
			return fmt.Errorf("failed to detect project type: %w", err)
		}

		// 3. Handle based on project type
		var worktreePath string
		switch projectType {
		case ProjectTypeSingleRepo:
			worktreePath, err = c.loadWorktreeForSingleRepo(remoteSource, branchName)
		case ProjectTypeWorkspace:
			return ErrWorkspaceModeNotSupported
		case ProjectTypeNone:
			return ErrNoGitRepositoryOrWorkspaceFound
		default:
			return fmt.Errorf("unknown project type")
		}

		if err != nil {
			return err
		}

		// Set worktreePath in params for the IDE opening hook
		params["worktreePath"] = worktreePath

		return nil
	})
}

// loadWorktreeForSingleRepo loads a worktree for single repository mode.
func (c *realCM) loadWorktreeForSingleRepo(remoteSource, branchName string) (string, error) {
	c.VerbosePrint("Loading worktree for single repository mode")

	worktreePath, err := c.repository.LoadWorktree(remoteSource, branchName)
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
