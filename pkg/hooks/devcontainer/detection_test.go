//go:build unit

package devcontainer

import (
	"testing"

	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDetector_DetectDevcontainer_WithDevcontainerDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	detector := NewDetector(mockFS)

	repoPath := "/test/repo"
	devcontainerPath := "/test/repo/.devcontainer/devcontainer.json"

	// Mock expectations
	mockFS.EXPECT().Exists(devcontainerPath).Return(true, nil)

	result, err := detector.DetectDevcontainer(repoPath)
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestDetector_DetectDevcontainer_WithRootDevcontainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	detector := NewDetector(mockFS)

	repoPath := "/test/repo"
	devcontainerPath := "/test/repo/.devcontainer/devcontainer.json"
	rootDevcontainerPath := "/test/repo/.devcontainer.json"

	// Mock expectations - first check returns false, second returns true
	mockFS.EXPECT().Exists(devcontainerPath).Return(false, nil)
	mockFS.EXPECT().Exists(rootDevcontainerPath).Return(true, nil)

	result, err := detector.DetectDevcontainer(repoPath)
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestDetector_DetectDevcontainer_NoDevcontainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	detector := NewDetector(mockFS)

	repoPath := "/test/repo"
	devcontainerPath := "/test/repo/.devcontainer/devcontainer.json"
	rootDevcontainerPath := "/test/repo/.devcontainer.json"

	// Mock expectations - both checks return false
	mockFS.EXPECT().Exists(devcontainerPath).Return(false, nil)
	mockFS.EXPECT().Exists(rootDevcontainerPath).Return(false, nil)

	result, err := detector.DetectDevcontainer(repoPath)
	assert.NoError(t, err)
	assert.False(t, result)
}

func TestDetector_DetectDevcontainer_ErrorOnFirstCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	detector := NewDetector(mockFS)

	repoPath := "/test/repo"
	devcontainerPath := "/test/repo/.devcontainer/devcontainer.json"

	// Mock expectations - first check returns error
	mockFS.EXPECT().Exists(devcontainerPath).Return(false, assert.AnError)

	result, err := detector.DetectDevcontainer(repoPath)
	assert.Error(t, err)
	assert.False(t, result)
}

func TestDetector_DetectDevcontainer_ErrorOnSecondCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	detector := NewDetector(mockFS)

	repoPath := "/test/repo"
	devcontainerPath := "/test/repo/.devcontainer/devcontainer.json"
	rootDevcontainerPath := "/test/repo/.devcontainer.json"

	// Mock expectations - first check returns false, second returns error
	mockFS.EXPECT().Exists(devcontainerPath).Return(false, nil)
	mockFS.EXPECT().Exists(rootDevcontainerPath).Return(false, assert.AnError)

	result, err := detector.DetectDevcontainer(repoPath)
	assert.Error(t, err)
	assert.False(t, result)
}
