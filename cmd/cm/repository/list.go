// Package repository provides repository management commands for the CM CLI.
package repository

import (
	"fmt"

	"github.com/lerenn/code-manager/cmd/cm/internal/config"
	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List all repositories in CM",
		Long: `List all repositories tracked by CM with visual indicators.

Repositories are displayed in a numbered list format. An asterisk (*) indicates 
repositories that are not within the configured base path.

Examples:
  cm repository list
  cm repo ls
  cm r l`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := config.CheckInitialization(); err != nil {
				return err
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			cmManager, err := cm.NewCM(cm.NewCMParams{
				Config: cfg,
			})
			if err != nil {
				return err
			}
			if config.Verbose {
				cmManager.SetLogger(logger.NewVerboseLogger())
			}

			repositories, err := cmManager.ListRepositories()
			if err != nil {
				return err
			}

			// Display repositories
			if len(repositories) == 0 {
				fmt.Println("No repositories found in status.yaml")
				return nil
			}

			for i, repo := range repositories {
				indicator := ""
				if !repo.InRepositoriesDir {
					indicator = "*"
				}
				fmt.Printf("  %d. %s%s\n", i+1, indicator, repo.Name)
			}

			return nil
		},
	}

	return listCmd
}
