//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
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

	var cm CM
	var err error

	cm, err = NewCM(NewCMParams{
		Config: createTestConfig(),
		Status: mockStatus,
		FS:     mockFS,
	})
	assert.NoError(t, err)

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
		assert.True(t, result[0].InBasePath)
		assert.Equal(t, "github.com/lerenn/example", result[1].Name)
		assert.Equal(t, "/test/base/path/example", result[1].Path)
		assert.True(t, result[1].InBasePath)
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
		assert.True(t, result[0].InBasePath)
		assert.Equal(t, "github.com/lerenn/outside", result[1].Name)
		assert.False(t, result[1].InBasePath)
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
		assert.False(t, result[0].InBasePath) // Defaults to false on error
	})

	t.Run("without logger", func(t *testing.T) {
		var cmNoLogger CM

		cmNoLogger, err = NewCM(NewCMParams{
			Config: createTestConfig(),
			Status: mockStatus,
			FS:     mockFS,
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
		assert.True(t, result[0].InBasePath)
	})
}

// createTestConfig creates a test configuration for use in tests.
func createTestConfig() config.Config {
	return config.Config{
		BasePath:   "/test/base/path",
		StatusFile: "/test/status.yaml",
	}
}
