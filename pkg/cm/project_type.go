package cm

// ProjectType represents the type of project detected.
type ProjectType int

// Project type constants.
const (
	ProjectTypeNone ProjectType = iota
	ProjectTypeSingleRepo
	ProjectTypeWorkspace
)
