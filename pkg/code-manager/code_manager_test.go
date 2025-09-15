//go:build unit

package codemanager

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/dependencies"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	"github.com/lerenn/code-manager/pkg/mode/workspace"
	"github.com/stretchr/testify/assert"
)

func TestNewCodeManager_WithProviders(t *testing.T) {
	// Create test config manager
	configManager := config.NewConfigManager("/test/config.yaml")

	// Create mock providers
	repoProvider := func(params repository.NewRepositoryParams) repository.Repository {
		// Return a mock repository
		return nil // In real tests, this would be a mock
	}

	workspaceProvider := func(params workspace.NewWorkspaceParams) workspace.Workspace {
		// Return a mock workspace
		return nil // In real tests, this would be a mock
	}

	// Test creating CodeManager with providers
	params := NewCodeManagerParams{
		Dependencies: dependencies.New().
			WithRepositoryProvider(repoProvider).
			WithWorkspaceProvider(workspaceProvider).
			WithConfig(configManager),
	}

	cm, err := NewCodeManager(params)
	assert.NoError(t, err)
	assert.NotNil(t, cm)
}

func TestNewCodeManager_WithoutProviders(t *testing.T) {
	// Create test config manager
	configManager := config.NewConfigManager("/test/config.yaml")

	// Test creating CodeManager without providers (should create defaults)
	params := NewCodeManagerParams{
		Dependencies: dependencies.New().
			WithConfig(configManager),
	}

	cm, err := NewCodeManager(params)
	assert.NoError(t, err)
	assert.NotNil(t, cm)
}
