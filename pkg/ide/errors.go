package ide

import "errors"

var (
	// ErrIDENotInstalled is returned when an IDE is not installed on the system.
	ErrIDENotInstalled = errors.New("IDE not installed")

	// ErrUnsupportedIDE is returned when an IDE is not supported.
	ErrUnsupportedIDE = errors.New("unsupported IDE")

	// ErrIDEExecutionFailed is returned when IDE command execution fails.
	ErrIDEExecutionFailed = errors.New("failed to execute IDE command")

	// ErrWorktreeNotFound is returned when a worktree is not found in status.yaml.
	ErrWorktreeNotFound = errors.New("worktree not found")
)
