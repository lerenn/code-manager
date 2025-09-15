//go:build unit

package cm

import (
	"errors"
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	hooksMocks "github.com/lerenn/code-manager/pkg/hooks/mocks"
	"github.com/lerenn/code-manager/pkg/logger"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestDeleteWorkspace_Success tests successful workspace deletion.
func TestDeleteWorkspace_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/test/base/path",
		StatusFile:      "/test/status.yaml",
		WorkspacesDir:   "/test/workspaces",
	}

	cm := &realCM{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
		prompt:        mockPrompt,
		hookManager:   mockHookManager,
	}

	params := DeleteWorkspaceParams{
		WorkspaceName: "test-workspace",
		Force:         false,
	}

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks("delete_workspace", gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks("delete_workspace", gomock.Any()).Return(nil)

	// Mock workspace exists
	workspace := &status.Workspace{
		Worktrees:    []string{"feature-1", "feature-2"},
		Repositories: []string{"repo1", "repo2"},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(workspace, nil).Times(2) // Once for initial check, once for removeWorktreesFromWorkspaceStatus

	// Mock worktree listing
	worktrees := []status.WorktreeInfo{
		{Remote: "origin", Branch: "feature-1"},
		{Remote: "origin", Branch: "feature-2"},
	}
	repo1 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:feature-1": worktrees[0],
		},
	}
	repo2 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:feature-2": worktrees[1],
		},
	}
	// Expect GetRepository calls for listWorkspaceWorktreesFromWorkspace, showDeletionConfirmation, deleteWorkspaceWorktrees, and removeWorktreeFromStatus
	mockStatus.EXPECT().GetRepository(gomock.Any()).DoAndReturn(func(repoName string) (*status.Repository, error) {
		switch repoName {
		case "repo1":
			return repo1, nil
		case "repo2":
			return repo2, nil
		default:
			return nil, errors.New("unknown repository")
		}
	}).AnyTimes()

	// Mock confirmation prompt
	mockPrompt.EXPECT().PromptForConfirmation(gomock.Any(), false).Return(true, nil)

	// Mock worktree path existence checks
	mockFS.EXPECT().Exists("/test/base/path/repo1/origin/feature-1").Return(true, nil)
	mockFS.EXPECT().Exists("/test/base/path/repo2/origin/feature-2").Return(true, nil)

	// Mock worktree existence checks
	mockGit.EXPECT().WorktreeExists(gomock.Any(), "feature-1").Return(true, nil)
	mockGit.EXPECT().WorktreeExists(gomock.Any(), "feature-2").Return(true, nil)

	// Mock worktree deletion
	mockGit.EXPECT().RemoveWorktree(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)

	// Mock status updates for worktree removal
	mockStatus.EXPECT().RemoveWorktree("repo1", "feature-1").Return(nil)
	mockStatus.EXPECT().RemoveWorktree("repo2", "feature-2").Return(nil)

	// Mock workspace update (for removeWorktreesFromWorkspaceStatus)
	mockStatus.EXPECT().UpdateWorkspace("test-workspace", gomock.Any()).Return(nil)

	// Mock workspace file deletion - main workspace file first, then worktree-specific files
	mockFS.EXPECT().Exists("/test/workspaces/test-workspace.code-workspace").Return(true, nil)
	mockFS.EXPECT().Remove("/test/workspaces/test-workspace.code-workspace").Return(nil)
	mockFS.EXPECT().Exists("/test/workspaces/test-workspace-feature-1.code-workspace").Return(true, nil)
	mockFS.EXPECT().Remove("/test/workspaces/test-workspace-feature-1.code-workspace").Return(nil)
	mockFS.EXPECT().Exists("/test/workspaces/test-workspace-feature-2.code-workspace").Return(true, nil)
	mockFS.EXPECT().Remove("/test/workspaces/test-workspace-feature-2.code-workspace").Return(nil)

	// Mock workspace removal from status
	mockStatus.EXPECT().RemoveWorkspace("test-workspace").Return(nil)

	// Execute
	err := cm.DeleteWorkspace(params)

	// Assert
	assert.NoError(t, err)
}

