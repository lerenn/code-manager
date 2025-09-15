// Package worktree provides worktree management functionality for CM.
package worktree

import (
	"github.com/lerenn/code-manager/pkg/logger"
)

// SetLogger sets the logger for this worktree instance.
func (w *realWorktree) SetLogger(logger logger.Logger) {
	w.logger = logger
}
