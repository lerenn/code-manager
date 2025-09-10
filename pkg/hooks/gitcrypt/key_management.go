// Package gitcrypt provides git-crypt functionality as a hook for worktree operations.
package gitcrypt

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/prompt"
)

// KeyManager handles git-crypt key location and validation.
type KeyManager struct {
	fs     fs.FS
	git    git.Git
	prompt prompt.Prompter
}

// NewKeyManager creates a new GitCryptKeyManager instance.
func NewKeyManager(fs fs.FS, git git.Git, prompt prompt.Prompter) *KeyManager {
	return &KeyManager{
		fs:     fs,
		git:    git,
		prompt: prompt,
	}
}

// FindGitCryptKey looks for git-crypt key in the repository.
func (k *KeyManager) FindGitCryptKey(repoPath string) (string, error) {
	// Check if repository has git-crypt key in the default location
	keyPath := filepath.Join(repoPath, ".git", "git-crypt", "keys", "default")

	exists, err := k.fs.Exists(keyPath)
	if err != nil {
		return "", err
	}

	if exists {
		return keyPath, nil
	}

	return "", nil // Key not found
}

// PromptUserForKeyPath prompts the user for the git-crypt key file path.
func (k *KeyManager) PromptUserForKeyPath() (string, error) {
	// For now, we'll use a simple approach since the prompt interface doesn't have a generic input method
	// We'll need to extend the prompt interface or use a different approach
	// For this implementation, we'll return an error asking the user to provide the key path
	return "", fmt.Errorf("git-crypt key not found in repository. " +
		"Please ensure the repository is unlocked with git-crypt unlock")
}

// ValidateKeyFile validates that the provided key file exists and is readable.
func (k *KeyManager) ValidateKeyFile(keyPath string) error {
	exists, err := k.fs.Exists(keyPath)
	if err != nil {
		return fmt.Errorf("failed to check key file existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("%w: %s", ErrKeyFileNotFound, keyPath)
	}

	// Try to read the key file to ensure it's accessible
	_, err = k.fs.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrKeyFileInvalid, keyPath)
	}

	return nil
}
