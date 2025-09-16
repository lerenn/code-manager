//go:build unit

package status

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	cfg := config.Config{
		RepositoriesDir: "/home/user/.cm",
		StatusFile:      "/home/user/.cmstatus.yaml",
	}

	// Mock expectations for initialization
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(false, nil)
	mockFS.EXPECT().FileLock("/home/user/.cmstatus.yaml").Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic("/home/user/.cmstatus.yaml", gomock.Any(), gomock.Any()).Return(nil)

	manager := NewManager(mockFS, cfg)

	assert.NotNil(t, manager)
	assert.Implements(t, (*Manager)(nil), manager)
}