// TestDeleteWorkspace_Force tests workspace deletion with force flag.
func TestDeleteWorkspace_Force(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/test/base/path",
		StatusFile:      "/test/status.yaml",
		WorkspacesDir:   "/test/workspaces",
	}

	cm := &realCM{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
	}

	params := DeleteWorkspaceParams{
		WorkspaceName: "test-workspace",
		Force:         true,
	}

	// Mock workspace exists
	workspace := &status.Workspace{
		Worktrees:    []string{"feature-1"},
		Repositories: []string{"repo1"},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(workspace, nil).Times(2) // Once for initial check, once for removeWorktreesFromWorkspaceStatus

	// Mock worktree listing
	worktrees := []status.WorktreeInfo{
		{Remote: "origin", Branch: "feature-1"},
	}
	repo1 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:feature-1": worktrees[0],
		},
	}
	// Expect GetRepository calls for listWorkspaceWorktreesFromWorkspace, deleteWorkspaceWorktrees, and removeWorktreeFromStatus (no confirmation for force)
	mockStatus.EXPECT().GetRepository("repo1").Return(repo1, nil).Times(3) // 1 for list, 1 for deletion, 1 for removeWorktreeFromStatus

	// No confirmation prompt expected due to force flag

	// Mock worktree path existence check
	mockFS.EXPECT().Exists("/test/base/path/repo1/origin/feature-1").Return(true, nil)

	// Mock worktree existence check
	mockGit.EXPECT().WorktreeExists(gomock.Any(), "feature-1").Return(true, nil)

	// Mock worktree deletion
	mockGit.EXPECT().RemoveWorktree(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	// Mock status updates for worktree removal
	mockStatus.EXPECT().RemoveWorktree("repo1", "feature-1").Return(nil)

	// Mock workspace update (for removeWorktreesFromWorkspaceStatus)
	mockStatus.EXPECT().UpdateWorkspace("test-workspace", gomock.Any()).Return(nil)

	// Mock workspace file deletion - main workspace file first, then worktree-specific files
	mockFS.EXPECT().Exists("/test/workspaces/test-workspace.code-workspace").Return(true, nil)
	mockFS.EXPECT().Remove("/test/workspaces/test-workspace.code-workspace").Return(nil)
	mockFS.EXPECT().Exists("/test/workspaces/test-workspace-feature-1.code-workspace").Return(true, nil)
	mockFS.EXPECT().Remove("/test/workspaces/test-workspace-feature-1.code-workspace").Return(nil)

	// Mock workspace removal from status
	mockStatus.EXPECT().RemoveWorkspace("test-workspace").Return(nil)

	// Execute
	err := cm.DeleteWorkspace(params)

	// Assert
	assert.NoError(t, err)
}

// TestDeleteWorkspace_NotFound tests workspace deletion when workspace doesn't exist.
func TestDeleteWorkspace_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/test/base/path",
		StatusFile:      "/test/status.yaml",
		WorkspacesDir:   "/test/workspaces",
	}

	cm := &realCM{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
	}

	params := DeleteWorkspaceParams{
		WorkspaceName: "nonexistent-workspace",
		Force:         true,
	}

	// Mock workspace doesn't exist
	mockStatus.EXPECT().GetWorkspace("nonexistent-workspace").Return(nil, errors.New("not found"))

	// Execute
	err := cm.DeleteWorkspace(params)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace 'nonexistent-workspace' not found")
}

// TestDeleteWorkspace_InvalidName tests workspace deletion with invalid workspace name.
func TestDeleteWorkspace_InvalidName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/test/base/path",
		StatusFile:      "/test/status.yaml",
		WorkspacesDir:   "/test/workspaces",
	}

	cm := &realCM{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
	}

	params := DeleteWorkspaceParams{
		WorkspaceName: "", // Empty name
		Force:         true,
	}

	// Execute
	err := cm.DeleteWorkspace(params)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid workspace name")
}

