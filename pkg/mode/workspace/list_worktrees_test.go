//go:build unit

package workspace

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/status"
	statusMocks "github.com/lerenn/code-manager/pkg/status/mocks"
)

func TestListWorktrees(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*statusMocks.MockManager)
		expectedResult []status.WorktreeInfo
		expectedError  string
	}{
		{
			name: "successful listing with worktrees",
			setupMocks: func(statusMock *statusMocks.MockManager) {
				workspace := &status.Workspace{
					Repositories: []string{"https://github.com/user/repo1", "https://github.com/user/repo2"},
				}
				statusMock.EXPECT().GetWorkspace("/path/to").Return(workspace, nil)

				repo1 := &status.Repository{
					Path: "/repos/repo1",
					Worktrees: map[string]status.WorktreeInfo{
						"feature-branch-1": {Remote: "origin", Branch: "feature-branch-1"},
						"feature-branch-2": {Remote: "origin", Branch: "feature-branch-2"},
					},
				}
				repo2 := &status.Repository{
					Path: "/repos/repo2",
					Worktrees: map[string]status.WorktreeInfo{
						"feature-branch-3": {Remote: "origin", Branch: "feature-branch-3"},
					},
				}
				statusMock.EXPECT().GetRepository("https://github.com/user/repo1").Return(repo1, nil)
				statusMock.EXPECT().GetRepository("https://github.com/user/repo2").Return(repo2, nil)
			},
			expectedResult: []status.WorktreeInfo{
				{Remote: "origin", Branch: "feature-branch-1"},
				{Remote: "origin", Branch: "feature-branch-2"},
				{Remote: "origin", Branch: "feature-branch-3"},
			},
		},
		{
			name: "workspace not found",
			setupMocks: func(statusMock *statusMocks.MockManager) {
				statusMock.EXPECT().GetWorkspace("/path/to/workspace.code-workspace").Return(nil, status.ErrWorkspaceNotFound)
			},
			expectedResult: []status.WorktreeInfo{},
		},
		{
			name: "empty workspace",
			setupMocks: func(statusMock *statusMocks.MockManager) {
				workspace := &status.Workspace{
					Repositories: []string{},
				}
				statusMock.EXPECT().GetWorkspace("/path/to").Return(workspace, nil)
			},
			expectedResult: []status.WorktreeInfo{},
		},
		{
			name: "repository not found",
			setupMocks: func(statusMock *statusMocks.MockManager) {
				workspace := &status.Workspace{
					Repositories: []string{"https://github.com/user/repo1"},
				}
				statusMock.EXPECT().GetWorkspace("/path/to").Return(workspace, nil)
				statusMock.EXPECT().GetRepository("https://github.com/user/repo1").Return(nil, errors.New("repository not found"))
			},
			expectedResult: []status.WorktreeInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			statusMock := statusMocks.NewMockManager(ctrl)
			tt.setupMocks(statusMock)

			workspace := &realWorkspace{
				statusManager: statusMock,
				logger:        logger.NewNoopLogger(),
				file:          "/path/to/workspace.code-workspace",
			}

			result, err := workspace.ListWorktrees()

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestListWorkspaceWorktrees(t *testing.T) {
	tests := []struct {
		name           string
		workspaceName  string
		setupMocks     func(*statusMocks.MockManager)
		expectedResult []WorktreeInfo
		expectedError  string
	}{
		{
			name:          "successful listing with worktrees",
			workspaceName: "test-workspace",
			setupMocks: func(statusMock *statusMocks.MockManager) {
				workspace := &status.Workspace{
					Repositories: []string{"https://github.com/user/repo1"},
				}
				statusMock.EXPECT().GetWorkspaceByName("test-workspace").Return(workspace, nil)

				workspaces := map[string]status.Workspace{
					"/path/to/test-workspace.code-workspace": *workspace,
				}
				statusMock.EXPECT().ListWorkspaces().Return(workspaces, nil)

				repo1 := &status.Repository{
					Path: "/repos/repo1",
					Worktrees: map[string]status.WorktreeInfo{
						"feature-branch-1": {Remote: "origin", Branch: "feature-branch-1"},
					},
				}
				statusMock.EXPECT().GetRepository("https://github.com/user/repo1").Return(repo1, nil)
			},
			expectedResult: []WorktreeInfo{
				{
					Repository:    "https://github.com/user/repo1",
					Branch:        "feature-branch-1",
					Remote:        "origin",
					WorktreePath:  "/repos/repo1-feature-branch-1",
					WorkspaceFile: "test-workspace-feature-branch-1.code-workspace",
					Issue:         &status.WorktreeInfo{Remote: "origin", Branch: "feature-branch-1"},
				},
			},
		},
		{
			name:          "workspace not found",
			workspaceName: "nonexistent",
			setupMocks: func(statusMock *statusMocks.MockManager) {
				statusMock.EXPECT().GetWorkspaceByName("nonexistent").Return(nil, status.ErrWorkspaceNotFound)
			},
			expectedResult: []WorktreeInfo{},
		},
		{
			name:          "workspace path not found",
			workspaceName: "test-workspace",
			setupMocks: func(statusMock *statusMocks.MockManager) {
				workspace := &status.Workspace{
					Repositories: []string{"https://github.com/user/repo1"},
				}
				statusMock.EXPECT().GetWorkspaceByName("test-workspace").Return(workspace, nil)
				statusMock.EXPECT().ListWorkspaces().Return(map[string]status.Workspace{}, nil)
			},
			expectedError: "failed to determine workspace path for: test-workspace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			statusMock := statusMocks.NewMockManager(ctrl)
			tt.setupMocks(statusMock)

			workspace := &realWorkspace{
				statusManager: statusMock,
				logger:        logger.NewNoopLogger(),
			}

			result, err := workspace.ListWorkspaceWorktrees(tt.workspaceName)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestExtractWorkspaceNameFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "workspace with extension",
			path:     "/path/to/test-workspace.code-workspace",
			expected: "test-workspace",
		},
		{
			name:     "workspace without extension",
			path:     "/path/to/test-workspace",
			expected: "test-workspace",
		},
		{
			name:     "workspace with multiple dots",
			path:     "/path/to/test.workspace.code-workspace",
			expected: "test.workspace",
		},
		{
			name:     "just filename",
			path:     "test-workspace.code-workspace",
			expected: "test-workspace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace := &realWorkspace{}
			result := workspace.extractWorkspaceNameFromPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
