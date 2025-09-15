//go:build unit

package codemanager

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	configmocks "github.com/lerenn/code-manager/pkg/config/mocks"
	"github.com/lerenn/code-manager/pkg/dependencies"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestListRepositories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStatus := statusmocks.NewMockManager(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockConfig := configmocks.NewMockManager(ctrl)

	var cm CodeManager
	var err error

	cm, err = NewCodeManager(NewCodeManagerParams{
		Dependencies: dependencies.New().
			WithConfig(mockConfig).
			WithStatusManager(mockStatus).
			WithFS(mockFS),
	})
	assert.NoError(t, err)

	// Mock config manager
	testConfig := config.Config{
		RepositoriesDir: "/test/base/path",
		WorkspacesDir:   "/test/workspaces",
		StatusFile:      "/test/status.yaml",
	}
	mockConfig.EXPECT().GetConfigWithFallback().Return(testConfig, nil).AnyTimes()

	t.Run("successful listing with repositories in base path", func(t *testing.T) {
		repositories := map[string]status.Repository{
			"github.com/lerenn/example": {
				Path: "/test/base/path/example",
			},
			"github.com/lerenn/another": {
				Path: "/test/base/path/another",
			},
		}

		mockStatus.EXPECT().ListRepositories().Return(repositories, nil)

		mockFS.EXPECT().IsPathWithinBase("/test/base/path", "/test/base/path/example").Return(true, nil)
		mockFS.EXPECT().IsPathWithinBase("/test/base/path", "/test/base/path/another").Return(true, nil)

		result, err := cm.ListRepositories()

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "github.com/lerenn/another", result[0].Name)
		assert.Equal(t, "/test/base/path/another", result[0].Path)
		assert.True(t, result[0].InRepositoriesDir)
		assert.Equal(t, "github.com/lerenn/example", result[1].Name)
		assert.Equal(t, "/test/base/path/example", result[1].Path)
		assert.True(t, result[1].InRepositoriesDir)
	})

	t.Run("successful listing with repositories outside base path", func(t *testing.T) {
		repositories := map[string]status.Repository{
			"github.com/lerenn/example": {
				Path: "/test/base/path/example",
			},
			"github.com/lerenn/outside": {
				Path: "/other/path/outside",
			},
		}

		mockStatus.EXPECT().ListRepositories().Return(repositories, nil)

		mockFS.EXPECT().IsPathWithinBase("/test/base/path", "/test/base/path/example").Return(true, nil)
		mockFS.EXPECT().IsPathWithinBase("/test/base/path", "/other/path/outside").Return(false, nil)

		result, err := cm.ListRepositories()

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "github.com/lerenn/example", result[0].Name)
		assert.True(t, result[0].InRepositoriesDir)
		assert.Equal(t, "github.com/lerenn/outside", result[1].Name)
		assert.False(t, result[1].InRepositoriesDir)
	})

	t.Run("empty repository list", func(t *testing.T) {
		repositories := map[string]status.Repository{}

		mockStatus.EXPECT().ListRepositories().Return(repositories, nil)

		result, err := cm.ListRepositories()

		assert.NoError(t, err)
		assert.Len(t, result, 0)
	})

	t.Run("status manager error", func(t *testing.T) {

		mockStatus.EXPECT().ListRepositories().Return(nil, assert.AnError)

		result, err := cm.ListRepositories()

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrFailedToLoadRepositories)
	})

	t.Run("base path validation error", func(t *testing.T) {
		repositories := map[string]status.Repository{
			"github.com/lerenn/example": {
				Path: "/test/base/path/example",
			},
		}

		mockStatus.EXPECT().ListRepositories().Return(repositories, nil)

		mockFS.EXPECT().IsPathWithinBase("/test/base/path", "/test/base/path/example").Return(false, assert.AnError)

		result, err := cm.ListRepositories()

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "github.com/lerenn/example", result[0].Name)
		assert.False(t, result[0].InRepositoriesDir) // Defaults to false on error
	})

	t.Run("without logger", func(t *testing.T) {
		var cmNoLogger CodeManager

		cmNoLogger, err = NewCodeManager(NewCodeManagerParams{
			Dependencies: dependencies.New().
				WithConfig(mockConfig).
				WithStatusManager(mockStatus).
				WithFS(mockFS),
		})

		repositories := map[string]status.Repository{
			"github.com/lerenn/example": {
				Path: "/test/base/path/example",
			},
		}

		mockStatus.EXPECT().ListRepositories().Return(repositories, nil)
		mockFS.EXPECT().IsPathWithinBase("/test/base/path", "/test/base/path/example").Return(true, nil)

		result, err := cmNoLogger.ListRepositories()

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "github.com/lerenn/example", result[0].Name)
		assert.True(t, result[0].InRepositoriesDir)
	})
}
