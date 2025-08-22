//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/ide"
	"github.com/lerenn/code-manager/pkg/repository"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/workspace"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCM_ListWorktrees_NoRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repository.NewMockRepository(ctrl)
	mockWorkspace := workspace.NewMockWorkspace(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	// Create CM with mocked dependencies
	cm := NewCMWithDependencies(NewCMParams{
		Repository: mockRepository,
		Workspace:  mockWorkspace,
		Config:     createTestConfig(),
	})

	// Override IDE manager with mock
	c := cm.(*realCM)
	c.ideManager = mockIDE

	// Mock repository detection to return false (no repository)
	mockRepository.EXPECT().IsGitRepository().Return(false, nil).AnyTimes()

	result, _, err := cm.ListWorktrees(false)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNoGitRepositoryOrWorkspaceFound)
	assert.Nil(t, result)
}

func TestCM_ListWorktrees_SingleRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repository.NewMockRepository(ctrl)
	mockWorkspace := workspace.NewMockWorkspace(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	// Create CM with mocked dependencies
	cm := NewCMWithDependencies(NewCMParams{
		Repository: mockRepository,
		Workspace:  mockWorkspace,
		Config:     createTestConfig(),
	})

	// Override IDE manager with mock
	c := cm.(*realCM)
	c.ideManager = mockIDE

	// Mock repository detection and list worktrees
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	expectedWorktrees := []status.WorktreeInfo{
		{Remote: "origin", Branch: "main"},
		{Remote: "origin", Branch: "feature"},
	}
	mockRepository.EXPECT().ListWorktrees().Return(expectedWorktrees, nil)

	result, projectType, err := cm.ListWorktrees(false)
	assert.NoError(t, err)
	assert.Equal(t, ProjectTypeSingleRepo, projectType)
	assert.Equal(t, expectedWorktrees, result)
}
