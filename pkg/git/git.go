package git

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

const defaultRemote = "origin"

//go:generate go run go.uber.org/mock/mockgen@v0.5.2 -source=git.go -destination=mockgit.gen.go -package=git

// Git interface provides Git command execution capabilities.
type Git interface {
	// Status executes `git status` in specified directory.
	Status(workDir string) (string, error)

	// ConfigGet executes `git config --get <key>` in specified directory.
	ConfigGet(workDir, key string) (string, error)

	// CreateWorktree creates a new worktree for the specified branch.
	CreateWorktree(repoPath, worktreePath, branch string) error

	// GetCurrentBranch gets the current branch name.
	GetCurrentBranch(repoPath string) (string, error)

	// GetRepositoryName gets the repository name from remote origin URL with fallback to local path.
	GetRepositoryName(repoPath string) (string, error)

	// IsClean checks if the repository is in a clean state (placeholder for future validation).
	IsClean(repoPath string) (bool, error)

	// BranchExists checks if a branch exists locally or remotely.
	BranchExists(repoPath, branch string) (bool, error)

	// CreateBranch creates a new branch from the current branch.
	CreateBranch(repoPath, branch string) error

	// WorktreeExists checks if a worktree exists for the specified branch.
	WorktreeExists(repoPath, branch string) (bool, error)

	// RemoveWorktree removes a worktree from Git's tracking.
	RemoveWorktree(repoPath, worktreePath string) error

	// GetWorktreePath gets the path of a worktree for a branch.
	GetWorktreePath(repoPath, branch string) (string, error)

	// AddRemote adds a new remote to the repository.
	AddRemote(repoPath, remoteName, remoteURL string) error

	// FetchRemote fetches from a specific remote.
	FetchRemote(repoPath, remoteName string) error

	// BranchExistsOnRemote checks if a branch exists on a specific remote.
	BranchExistsOnRemote(repoPath, remoteName, branch string) (bool, error)

	// GetRemoteURL gets the URL of a remote.
	GetRemoteURL(repoPath, remoteName string) (string, error)

	// RemoteExists checks if a remote exists.
	RemoteExists(repoPath, remoteName string) (bool, error)

	// GetBranchRemote gets the remote name for a branch (e.g., "origin", "justenstall").
	GetBranchRemote(repoPath, branch string) (string, error)
}

type realGit struct {
	// No fields needed for basic Git operations
}

// NewGit creates a new Git instance.
func NewGit() Git {
	return &realGit{}
}

// Status executes `git status` in specified directory.
func (g *realGit) Status(workDir string) (string, error) {
	cmd := exec.Command("git", "status")
	cmd.Dir = workDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git command failed: %w (command: git status, output: %s)", err, string(output))
	}

	return string(output), nil
}

// ConfigGet executes `git config --get <key>` in specified directory.
func (g *realGit) ConfigGet(workDir, key string) (string, error) {
	cmd := exec.Command("git", "config", "--get", key)
	cmd.Dir = workDir

	output, err := cmd.Output()
	if err != nil {
		// Return empty string for missing config keys (exit code 1)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", nil
		}
		return "", fmt.Errorf("git command failed: %w (command: git config --get %s, output: %s)", err, key, string(output))
	}

	return string(output), nil
}

// CreateWorktree creates a new worktree for the specified branch.
func (g *realGit) CreateWorktree(repoPath, worktreePath, branch string) error {
	cmd := exec.Command("git", "worktree", "add", worktreePath, branch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add failed: %w (command: git worktree add %s %s, output: %s)",
			err, worktreePath, branch, string(output))
	}
	return nil
}

// GetCurrentBranch gets the current branch name.
func (g *realGit) GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git branch --show-current failed: %w (command: git branch --show-current, output: %s)",
			err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

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
		// Handle different URL formats: https://github.com/user/repo.git, git@github.com:user/repo.git
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

// extractRepoNameFromURL extracts the repository name from a Git remote URL.
func (g *realGit) extractRepoNameFromURL(url string) string {
	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Handle SSH format: git@github.com:user/repo
	if strings.Contains(url, "@") && strings.Contains(url, ":") {
		parts := strings.Split(url, ":")
		if len(parts) == 2 {
			hostParts := strings.Split(parts[0], "@")
			if len(hostParts) == 2 {
				return hostParts[1] + "/" + parts[1]
			}
		}
	}

	// Handle HTTPS format: https://github.com/user/repo
	if strings.HasPrefix(url, "http") {
		return g.extractHTTPSRepoName(url)
	}

	return ""
}

// extractHTTPSRepoName extracts repository name from HTTPS URLs.
func (g *realGit) extractHTTPSRepoName(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) < 3 {
		return ""
	}

	// Extract host and path: github.com/user/repo
	host := parts[2] // github.com
	if len(parts) < 4 {
		return host
	}

	user := parts[3] // user
	if len(parts) < 5 {
		return host + "/" + user
	}

	repo := parts[4] // repo
	return host + "/" + user + "/" + repo
}

// IsClean checks if the repository is in a clean state (placeholder for future validation).
func (g *realGit) IsClean(_ string) (bool, error) {
	// TODO: Implement actual clean state check
	// For now, always return true as placeholder
	return true, nil
}

// BranchExists checks if a branch exists locally or remotely.
func (g *realGit) BranchExists(repoPath, branch string) (bool, error) {
	// Check local branches
	cmd := exec.Command("git", "branch", "--list", branch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("git branch --list failed: %w (command: git branch --list %s, output: %s)",
			err, branch, string(output))
	}
	if strings.TrimSpace(string(output)) != "" {
		return true, nil
	}

	// Check remote branches
	cmd = exec.Command("git", "branch", "-r", "--list", "origin/"+branch)
	cmd.Dir = repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("git branch -r --list failed: %w (command: git branch -r --list origin/%s, output: %s)",
			err, branch, string(output))
	}
	return strings.TrimSpace(string(output)) != "", nil
}

