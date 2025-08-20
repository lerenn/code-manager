package cm

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/cm/pkg/config"
)

// IsInitialized checks if CM is initialized.
func (c *realCM) IsInitialized() (bool, error) {
	return c.StatusManager.IsInitialized()
}

// InitOpts contains optional parameters for Init.
type InitOpts struct {
	Force    bool
	Reset    bool
	BasePath string
}

// Init initializes CM configuration.
func (c *realCM) Init(opts InitOpts) error {
	c.VerbosePrint("Starting CM initialization")

	// Check if already initialized (unless reset is requested)
	if err := c.checkInitializationStatus(opts.Reset); err != nil {
		return err
	}

	// Handle reset flag
	if opts.Reset {
		if err := c.handleReset(opts.Force); err != nil {
			return err
		}
	}

	// Get and validate base path
	expandedBasePath, err := c.getAndValidateBasePath(opts.BasePath)
	if err != nil {
		return err
	}

	// Create base path directory if it doesn't exist
	if err := c.FS.CreateDirectory(expandedBasePath, 0755); err != nil {
		return fmt.Errorf("failed to create base path directory: %w", err)
	}

	// Update and save configuration
	if err := c.updateConfiguration(expandedBasePath); err != nil {
		return err
	}

	// Initialize status
	if err := c.initializeStatus(); err != nil {
		return err
	}

	c.VerbosePrint("CM initialization completed successfully")
	fmt.Printf("CM initialized successfully!\n")
	fmt.Printf("Base path: %s\n", expandedBasePath)
	fmt.Printf("Configuration: %s\n", c.getConfigPath())
	fmt.Printf("Status file: %s\n", c.Config.StatusFile)

	return nil
}

// checkInitializationStatus checks if CM is already initialized.
func (c *realCM) checkInitializationStatus(reset bool) error {
	if reset {
		return nil
	}

	initialized, err := c.IsInitialized()
	if err != nil {
		return fmt.Errorf("failed to check initialization status: %w", err)
	}
	if initialized {
		return fmt.Errorf("%w. Use --reset to clear existing configuration and start fresh", ErrAlreadyInitialized)
	}

	return nil
}

// getAndValidateBasePath gets and validates the base path.
func (c *realCM) getAndValidateBasePath(flagBasePath string) (string, error) {
	basePath, err := c.getBasePath(flagBasePath)
	if err != nil {
		return "", err
	}

	// Validate and expand base path
	expandedBasePath, err := c.FS.ExpandPath(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to expand base path: %w", err)
	}

	// Validate base path
	configManager := config.NewManager()
	if err := configManager.ValidateBasePath(expandedBasePath); err != nil {
		return "", fmt.Errorf("invalid base path: %w", err)
	}

	return expandedBasePath, nil
}

// updateConfiguration updates and saves the configuration.
func (c *realCM) updateConfiguration(expandedBasePath string) error {
	newConfig := &config.Config{
		BasePath:   expandedBasePath,
		StatusFile: c.Config.StatusFile, // Keep existing status file path
	}

	configPath := c.getConfigPath()
	configManager := config.NewManager()

	if err := configManager.SaveConfig(newConfig, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// initializeStatus creates initial status and sets initialized flag.
func (c *realCM) initializeStatus() error {
	// Create initial status file
	if err := c.StatusManager.CreateInitialStatus(); err != nil {
		return fmt.Errorf("failed to create initial status: %w", err)
	}

	// Set initialized flag to true
	if err := c.StatusManager.SetInitialized(true); err != nil {
		return fmt.Errorf("failed to set initialization status: %w", err)
	}

	return nil
}

// getConfigPath returns the config file path.
func (c *realCM) getConfigPath() string {
	homeDir, err := c.FS.GetHomeDir()
	if err != nil {
		// Fallback to default path if home directory cannot be determined
		return filepath.Join("~", ".cm", "config.yaml")
	}
	return filepath.Join(homeDir, ".cm", "config.yaml")
}

// handleReset handles the reset functionality.
func (c *realCM) handleReset(force bool) error {
	if !force {
		confirmed, err := c.Prompt.PromptForConfirmation(
			"This will reset your CM configuration and remove all existing worktrees. Are you sure?", false)
		if err != nil {
			return fmt.Errorf("failed to get user confirmation: %w", err)
		}
		if !confirmed {
			return fmt.Errorf("initialization cancelled by user")
		}
	}

	c.VerbosePrint("Resetting CM configuration")

	// Clear status file
	if err := c.StatusManager.CreateInitialStatus(); err != nil {
		return fmt.Errorf("failed to reset status: %w", err)
	}

	return nil
}

// getBasePath gets the base path from user input, flag, or default.
func (c *realCM) getBasePath(flagBasePath string) (string, error) {
	if flagBasePath != "" {
		return flagBasePath, nil
	}

	// Interactive prompt
	return c.Prompt.PromptForBasePath()
}
