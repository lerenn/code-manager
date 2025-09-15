//go:build unit

package codemanager

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/dependencies"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	hooksMocks "github.com/lerenn/code-manager/pkg/hooks/mocks"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	repositoryMocks "github.com/lerenn/code-manager/pkg/mode/repository/mocks"
	"github.com/lerenn/code-manager/pkg/mode/workspace"
	workspaceMocks "github.com/lerenn/code-manager/pkg/mode/workspace/mocks"
	promptMocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusMocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/lerenn/code-manager/pkg/worktree"
	worktreemocks "github.com/lerenn/code-manager/pkg/worktree/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCM_CreateWorkTree_SingleRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	mockPrompt := promptMocks.NewMockPrompter(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCodeManager(NewCodeManagerParams{
		Dependencies: dependencies.New().
			WithHookManager(mockHookManager).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithFS(mockFS).
			WithGit(mockGit).
			WithStatusManager(mockStatus).
			WithPrompt(mockPrompt).
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository {
				return mockRepository
			}).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace {
				return mockWorkspace
			}).
			WithWorktreeProvider(func(params worktree.NewWorktreeParams) worktree.Worktree {
				return mockWorktree
			}).
			WithHookManager(mockHookManager),
	})
	assert.NoError(t, err)

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree creation
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().Validate().Return(nil)
	mockRepository.EXPECT().CreateWorktree("test-branch", gomock.Any()).Return("/test/base/path/test-repo/origin/test-branch", nil)

	err = cm.CreateWorkTree("test-branch")
	assert.NoError(t, err)
}

func TestCM_CreateWorkTreeWithIDE(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	mockPrompt := promptMocks.NewMockPrompter(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCodeManager(NewCodeManagerParams{
		Dependencies: dependencies.New().
			WithHookManager(mockHookManager).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithFS(mockFS).
			WithGit(mockGit).
			WithStatusManager(mockStatus).
			WithPrompt(mockPrompt).
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository {
				return mockRepository
			}).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace {
				return mockWorkspace
			}).
			WithWorktreeProvider(func(params worktree.NewWorktreeParams) worktree.Worktree {
				return mockWorktree
			}).
			WithHookManager(mockHookManager),
	})
	assert.NoError(t, err)

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree creation
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().Validate().Return(nil)
	mockRepository.EXPECT().CreateWorktree("test-branch", gomock.Any()).Return("/test/base/path/test-repo/origin/test-branch", nil)

	// Note: IDE opening is now handled by the hook system, not tested here
	err = cm.CreateWorkTree("test-branch", CreateWorkTreeOpts{IDEName: "vscode"})
	assert.NoError(t, err)
}

func TestCM_CreateWorkTree_WorkspaceMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	mockPrompt := promptMocks.NewMockPrompter(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCodeManager(NewCodeManagerParams{
		Dependencies: dependencies.New().
			WithHookManager(mockHookManager).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithFS(mockFS).
			WithGit(mockGit).
			WithStatusManager(mockStatus).
			WithPrompt(mockPrompt).
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository {
				return mockRepository
			}).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace {
				return mockWorkspace
			}).
			WithWorktreeProvider(func(params worktree.NewWorktreeParams) worktree.Worktree {
				return mockWorktree
			}).
			WithHookManager(mockHookManager),
	})
	assert.NoError(t, err)

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)

	// Mock workspace worktree creation
	workspaceName := "test-workspace"
	expectedPath := "/test/workspaces/test-workspace-feature-branch.code-workspace"
	mockWorkspace.EXPECT().CreateWorktree("feature-branch", gomock.Any()).Return(expectedPath, nil)

	opts := CreateWorkTreeOpts{WorkspaceName: workspaceName}
	err = cm.CreateWorkTree("feature-branch", opts)
	assert.NoError(t, err)
}

func TestCM_CreateWorkTree_WorkspaceModeWithIDE(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	mockPrompt := promptMocks.NewMockPrompter(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCodeManager(NewCodeManagerParams{
		Dependencies: dependencies.New().
			WithHookManager(mockHookManager).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithFS(mockFS).
			WithGit(mockGit).
			WithStatusManager(mockStatus).
			WithPrompt(mockPrompt).
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository {
				return mockRepository
			}).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace {
				return mockWorkspace
			}).
			WithWorktreeProvider(func(params worktree.NewWorktreeParams) worktree.Worktree {
				return mockWorktree
			}).
			WithHookManager(mockHookManager),
	})
	assert.NoError(t, err)

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)

	// Mock workspace worktree creation
	workspaceName := "test-workspace"
	expectedPath := "/test/workspaces/test-workspace-feature-branch.code-workspace"
	mockWorkspace.EXPECT().CreateWorktree("feature-branch", gomock.Any()).Return(expectedPath, nil)

	opts := CreateWorkTreeOpts{
		WorkspaceName: workspaceName,
		IDEName:       "vscode",
	}
	err = cm.CreateWorkTree("feature-branch", opts)
	assert.NoError(t, err)
}

