//go:build unit

package gitcrypt

import (
	"errors"
	"testing"

	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGitCryptKeyManager_FindGitCryptKey_KeyFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fsMock := fsmocks.NewMockFS(ctrl)
	gitMock := gitmocks.NewMockGit(ctrl)
	promptMock := promptmocks.NewMockPrompter(ctrl)

	keyManager := NewKeyManager(fsMock, gitMock, promptMock)

	// Mock key exists
	fsMock.EXPECT().Exists("/path/to/repo/.git/git-crypt/keys/default").Return(true, nil)

	// Test finding key
	keyPath, err := keyManager.FindGitCryptKey("/path/to/repo")
	assert.NoError(t, err)
	assert.Equal(t, "/path/to/repo/.git/git-crypt/keys/default", keyPath)
}

func TestGitCryptKeyManager_FindGitCryptKey_KeyNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fsMock := fsmocks.NewMockFS(ctrl)
	gitMock := gitmocks.NewMockGit(ctrl)
	promptMock := promptmocks.NewMockPrompter(ctrl)

	keyManager := NewKeyManager(fsMock, gitMock, promptMock)

	// Mock key doesn't exist
	fsMock.EXPECT().Exists("/path/to/repo/.git/git-crypt/keys/default").Return(false, nil)

	// Test finding key
	keyPath, err := keyManager.FindGitCryptKey("/path/to/repo")
	assert.NoError(t, err)
	assert.Empty(t, keyPath)
}

func TestGitCryptKeyManager_FindGitCryptKey_ExistsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fsMock := fsmocks.NewMockFS(ctrl)
	gitMock := gitmocks.NewMockGit(ctrl)
	promptMock := promptmocks.NewMockPrompter(ctrl)

	keyManager := NewKeyManager(fsMock, gitMock, promptMock)

	// Mock exists check fails
	fsMock.EXPECT().Exists("/path/to/repo/.git/git-crypt/keys/default").Return(false, errors.New("permission denied"))

	// Test finding key
	keyPath, err := keyManager.FindGitCryptKey("/path/to/repo")
	assert.Error(t, err)
	assert.Empty(t, keyPath)
}

func TestGitCryptKeyManager_PromptUserForKeyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fsMock := fsmocks.NewMockFS(ctrl)
	gitMock := gitmocks.NewMockGit(ctrl)
	promptMock := promptmocks.NewMockPrompter(ctrl)

	keyManager := NewKeyManager(fsMock, gitMock, promptMock)

	// Test prompting user for key path
	keyPath, err := keyManager.PromptUserForKeyPath()
	assert.Error(t, err)
	assert.Empty(t, keyPath)
	assert.Contains(t, err.Error(), "git-crypt key not found")
}

func TestGitCryptKeyManager_ValidateKeyFile_ValidKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fsMock := fsmocks.NewMockFS(ctrl)
	gitMock := gitmocks.NewMockGit(ctrl)
	promptMock := promptmocks.NewMockPrompter(ctrl)

	keyManager := NewKeyManager(fsMock, gitMock, promptMock)

	// Mock key file exists and is readable
	fsMock.EXPECT().Exists("/path/to/key").Return(true, nil)
	fsMock.EXPECT().ReadFile("/path/to/key").Return([]byte("key content"), nil)

	// Test validation
	err := keyManager.ValidateKeyFile("/path/to/key")
	assert.NoError(t, err)
}

func TestGitCryptKeyManager_ValidateKeyFile_KeyNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fsMock := fsmocks.NewMockFS(ctrl)
	gitMock := gitmocks.NewMockGit(ctrl)
	promptMock := promptmocks.NewMockPrompter(ctrl)

	keyManager := NewKeyManager(fsMock, gitMock, promptMock)

	// Mock key file doesn't exist
	fsMock.EXPECT().Exists("/path/to/key").Return(false, nil)

	// Test validation
	err := keyManager.ValidateKeyFile("/path/to/key")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrKeyFileNotFound)
}

func TestGitCryptKeyManager_ValidateKeyFile_ExistsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fsMock := fsmocks.NewMockFS(ctrl)
	gitMock := gitmocks.NewMockGit(ctrl)
	promptMock := promptmocks.NewMockPrompter(ctrl)

	keyManager := NewKeyManager(fsMock, gitMock, promptMock)

	// Mock exists check fails
	fsMock.EXPECT().Exists("/path/to/key").Return(false, errors.New("permission denied"))

	// Test validation
	err := keyManager.ValidateKeyFile("/path/to/key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check key file existence")
}

func TestGitCryptKeyManager_ValidateKeyFile_ReadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fsMock := fsmocks.NewMockFS(ctrl)
	gitMock := gitmocks.NewMockGit(ctrl)
	promptMock := promptmocks.NewMockPrompter(ctrl)

	keyManager := NewKeyManager(fsMock, gitMock, promptMock)

	// Mock key file exists but read fails
	fsMock.EXPECT().Exists("/path/to/key").Return(true, nil)
	fsMock.EXPECT().ReadFile("/path/to/key").Return(nil, errors.New("read error"))

	// Test validation
	err := keyManager.ValidateKeyFile("/path/to/key")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrKeyFileInvalid)
}
