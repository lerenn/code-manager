package workspace

import (
	"github.com/lerenn/code-manager/pkg/logger"
)

// SetLogger sets the logger for this workspace instance.
func (w *realWorkspace) SetLogger(logger logger.Logger) {
	w.logger = logger
}
