package git

//go:generate go run go.uber.org/mock/mockgen@latest  -source=git.go -destination=mocks/git.gen.go -package=mocks

// Git interface provides Git command execution capabilities.
type Git interface {
	// Status executes `git status` in specified directory.
	Status(workDir string) (string, error)

	// ConfigGet executes `git config --get <key>` in specified directory.
	ConfigGet(workDir, key string) (string, error)

	// CreateWorktree creates a new worktree for the specified branch.
	CreateWorktree(repoPath, worktreePath, branch string) error

	// CreateWorktreeWithNoCheckout creates a new worktree without checking out files.
	CreateWorktreeWithNoCheckout(repoPath, worktreePath, branch string) error

	// CheckoutBranch checks out a branch in the specified worktree.
	CheckoutBranch(worktreePath, branch string) error

	// GetCurrentBranch gets the current branch name.
	GetCurrentBranch(repoPath string) (string, error)

	// GetRepositoryName gets the repository name from remote origin URL with fallback to local path.
	GetRepositoryName(repoPath string) (string, error)

	// IsClean checks if the repository is in a clean state (placeholder for future validation).
	IsClean(repoPath string) (bool, error)

	// BranchExists checks if a branch exists locally or remotely.
	BranchExists(repoPath, branch string) (bool, error)

	// CreateBranchFrom creates a new branch from a specific branch.
	CreateBranchFrom(params CreateBranchFromParams) error

	// CheckReferenceConflict checks if creating a branch would conflict with existing references.
	CheckReferenceConflict(repoPath, branch string) error

	// WorktreeExists checks if a worktree exists for the specified branch.
	WorktreeExists(repoPath, branch string) (bool, error)

	// RemoveWorktree removes a worktree from Git's tracking.
	RemoveWorktree(repoPath, worktreePath string, force bool) error

	// GetWorktreePath gets the path of a worktree for a branch.
	GetWorktreePath(repoPath, branch string) (string, error)

	// AddRemote adds a new remote to the repository.
	AddRemote(repoPath, remoteName, remoteURL string) error

	// FetchRemote fetches from a specific remote.
	FetchRemote(repoPath, remoteName string) error

	// BranchExistsOnRemote checks if a branch exists on a specific remote.
	BranchExistsOnRemote(params BranchExistsOnRemoteParams) (bool, error)

	// GetRemoteURL gets the URL of a remote.
	GetRemoteURL(repoPath, remoteName string) (string, error)

	// RemoteExists checks if a remote exists.
	RemoteExists(repoPath, remoteName string) (bool, error)

	// Clone clones a repository to the specified path.
	Clone(params CloneParams) error

	// GetDefaultBranch gets the default branch name from a remote repository.
	GetDefaultBranch(remoteURL string) (string, error)

	// Add adds files to the Git staging area.
	Add(repoPath string, files ...string) error

	// CreateBranch creates a new branch from the current branch.
	CreateBranch(repoPath, branch string) error

	// Commit creates a new commit with the specified message.
	Commit(repoPath, message string) error

	// GetBranchRemote gets the remote name for a branch (e.g., "origin", "justenstall").
	GetBranchRemote(repoPath, branch string) (string, error)

	// SetUpstreamBranch sets the upstream branch for the current branch.
	SetUpstreamBranch(repoPath, remote, branch string) error
}

type realGit struct {
	// No fields needed for basic Git operations
}

// NewGit creates a new Git instance.
func NewGit() Git {
	return &realGit{}
}
