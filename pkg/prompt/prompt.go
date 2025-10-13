package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

//go:generate go run go.uber.org/mock/mockgen@latest  -source=prompt.go -destination=mocks/prompt.gen.go -package=mocks

// Target type string constants.
const (
	// TargetRepository is the string representation of repository target type.
	TargetRepository = "repository"
	// TargetWorkspace is the string representation of workspace target type.
	TargetWorkspace = "workspace"
)

// TargetChoice represents a selectable target with optional worktree information.
type TargetChoice struct {
	Type     string
	Name     string
	Worktree string // optional label for display only
}

// Prompter interface provides user interaction functionality.
type Prompter interface {
	// PromptForRepositoriesDir prompts the user for the repositories directory with examples.
	PromptForRepositoriesDir(defaultRepositoriesDir string) (string, error)

	// PromptForWorkspacesDir prompts the user for the workspaces directory with examples.
	PromptForWorkspacesDir(defaultWorkspacesDir string) (string, error)

	// PromptForStatusFile prompts the user for the status file location with examples.
	PromptForStatusFile(defaultStatusFile string) (string, error)

	// PromptForConfirmation prompts the user for confirmation with a default value.
	PromptForConfirmation(message string, defaultYes bool) (bool, error)

	// PromptSelectTarget prompts the user to select a repository or workspace from a list.
	// showWorktreeLabel controls rendering of ": worktree" suffix.
	PromptSelectTarget(choices []TargetChoice, showWorktreeLabel bool) (TargetChoice, error)

	// PromptForBranchName prompts the user for a branch name.
	PromptForBranchName() (string, error)
}

type realPrompt struct {
	reader *bufio.Reader
}

// NewPrompt creates a new Prompt instance.
func NewPrompt() Prompter {
	return &realPrompt{
		reader: bufio.NewReader(os.Stdin),
	}
}

// PromptForRepositoriesDir prompts the user for the repositories directory with examples.
func (p *realPrompt) PromptForRepositoriesDir(defaultRepositoriesDir string) (string, error) {
	if defaultRepositoriesDir == "" {
		defaultRepositoriesDir = "~/Code/repos"
	}
	fmt.Printf("Choose the location of the repositories directory "+
		"(ex: ~/Code/repos, ~/Projects/repos, ~/Worktrees/repos): "+
		"[default: %s]: ", defaultRepositoriesDir)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read user input: %w", err)
	}

	// Trim whitespace and newlines
	input = strings.TrimSpace(input)

	// Use default if input is empty
	if input == "" {
		return defaultRepositoriesDir, nil
	}

	return input, nil
}

// PromptForWorkspacesDir prompts the user for the workspaces directory with examples.
func (p *realPrompt) PromptForWorkspacesDir(defaultWorkspacesDir string) (string, error) {
	if defaultWorkspacesDir == "" {
		defaultWorkspacesDir = "~/Code/workspaces"
	}
	fmt.Printf("Choose the location of the workspaces directory "+
		"(ex: ~/Code/workspaces, ~/Projects/workspaces, ~/Worktrees/workspaces): "+
		"[default: %s]: ", defaultWorkspacesDir)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read user input: %w", err)
	}

	// Trim whitespace and newlines
	input = strings.TrimSpace(input)

	// Use default if input is empty
	if input == "" {
		return defaultWorkspacesDir, nil
	}

	return input, nil
}

// PromptForStatusFile prompts the user for the status file location with examples.
func (p *realPrompt) PromptForStatusFile(defaultStatusFile string) (string, error) {
	if defaultStatusFile == "" {
		defaultStatusFile = "~/.cm/status.yaml"
	}
	fmt.Printf("Choose the location of the status file "+
		"(ex: ~/.cm/status.yaml, ~/.config/cm/status.yaml, ./cm-status.yaml): "+
		"[default: %s]: ", defaultStatusFile)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read user input: %w", err)
	}

	// Trim whitespace and newlines
	input = strings.TrimSpace(input)

	// Use default if input is empty
	if input == "" {
		return defaultStatusFile, nil
	}

	return input, nil
}

// PromptForConfirmation prompts the user for confirmation with a default value.
func (p *realPrompt) PromptForConfirmation(message string, defaultYes bool) (bool, error) {
	var defaultText string
	if defaultYes {
		defaultText = "[Y/n]"
	} else {
		defaultText = "[y/N]"
	}

	fmt.Printf("%s %s: ", message, defaultText)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}

	// Trim whitespace and newlines
	input = strings.TrimSpace(strings.ToLower(input))

	// Use default if input is empty
	if input == "" {
		return defaultYes, nil
	}

	// Check for yes/no responses
	switch input {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, ErrInvalidConfirmationInput
	}
}

// PromptSelectTarget prompts the user to select a repository or workspace from a list.
func (p *realPrompt) PromptSelectTarget(choices []TargetChoice, showWorktreeLabel bool) (TargetChoice, error) {
	if len(choices) == 0 {
		return TargetChoice{}, fmt.Errorf("no choices available")
	}

	// Use Bubble Tea selector for interactive selection
	return promptSelectTargetBubbleTea(choices, showWorktreeLabel)
}

// PromptForBranchName prompts the user for a branch name.
func (p *realPrompt) PromptForBranchName() (string, error) {
	fmt.Print("Enter branch name: ")

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read user input: %w", err)
	}

	// Trim whitespace and newlines
	branchName := strings.TrimSpace(input)

	if branchName == "" {
		return "", fmt.Errorf("branch name cannot be empty")
	}

	return branchName, nil
}
