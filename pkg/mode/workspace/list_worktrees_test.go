//go:build unit

package workspace

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/lerenn/code-manager/pkg/dependencies"
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
						"origin:feature-branch-1": {Remote: "origin", Branch: "feature-branch-1"},
						"origin:feature-branch-2": {Remote: "origin", Branch: "feature-branch-2"},
					},
				}
				repo2 := &status.Repository{
					Path: "/repos/repo2",
					Worktrees: map[string]status.WorktreeInfo{
						"origin:feature-branch-3": {Remote: "origin", Branch: "feature-branch-3"},
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
				statusMock.EXPECT().GetWorkspace("/path/to").Return(nil, status.ErrWorkspaceNotFound)
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
				deps: &dependencies.Dependencies{
					StatusManager: statusMock,
					Logger:        logger.NewNoopLogger(),
				},
				file: "/path/to/workspace.code-workspace",
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
