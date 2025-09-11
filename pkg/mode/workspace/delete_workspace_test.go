//go:build unit

package workspace

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/lerenn/code-manager/pkg/config"
	fsMocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	"github.com/lerenn/code-manager/pkg/logger"
	promptMocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusMocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/lerenn/code-manager/pkg/worktree"
	worktreeMocks "github.com/lerenn/code-manager/pkg/worktree/mocks"
)

func TestDeleteWorkspace(t *testing.T) {
	tests := []struct {
		name          string
		workspaceName string
		force         bool
		setupMocks    func(*statusMocks.MockManager, *promptMocks.MockPrompter, *worktreeMocks.MockWorktree)
		expectedError string
	}{
		{
			name:          "successful deletion with confirmation",
			workspaceName: "test-workspace",
			force:         false,
			setupMocks: func(statusMock *statusMocks.MockManager, promptMock *promptMocks.MockPrompter, worktreeMock *worktreeMocks.MockWorktree) {
				// Validate workspace exists (called twice - once in DeleteWorkspace, once in ListWorkspaceWorktrees)
				workspace := &status.Workspace{
					Repositories: []string{"https://github.com/user/repo1"},
				}
				statusMock.EXPECT().GetWorkspaceByName("test-workspace").Return(workspace, nil).Times(2).Times(2)

				// List workspace worktrees
				workspaces := map[string]status.Workspace{
					"/path/to/test-workspace.code-workspace": *workspace,
				}
				statusMock.EXPECT().ListWorkspaces().Return(workspaces, nil).Times(2)

				repo1 := &status.Repository{
					Path: "/repos/repo1",
					Worktrees: map[string]status.WorktreeInfo{
						"feature-branch-1": {Remote: "origin", Branch: "feature-branch-1"},
					},
				}
				statusMock.EXPECT().GetRepository("https://github.com/user/repo1").Return(repo1, nil).Times(2)

				// Show confirmation
				promptMock.EXPECT().PromptForConfirmation(gomock.Any(), false).Return(true, nil)

				// Delete worktree
				worktreeMock.EXPECT().Delete(worktree.DeleteParams{
					RepoURL:      "https://github.com/user/repo1",
					Branch:       "feature-branch-1",
					WorktreePath: "/repos/repo1-feature-branch-1",
					RepoPath:     "/repos/repo1",
					Force:        true,
				}).Return(nil)

				// Remove workspace from status
				statusMock.EXPECT().RemoveWorkspace("test-workspace").Return(nil)
			},
		},
		{
			name:          "successful deletion with force flag",
			workspaceName: "test-workspace",
			force:         true,
			setupMocks: func(statusMock *statusMocks.MockManager, promptMock *promptMocks.MockPrompter, worktreeMock *worktreeMocks.MockWorktree) {
				// Validate workspace exists
				workspace := &status.Workspace{
					Repositories: []string{"https://github.com/user/repo1"},
				}
				statusMock.EXPECT().GetWorkspaceByName("test-workspace").Return(workspace, nil).Times(2)

				// List workspace worktrees
				workspaces := map[string]status.Workspace{
					"/path/to/test-workspace.code-workspace": *workspace,
				}
				statusMock.EXPECT().ListWorkspaces().Return(workspaces, nil).Times(2)

				repo1 := &status.Repository{
					Path: "/repos/repo1",
					Worktrees: map[string]status.WorktreeInfo{
						"feature-branch-1": {Remote: "origin", Branch: "feature-branch-1"},
					},
				}
				statusMock.EXPECT().GetRepository("https://github.com/user/repo1").Return(repo1, nil).Times(2)

				// No confirmation prompt expected with force flag

				// Delete worktree
				worktreeMock.EXPECT().Delete(worktree.DeleteParams{
					RepoURL:      "https://github.com/user/repo1",
					Branch:       "feature-branch-1",
					WorktreePath: "/repos/repo1-feature-branch-1",
					RepoPath:     "/repos/repo1",
					Force:        true,
				}).Return(nil)

				// Remove workspace from status
				statusMock.EXPECT().RemoveWorkspace("test-workspace").Return(nil)
			},
		},
		{
			name:          "workspace not found",
			workspaceName: "nonexistent",
			force:         false,
			setupMocks: func(statusMock *statusMocks.MockManager, promptMock *promptMocks.MockPrompter, worktreeMock *worktreeMocks.MockWorktree) {
				statusMock.EXPECT().GetWorkspaceByName("nonexistent").Return(nil, status.ErrWorkspaceNotFound)
			},
			expectedError: "workspace 'nonexistent' not found",
		},
		{
			name:          "user cancels confirmation",
			workspaceName: "test-workspace",
			force:         false,
			setupMocks: func(statusMock *statusMocks.MockManager, promptMock *promptMocks.MockPrompter, worktreeMock *worktreeMocks.MockWorktree) {
				// Validate workspace exists
				workspace := &status.Workspace{
					Repositories: []string{"https://github.com/user/repo1"},
				}
				statusMock.EXPECT().GetWorkspaceByName("test-workspace").Return(workspace, nil).Times(2)

				// List workspace worktrees
				workspaces := map[string]status.Workspace{
					"/path/to/test-workspace.code-workspace": *workspace,
				}
				statusMock.EXPECT().ListWorkspaces().Return(workspaces, nil).Times(2)

				repo1 := &status.Repository{
					Path: "/repos/repo1",
					Worktrees: map[string]status.WorktreeInfo{
						"feature-branch-1": {Remote: "origin", Branch: "feature-branch-1"},
					},
				}
				statusMock.EXPECT().GetRepository("https://github.com/user/repo1").Return(repo1, nil).Times(2)

				// User cancels confirmation
				promptMock.EXPECT().PromptForConfirmation(gomock.Any(), false).Return(false, nil)
			},
			expectedError: "user cancelled deletion",
		},
		{
			name:          "worktree deletion fails",
			workspaceName: "test-workspace",
			force:         true,
			setupMocks: func(statusMock *statusMocks.MockManager, promptMock *promptMocks.MockPrompter, worktreeMock *worktreeMocks.MockWorktree) {
				// Validate workspace exists
				workspace := &status.Workspace{
					Repositories: []string{"https://github.com/user/repo1"},
				}
				statusMock.EXPECT().GetWorkspaceByName("test-workspace").Return(workspace, nil).Times(2)

				// List workspace worktrees
				workspaces := map[string]status.Workspace{
					"/path/to/test-workspace.code-workspace": *workspace,
				}
				statusMock.EXPECT().ListWorkspaces().Return(workspaces, nil).Times(2)

				repo1 := &status.Repository{
					Path: "/repos/repo1",
					Worktrees: map[string]status.WorktreeInfo{
						"feature-branch-1": {Remote: "origin", Branch: "feature-branch-1"},
					},
				}
				statusMock.EXPECT().GetRepository("https://github.com/user/repo1").Return(repo1, nil).Times(2)

				// Worktree deletion fails
				worktreeMock.EXPECT().Delete(gomock.Any()).Return(errors.New("worktree deletion failed"))
			},
			expectedError: "failed to delete worktree https://github.com/user/repo1/feature-branch-1: worktree deletion failed",
		},
		{
			name:          "status removal fails",
			workspaceName: "test-workspace",
			force:         true,
			setupMocks: func(statusMock *statusMocks.MockManager, promptMock *promptMocks.MockPrompter, worktreeMock *worktreeMocks.MockWorktree) {
				// Validate workspace exists
				workspace := &status.Workspace{
					Repositories: []string{"https://github.com/user/repo1"},
				}
				statusMock.EXPECT().GetWorkspaceByName("test-workspace").Return(workspace, nil).Times(2)

				// List workspace worktrees
				workspaces := map[string]status.Workspace{
					"/path/to/test-workspace.code-workspace": *workspace,
				}
				statusMock.EXPECT().ListWorkspaces().Return(workspaces, nil).Times(2)

				repo1 := &status.Repository{
					Path: "/repos/repo1",
					Worktrees: map[string]status.WorktreeInfo{
						"feature-branch-1": {Remote: "origin", Branch: "feature-branch-1"},
					},
				}
				statusMock.EXPECT().GetRepository("https://github.com/user/repo1").Return(repo1, nil).Times(2)

				// Delete worktree successfully
				worktreeMock.EXPECT().Delete(gomock.Any()).Return(nil)

				// Status removal fails
				statusMock.EXPECT().RemoveWorkspace("test-workspace").Return(errors.New("status removal failed"))
			},
			expectedError: "failed to remove workspace from status: status removal failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			statusMock := statusMocks.NewMockManager(ctrl)
			promptMock := promptMocks.NewMockPrompter(ctrl)
			worktreeMock := worktreeMocks.NewMockWorktree(ctrl)
			fsMock := fsMocks.NewMockFS(ctrl)

			tt.setupMocks(statusMock, promptMock, worktreeMock)

			// Setup worktree provider
			worktreeProvider := func(params worktree.NewWorktreeParams) worktree.Worktree {
				return worktreeMock
			}

			workspace := &realWorkspace{
				fs:               fsMock,
				statusManager:    statusMock,
				logger:           logger.NewNoopLogger(),
				prompt:           promptMock,
				worktreeProvider: worktreeProvider,
				config:           config.Config{RepositoriesDir: "/repos"},
			}

			err := workspace.DeleteWorkspace(tt.workspaceName, tt.force)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeleteWorkspaceFiles(t *testing.T) {
	tests := []struct {
		name          string
		workspaceName string
		worktrees     []WorktreeInfo
		setupMocks    func(*fsMocks.MockFS, *statusMocks.MockManager)
		expectedError string
	}{
		{
			name:          "successful file deletion",
			workspaceName: "test-workspace",
			worktrees: []WorktreeInfo{
				{
					WorkspaceFile: "test-workspace-feature-branch-1.code-workspace",
				},
			},
			setupMocks: func(fsMock *fsMocks.MockFS, statusMock *statusMocks.MockManager) {
				workspaces := map[string]status.Workspace{
					"/path/to/test-workspace.code-workspace": {},
				}
				statusMock.EXPECT().ListWorkspaces().Return(workspaces, nil).Times(2)

				// Main workspace file exists and gets deleted
				fsMock.EXPECT().Exists("/path/to/test-workspace.code-workspace").Return(true, nil)
				fsMock.EXPECT().RemoveAll("/path/to/test-workspace.code-workspace").Return(nil)

				// Worktree workspace file exists and gets deleted
				fsMock.EXPECT().Exists("test-workspace-feature-branch-1.code-workspace").Return(true, nil)
				fsMock.EXPECT().RemoveAll("test-workspace-feature-branch-1.code-workspace").Return(nil)
			},
		},
		{
			name:          "files don't exist",
			workspaceName: "test-workspace",
			worktrees:     []WorktreeInfo{},
			setupMocks: func(fsMock *fsMocks.MockFS, statusMock *statusMocks.MockManager) {
				workspaces := map[string]status.Workspace{
					"/path/to/test-workspace.code-workspace": {},
				}
				statusMock.EXPECT().ListWorkspaces().Return(workspaces, nil).Times(2)

				// Main workspace file doesn't exist
				fsMock.EXPECT().Exists("/path/to/test-workspace.code-workspace").Return(false, nil)
			},
		},
		{
			name:          "workspace path not found",
			workspaceName: "test-workspace",
			worktrees:     []WorktreeInfo{},
			setupMocks: func(fsMock *fsMocks.MockFS, statusMock *statusMocks.MockManager) {
				statusMock.EXPECT().ListWorkspaces().Return(map[string]status.Workspace{}, nil)
			},
			expectedError: "failed to determine workspace path for: test-workspace",
		},
		{
			name:          "file deletion fails",
			workspaceName: "test-workspace",
			worktrees:     []WorktreeInfo{},
			setupMocks: func(fsMock *fsMocks.MockFS, statusMock *statusMocks.MockManager) {
				workspaces := map[string]status.Workspace{
					"/path/to/test-workspace.code-workspace": {},
				}
				statusMock.EXPECT().ListWorkspaces().Return(workspaces, nil).Times(2)

				// Main workspace file exists but deletion fails
				fsMock.EXPECT().Exists("/path/to/test-workspace.code-workspace").Return(true, nil)
				fsMock.EXPECT().RemoveAll("/path/to/test-workspace.code-workspace").Return(errors.New("deletion failed"))
			},
			expectedError: "failed to delete main workspace file /path/to/test-workspace.code-workspace: deletion failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			fsMock := fsMocks.NewMockFS(ctrl)
			statusMock := statusMocks.NewMockManager(ctrl)

			tt.setupMocks(fsMock, statusMock)

			workspace := &realWorkspace{
				fs:            fsMock,
				statusManager: statusMock,
				logger:        logger.NewNoopLogger(),
			}

			err := workspace.deleteWorkspaceFiles(tt.workspaceName, tt.worktrees)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
