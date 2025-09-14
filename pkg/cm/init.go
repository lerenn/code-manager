package cm

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/config"
)

// InitOpts contains optional parameters for Init.
type InitOpts struct {
	Force           bool
	Reset           bool
	RepositoriesDir string
	WorkspacesDir   string
	StatusFile      string
	NonInteractive  bool
}

// Init initializes CM configuration.
func (c *realCM) Init(opts InitOpts) error {
	// Prepare parameters for hooks
	params := map[string]interface{}{
		"reset":           opts.Reset,
		"force":           opts.Force,
		"repositoriesDir": opts.RepositoriesDir,
		"workspacesDir":   opts.WorkspacesDir,
		"statusFile":      opts.StatusFile,
		"nonInteractive":  opts.NonInteractive,
	}

	// Execute with hooks
	return c.executeWithHooks(consts.Init, params, func() error {
		return c.performInitialization(opts)
	})
}

// performInitialization performs the actual initialization logic.
func (c *realCM) performInitialization(opts InitOpts) error {
	c.VerbosePrint("Starting CM initialization")

	// Handle reset flag
	if opts.Reset {
		if err := c.handleReset(opts.Force); err != nil {
			return err
		}
	}

	// Get and validate directories and status file
	expandedRepositoriesDir, expandedWorkspacesDir, expandedStatusFile, err := c.setupDirectories(opts)
	if err != nil {
		return err
	}

	// Create directories if they don't exist
	if err := c.createDirectories(expandedRepositoriesDir, expandedWorkspacesDir); err != nil {
		return err
	}

	// Update and save configuration
	if err := c.updateConfiguration(expandedRepositoriesDir, expandedWorkspacesDir, expandedStatusFile); err != nil {
		return err
	}

	// Ensure status exists
	if err := c.ensureStatusExists(); err != nil {
		return err
	}

	// Print success message
	c.printInitializationSuccess(expandedRepositoriesDir, expandedWorkspacesDir)
	return nil
}

// setupDirectories gets and validates repositories, workspaces directories, and status file.
func (c *realCM) setupDirectories(opts InitOpts) (string, string, string, error) {
	// Get and validate repositories directory
	expandedRepositoriesDir, err := c.getAndValidateRepositoriesDir(opts.RepositoriesDir, opts.NonInteractive)
	if err != nil {
		return "", "", "", err
	}

	// Get and validate workspaces directory
	expandedWorkspacesDir, err := c.getAndValidateWorkspacesDir(
		opts.WorkspacesDir,
		expandedRepositoriesDir,
		opts.NonInteractive)
	if err != nil {
		return "", "", "", err
	}

	// Get and validate status file
	expandedStatusFile, err := c.getAndValidateStatusFile(opts.StatusFile, opts.NonInteractive)
	if err != nil {
		return "", "", "", err
	}

	return expandedRepositoriesDir, expandedWorkspacesDir, expandedStatusFile, nil
}

// createDirectories creates the repositories and workspaces directories.
func (c *realCM) createDirectories(expandedRepositoriesDir, expandedWorkspacesDir string) error {
	if err := c.fs.CreateDirectory(expandedRepositoriesDir, 0755); err != nil {
		return fmt.Errorf("failed to create repositories directory: %w", err)
	}
	if err := c.fs.CreateDirectory(expandedWorkspacesDir, 0755); err != nil {
		return fmt.Errorf("failed to create workspaces directory: %w", err)
	}
	return nil
}

// ensureStatusExists ensures the status file exists, creating it if necessary.
func (c *realCM) ensureStatusExists() error {
	exists, err := c.fs.Exists(c.config.StatusFile)
	if err != nil {
		return fmt.Errorf("failed to check status existence: %w", err)
	}
	if !exists {
		if err := c.statusManager.CreateInitialStatus(); err != nil {
			return fmt.Errorf("failed to create initial status: %w", err)
		}
	}
	return nil
}

// printInitializationSuccess prints the success message and configuration details.
func (c *realCM) printInitializationSuccess(expandedRepositoriesDir, expandedWorkspacesDir string) {
	c.VerbosePrint("CM initialization completed successfully")
	fmt.Printf("CM initialized successfully!\n")
	fmt.Printf("Repositories directory: %s\n", expandedRepositoriesDir)
	fmt.Printf("Workspaces directory: %s\n", expandedWorkspacesDir)
	fmt.Printf("Configuration: %s\n", c.getConfigPath())
	fmt.Printf("Status file: %s\n", c.config.StatusFile)
}

// getAndValidateRepositoriesDir gets and validates the repositories directory.
func (c *realCM) getAndValidateRepositoriesDir(flagRepositoriesDir string, nonInteractive bool) (string, error) {
	repositoriesDir, err := c.getRepositoriesDir(flagRepositoriesDir, nonInteractive)
	if err != nil {
		return "", err
	}

	// Validate and expand repositories directory
	expandedRepositoriesDir, err := c.fs.ExpandPath(repositoriesDir)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrFailedToExpandRepositoriesDir, err)
	}

	// Validate repositories directory
	configManager := config.NewManager()
	if err := configManager.ValidateRepositoriesDir(expandedRepositoriesDir); err != nil {
		return "", fmt.Errorf("invalid repositories directory: %w", err)
	}

	return expandedRepositoriesDir, nil
}

