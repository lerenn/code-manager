//go:build unit

package cgwt

import (
	"testing"

	"github.com/lerenn/cgwt/pkg/fs"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCGWT_Run_QuietMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWTWithMode(mockFS, OutputModeQuiet)

	// Mock single repo detection - .git found
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)

	err := cgwt.Run()
	assert.NoError(t, err)
}

func TestCGWT_Run_VerboseMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWTWithMode(mockFS, OutputModeVerbose)

	// Mock single repo detection - .git found
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)

	err := cgwt.Run()
	assert.NoError(t, err)
}

func TestCGWT_Run_NormalMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWTWithMode(mockFS, OutputModeNormal)

	// Mock single repo detection - .git found
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)

	err := cgwt.Run()
	assert.NoError(t, err)
}
