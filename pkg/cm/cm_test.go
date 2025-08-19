package cm

import (
	"github.com/lerenn/cm/pkg/config"
)

// createTestConfig creates a test configuration.
//
//nolint:unused // This function is used by other test files in the package
func createTestConfig() *config.Config {
	return &config.Config{
		BasePath:   "/test/base/path",
		StatusFile: "/test/status.yaml",
	}
}