// TestDeleteWorkspace_ConfirmationCancelled tests workspace deletion when user cancels confirmation.
func TestDeleteWorkspace_ConfirmationCancelled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/test/base/path",
		StatusFile:      "/test/status.yaml",
		WorkspacesDir:   "/test/workspaces",
	}

	cm := &realCM{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
		prompt:        mockPrompt,
	}

	params := DeleteWorkspaceParams{
		WorkspaceName: "test-workspace",
		Force:         false,
	}

	// Mock workspace exists
	workspace := &status.Workspace{
		Worktrees:    []string{"feature-1"},
		Repositories: []string{"repo1"},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(workspace, nil) // Only once since confirmation is cancelled

	// Mock worktree listing
	worktrees := []status.WorktreeInfo{
		{Remote: "origin", Branch: "feature-1"},
	}
	repo1 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:feature-1": worktrees[0],
		},
	}
	// Expect GetRepository calls for listWorkspaceWorktreesFromWorkspace and showDeletionConfirmation
	// listWorkspaceWorktreesFromWorkspace: feature-1 in repo1
	// showDeletionConfirmation: repo1
	mockStatus.EXPECT().GetRepository("repo1").Return(repo1, nil).Times(2)

	// Mock confirmation prompt returns "no"
	mockPrompt.EXPECT().PromptForConfirmation(gomock.Any(), false).Return(false, nil)

	// Execute
	err := cm.DeleteWorkspace(params)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deletion cancelled")
}

// TestDeleteWorkspace_WorktreeDeletionFailure tests workspace deletion when worktree deletion fails.
func TestDeleteWorkspace_WorktreeDeletionFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/test/base/path",
		StatusFile:      "/test/status.yaml",
		WorkspacesDir:   "/test/workspaces",
	}

	cm := &realCM{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
	}

	params := DeleteWorkspaceParams{
		WorkspaceName: "test-workspace",
		Force:         true,
	}

	// Mock workspace exists
	workspace := &status.Workspace{
		Worktrees:    []string{"feature-1"},
		Repositories: []string{"repo1"},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(workspace, nil) // Only once since worktree deletion fails

	// Mock worktree listing
	worktrees := []status.WorktreeInfo{
		{Remote: "origin", Branch: "feature-1"},
	}
	repo1 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:feature-1": worktrees[0],
		},
	}
	// Expect GetRepository calls for listWorkspaceWorktreesFromWorkspace and deleteWorkspaceWorktrees (no confirmation for force)
	mockStatus.EXPECT().GetRepository("repo1").Return(repo1, nil).Times(2) // 1 for list, 1 for deletion

	// Mock worktree path existence check
	mockFS.EXPECT().Exists("/test/base/path/repo1/origin/feature-1").Return(true, nil)

	// Mock worktree existence check
	mockGit.EXPECT().WorktreeExists(gomock.Any(), "feature-1").Return(true, nil)

	// Mock worktree deletion failure
	mockGit.EXPECT().RemoveWorktree(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("deletion failed"))

	// No file operations expected since worktree deletion fails and stops the process

	// Execute
	err := cm.DeleteWorkspace(params)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove worktree")
}

