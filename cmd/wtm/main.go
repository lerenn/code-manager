// Package main provides the command-line interface for the WTM application.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/status"
	"github.com/lerenn/wtm/pkg/wtm"
	"github.com/spf13/cobra"
)

var (
	quiet      bool
	verbose    bool
	configPath string
	ideName    string
	fromIssue  string
)

// loadConfig loads the configuration with fallback to default.
func loadConfig() *config.Config {
	var cfg *config.Config
	var err error

	if configPath != "" {
		// Use custom config path if provided
		manager := config.NewManager()
		cfg, err = manager.LoadConfig(configPath)
		if err != nil {
			log.Printf("Failed to load custom config from %s: %v", configPath, err)
			// Fall back to default config
			cfg = manager.DefaultConfig()
		}
	} else {
		// Use default config loading logic
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory if home directory cannot be determined
			homeDir = "."
		}

		defaultConfigPath := filepath.Join(homeDir, ".wtm", "config.yaml")
		cfg, err = config.LoadConfigWithFallback(defaultConfigPath)
		if err != nil {
			// If there's an error, use default config
			cfg = config.NewManager().DefaultConfig()
		}
	}

	return cfg
}

func createCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create [branch]",
		Short: "Create worktree(s) for the specified branch",
		Long:  `Create worktree(s) for the specified branch. Currently supports single repository mode.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var branchName string
			if len(args) > 0 {
				branchName = args[0]
			}

			cfg := loadConfig()
			wtmManager := wtm.NewWTM(cfg)
			wtmManager.SetVerbose(verbose)

			// Create worktree with options
			opts := wtm.CreateWorkTreeOpts{}

			// Set IDE name if specified
			if ideName != "" {
				opts.IDEName = ideName
			}

			// Set issue reference if specified
			if fromIssue != "" {
				opts.IssueRef = fromIssue
			}

			// If no branch name provided and no issue reference, return error
			if branchName == "" && fromIssue == "" {
				return fmt.Errorf("branch name is required when not using --from-issue")
			}

			// Create worktree with options
			return wtmManager.CreateWorkTree(branchName, opts)
		},
	}

	// Add --from-issue flag
	createCmd.Flags().StringVar(&fromIssue, "from-issue", "",
		"Create worktree from forge issue (GitHub issue URL, issue number, or owner/repo#issue format)")

	return createCmd
}

func createOpenCmd() *cobra.Command {
	openCmd := &cobra.Command{
		Use:   "open [worktree-name]",
		Short: "Open existing worktree in IDE",
		Long:  "Open an existing worktree in the specified IDE",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			worktreeName := args[0]

			if ideName == "" {
				return fmt.Errorf("IDE name is required. Use -i or --ide flag")
			}

			// Load configuration and create WTM instance
			cfg := loadConfig()
			wtmManager := wtm.NewWTM(cfg)
			wtmManager.SetVerbose(verbose)

			// Open worktree in IDE
			return wtmManager.OpenWorktree(worktreeName, ideName)
		},
	}

	// Add IDE flag to open command
	openCmd.Flags().StringVarP(&ideName, "ide", "i", "", "IDE to open worktree in")
	err := openCmd.MarkFlagRequired("ide")
	if err != nil {
		log.Fatal(err)
	}

	return openCmd
}

func createDeleteCmd() *cobra.Command {
	var force bool

	deleteCmd := &cobra.Command{
		Use:   "delete [branch-name]",
		Short: "Delete a worktree",
		Long:  "Delete a worktree and clean up Git state",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			branch := args[0]

			// Load configuration and create WTM instance
			cfg := loadConfig()
			wtmManager := wtm.NewWTM(cfg)
			wtmManager.SetVerbose(verbose)

			// Delete worktree
			return wtmManager.DeleteWorkTree(branch, force)
		},
	}

	// Add force flag to delete command
	deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "Force deletion without confirmation")

	return deleteCmd
}

// getWorkspaceName extracts the workspace name from the workspace file path.
func getWorkspaceName(workspacePath string) string {
	// Extract filename without extension
	filename := filepath.Base(workspacePath)
	return strings.TrimSuffix(filename, ".code-workspace")
}

// displayWorktrees displays worktrees based on project type.
func displayWorktrees(worktrees []status.Repository, projectType wtm.ProjectType) {
	switch projectType {
	case wtm.ProjectTypeSingleRepo:
		displaySingleRepoWorktrees(worktrees)
	case wtm.ProjectTypeWorkspace:
		displayWorkspaceWorktrees(worktrees)
	case wtm.ProjectTypeNone:
		displayFallbackWorktrees(worktrees)
	default:
		displayFallbackWorktrees(worktrees)
	}
}

// displaySingleRepoWorktrees displays worktrees for single repository mode.
func displaySingleRepoWorktrees(worktrees []status.Repository) {
	repoName := worktrees[0].URL
	fmt.Printf("Worktrees for %s repository:\n", repoName)
	displayUniqueBranches(worktrees)
}

// displayWorkspaceWorktrees displays worktrees for workspace mode.
func displayWorkspaceWorktrees(worktrees []status.Repository) {
	workspaceName := getWorkspaceName(worktrees[0].Workspace)
	fmt.Printf("Worktrees for %s workspace:\n", workspaceName)
	displayUniqueBranches(worktrees)
}

// displayFallbackWorktrees displays worktrees in fallback format.
func displayFallbackWorktrees(worktrees []status.Repository) {
	repoName := worktrees[0].URL
	fmt.Printf("Worktrees for %s:\n", repoName)
	for _, worktree := range worktrees {
		fmt.Printf("  %s: %s\n", worktree.Branch, worktree.Path)
	}
}

// displayUniqueBranches displays unique branches from worktrees.
func displayUniqueBranches(worktrees []status.Repository) {
	branches := make(map[string]bool)
	for _, worktree := range worktrees {
		if !branches[worktree.Branch] {
			// Display branch with remote information if available
			if worktree.Remote != "" {
				fmt.Printf("  [%s] %s\n", worktree.Remote, worktree.Branch)
			} else {
				fmt.Printf("  %s\n", worktree.Branch)
			}
			branches[worktree.Branch] = true
		}
	}
}

func createListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List worktrees for the current repository",
		Long:  `List all worktrees for the current Git repository. Currently supports single repository mode.`,
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			// Load configuration and create WTM instance
			cfg := loadConfig()
			wtmManager := wtm.NewWTM(cfg)
			wtmManager.SetVerbose(verbose)

			// List worktrees
			worktrees, projectType, err := wtmManager.ListWorktrees()
			if err != nil {
				return err
			}

			// Display worktrees in simple text format
			if len(worktrees) == 0 {
				switch projectType {
				case wtm.ProjectTypeSingleRepo:
					fmt.Println("No worktrees found for current repository")
				case wtm.ProjectTypeWorkspace:
					fmt.Println("No worktrees found for current workspace")
				case wtm.ProjectTypeNone:
					fmt.Println("No worktrees found")
				default:
					fmt.Println("No worktrees found")
				}
				return nil
			}

			// Display worktrees based on project type
			displayWorktrees(worktrees, projectType)

			return nil
		},
	}
}

func createLoadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "load [remote-source:]<branch-name>",
		Short: "Load branch from remote source",
		Long: `Load a branch from a remote source and create a worktree. Supports loading from 
