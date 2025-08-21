package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

//go:generate mockgen -source=prompt.go -destination=mockprompt.gen.go -package=prompt

// Prompt interface provides user interaction functionality.
type Prompt interface {
	// PromptForBasePath prompts the user for the base path with examples.
	PromptForBasePath(defaultBasePath string) (string, error)

	// PromptForConfirmation prompts the user for confirmation with a default value.
	PromptForConfirmation(message string, defaultYes bool) (bool, error)
}

type realPrompt struct {
	reader *bufio.Reader
}

// NewPrompt creates a new Prompt instance.
func NewPrompt() Prompt {
	return &realPrompt{
		reader: bufio.NewReader(os.Stdin),
	}
}

// PromptForBasePath prompts the user for the base path with examples.
func (p *realPrompt) PromptForBasePath(defaultBasePath string) (string, error) {
	if defaultBasePath == "" {
		defaultBasePath = "~/Code"
	}
	fmt.Printf("Choose the location of the repositories (ex: ~/Code, ~/Projects, ~/Worktrees): "+
		"[default: %s]: ", defaultBasePath)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read user input: %w", err)
	}

	// Trim whitespace and newlines
	input = strings.TrimSpace(input)

	// Use default if input is empty
	if input == "" {
		return defaultBasePath, nil
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
