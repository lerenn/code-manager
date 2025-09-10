//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	"github.com/lerenn/code-manager/pkg/mode/workspace"
	"github.com/stretchr/testify/assert"
)

func TestNewCM_WithProviders(t *testing.T) {
	// Create test config
	cfg := config.Config{
		BasePath:   "/test/base/path",
		StatusFile: "/test/status.yaml",
	}

	// Create mock providers
	repoProvider := func(params repository.NewRepositoryParams) repository.Repository {
		// Return a mock repository
		return nil // In real tests, this would be a mock
	}

	workspaceProvider := func(params workspace.NewWorkspaceParams) workspace.Workspace {
		// Return a mock workspace
		return nil // In real tests, this would be a mock
	}

	// Test creating CM with providers
	params := NewCMParams{
		RepositoryProvider: repoProvider,
		WorkspaceProvider:  workspaceProvider,
		Config:             cfg,
	}

	cm, err := NewCM(params)
	assert.NoError(t, err)
	assert.NotNil(t, cm)
}

func TestNewCM_WithoutProviders(t *testing.T) {
	// Create test config
	cfg := config.Config{
		BasePath:   "/test/base/path",
		StatusFile: "/test/status.yaml",
	}

	// Test creating CM without providers (should create defaults)
	params := NewCMParams{
		Config: cfg,
	}

	cm, err := NewCM(params)
	assert.NoError(t, err)
	assert.NotNil(t, cm)
}
