// Package gitcrypt provides git-crypt functionality as a hook for worktree operations.
package gitcrypt

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/fs"
)

// WorktreeSetup handles git-crypt setup in worktree directories.
type WorktreeSetup struct {
	fs fs.FS
}

// NewWorktreeSetup creates a new GitCryptWorktreeSetup instance.
func NewWorktreeSetup(fs fs.FS) *WorktreeSetup {
	return &WorktreeSetup{
		fs: fs,
	}
}

// SetupGitCryptForWorktree sets up git-crypt in the worktree by copying the key.
func (s *WorktreeSetup) SetupGitCryptForWorktree(repoPath, worktreePath, keyPath string) error {
	// Get the worktree's git directory
	worktreeGitDir := filepath.Join(repoPath, ".git", "worktrees", filepath.Base(worktreePath))

	// Create git-crypt directory in worktree's git directory
	gitCryptDir := filepath.Join(worktreeGitDir, "git-crypt")
	if err := s.fs.MkdirAll(gitCryptDir, 0755); err != nil {
		return fmt.Errorf("failed to create git-crypt directory: %w", err)
	}

	// Create keys directory
	keysDir := filepath.Join(gitCryptDir, "keys")
	if err := s.fs.MkdirAll(keysDir, 0755); err != nil {
		return fmt.Errorf("failed to create keys directory: %w", err)
	}

	// Copy the key to the worktree's git-crypt directory
	destKeyPath := filepath.Join(keysDir, "default")
	if err := s.copyFile(keyPath, destKeyPath); err != nil {
		return fmt.Errorf("failed to copy git-crypt key: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst.
func (s *WorktreeSetup) copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		_ = srcFile.Close()
	}()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		_ = dstFile.Close()
	}()

	// Copy file contents
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Get source file info to copy permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	// Set destination file permissions
	if err := dstFile.Chmod(srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set destination file permissions: %w", err)
	}

	return nil
}
