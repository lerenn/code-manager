package status

// CreateInitialStatus creates the initial status file structure.
func (s *realManager) CreateInitialStatus() error {
	initialStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	return s.saveStatus(initialStatus)
}
