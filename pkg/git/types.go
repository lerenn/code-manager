package git

// BranchExistsOnRemoteParams contains parameters for BranchExistsOnRemote.
type BranchExistsOnRemoteParams struct {
	RepoPath   string
	RemoteName string
	Branch     string
}

// CreateBranchFromParams contains parameters for CreateBranchFrom.
type CreateBranchFromParams struct {
	RepoPath   string
	NewBranch  string
	FromBranch string
}

// CloneParams contains parameters for Clone.
type CloneParams struct {
	RepoURL    string
	TargetPath string
	Recursive  bool
}
