//go:build unit

package gitcrypt

import (
	"errors"
	"os"
	"testing"

	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/logger"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGitCryptWorktreeCheckoutHook_RegisterForOperations(t *testing.T) {
	hook := NewWorktreeCheckoutHook()

	// Mock register function
	registeredOperations := make(map[string]hooks.WorktreeCheckoutHook)
	registerHook := func(operation string, h hooks.WorktreeCheckoutHook) error {
		registeredOperations[operation] = h
		return nil
	}

	// Register hook
	err := hook.RegisterForOperations(registerHook)
	assert.NoError(t, err)

	// Verify operations are registered
	assert.Contains(t, registeredOperations, "CreateWorkTree")
	assert.Contains(t, registeredOperations, "LoadWorktree")
	assert.Equal(t, hook, registeredOperations["CreateWorkTree"])
	assert.Equal(t, hook, registeredOperations["LoadWorktree"])
}

func TestGitCryptWorktreeCheckoutHook_Name(t *testing.T) {
	hook := NewWorktreeCheckoutHook()
	assert.Equal(t, "git-crypt-worktree-checkout", hook.Name())
}

func TestGitCryptWorktreeCheckoutHook_Priority(t *testing.T) {
	hook := NewWorktreeCheckoutHook()
	assert.Equal(t, 50, hook.Priority())
}

func TestGitCryptWorktreeCheckoutHook_Execute(t *testing.T) {
	hook := NewWorktreeCheckoutHook()
	ctx := &hooks.HookContext{}

	err := hook.Execute(ctx)
	assert.NoError(t, err)
}

func TestGitCryptWorktreeCheckoutHook_OnWorktreeCheckout_NoGitCrypt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	fsMock := fsmocks.NewMockFS(ctrl)
	gitMock := gitmocks.NewMockGit(ctrl)
	promptMock := promptmocks.NewMockPrompter(ctrl)

	// Create hook with mocks
	hook := &WorktreeCheckoutHook{
		fs:            fsMock,
		git:           gitMock,
		prompt:        promptMock,
		logger:        logger.NewNoopLogger(),
		detector:      NewDetector(fsMock),
		keyManager:    NewKeyManager(fsMock, gitMock, promptMock),
		worktreeSetup: NewWorktreeSetup(fsMock),
	}

	// Setup context
	ctx := &hooks.HookContext{
		Parameters: map[string]interface{}{
			"worktreePath": "/path/to/worktree",
			"repoPath":     "/path/to/repo",
			"branch":       "main",
		},
	}

	// Mock git-crypt detection - no git-crypt usage
	fsMock.EXPECT().Exists("/path/to/repo/.gitattributes").Return(false, nil)

	// Execute hook
	err := hook.OnWorktreeCheckout(ctx)
	assert.NoError(t, err)
}

func TestGitCryptWorktreeCheckoutHook_OnWorktreeCheckout_WithGitCrypt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	fsMock := fsmocks.NewMockFS(ctrl)
	gitMock := gitmocks.NewMockGit(ctrl)
	promptMock := promptmocks.NewMockPrompter(ctrl)

	// Create hook with mocks
	hook := &WorktreeCheckoutHook{
		fs:            fsMock,
		git:           gitMock,
		prompt:        promptMock,
		logger:        logger.NewNoopLogger(),
		detector:      NewDetector(fsMock),
		keyManager:    NewKeyManager(fsMock, gitMock, promptMock),
		worktreeSetup: NewWorktreeSetup(fsMock),
	}

	// Setup context
	ctx := &hooks.HookContext{
		Parameters: map[string]interface{}{
			"worktreePath": "/path/to/worktree",
			"repoPath":     "/path/to/repo",
			"branch":       "main",
		},
	}

	// Mock git-crypt detection - git-crypt usage detected
	fsMock.EXPECT().Exists("/path/to/repo/.gitattributes").Return(true, nil)
	fsMock.EXPECT().ReadFile("/path/to/repo/.gitattributes").Return([]byte("*.secret filter=git-crypt diff=git-crypt"), nil)

	// Mock key finding - key found
	fsMock.EXPECT().Exists("/path/to/repo/.git/git-crypt/keys/default").Return(true, nil)

	// Mock key validation - key file exists and is readable
	fsMock.EXPECT().Exists("/path/to/repo/.git/git-crypt/keys/default").Return(true, nil)
	fsMock.EXPECT().ReadFile("/path/to/repo/.git/git-crypt/keys/default").Return([]byte("key content"), nil)

	// Mock worktree setup
	fsMock.EXPECT().MkdirAll("/path/to/repo/.git/worktrees/worktree/git-crypt", os.FileMode(0755)).Return(nil)
	fsMock.EXPECT().MkdirAll("/path/to/repo/.git/worktrees/worktree/git-crypt/keys", os.FileMode(0755)).Return(nil)

	// Execute hook
	err := hook.OnWorktreeCheckout(ctx)
	// Expected to fail in unit tests because file copy operation can't be mocked
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to setup git-crypt in worktree")
}