// TestDeleteWorkspace_FileDeletionFailure tests workspace deletion when file deletion fails.
func TestDeleteWorkspace_FileDeletionFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/test/base/path",
		StatusFile:      "/test/status.yaml",
		WorkspacesDir:   "/test/workspaces",
	}

	cm := &realCM{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
	}

	params := DeleteWorkspaceParams{
		WorkspaceName: "test-workspace",
		Force:         true,
	}

	// Mock workspace exists
	workspace := &status.Workspace{
		Worktrees:    []string{"feature-1"},
		Repositories: []string{"repo1"},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(workspace, nil).Times(2) // Once for initial check, once for removeWorktreesFromWorkspaceStatus

	// Mock worktree listing
	worktrees := []status.WorktreeInfo{
		{Remote: "origin", Branch: "feature-1"},
	}
	repo1 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:feature-1": worktrees[0],
		},
	}
	// Expect GetRepository calls for listWorkspaceWorktreesFromWorkspace and deleteWorkspaceWorktrees (no confirmation for force)
	// Note: deleteWorkspaceWorktrees calls GetRepository again for each repo
	mockStatus.EXPECT().GetRepository(gomock.Any()).DoAndReturn(func(repoName string) (*status.Repository, error) {
		switch repoName {
		case "repo1":
			return repo1, nil
		default:
			return nil, errors.New("unknown repository")
		}
	}).AnyTimes()

	// Mock worktree path existence check
	mockFS.EXPECT().Exists("/test/base/path/repo1/origin/feature-1").Return(true, nil)

	// Mock worktree existence check
	mockGit.EXPECT().WorktreeExists(gomock.Any(), "feature-1").Return(true, nil)

	// Mock worktree deletion
	mockGit.EXPECT().RemoveWorktree(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	// Mock status updates for worktree removal
	mockStatus.EXPECT().RemoveWorktree("repo1", "feature-1").Return(nil)

	// Mock workspace update (for removeWorktreesFromWorkspaceStatus)
	mockStatus.EXPECT().UpdateWorkspace("test-workspace", gomock.Any()).Return(nil)

	// Mock workspace file deletion failure
	mockFS.EXPECT().Exists("/test/workspaces/test-workspace.code-workspace").Return(true, nil)
	mockFS.EXPECT().Remove("/test/workspaces/test-workspace.code-workspace").Return(errors.New("file deletion failed"))

	// Execute
	err := cm.DeleteWorkspace(params)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete workspace files")
}

// TestDeleteWorkspace_StatusRemovalFailure tests workspace deletion when status removal fails.
func TestDeleteWorkspace_StatusRemovalFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/test/base/path",
		StatusFile:      "/test/status.yaml",
		WorkspacesDir:   "/test/workspaces",
	}

	cm := &realCM{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
	}

	params := DeleteWorkspaceParams{
		WorkspaceName: "test-workspace",
		Force:         true,
	}

	// Mock workspace exists
	workspace := &status.Workspace{
		Worktrees:    []string{"feature-1"},
		Repositories: []string{"repo1"},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(workspace, nil).Times(2) // Once for initial check, once for removeWorktreesFromWorkspaceStatus

	// Mock worktree listing
	worktrees := []status.WorktreeInfo{
		{Remote: "origin", Branch: "feature-1"},
	}
	repo1 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:feature-1": worktrees[0],
		},
	}
	// Expect GetRepository calls for listWorkspaceWorktreesFromWorkspace, deleteWorkspaceWorktrees, and removeWorktreeFromStatus (no confirmation for force)
	mockStatus.EXPECT().GetRepository("repo1").Return(repo1, nil).Times(3) // 1 for list, 1 for deletion, 1 for removeWorktreeFromStatus

	// Mock worktree path existence check
	mockFS.EXPECT().Exists("/test/base/path/repo1/origin/feature-1").Return(true, nil)

	// Mock worktree existence check
	mockGit.EXPECT().WorktreeExists(gomock.Any(), "feature-1").Return(true, nil)

	// Mock worktree deletion
	mockGit.EXPECT().RemoveWorktree(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	// Mock status updates for worktree removal
	mockStatus.EXPECT().RemoveWorktree("repo1", "feature-1").Return(nil)

	// Mock workspace update (for removeWorktreesFromWorkspaceStatus)
	mockStatus.EXPECT().UpdateWorkspace("test-workspace", gomock.Any()).Return(nil)

	// Mock workspace file deletion
	mockFS.EXPECT().Exists("/test/workspaces/test-workspace.code-workspace").Return(true, nil)
	mockFS.EXPECT().Remove("/test/workspaces/test-workspace.code-workspace").Return(nil)
	mockFS.EXPECT().Exists("/test/workspaces/test-workspace-feature-1.code-workspace").Return(true, nil)
	mockFS.EXPECT().Remove("/test/workspaces/test-workspace-feature-1.code-workspace").Return(nil)

	// Mock workspace removal from status failure
	mockStatus.EXPECT().RemoveWorkspace("test-workspace").Return(errors.New("status removal failed"))

	// Execute
	err := cm.DeleteWorkspace(params)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove workspace from status")
}

