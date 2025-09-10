//go:build unit

package gitcrypt

import (
	"errors"
	"testing"

	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGitCryptDetector_DetectGitCryptUsage_NoGitAttributes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fsMock := fsmocks.NewMockFS(ctrl)
	detector := NewDetector(fsMock)

	// Mock .gitattributes doesn't exist
	fsMock.EXPECT().Exists("/path/to/repo/.gitattributes").Return(false, nil)

	// Test detection
	usesGitCrypt, err := detector.DetectGitCryptUsage("/path/to/repo")
	assert.NoError(t, err)
	assert.False(t, usesGitCrypt)
}

func TestGitCryptDetector_DetectGitCryptUsage_WithGitCrypt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fsMock := fsmocks.NewMockFS(ctrl)
	detector := NewDetector(fsMock)

	// Mock .gitattributes exists and contains git-crypt filter
	fsMock.EXPECT().Exists("/path/to/repo/.gitattributes").Return(true, nil)
	fsMock.EXPECT().ReadFile("/path/to/repo/.gitattributes").Return([]byte("*.secret filter=git-crypt diff=git-crypt"), nil)

	// Test detection
	usesGitCrypt, err := detector.DetectGitCryptUsage("/path/to/repo")
	assert.NoError(t, err)
	assert.True(t, usesGitCrypt)
}

func TestGitCryptDetector_DetectGitCryptUsage_WithoutGitCrypt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fsMock := fsmocks.NewMockFS(ctrl)
	detector := NewDetector(fsMock)

	// Mock .gitattributes exists but doesn't contain git-crypt filter
	fsMock.EXPECT().Exists("/path/to/repo/.gitattributes").Return(true, nil)
	fsMock.EXPECT().ReadFile("/path/to/repo/.gitattributes").Return([]byte("*.txt text"), nil)

	// Test detection
	usesGitCrypt, err := detector.DetectGitCryptUsage("/path/to/repo")
	assert.NoError(t, err)
	assert.False(t, usesGitCrypt)
}

func TestGitCryptDetector_DetectGitCryptUsage_ExistsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fsMock := fsmocks.NewMockFS(ctrl)
	detector := NewDetector(fsMock)

	// Mock .gitattributes existence check fails
	fsMock.EXPECT().Exists("/path/to/repo/.gitattributes").Return(false, errors.New("permission denied"))

	// Test detection
	usesGitCrypt, err := detector.DetectGitCryptUsage("/path/to/repo")
	assert.Error(t, err)
	assert.False(t, usesGitCrypt)
}

func TestGitCryptDetector_DetectGitCryptUsage_ReadFileError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fsMock := fsmocks.NewMockFS(ctrl)
	detector := NewDetector(fsMock)

	// Mock .gitattributes exists but read fails
	fsMock.EXPECT().Exists("/path/to/repo/.gitattributes").Return(true, nil)
	fsMock.EXPECT().ReadFile("/path/to/repo/.gitattributes").Return(nil, errors.New("read error"))

	// Test detection
	usesGitCrypt, err := detector.DetectGitCryptUsage("/path/to/repo")
	assert.Error(t, err)
	assert.False(t, usesGitCrypt)
}
