package main

import (
	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/spf13/cobra"
)

func createCloneCmd() *cobra.Command {
	var shallow bool

	cloneCmd := &cobra.Command{
		Use:   "clone <repository-url> [--shallow]",
		Short: "Clone a repository and initialize it in CM",
		Long: `Clone a repository from a remote source and initialize it in CM.

The repository will be cloned to $base_path/<repo_url>/<remote_name>/<default_branch> 
and automatically initialized in CM with the detected default branch.

Examples:
  cm clone https://github.com/lerenn/example.git
  cm clone git@github.com:lerenn/example.git
  cm clone https://github.com/lerenn/example.git --shallow`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg := loadConfig()
			cmManager := cm.NewCM(cfg)
			cmManager.SetVerbose(verbose)

			// Create clone options
			opts := cm.CloneOpts{
				Recursive: !shallow, // --shallow means not recursive
			}

			return cmManager.Clone(args[0], opts)
		},
	}

	// Add flags
	cloneCmd.Flags().BoolVarP(&shallow, "shallow", "s", false, "Perform a shallow clone (non-recursive)")

	return cloneCmd
}