// TestDeleteWorkspace_EmptyWorkspace tests workspace deletion with no worktrees.
func TestDeleteWorkspace_EmptyWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/test/base/path",
		StatusFile:      "/test/status.yaml",
		WorkspacesDir:   "/test/workspaces",
	}

	cm := &realCM{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
		prompt:        mockPrompt,
	}

	params := DeleteWorkspaceParams{
		WorkspaceName: "empty-workspace",
		Force:         false,
	}

	// Mock workspace exists with no worktrees
	workspace := &status.Workspace{
		Worktrees:    []string{},
		Repositories: []string{"repo1"},
	}
	mockStatus.EXPECT().GetWorkspace("empty-workspace").Return(workspace, nil).Times(2) // Once for initial check, once for removeWorktreesFromWorkspaceStatus

	// Mock worktree listing returns empty
	repo1 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{},
	}
	// Expect GetRepository calls for listWorkspaceWorktreesFromWorkspace and showDeletionConfirmation
	mockStatus.EXPECT().GetRepository(gomock.Any()).DoAndReturn(func(repoName string) (*status.Repository, error) {
		switch repoName {
		case "repo1":
			return repo1, nil
		default:
			return nil, errors.New("unknown repository")
		}
	}).AnyTimes()

	// Mock confirmation prompt
	mockPrompt.EXPECT().PromptForConfirmation(gomock.Any(), false).Return(true, nil)

	// Mock workspace file deletion
	mockFS.EXPECT().Exists("/test/workspaces/empty-workspace.code-workspace").Return(true, nil)
	mockFS.EXPECT().Remove("/test/workspaces/empty-workspace.code-workspace").Return(nil)

	// Mock workspace update (for removeWorktreesFromWorkspaceStatus)
	mockStatus.EXPECT().UpdateWorkspace("empty-workspace", gomock.Any()).Return(nil)

	// Mock workspace removal from status
	mockStatus.EXPECT().RemoveWorkspace("empty-workspace").Return(nil)

	// Execute
	err := cm.DeleteWorkspace(params)

	// Assert
	assert.NoError(t, err)
}

// TestDeleteWorkspace_ConfirmationPromptError tests workspace deletion when confirmation prompt fails.
func TestDeleteWorkspace_ConfirmationPromptError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/test/base/path",
		StatusFile:      "/test/status.yaml",
		WorkspacesDir:   "/test/workspaces",
	}

	cm := &realCM{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
		prompt:        mockPrompt,
	}

	params := DeleteWorkspaceParams{
		WorkspaceName: "test-workspace",
		Force:         false,
	}

	// Mock workspace exists
	workspace := &status.Workspace{
		Worktrees:    []string{"feature-1"},
		Repositories: []string{"repo1"},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(workspace, nil) // Only once since prompt fails

	// Mock worktree listing
	worktrees := []status.WorktreeInfo{
		{Remote: "origin", Branch: "feature-1"},
	}
	repo1 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:feature-1": worktrees[0],
		},
	}
	// Expect GetRepository calls for listWorkspaceWorktreesFromWorkspace and showDeletionConfirmation
	mockStatus.EXPECT().GetRepository("repo1").Return(repo1, nil).Times(2)

	// Mock confirmation prompt error
	mockPrompt.EXPECT().PromptForConfirmation(gomock.Any(), false).Return(false, errors.New("prompt failed"))

	// Execute
	err := cm.DeleteWorkspace(params)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user confirmation")
}

