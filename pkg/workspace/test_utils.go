//go:build unit

package workspace

import (
	"github.com/lerenn/code-manager/pkg/config"
)

// createTestConfig creates a test configuration.
func createTestConfig() *config.Config {
	return &config.Config{
		BasePath:   "/test/base/path",
		StatusFile: "/test/status.yaml",
	}
}
