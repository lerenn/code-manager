package workspace

// Validate validates all repositories in a workspace.
func (w *realWorkspace) Validate() error {
	w.deps.Logger.Logf("Validating workspace: %s", w.file)

	// Use the new workspace validation logic that ensures repositories are in status
	// and have default branch worktrees
	return w.ValidateWorkspaceReferences()
}
