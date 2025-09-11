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

func TestListRepositories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/home/user/.cm",
		StatusFile:      "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Expected repositories
	expectedRepos := map[string]Repository{
		"github.com/octocat/Hello-World": {
			Path: "/home/user/.cmrepos/github.com/octocat/Hello-World/origin/main",
			Remotes: map[string]Remote{
				"origin": {
					DefaultBranch: "main",
				},
			},
			Worktrees: make(map[string]WorktreeInfo),
		},
		"github.com/lerenn/other": {
			Path: "/home/user/repos/other",
			Remotes: map[string]Remote{
				"origin": {
					DefaultBranch: "master",
				},
			},
			Worktrees: make(map[string]WorktreeInfo),
		},
	}

	// Existing status
	existingStatus := &Status{
		Repositories: expectedRepos,
		Workspaces:   make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	repos, err := manager.ListRepositories()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedRepos, repos)
}
