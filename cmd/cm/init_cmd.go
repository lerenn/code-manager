package main

import (
	"os"
	"path/filepath"

	"github.com/lerenn/code-manager/cmd/cm/internal/config"
	cm "github.com/lerenn/code-manager/pkg/cm"
	pkgconfig "github.com/lerenn/code-manager/pkg/config"
	"github.com/spf13/cobra"
)

var (
	force    bool
	reset    bool
	basePath string
)

func createInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init [--force] [--base-path <path>] [--reset]",
		Short: "Initialize CM configuration",
		Long: `Initialize CM configuration with interactive prompts or direct path specification.

Flags:
  --force       Skip interactive confirmation when using --reset flag
  --base-path   Set the base path for code storage directly (skips interactive prompt)
  --reset       Reset existing CM configuration and start fresh`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			// Resolve config path
			var path string
			if config.ConfigPath != "" {
				path = config.ConfigPath
			} else {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					homeDir = "."
				}
				path = filepath.Join(homeDir, ".cm", "config.yaml")
			}

			// Ensure config file exists (copy embedded default if missing)
			manager := pkgconfig.NewManager()
			cfg, _, err := manager.EnsureConfigFile(path)
			if err != nil {
				return err
			}

			cmManager, err := cm.NewCM(cfg)
			if err != nil {
				return err
			}
			cmManager.SetVerbose(config.Verbose)

			opts := cm.InitOpts{
				Force:    force,
				Reset:    reset,
				BasePath: basePath,
			}

			return cmManager.Init(opts)
		},
	}

	// Add flags
	initCmd.Flags().BoolVar(&force, "force", false, "Skip interactive confirmation when using --reset flag")
	initCmd.Flags().StringVar(&basePath, "base-path", "",
		"Set the base path for code storage directly (skips interactive prompt)")
	initCmd.Flags().BoolVar(&reset, "reset", false, "Reset existing CM configuration and start fresh")

	return initCmd
}