// CreateBranch creates a new branch from the current branch.
func (g *realGit) CreateBranch(repoPath, branch string) error {
	cmd := exec.Command("git", "branch", branch)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git branch failed: %w (command: git branch %s, output: %s)", err, branch, string(output))
	}

	return nil
}

// WorktreeExists checks if a worktree exists for the specified branch.
func (g *realGit) WorktreeExists(repoPath, branch string) (bool, error) {
	cmd := exec.Command("git", "worktree", "list")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git worktree list failed: %w (command: git worktree list, output: %s)", err, string(output))
	}

	// Check if the branch is mentioned in the worktree list
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, branch) {
			return true, nil
		}
	}

	return false, nil
}

// RemoveWorktree removes a worktree from Git's tracking.
func (g *realGit) RemoveWorktree(repoPath, worktreePath string) error {
	cmd := exec.Command("git", "worktree", "remove", worktreePath)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree remove failed: %w (command: git worktree remove %s, output: %s)",
			err, worktreePath, string(output))
	}
	return nil
}

// GetWorktreePath gets the path of a worktree for a branch.
func (g *realGit) GetWorktreePath(repoPath, branch string) (string, error) {
	cmd := exec.Command("git", "worktree", "list")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git worktree list failed: %w (command: git worktree list, output: %s)", err, string(output))
	}

	// Parse the worktree list output to find the path for the specified branch
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse worktree list format: "worktree-path [branch-name]"
		// Example: "/path/to/worktree [feature/branch]"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			worktreePath := parts[0]
			// Check if the branch name matches (it's in brackets)
			branchPart := strings.Join(parts[1:], " ")
			if strings.Contains(branchPart, branch) {
				return worktreePath, nil
			}
		}
	}

	return "", fmt.Errorf("worktree path not found for branch %s", branch)
}

// AddRemote adds a new remote to the repository.
func (g *realGit) AddRemote(repoPath, remoteName, remoteURL string) error {
	cmd := exec.Command("git", "remote", "add", remoteName, remoteURL)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git remote add failed: %w (command: git remote add %s %s, output: %s)",
			err, remoteName, remoteURL, string(output))
	}
	return nil
}

// FetchRemote fetches from a specific remote.
func (g *realGit) FetchRemote(repoPath, remoteName string) error {
	cmd := exec.Command("git", "fetch", remoteName)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch failed: %w (command: git fetch %s, output: %s)",
			err, remoteName, string(output))
	}
	return nil
}

// BranchExistsOnRemote checks if a branch exists on a specific remote.
func (g *realGit) BranchExistsOnRemote(repoPath, remoteName, branch string) (bool, error) {
	cmd := exec.Command("git", "ls-remote", "--heads", remoteName, branch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("git ls-remote failed: %w (command: git ls-remote --heads %s %s, output: %s)",
			err, remoteName, branch, string(output))
	}
	return strings.TrimSpace(string(output)) != "", nil
}

// GetRemoteURL gets the URL of a remote.
func (g *realGit) GetRemoteURL(repoPath, remoteName string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", remoteName)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git remote get-url failed: %w (command: git remote get-url %s, output: %s)",
			err, remoteName, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

// RemoteExists checks if a remote exists.
func (g *realGit) RemoteExists(repoPath, remoteName string) (bool, error) {
	cmd := exec.Command("git", "remote")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("git remote failed: %w (command: git remote, output: %s)",
			err, string(output))
	}

	// Check if the remote name exists in the list
	remotes := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, remote := range remotes {
		if strings.TrimSpace(remote) == remoteName {
			return true, nil
		}
	}
	return false, nil
}

// GetBranchRemote gets the remote name for a branch (e.g., "origin", "justenstall").
func (g *realGit) GetBranchRemote(repoPath, branch string) (string, error) {
	// First, try to get the upstream branch information
	remote, err := g.getUpstreamRemote(repoPath, branch)
	if err == nil {
		return remote, nil
	}

	// If the branch doesn't have an upstream, try to find which remote has this branch
	return g.findRemoteFromBranchList(repoPath, branch)
}

// getUpstreamRemote tries to get the remote from the branch's upstream configuration.
func (g *realGit) getUpstreamRemote(repoPath, branch string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", branch+"@{upstream}")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("no upstream configured: %w", err)
	}

	// Parse the upstream branch name to extract remote
	// Format: "refs/remotes/remote/branch"
	upstream := strings.TrimSpace(string(output))
	parts := strings.Split(upstream, "/")
	if len(parts) >= 3 && parts[1] == "remotes" {
		return parts[2], nil
	}

	return defaultRemote, nil
}

// findRemoteFromBranchList searches through remote branches to find which remote has the specified branch.
func (g *realGit) findRemoteFromBranchList(repoPath, branch string) (string, error) {
	cmd := exec.Command("git", "branch", "-r")
	cmd.Dir = repoPath
	remoteOutput, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git branch -r failed: %w (command: git branch -r, output: %s)",
			err, string(remoteOutput))
	}

	// Parse remote branches to find which remote has this branch
	lines := strings.Split(string(remoteOutput), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Remote branch format: "remote/branch"
		if strings.HasSuffix(line, "/"+branch) {
			parts := strings.SplitN(line, "/", 2)
			if len(parts) == 2 {
				return parts[0], nil
			}
		}
	}

	// If we can't find the remote, return "origin" as default
	return defaultRemote, nil
}
