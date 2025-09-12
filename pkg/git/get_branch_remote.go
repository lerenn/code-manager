package git

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