// TestDeleteWorkspace_WorktreeStatusRemovalFailure tests workspace deletion when worktree status removal fails.
func TestDeleteWorkspace_WorktreeStatusRemovalFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/test/base/path",
		StatusFile:      "/test/status.yaml",
		WorkspacesDir:   "/test/workspaces",
	}

	cm := &realCM{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
	}

	params := DeleteWorkspaceParams{
		WorkspaceName: "test-workspace",
		Force:         true,
	}

	// Mock workspace exists
	workspace := &status.Workspace{
		Worktrees:    []string{"feature-1"},
		Repositories: []string{"repo1"},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(workspace, nil) // Only once since status removal fails

	// Mock worktree listing
	worktrees := []status.WorktreeInfo{
		{Remote: "origin", Branch: "feature-1"},
	}
	repo1 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:feature-1": worktrees[0],
		},
	}
	mockStatus.EXPECT().GetRepository("repo1").Return(repo1, nil).Times(3) // 1 for list, 1 for deletion, 1 for removeWorktreeFromStatus

	// Mock worktree path existence check
	mockFS.EXPECT().Exists("/test/base/path/repo1/origin/feature-1").Return(true, nil)

	// Mock worktree existence check
	mockGit.EXPECT().WorktreeExists(gomock.Any(), "feature-1").Return(true, nil)

	// Mock worktree deletion
	mockGit.EXPECT().RemoveWorktree(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	// Mock status updates for worktree removal failure
	mockStatus.EXPECT().RemoveWorktree("repo1", "feature-1").Return(errors.New("status removal failed"))

	// Execute
	err := cm.DeleteWorkspace(params)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete workspace worktrees")
}

// TestDeleteWorkspace_MultipleRepositories tests workspace deletion with multiple repositories.
func TestDeleteWorkspace_MultipleRepositories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/test/base/path",
		StatusFile:      "/test/status.yaml",
		WorkspacesDir:   "/test/workspaces",
	}

	cm := &realCM{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
		hookManager:   mockHookManager,
	}

	params := DeleteWorkspaceParams{
		WorkspaceName: "multi-repo-workspace",
		Force:         true,
	}

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks("delete_workspace", gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks("delete_workspace", gomock.Any()).Return(nil)

	// Mock workspace exists with multiple repositories
	workspace := &status.Workspace{
		Worktrees:    []string{"feature-1", "feature-2"},
		Repositories: []string{"repo1", "repo2"},
	}
	mockStatus.EXPECT().GetWorkspace("multi-repo-workspace").Return(workspace, nil).Times(2) // Once for initial check, once for removeWorktreesFromWorkspaceStatus

	// Mock worktree listing
	worktrees := []status.WorktreeInfo{
		{Remote: "origin", Branch: "feature-1"},
		{Remote: "origin", Branch: "feature-2"},
	}
	repo1 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:feature-1": worktrees[0],
		},
	}
	repo2 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:feature-2": worktrees[1],
		},
	}
	// Expect GetRepository calls for listWorkspaceWorktreesFromWorkspace, deleteWorkspaceWorktrees, and removeWorktreeFromStatus
	mockStatus.EXPECT().GetRepository(gomock.Any()).DoAndReturn(func(repoName string) (*status.Repository, error) {
		switch repoName {
		case "repo1":
			return repo1, nil
		case "repo2":
			return repo2, nil
		default:
			return nil, errors.New("unknown repository")
		}
	}).AnyTimes()

	// Mock worktree path existence checks
	mockFS.EXPECT().Exists("/test/base/path/repo1/origin/feature-1").Return(true, nil)
	mockFS.EXPECT().Exists("/test/base/path/repo2/origin/feature-2").Return(true, nil)

	// Mock worktree existence checks
	mockGit.EXPECT().WorktreeExists(gomock.Any(), "feature-1").Return(true, nil)
	mockGit.EXPECT().WorktreeExists(gomock.Any(), "feature-2").Return(true, nil)

	// Mock worktree deletion
	mockGit.EXPECT().RemoveWorktree(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)

	// Mock status updates for worktree removal
	mockStatus.EXPECT().RemoveWorktree("repo1", "feature-1").Return(nil)
	mockStatus.EXPECT().RemoveWorktree("repo2", "feature-2").Return(nil)

	// Mock workspace update (for removeWorktreesFromWorkspaceStatus)
	mockStatus.EXPECT().UpdateWorkspace("multi-repo-workspace", gomock.Any()).Return(nil)

	// Mock workspace file deletion
	mockFS.EXPECT().Exists("/test/workspaces/multi-repo-workspace.code-workspace").Return(true, nil)
	mockFS.EXPECT().Remove("/test/workspaces/multi-repo-workspace.code-workspace").Return(nil)
	mockFS.EXPECT().Exists("/test/workspaces/multi-repo-workspace-feature-1.code-workspace").Return(true, nil)
	mockFS.EXPECT().Remove("/test/workspaces/multi-repo-workspace-feature-1.code-workspace").Return(nil)
	mockFS.EXPECT().Exists("/test/workspaces/multi-repo-workspace-feature-2.code-workspace").Return(true, nil)
	mockFS.EXPECT().Remove("/test/workspaces/multi-repo-workspace-feature-2.code-workspace").Return(nil)

	// Mock workspace removal from status
	mockStatus.EXPECT().RemoveWorkspace("multi-repo-workspace").Return(nil)

	// Execute
	err := cm.DeleteWorkspace(params)

	// Assert
	assert.NoError(t, err)
}

