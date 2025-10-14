// Package repository provides repository management commands for the CM CLI.
package repository

import (
	"fmt"

	"github.com/lerenn/code-manager/cmd/cm/internal/cli"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List all repositories in CM",
		Long: `List all repositories tracked by CM with visual indicators.

An asterisk (*) indicates repositories that are not within the configured base path.

Examples:
  cm repository list
  cm repo ls
  cm r l`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := cli.CheckInitialization(); err != nil {
				return err
			}

			cmManager, err := cli.NewCodeManager()
			if err != nil {
				return err
			}
			if cli.Verbose {
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

			for _, repo := range repositories {
				indicator := ""
				if !repo.InRepositoriesDir {
					indicator = "*"
				}
				fmt.Printf("  %s%s\n", indicator, repo.Name)
			}

			return nil
		},
	}

	return listCmd
}