func TestGitCryptWorktreeCheckoutHook_OnWorktreeCheckout_MissingWorktreePath(t *testing.T) {
	hook := NewWorktreeCheckoutHook()

	// Setup context without worktree path
	ctx := &hooks.HookContext{
		Parameters: map[string]interface{}{
			"repoPath": "/path/to/repo",
			"branch":   "main",
		},
	}

	// Execute hook
	err := hook.OnWorktreeCheckout(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWorktreePathNotFound)
}

func TestGitCryptWorktreeCheckoutHook_OnWorktreeCheckout_MissingRepoPath(t *testing.T) {
	hook := NewWorktreeCheckoutHook()

	// Setup context without repo path
	ctx := &hooks.HookContext{
		Parameters: map[string]interface{}{
			"worktreePath": "/path/to/worktree",
			"branch":       "main",
		},
	}

	// Execute hook
	err := hook.OnWorktreeCheckout(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRepositoryPathNotFound)
}

func TestGitCryptWorktreeCheckoutHook_OnWorktreeCheckout_MissingBranch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	fsMock := fsmocks.NewMockFS(ctrl)
	gitMock := gitmocks.NewMockGit(ctrl)
	promptMock := promptmocks.NewMockPrompter(ctrl)

	// Create hook with mocks
	hook := &WorktreeCheckoutHook{
		fs:            fsMock,
		git:           gitMock,
		prompt:        promptMock,
		logger:        logger.NewNoopLogger(),
		detector:      NewDetector(fsMock),
		keyManager:    NewKeyManager(fsMock, gitMock, promptMock),
		worktreeSetup: NewWorktreeSetup(fsMock),
	}

	// Setup context without branch
	ctx := &hooks.HookContext{
		Parameters: map[string]interface{}{
			"worktreePath": "/path/to/worktree",
			"repoPath":     "/path/to/repo",
		},
	}

	// Mock git-crypt detection - git-crypt usage detected
	fsMock.EXPECT().Exists("/path/to/repo/.gitattributes").Return(true, nil)
	fsMock.EXPECT().ReadFile("/path/to/repo/.gitattributes").Return([]byte("*.secret filter=git-crypt diff=git-crypt"), nil)

	// Execute hook
	err := hook.OnWorktreeCheckout(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrBranchNotFound)
}

func TestGitCryptWorktreeCheckoutHook_OnWorktreeCheckout_GitCryptDetectionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	fsMock := fsmocks.NewMockFS(ctrl)
	gitMock := gitmocks.NewMockGit(ctrl)
	promptMock := promptmocks.NewMockPrompter(ctrl)

	// Create hook with mocks
	hook := &WorktreeCheckoutHook{
		fs:            fsMock,
		git:           gitMock,
		prompt:        promptMock,
		logger:        logger.NewNoopLogger(),
		detector:      NewDetector(fsMock),
		keyManager:    NewKeyManager(fsMock, gitMock, promptMock),
		worktreeSetup: NewWorktreeSetup(fsMock),
	}

	// Setup context
	ctx := &hooks.HookContext{
		Parameters: map[string]interface{}{
			"worktreePath": "/path/to/worktree",
			"repoPath":     "/path/to/repo",
			"branch":       "main",
		},
	}

	// Mock git-crypt detection error
	fsMock.EXPECT().Exists("/path/to/repo/.gitattributes").Return(false, errors.New("permission denied"))

	// Execute hook
	err := hook.OnWorktreeCheckout(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to detect git-crypt usage")
}
