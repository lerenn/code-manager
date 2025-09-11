// Package worktree provides worktree management commands for the CM CLI.
package worktree

import (
	"github.com/lerenn/code-manager/cmd/cm/internal/config"
	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/hooks/ide"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createCreateCmd() *cobra.Command {
	var ideName string
	var force bool
	var fromIssue string
	var workspaceName string

	createCmd := &cobra.Command{
		Use:   "create [branch] [--from-issue <issue-reference>] [--ide <ide-name>] [--workspace <workspace-name>]",
		Short: "Create a worktree for the specified branch or from a GitHub issue",
		Long:  getCreateCommandLongDescription(),
		Args:  createCreateCmdArgsValidator(&fromIssue, &workspaceName),
		RunE: createCreateCmdRunE(createCreateCmdRunEParams{
			IDEName:       &ideName,
			Force:         &force,
			FromIssue:     &fromIssue,
			WorkspaceName: &workspaceName,
		}),
	}

	// Add flags
	createCmd.Flags().StringVarP(&ideName, "ide", "i", "", "Open in specified IDE after creation")
	createCmd.Flags().BoolVarP(&force, "force", "f", false, "Force creation without prompts")
	createCmd.Flags().StringVar(&fromIssue, "from-issue", "",
		"Create worktree from GitHub issue (URL, number, or owner/repo#issue format)")
	createCmd.Flags().StringVarP(&workspaceName, "workspace", "w", "",
		"Create worktrees from workspace definition in status.yaml")

	return createCmd
}

// getCreateCommandLongDescription returns the long description for the create command.
func getCreateCommandLongDescription() string {
	return `Create a worktree for the specified branch in the current repository or workspace.
When using --from-issue, the branch name becomes optional and will be inferred from the issue title.
When using --workspace, worktrees will be created in all repositories defined in the workspace.

Issue Reference Formats:
  - GitHub issue URL: https://github.com/owner/repo/issues/123
  - Issue number (requires remote origin to be GitHub): 123
  - Owner/repo#issue format: owner/repo#123

Examples:
  cm worktree create feature-branch
  cm wt create feature-branch --ide ` + ide.DefaultIDE + `
  cm w create feature-branch --ide cursor
  cm worktree create --from-issue https://github.com/owner/repo/issues/123
  cm worktree create custom-branch --from-issue 456
  cm worktree create --from-issue owner/repo#789 --ide cursor
  cm worktree create feature-branch --workspace my-workspace
  cm worktree create feature-branch --workspace my-workspace --ide cursor
  cm worktree create --from-issue 123 --workspace my-workspace`
}

// createCreateCmdArgsValidator creates the argument validator for the create command.
func createCreateCmdArgsValidator(fromIssue *string, workspaceName *string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// If --from-issue is provided, branch name is optional
		if *fromIssue != "" {
			return cobra.MaximumNArgs(1)(cmd, args)
		}
		// If --workspace is provided, branch name is required
		if *workspaceName != "" {
			return cobra.ExactArgs(1)(cmd, args)
		}
		// Otherwise, branch name is required
		return cobra.ExactArgs(1)(cmd, args)
	}
}

// createCreateCmdRunEParams contains parameters for createCreateCmdRunE.
type createCreateCmdRunEParams struct {
	IDEName       *string
	Force         *bool
	FromIssue     *string
	WorkspaceName *string
}

// createCreateCmdRunE creates the RunE function for the create command.
func createCreateCmdRunE(params createCreateCmdRunEParams) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
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

		// Determine branch name
		var branchName string
		if len(args) > 0 {
			branchName = args[0]
		}

		var opts cm.CreateWorkTreeOpts
		if *params.IDEName != "" {
			opts.IDEName = *params.IDEName
		}
		if *params.FromIssue != "" {
			opts.IssueRef = *params.FromIssue
		}
		if *params.WorkspaceName != "" {
			opts.WorkspaceName = *params.WorkspaceName
		}
		opts.Force = *params.Force

		return cmManager.CreateWorkTree(branchName, opts)
	}
}