// getAndValidateWorkspacesDir gets and validates the workspaces directory.
func (c *realCM) getAndValidateWorkspacesDir(
	flagWorkspacesDir, expandedRepositoriesDir string,
	nonInteractive bool,
) (string, error) {
	workspacesDir, err := c.getWorkspacesDir(flagWorkspacesDir, expandedRepositoriesDir, nonInteractive)
	if err != nil {
		return "", err
	}

	// Validate and expand workspaces directory
	expandedWorkspacesDir, err := c.fs.ExpandPath(workspacesDir)
	if err != nil {
		return "", fmt.Errorf("failed to expand workspaces directory: %w", err)
	}

	// Validate workspaces directory
	configManager := config.NewManager()
	if err := configManager.ValidateWorkspacesDir(expandedWorkspacesDir); err != nil {
		return "", fmt.Errorf("invalid workspaces directory: %w", err)
	}

	return expandedWorkspacesDir, nil
}

// getWorkspacesDir gets the workspaces directory from flag, prompt, or default.
func (c *realCM) getWorkspacesDir(
	flagWorkspacesDir, expandedRepositoriesDir string,
	nonInteractive bool,
) (string, error) {
	if flagWorkspacesDir != "" {
		return flagWorkspacesDir, nil
	}

	if nonInteractive {
		// Use default workspaces directory instead of prompting
		return filepath.Join(filepath.Dir(expandedRepositoriesDir), "workspaces"), nil
	}

	// Interactive prompt
	defaultWorkspacesDir := filepath.Join(filepath.Dir(expandedRepositoriesDir), "workspaces")
	return c.prompt.PromptForWorkspacesDir(defaultWorkspacesDir)
}

// getAndValidateStatusFile gets and validates the status file path.
func (c *realCM) getAndValidateStatusFile(flagStatusFile string, nonInteractive bool) (string, error) {
	statusFile, err := c.getStatusFile(flagStatusFile, nonInteractive)
	if err != nil {
		return "", err
	}

	// Validate and expand status file path
	expandedStatusFile, err := c.fs.ExpandPath(statusFile)
	if err != nil {
		return "", fmt.Errorf("failed to expand status file path: %w", err)
	}

	// Validate status file path
	configManager := config.NewManager()
	if err := configManager.ValidateStatusFile(expandedStatusFile); err != nil {
		return "", fmt.Errorf("invalid status file path: %w", err)
	}

	return expandedStatusFile, nil
}

// getStatusFile gets the status file path from flag, prompt, or default.
func (c *realCM) getStatusFile(flagStatusFile string, nonInteractive bool) (string, error) {
	if flagStatusFile != "" {
		return flagStatusFile, nil
	}

	if nonInteractive {
		// Use default status file path instead of prompting
		return c.config.StatusFile, nil
	}

	// Interactive prompt
	return c.prompt.PromptForStatusFile(c.config.StatusFile)
}

// updateConfiguration updates and saves the configuration.
func (c *realCM) updateConfiguration(expandedRepositoriesDir, expandedWorkspacesDir, expandedStatusFile string) error {
	newConfig := config.Config{
		RepositoriesDir: expandedRepositoriesDir,
		WorkspacesDir:   expandedWorkspacesDir,
		StatusFile:      expandedStatusFile,
	}

	configPath := c.getConfigPath()
	configManager := config.NewManager()

	if err := configManager.SaveConfig(newConfig, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// getConfigPath returns the config file path.
func (c *realCM) getConfigPath() string {
	// If a custom config path was provided, use it
	if c.configPath != "" {
		return c.configPath
	}

	// Otherwise, use the default global path
	homeDir, err := c.fs.GetHomeDir()
	if err != nil {
		// Fallback to default path if home directory cannot be determined
		return filepath.Join("~", ".cm", "config.yaml")
	}
	return filepath.Join(homeDir, ".cm", "config.yaml")
}

// handleReset handles the reset functionality.
func (c *realCM) handleReset(force bool) error {
	if !force {
		confirmed, err := c.prompt.PromptForConfirmation(
			"This will reset your CM configuration and remove all existing worktrees. Are you sure?", false)
		if err != nil {
			return fmt.Errorf("failed to get user confirmation: %w", err)
		}
		if !confirmed {
			return fmt.Errorf("initialization cancelled by user")
		}
	}

	c.VerbosePrint("Resetting CM configuration")

	// Clear status file by recreating empty structure
	if err := c.statusManager.CreateInitialStatus(); err != nil {
		return fmt.Errorf("failed to reset status: %w", err)
	}

	return nil
}

// getRepositoriesDir gets the repositories directory from user input, flag, or default.
func (c *realCM) getRepositoriesDir(flagRepositoriesDir string, nonInteractive bool) (string, error) {
	if flagRepositoriesDir != "" {
		return flagRepositoriesDir, nil
	}

	if nonInteractive {
		// Use default repositories directory instead of prompting
		return c.config.RepositoriesDir, nil
	}

	// Interactive prompt
	return c.prompt.PromptForRepositoriesDir(c.config.RepositoriesDir)
}