origin or other users/organizations.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			// Load configuration and create WTM instance
			cfg := loadConfig()
			wtmManager := wtm.NewWTM(cfg)
			wtmManager.SetVerbose(verbose)

			// Load worktree with IDE if specified
			if ideName != "" {
				return wtmManager.LoadWorktree(args[0], wtm.LoadWorktreeOpts{IDEName: ideName})
			}

			// Just load worktree without IDE
			return wtmManager.LoadWorktree(args[0])
		},
	}
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "wtm",
		Short: "Cursor Git WorkTree Manager",
		Long:  `A powerful CLI tool for managing Git worktrees specifically designed for Cursor IDE.`,
	}

	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress all output except errors")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Specify a custom config file path")

	// Create commands
	createCmd := createCreateCmd()
	openCmd := createOpenCmd()
	deleteCmd := createDeleteCmd()
	listCmd := createListCmd()
	loadCmd := createLoadCmd()

	// Add IDE flag to create command
	createCmd.Flags().StringVarP(&ideName, "ide", "i", "", "Open in specified IDE after creation")

	// Add IDE flag to load command
	loadCmd.Flags().StringVarP(&ideName, "ide", "i", "", "Open in specified IDE after loading")

	// Add subcommands
	rootCmd.AddCommand(createCmd, openCmd, deleteCmd, listCmd, loadCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
