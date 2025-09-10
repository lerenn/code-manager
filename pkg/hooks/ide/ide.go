package ide

//go:generate mockgen -source=ide.go -destination=mocks/ide.gen.go -package=mocks

// DefaultIDE is the default IDE name used when no IDE is specified.
const DefaultIDE = VSCodeName

// IDE interface defines the methods that all IDE implementations must provide.
type IDE interface {
	// Name returns the name of the IDE
	Name() string

	// IsInstalled checks if the IDE is installed on the system
	IsInstalled() bool

	// OpenRepository opens the IDE with the specified repository path
	OpenRepository(path string) error
}
