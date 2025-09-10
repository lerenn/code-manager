//go:build unit

package status

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"
)

func TestListWorkspaces(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)

	cfg := config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Expected workspaces
	expectedWorkspaces := map[string]Workspace{
		"/home/user/workspace1.code-workspace": {
			Worktree:     []string{"origin:feature-a"},
			Repositories: []string{"github.com/octocat/Hello-World"},
		},
		"/home/user/workspace2.code-workspace": {
			Worktree:     []string{"origin:feature-b"},
			Repositories: []string{"github.com/lerenn/other"},
		},
	}

	// Existing status
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   expectedWorkspaces,
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	workspaces, err := manager.ListWorkspaces()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedWorkspaces, workspaces)
}