func TestCM_CreateWorkTree_WorkspaceModeFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	mockPrompt := promptMocks.NewMockPrompter(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCodeManager(NewCodeManagerParams{
		Dependencies: dependencies.New().
			WithHookManager(mockHookManager).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithFS(mockFS).
			WithGit(mockGit).
			WithStatusManager(mockStatus).
			WithPrompt(mockPrompt).
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository {
				return mockRepository
			}).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace {
				return mockWorkspace
			}).
			WithWorktreeProvider(func(params worktree.NewWorktreeParams) worktree.Worktree {
				return mockWorktree
			}).
			WithHookManager(mockHookManager),
	})
	assert.NoError(t, err)

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecuteErrorHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)

	// Mock workspace worktree creation failure
	workspaceName := "test-workspace"
	mockWorkspace.EXPECT().CreateWorktree("feature-branch", gomock.Any()).Return("", assert.AnError)

	opts := CreateWorkTreeOpts{WorkspaceName: workspaceName}
	err = cm.CreateWorkTree("feature-branch", opts)
	assert.Error(t, err)
}

func TestCreateWorkTreeWithRepositoryName(t *testing.T) {
	tests := []struct {
		name          string
		branch        string
		opts          CreateWorkTreeOpts
		expectedError string
	}{
		{
			name:   "error when both workspace and repository specified",
			branch: "feature-branch",
			opts: CreateWorkTreeOpts{
				RepositoryName: "/path/to/repo",
				WorkspaceName:  "test-workspace",
			},
			expectedError: "cannot specify both WorkspaceName and RepositoryName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := fsmocks.NewMockFS(ctrl)
			mockGit := gitmocks.NewMockGit(ctrl)
			mockStatus := statusMocks.NewMockManager(ctrl)
			mockRepo := repositoryMocks.NewMockRepository(ctrl)
			mockHooks := hooksMocks.NewMockHookManagerInterface(ctrl)

			// Setup default mocks
			mockHooks.EXPECT().ExecutePreHooks(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			mockHooks.EXPECT().ExecutePostHooks(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			// Create CM instance
			cmInstance := &realCodeManager{
				deps: dependencies.New().
					WithFS(mockFS).
					WithGit(mockGit).
					WithStatusManager(mockStatus).
					WithLogger(logger.NewNoopLogger()).
					WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository {
						return mockRepo
					}).
					WithHookManager(mockHooks).
					WithConfig(config.NewConfigManager("/test/config.yaml")),
			}

			// Execute
			err := cmInstance.CreateWorkTree(tt.branch, tt.opts)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateWorkTreeRepositoryResolution(t *testing.T) {
	tests := []struct {
		name           string
		repositoryName string
		expectedError  string
		setupMocks     func(*fsmocks.MockFS, *gitmocks.MockGit, *statusMocks.MockManager)
	}{
		{
			name:           "repository name from status",
			repositoryName: "my-repo",
			setupMocks: func(fs *fsmocks.MockFS, git *gitmocks.MockGit, statusMgr *statusMocks.MockManager) {
				// Mock repository found in status
				mockRepo := &status.Repository{
					Path: "/path/to/my-repo",
				}
				statusMgr.EXPECT().GetRepository("my-repo").Return(mockRepo, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := fsmocks.NewMockFS(ctrl)
			mockGit := gitmocks.NewMockGit(ctrl)
			mockStatus := statusMocks.NewMockManager(ctrl)
			mockHooks := hooksMocks.NewMockHookManagerInterface(ctrl)

			// Setup default mocks
			mockHooks.EXPECT().ExecutePreHooks(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			mockHooks.EXPECT().ExecutePostHooks(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, mockGit, mockStatus)
			}

			// Create CM instance
			cmInstance := &realCodeManager{
				deps: dependencies.New().
					WithFS(mockFS).
					WithGit(mockGit).
					WithStatusManager(mockStatus).
					WithLogger(logger.NewNoopLogger()).
					WithHookManager(mockHooks).
					WithConfig(config.NewConfigManager("/test/config.yaml")),
			}

			// Test repository resolution
			_, err := cmInstance.resolveRepository(tt.repositoryName)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
