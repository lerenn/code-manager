//go:build unit

package codemanager

import (
	"errors"
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestListWorkspaces(t *testing.T) {
	tests := []struct {
		name           string
		workspaces     map[string]status.Workspace
		expectedResult []WorkspaceInfo
		expectedError  error
	}{
		{
			name: "success with multiple workspaces",
			workspaces: map[string]status.Workspace{
				"workspace1": {
					Repositories: []string{"repo1", "repo2"},
					Worktrees:    []string{"branch1", "branch2"},
				},
				"workspace2": {
					Repositories: []string{"repo3"},
					Worktrees:    []string{},
				},
			},
			expectedResult: []WorkspaceInfo{
				{
					Name:         "workspace1",
					Repositories: []string{"repo1", "repo2"},
					Worktrees:    []string{"branch1", "branch2"},
				},
				{
					Name:         "workspace2",
					Repositories: []string{"repo3"},
					Worktrees:    []string{},
				},
			},
			expectedError: nil,
		},
		{
			name:           "success with no workspaces",
			workspaces:     map[string]status.Workspace{},
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name: "success with single workspace",
			workspaces: map[string]status.Workspace{
				"single-workspace": {
					Repositories: []string{"single-repo"},
					Worktrees:    []string{"single-branch"},
				},
			},
			expectedResult: []WorkspaceInfo{
				{
					Name:         "single-workspace",
					Repositories: []string{"single-repo"},
					Worktrees:    []string{"single-branch"},
				},
			},
			expectedError: nil,
		},
		{
			name:           "error from status manager",
			workspaces:     nil,
			expectedResult: nil,
			expectedError:  errors.New("status manager error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mocks
			mockStatusManager := statusmocks.NewMockManager(ctrl)

			// Setup status manager mock
			if tt.expectedError != nil {
				mockStatusManager.EXPECT().ListWorkspaces().Return(nil, tt.expectedError)
			} else {
				mockStatusManager.EXPECT().ListWorkspaces().Return(tt.workspaces, nil)
			}

			// Create CM instance
			cm := &realCodeManager{
				statusManager: mockStatusManager,
				hookManager:   nil, // No hooks for this test
				configManager: config.NewConfigManager("/test/config.yaml"),
			}

			// Execute
			result, err := cm.ListWorkspaces()

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestListWorkspaces_Sorting(t *testing.T) {
	// Test that workspaces are sorted by name
	workspaces := map[string]status.Workspace{
		"z-workspace": {
			Repositories: []string{"repo1"},
			Worktrees:    []string{"branch1"},
		},
		"a-workspace": {
			Repositories: []string{"repo2"},
			Worktrees:    []string{"branch2"},
		},
		"m-workspace": {
			Repositories: []string{"repo3"},
			Worktrees:    []string{"branch3"},
		},
	}

	expectedResult := []WorkspaceInfo{
		{
			Name:         "a-workspace",
			Repositories: []string{"repo2"},
			Worktrees:    []string{"branch2"},
		},
		{
			Name:         "m-workspace",
			Repositories: []string{"repo3"},
			Worktrees:    []string{"branch3"},
		},
		{
			Name:         "z-workspace",
			Repositories: []string{"repo1"},
			Worktrees:    []string{"branch1"},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockStatusManager := statusmocks.NewMockManager(ctrl)

	// Setup mocks
	mockStatusManager.EXPECT().ListWorkspaces().Return(workspaces, nil)

	// Create CM instance
	cm := &realCodeManager{
		statusManager: mockStatusManager,
		hookManager:   nil, // No hooks for this test
		configManager: config.NewConfigManager("/test/config.yaml"),
	}

	// Execute
	result, err := cm.ListWorkspaces()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestListWorkspaces_EmptyWorkspace(t *testing.T) {
	// Test workspace with empty repositories and worktrees
	workspaces := map[string]status.Workspace{
		"empty-workspace": {
			Repositories: []string{},
			Worktrees:    []string{},
		},
	}

	expectedResult := []WorkspaceInfo{
		{
			Name:         "empty-workspace",
			Repositories: []string{},
			Worktrees:    []string{},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockStatusManager := statusmocks.NewMockManager(ctrl)

	// Setup mocks
	mockStatusManager.EXPECT().ListWorkspaces().Return(workspaces, nil)

	// Create CM instance
	cm := &realCodeManager{
		statusManager: mockStatusManager,
		hookManager:   nil, // No hooks for this test
		configManager: config.NewConfigManager("/test/config.yaml"),
	}

	// Execute
	result, err := cm.ListWorkspaces()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}