// TestDeleteWorkspace_WorktreeNotInGit tests workspace deletion when worktree doesn't exist in Git
func TestDeleteWorkspace_WorktreeNotInGit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/test/base/path",
		StatusFile:      "/test/status.yaml",
		WorkspacesDir:   "/test/workspaces",
	}

	cm := &realCM{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
		prompt:        mockPrompt,
		hookManager:   mockHookManager,
	}

	params := DeleteWorkspaceParams{
		WorkspaceName: "test-workspace",
		Force:         true,
	}

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks("delete_workspace", gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks("delete_workspace", gomock.Any()).Return(nil)

	// Mock workspace exists
	workspace := &status.Workspace{
		Worktrees:    []string{"feature-1"},
		Repositories: []string{"repo1"},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(workspace, nil).Times(2) // Once for initial check, once for removeWorktreesFromWorkspaceStatus

	// Mock worktree listing
	worktrees := []status.WorktreeInfo{
		{Remote: "origin", Branch: "feature-1"},
	}
	repo1 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:feature-1": worktrees[0],
		},
	}
	// Expect GetRepository calls for listWorkspaceWorktreesFromWorkspace, deleteWorkspaceWorktrees, and removeWorktreeFromStatus (no confirmation for force)
	mockStatus.EXPECT().GetRepository("repo1").Return(repo1, nil).Times(3) // 1 for list, 1 for deletion, 1 for removeWorktreeFromStatus

	// Mock worktree path existence check
	mockFS.EXPECT().Exists("/test/base/path/repo1/origin/feature-1").Return(true, nil)

	// Mock worktree existence check - worktree doesn't exist in Git
	mockGit.EXPECT().WorktreeExists(gomock.Any(), "feature-1").Return(false, nil)

	// No RemoveWorktree call expected since worktree doesn't exist in Git

	// Mock status updates for worktree removal
	mockStatus.EXPECT().RemoveWorktree("repo1", "feature-1").Return(nil)

	// Mock workspace update (for removeWorktreesFromWorkspaceStatus)
	mockStatus.EXPECT().UpdateWorkspace("test-workspace", gomock.Any()).Return(nil)

	// Mock workspace file deletion - main workspace file first, then worktree-specific files
	mockFS.EXPECT().Exists("/test/workspaces/test-workspace.code-workspace").Return(true, nil)
	mockFS.EXPECT().Remove("/test/workspaces/test-workspace.code-workspace").Return(nil)
	mockFS.EXPECT().Exists("/test/workspaces/test-workspace-feature-1.code-workspace").Return(true, nil)
	mockFS.EXPECT().Remove("/test/workspaces/test-workspace-feature-1.code-workspace").Return(nil)

	// Mock workspace removal from status
	mockStatus.EXPECT().RemoveWorkspace("test-workspace").Return(nil)

	// Execute
	err := cm.DeleteWorkspace(params)

	// Assert
	assert.NoError(t, err)
}
