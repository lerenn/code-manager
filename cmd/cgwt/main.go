package main

import (
	"log"

	"github.com/lerenn/cgwt/pkg/cgwt"
	"github.com/lerenn/cgwt/pkg/fs"
	"github.com/spf13/cobra"
)

var (
	quiet   bool
	verbose bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cgwt",
		Short: "Cursor Git WorkTree Manager",
		Long:  `A powerful CLI tool for managing Git worktrees specifically designed for Cursor IDE.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			outputMode := getOutputMode()
			fsAdapter := fs.NewFS()
			cgwtManager := cgwt.NewCGWTWithMode(fsAdapter, outputMode)
			return cgwtManager.Run()
		},
	}

	var createCmd = &cobra.Command{
		Use:   "create [branch]",
		Short: "Create worktree(s) for the specified branch",
		Long:  `Create worktree(s) for the specified branch. Currently only detects Git repository mode.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			// For now, just call the detection logic
			// The branch argument is not used yet
			outputMode := getOutputMode()
			fsAdapter := fs.NewFS()
			cgwtManager := cgwt.NewCGWTWithMode(fsAdapter, outputMode)
			return cgwtManager.Run()
		},
	}

	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress all output except errors")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Add subcommands
	rootCmd.AddCommand(createCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// getOutputMode determines the output mode based on flags.
func getOutputMode() cgwt.OutputMode {
	if quiet {
		return cgwt.OutputModeQuiet
	}
	if verbose {
		return cgwt.OutputModeVerbose
	}
	return cgwt.OutputModeNormal
}
