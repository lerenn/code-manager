package cgwt

import (
	"fmt"

	"github.com/lerenn/cgwt/pkg/fs"
)

// CGWT interface provides Git repository detection functionality.
type CGWT interface {
	// Run executes the main application logic.
	Run() error
}

// OutputMode represents the different output modes.
type OutputMode int

const (
	OutputModeNormal OutputMode = iota
	OutputModeQuiet
	OutputModeVerbose
)

type cgwt struct {
	fs         fs.FS
	outputMode OutputMode
}

// NewCGWT creates a new CGWT instance.
func NewCGWT(fs fs.FS) CGWT {
	return &cgwt{
		fs:         fs,
		outputMode: OutputModeNormal,
	}
}

// NewCGWTWithMode creates a new CGWT instance with specified output mode.
func NewCGWTWithMode(fs fs.FS, mode OutputMode) CGWT {
	return &cgwt{
		fs:         fs,
		outputMode: mode,
	}
}

// Run executes the main application logic.
func (c *cgwt) Run() error {
	// First check for single repository mode
	isSingleRepo, err := c.detectSingleRepoMode()
	if err != nil {
		return fmt.Errorf("failed to detect repository mode: %w", err)
	}

	if isSingleRepo {
		c.handleSingleRepoMode()
		return nil
	}

	// If no single repo found, check for workspace mode
	workspaceFiles, err := c.detectWorkspaceMode()
	if err != nil {
		return fmt.Errorf("failed to detect workspace mode: %w", err)
	}

	if len(workspaceFiles) > 1 {
		return fmt.Errorf("%d .code-workspace files found in current directory", len(workspaceFiles))
	}

	if len(workspaceFiles) == 1 {
		return c.handleWorkspaceMode(workspaceFiles[0])
	}

	// No repository or workspace found
	c.handleNoProjectFound()
	return nil
}

// verbosePrint prints a message only in verbose mode.
func (c *cgwt) verbosePrint(message string) {
	if c.outputMode == OutputModeVerbose {
		fmt.Printf("[VERBOSE] %s\n", message)
	}
}

// detectSingleRepoMode checks if the current directory is a single Git repository.
func (c *cgwt) detectSingleRepoMode() (bool, error) {
	c.verbosePrint("Checking for .git directory...")

	// Check if .git exists
	exists, err := c.fs.Exists(".git")
	if err != nil {
		return false, fmt.Errorf("failed to check .git existence: %w", err)
	}

	if !exists {
		c.verbosePrint("No .git directory found")
		return false, nil
	}

	c.verbosePrint("Verifying .git is a directory...")

	// Check if .git is a directory
	isDir, err := c.fs.IsDir(".git")
	if err != nil {
		return false, fmt.Errorf("failed to check .git directory: %w", err)
	}

	if !isDir {
		c.verbosePrint(".git exists but is not a directory")
		return false, nil
	}

	c.verbosePrint("Git repository detected")
	return true, nil
}

// handleSingleRepoMode handles the output for single repository mode.
func (c *cgwt) handleSingleRepoMode() {
	if c.outputMode != OutputModeQuiet {
		fmt.Println("Single repository mode detected")
	}
}

// handleWorkspaceMode handles the output for workspace mode.
func (c *cgwt) handleWorkspaceMode(workspaceFile string) error {
	workspaceConfig, err := c.getWorkspaceInfo(workspaceFile)
	if err != nil {
		return fmt.Errorf("failed to get workspace info: %w", err)
	}

	workspaceName := c.getWorkspaceName(workspaceConfig, workspaceFile)
	if c.outputMode != OutputModeQuiet {
		fmt.Printf("Found workspace: %s\n", workspaceName)
	}
	if c.outputMode == OutputModeVerbose {
		c.verbosePrint("Workspace configuration:")
		c.verbosePrint(fmt.Sprintf("  Folders: %d", len(workspaceConfig.Folders)))
		for _, folder := range workspaceConfig.Folders {
			c.verbosePrint(fmt.Sprintf("    - %s: %s", folder.Name, folder.Path))
		}
	}
	return nil
}

// handleNoProjectFound handles the output when no project is found.
func (c *cgwt) handleNoProjectFound() {
	if c.outputMode != OutputModeQuiet {
		fmt.Println("No Git repository or workspace found")
	}
}
