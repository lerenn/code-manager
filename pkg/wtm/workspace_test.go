//go:build unit

package wtm

import (
	"path/filepath"
	"testing"

	"github.com/lerenn/wtm/pkg/fs"
	"github.com/lerenn/wtm/pkg/git"
	"github.com/lerenn/wtm/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWTM_Run_WorkspaceMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	wtm := NewWTM(createTestConfig())
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus

	// Mock single repo detection - no .git found (called once: detectProjectMode)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file (called twice: once in detectProjectMode, once in handleWorkspaceMode)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(2)

	// Mock reading workspace file (called multiple times: Load, Validate, validateWorkspaceForWorktreeCreation, createWorktreesForWorkspace)
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			},
			{
				"name": "Backend",
				"path": "./backend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil).Times(4)

	// Mock repository validation for frontend
	mockFS.EXPECT().Exists("frontend").Return(true, nil).AnyTimes()
	mockFS.EXPECT().Exists("frontend/.git").Return(true, nil).AnyTimes()

	// Mock repository validation for backend
	mockFS.EXPECT().Exists("backend").Return(true, nil).AnyTimes()
	mockFS.EXPECT().Exists("backend/.git").Return(true, nil).AnyTimes()

	// Mock Git status for validation (called for each repository, order may vary)
	mockGit.EXPECT().Status("frontend").Return("On branch main", nil).AnyTimes()
	mockGit.EXPECT().Status("backend").Return("On branch main", nil).AnyTimes()

	// Mock Git operations for worktree creation
	mockGit.EXPECT().GetRepositoryName("frontend").Return("github.com/lerenn/frontend", nil).AnyTimes()
	mockGit.EXPECT().GetRepositoryName("backend").Return("github.com/lerenn/backend", nil).AnyTimes()
	mockGit.EXPECT().BranchExists("frontend", "test-branch").Return(false, nil).AnyTimes()
	mockGit.EXPECT().BranchExists("backend", "test-branch").Return(false, nil).AnyTimes()
	mockGit.EXPECT().CreateBranch("frontend", "test-branch").Return(nil).AnyTimes()
	mockGit.EXPECT().CreateBranch("backend", "test-branch").Return(nil).AnyTimes()
	mockGit.EXPECT().CreateWorktree("frontend", "/test/base/path/github.com/lerenn/frontend/test-branch", "test-branch").Return(nil).AnyTimes()
	mockGit.EXPECT().CreateWorktree("backend", "/test/base/path/github.com/lerenn/backend/test-branch", "test-branch").Return(nil).AnyTimes()

	// Mock file system operations for worktree creation
	mockFS.EXPECT().Exists("/test/base/path/github.com/lerenn/frontend/test-branch").Return(false, nil).AnyTimes()
	mockFS.EXPECT().Exists("/test/base/path/github.com/lerenn/backend/test-branch").Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll("/test/base/path/github.com/lerenn/frontend/test-branch", gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().MkdirAll("/test/base/path/github.com/lerenn/backend/test-branch", gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().MkdirAll("/test/base/path/workspaces", gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().WriteFileAtomic("/test/base/path/workspaces/project-test-branch.code-workspace", gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Mock status manager operations
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/frontend", "test-branch").Return(nil, status.ErrWorktreeNotFound).AnyTimes()
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/backend", "test-branch").Return(nil, status.ErrWorktreeNotFound).AnyTimes()
	// Get absolute path for workspace file
	absPath, _ := filepath.Abs("project.code-workspace")
	mockStatus.EXPECT().AddWorktree("github.com/lerenn/frontend", "test-branch", "frontend", absPath).Return(nil).AnyTimes()
	mockStatus.EXPECT().AddWorktree("github.com/lerenn/backend", "test-branch", "backend", absPath).Return(nil).AnyTimes()

	err := wtm.CreateWorkTree("test-branch", nil)
	assert.NoError(t, err)
}

func TestWTM_Run_InvalidWorkspaceJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	wtm := NewWTM(createTestConfig())
	c := wtm.(*realWTM)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called once: detectProjectMode)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file (called twice: once in detectProjectMode, once in handleWorkspaceMode)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(2)

	// Mock reading workspace file with invalid JSON (called once: handleWorkspaceMode)
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(`{invalid json`), nil).Times(1)

	err := wtm.CreateWorkTree("test-branch", nil)
	assert.ErrorIs(t, err, ErrWorkspaceFileMalformed)
}

func TestWTM_Run_MissingRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	wtm := NewWTM(createTestConfig())
	c := wtm.(*realWTM)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called once: detectProjectMode)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file (called twice: once in detectProjectMode, once in handleWorkspaceMode)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(2)

	// Mock reading workspace file (called twice: once for display, once for validation)
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil).Times(2)

	// Mock repository validation - repository not found
	mockFS.EXPECT().Exists("frontend").Return(false, nil).AnyTimes()

	err := wtm.CreateWorkTree("test-branch", nil)
	assert.ErrorIs(t, err, ErrRepositoryNotFoundInWorkspace)
}

func TestWTM_Run_InvalidRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	wtm := NewWTM(createTestConfig())
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit

	// Mock single repo detection - no .git found (called once: detectProjectMode)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file (called twice: once in detectProjectMode, once in handleWorkspaceMode)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(2)

	// Mock reading workspace file (called twice: once for display, once for validation)
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil).Times(2)

	// Mock repository validation - repository exists but no .git
	mockFS.EXPECT().Exists("frontend").Return(true, nil).AnyTimes()
	mockFS.EXPECT().Exists("frontend/.git").Return(false, nil).AnyTimes()

	err := wtm.CreateWorkTree("test-branch", nil)
	assert.ErrorIs(t, err, ErrInvalidRepositoryInWorkspaceNoGit)
}

func TestWTM_Run_GitStatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	wtm := NewWTM(createTestConfig())
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit

	// Mock single repo detection - no .git found (called once: detectProjectMode)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file (called twice: once in detectProjectMode, once in handleWorkspaceMode)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(2)

	// Mock reading workspace file (called twice: once for display, once for validation)
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil).Times(2)

	// Mock repository validation - repository exists and has .git
	mockFS.EXPECT().Exists("frontend").Return(true, nil).AnyTimes()
	mockFS.EXPECT().Exists("frontend/.git").Return(true, nil).AnyTimes()

	// Mock Git status error
	mockGit.EXPECT().Status("frontend").Return("", assert.AnError).AnyTimes()

	err := wtm.CreateWorkTree("test-branch", nil)
	assert.ErrorIs(t, err, ErrInvalidRepositoryInWorkspace)
}

func TestWTM_Run_MultipleWorkspaceFiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	wtm := NewWTM(createTestConfig())
	c := wtm.(*realWTM)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called once: detectProjectMode)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find multiple workspace files (called twice: once in detectProjectMode, once in handleWorkspaceMode)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project1.code-workspace", "project2.code-workspace"}, nil).Times(2)

	err := wtm.CreateWorkTree("test-branch", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user cancelled selection")
}

func TestWTM_Run_WorkspaceFileReadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	wtm := NewWTM(createTestConfig())
	c := wtm.(*realWTM)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called once: detectProjectMode)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file (called twice: once in detectProjectMode, once in handleWorkspaceMode)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(2)

	// Mock reading workspace file error (called once: handleWorkspaceMode)
	mockFS.EXPECT().ReadFile("project.code-workspace").Return(nil, assert.AnError).Times(1)

	err := wtm.CreateWorkTree("test-branch", nil)
	assert.ErrorIs(t, err, ErrWorkspaceFileReadError)
}

func TestWTM_Run_WorkspaceGlobError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	wtm := NewWTM(createTestConfig())
	c := wtm.(*realWTM)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called once: detectProjectMode)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection error (called once: detectProjectMode)
	mockFS.EXPECT().Glob("*.code-workspace").Return(nil, assert.AnError).Times(1)

	err := wtm.CreateWorkTree("test-branch", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check for workspace files")
}

func TestWTM_Run_WorkspaceVerboseMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	wtm := NewWTM(createTestConfig())
	wtm.SetVerbose(true)
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus

	// Mock single repo detection - no .git found (called once: detectProjectMode)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file (called twice: once in detectProjectMode, once in handleWorkspaceMode)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(2)

	// Mock reading workspace file (called multiple times: Load, Validate, validateWorkspaceForWorktreeCreation, createWorktreesForWorkspace)
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil).Times(4)

	// Mock repository validation - repository exists and has .git
	mockFS.EXPECT().Exists("frontend").Return(true, nil).AnyTimes()
	mockFS.EXPECT().Exists("frontend/.git").Return(true, nil).AnyTimes()

	// Mock Git status for validation
	mockGit.EXPECT().Status("frontend").Return("On branch main", nil).AnyTimes()

	// Mock Git operations for worktree creation
	mockGit.EXPECT().GetRepositoryName("frontend").Return("github.com/lerenn/frontend", nil).AnyTimes()
	mockGit.EXPECT().BranchExists("frontend", "test-branch").Return(false, nil).AnyTimes()
	mockGit.EXPECT().CreateBranch("frontend", "test-branch").Return(nil).AnyTimes()
	mockGit.EXPECT().CreateWorktree("frontend", "/test/base/path/github.com/lerenn/frontend/test-branch", "test-branch").Return(nil).AnyTimes()

	// Mock file system operations for worktree creation
	mockFS.EXPECT().Exists("/test/base/path/github.com/lerenn/frontend/test-branch").Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll("/test/base/path/github.com/lerenn/frontend/test-branch", gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().MkdirAll("/test/base/path/workspaces", gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().WriteFileAtomic("/test/base/path/workspaces/project-test-branch.code-workspace", gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Mock status manager operations
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/frontend", "test-branch").Return(nil, status.ErrWorktreeNotFound).AnyTimes()
	mockStatus.EXPECT().AddWorktree("github.com/lerenn/frontend", "test-branch", "frontend", "project.code-workspace").Return(nil).AnyTimes()

	err := wtm.CreateWorkTree("test-branch", nil)
	assert.NoError(t, err)
}

func TestWTM_Run_EmptyWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	wtm := NewWTM(createTestConfig())
	c := wtm.(*realWTM)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called once: detectProjectMode)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file (called twice: once in detectProjectMode, once in handleWorkspaceMode)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(2)

	// Mock reading workspace file with empty folders (called multiple times for display and validation)
	workspaceJSON := `{
		"folders": []
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil).AnyTimes()

	err := wtm.CreateWorkTree("test-branch", nil)
	assert.ErrorIs(t, err, ErrWorkspaceEmptyFolders)
}

func TestWorkspace_DeleteWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	workspace := newWorkspace(mockFS, mockGit, createTestConfig(), mockStatus, nil, false)
	workspace.originalFile = "project.code-workspace"

	// Mock workspace file parsing
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil).Times(1)

	// Mock worktree retrieval
	worktrees := []status.Repository{
		{
			URL:       "github.com/lerenn/frontend",
			Branch:    "test-branch",
			Path:      "frontend",
			Workspace: "project.code-workspace",
		},
	}
	mockStatus.EXPECT().ListAllWorktrees().Return(worktrees, nil).Times(1)

	// Mock Git worktree deletion
	mockGit.EXPECT().RemoveWorktree("frontend", "/test/base/path/github.com/lerenn/frontend/test-branch").Return(nil).Times(1)

	// Mock file system operations
	mockFS.EXPECT().RemoveAll("/test/base/path/github.com/lerenn/frontend/test-branch").Return(nil).Times(1)
	mockFS.EXPECT().RemoveAll("/test/base/path/workspaces/project-test-branch.code-workspace").Return(nil).Times(1)

	// Mock status removal
	mockStatus.EXPECT().RemoveWorktree("github.com/lerenn/frontend", "test-branch").Return(nil).Times(1)

	err := workspace.DeleteWorktree("test-branch", false)
	assert.NoError(t, err)
}

func TestWorkspace_DeleteWorktree_NoWorktreesFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	workspace := newWorkspace(mockFS, mockGit, createTestConfig(), mockStatus, nil, false)
	workspace.originalFile = "project.code-workspace"

	// Mock worktree retrieval - no worktrees found
	mockStatus.EXPECT().ListAllWorktrees().Return([]status.Repository{}, nil).Times(1)

	err := workspace.DeleteWorktree("test-branch", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no worktrees found for workspace branch")
}

func TestWorkspace_DeleteWorktree_ForceMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	workspace := newWorkspace(mockFS, mockGit, createTestConfig(), mockStatus, nil, false)
	workspace.originalFile = "project.code-workspace"

	// Mock workspace file parsing
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil).Times(1)

	// Mock worktree retrieval
	worktrees := []status.Repository{
		{
			URL:       "github.com/lerenn/frontend",
			Branch:    "test-branch",
			Path:      "frontend",
			Workspace: "project.code-workspace",
		},
	}
	mockStatus.EXPECT().ListAllWorktrees().Return(worktrees, nil).Times(1)

	// Mock Git worktree deletion with error (should be ignored in force mode)
	mockGit.EXPECT().RemoveWorktree("frontend", "/test/base/path/github.com/lerenn/frontend/test-branch").Return(assert.AnError).Times(1)

	// Mock file system operations with error (should be ignored in force mode)
	mockFS.EXPECT().RemoveAll("/test/base/path/github.com/lerenn/frontend/test-branch").Return(assert.AnError).Times(1)
	mockFS.EXPECT().RemoveAll("/test/base/path/workspaces/project-test-branch.code-workspace").Return(assert.AnError).Times(1)

	// Mock status removal with error (should be ignored in force mode)
	mockStatus.EXPECT().RemoveWorktree("github.com/lerenn/frontend", "test-branch").Return(assert.AnError).Times(1)

	err := workspace.DeleteWorktree("test-branch", true)
	assert.NoError(t, err)
}

func TestWorkspace_ListWorktrees(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	workspace := newWorkspace(mockFS, mockGit, createTestConfig(), mockStatus, nil, false)
	workspace.originalFile = "project.code-workspace"

	// Mock all worktrees retrieval
	allWorktrees := []status.Repository{
		{
			URL:       "github.com/lerenn/frontend",
			Branch:    "test-branch",
			Path:      "frontend",
			Workspace: "project.code-workspace",
		},
		{
			URL:       "github.com/lerenn/backend",
			Branch:    "test-branch",
			Path:      "backend",
			Workspace: "project.code-workspace",
		},
		{
			URL:       "github.com/lerenn/other",
			Branch:    "main",
			Path:      "other",
			Workspace: "other.code-workspace",
		},
	}
	mockStatus.EXPECT().ListAllWorktrees().Return(allWorktrees, nil).Times(1)

	worktrees, err := workspace.ListWorktrees()
	assert.NoError(t, err)
	assert.Len(t, worktrees, 2)
	assert.Equal(t, "github.com/lerenn/frontend", worktrees[0].URL)
	assert.Equal(t, "github.com/lerenn/backend", worktrees[1].URL)
}

func TestWorkspace_ListWorktrees_LoadWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	workspace := newWorkspace(mockFS, mockGit, createTestConfig(), mockStatus, nil, false)
	// originalFile is empty, so Load() will be called

	// Mock workspace detection and loading
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(1)
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil).Times(1)

	// Mock all worktrees retrieval
	allWorktrees := []status.Repository{
		{
			URL:       "github.com/lerenn/frontend",
			Branch:    "test-branch",
			Path:      "frontend",
			Workspace: "project.code-workspace",
		},
	}
	mockStatus.EXPECT().ListAllWorktrees().Return(allWorktrees, nil).Times(1)

	worktrees, err := workspace.ListWorktrees()
	assert.NoError(t, err)
	assert.Len(t, worktrees, 1)
	assert.Equal(t, "github.com/lerenn/frontend", worktrees[0].URL)
}

func TestWorkspace_ValidateWorkspaceForWorktreeCreation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	workspace := newWorkspace(mockFS, mockGit, createTestConfig(), mockStatus, nil, false)
	workspace.originalFile = "project.code-workspace"

	// Mock workspace file parsing
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil).Times(1)

	// Mock Git operations
	mockGit.EXPECT().GetRepositoryName("frontend").Return("github.com/lerenn/frontend", nil).Times(1)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/frontend", "test-branch").Return(nil, status.ErrWorktreeNotFound).Times(1)
	mockGit.EXPECT().BranchExists("frontend", "test-branch").Return(false, nil).Times(1)

	// Mock worktree directory check
	mockFS.EXPECT().Exists("/test/base/path/github.com/lerenn/frontend/test-branch").Return(false, nil).Times(1)

	err := workspace.validateWorkspaceForWorktreeCreation("test-branch")
	assert.NoError(t, err)
}

func TestWorkspace_ValidateWorkspaceForWorktreeCreation_WorktreeExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	workspace := newWorkspace(mockFS, mockGit, createTestConfig(), mockStatus, nil, false)
	workspace.originalFile = "project.code-workspace"

	// Mock workspace file parsing
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil).Times(1)

	// Mock Git operations - worktree already exists
	mockGit.EXPECT().GetRepositoryName("frontend").Return("github.com/lerenn/frontend", nil).Times(1)
	existingWorktree := &status.Repository{
		URL:       "github.com/lerenn/frontend",
		Branch:    "test-branch",
		Path:      "frontend",
		Workspace: "project.code-workspace",
	}
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/frontend", "test-branch").Return(existingWorktree, nil).Times(1)

	err := workspace.validateWorkspaceForWorktreeCreation("test-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worktree already exists")
}

func TestWorkspace_ValidateWorkspaceForWorktreeCreation_DirectoryExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	workspace := newWorkspace(mockFS, mockGit, createTestConfig(), mockStatus, nil, false)
	workspace.originalFile = "project.code-workspace"

	// Mock workspace file parsing
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil).Times(1)

	// Mock Git operations
	mockGit.EXPECT().GetRepositoryName("frontend").Return("github.com/lerenn/frontend", nil).Times(1)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/frontend", "test-branch").Return(nil, status.ErrWorktreeNotFound).Times(1)
	mockGit.EXPECT().BranchExists("frontend", "test-branch").Return(false, nil).Times(1)

	// Mock worktree directory check - directory already exists
	mockFS.EXPECT().Exists("/test/base/path/github.com/lerenn/frontend/test-branch").Return(true, nil).Times(1)

	err := workspace.validateWorkspaceForWorktreeCreation("test-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worktree directory already exists")
}

func TestWorkspace_GetName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	workspace := newWorkspace(mockFS, mockGit, createTestConfig(), mockStatus, nil, false)

	// Test with workspace name in config
	config := &WorkspaceConfig{
		Name: "My Workspace",
		Folders: []WorkspaceFolder{
			{Path: "./frontend"},
		},
	}
	name := workspace.getName(config, "project.code-workspace")
	assert.Equal(t, "My Workspace", name)

	// Test fallback to filename
	config = &WorkspaceConfig{
		Folders: []WorkspaceFolder{
			{Path: "./frontend"},
		},
	}
	name = workspace.getName(config, "my-project.code-workspace")
	assert.Equal(t, "my-project", name)
}

func TestWorkspace_IsQuitCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	workspace := newWorkspace(mockFS, mockGit, createTestConfig(), mockStatus, nil, false)

	// Test quit commands
	assert.True(t, workspace.isQuitCommand("q"))
	assert.True(t, workspace.isQuitCommand("quit"))
	assert.True(t, workspace.isQuitCommand("exit"))
	assert.True(t, workspace.isQuitCommand("cancel"))

	// Test non-quit commands
	assert.False(t, workspace.isQuitCommand("1"))
	assert.False(t, workspace.isQuitCommand("yes"))
	assert.False(t, workspace.isQuitCommand(""))
}

func TestWorkspace_ParseNumericInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	workspace := newWorkspace(mockFS, mockGit, createTestConfig(), mockStatus, nil, false)

	// Test valid numeric input
	result, err := workspace.parseNumericInput("1")
	assert.NoError(t, err)
	assert.Equal(t, 1, result)

	result, err = workspace.parseNumericInput("42")
	assert.NoError(t, err)
	assert.Equal(t, 42, result)

	// Test invalid numeric input
	_, err = workspace.parseNumericInput("abc")
	assert.Error(t, err)

	_, err = workspace.parseNumericInput("")
	assert.Error(t, err)
}

func TestWorkspace_IsValidChoice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	workspace := newWorkspace(mockFS, mockGit, createTestConfig(), mockStatus, nil, false)

	// Test valid choices
	assert.True(t, workspace.isValidChoice(1, 3))
	assert.True(t, workspace.isValidChoice(2, 3))
	assert.True(t, workspace.isValidChoice(3, 3))

	// Test invalid choices
	assert.False(t, workspace.isValidChoice(0, 3))
	assert.False(t, workspace.isValidChoice(4, 3))
	assert.False(t, workspace.isValidChoice(-1, 3))
}

func TestWorkspace_ParseConfirmationInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	workspace := newWorkspace(mockFS, mockGit, createTestConfig(), mockStatus, nil, false)

	// Test yes inputs
	result, err := workspace.parseConfirmationInput("y")
	assert.NoError(t, err)
	assert.True(t, result)

	result, err = workspace.parseConfirmationInput("yes")
	assert.NoError(t, err)
	assert.True(t, result)

	result, err = workspace.parseConfirmationInput("Y")
	assert.NoError(t, err)
	assert.True(t, result)

	// Test no inputs
	result, err = workspace.parseConfirmationInput("n")
	assert.NoError(t, err)
	assert.False(t, result)

	result, err = workspace.parseConfirmationInput("no")
	assert.NoError(t, err)
	assert.False(t, result)

	// Test quit inputs
	_, err = workspace.parseConfirmationInput("q")
	assert.Error(t, err)

	_, err = workspace.parseConfirmationInput("quit")
	assert.Error(t, err)

	// Test invalid inputs
	_, err = workspace.parseConfirmationInput("maybe")
	assert.Error(t, err)

	_, err = workspace.parseConfirmationInput("")
	assert.Error(t, err)
}

func TestWTM_ListWorktrees_WorkspaceMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	wtm := NewWTM(createTestConfig())
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus

	// Mock single repo detection - no .git found
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).AnyTimes()

	// Mock workspace loading
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil).AnyTimes()

	// Mock worktree listing
	worktrees := []status.Repository{
		{
			URL:       "github.com/lerenn/frontend",
			Branch:    "test-branch",
			Path:      "frontend",
			Workspace: "project.code-workspace",
		},
	}
	mockStatus.EXPECT().ListAllWorktrees().Return(worktrees, nil).AnyTimes()

	result, _, err := wtm.ListWorktrees()
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "github.com/lerenn/frontend", result[0].URL)
}

func TestWTM_DeleteWorkTree_WorkspaceMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	wtm := NewWTM(createTestConfig())
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus

	// Mock single repo detection - no .git found
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).AnyTimes()

	// Mock workspace loading
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil).AnyTimes()

	// Mock repository validation
	mockFS.EXPECT().Exists("frontend").Return(true, nil).AnyTimes()
	mockFS.EXPECT().Exists("frontend/.git").Return(true, nil).AnyTimes()
	mockGit.EXPECT().Status("frontend").Return("On branch main", nil).AnyTimes()

	// Mock worktree retrieval
	worktrees := []status.Repository{
		{
			URL:       "github.com/lerenn/frontend",
			Branch:    "test-branch",
			Path:      "frontend",
			Workspace: "project.code-workspace",
		},
	}
	mockStatus.EXPECT().ListAllWorktrees().Return(worktrees, nil).AnyTimes()

	// Mock Git worktree deletion
	mockGit.EXPECT().RemoveWorktree("frontend", "/test/base/path/github.com/lerenn/frontend/test-branch").Return(nil).AnyTimes()

	// Mock file system operations
	mockFS.EXPECT().RemoveAll("/test/base/path/github.com/lerenn/frontend/test-branch").Return(nil).AnyTimes()
	mockFS.EXPECT().RemoveAll("/test/base/path/workspaces/project-test-branch.code-workspace").Return(nil).AnyTimes()

	// Mock status removal
	mockStatus.EXPECT().RemoveWorktree("github.com/lerenn/frontend", "test-branch").Return(nil).AnyTimes()

	err := wtm.DeleteWorkTree("test-branch", false)
	assert.NoError(t, err)
}

func TestWTM_ListWorktrees_NoProjectFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	wtm := NewWTM(createTestConfig())
	c := wtm.(*realWTM)
	c.fs = mockFS

	// Mock single repo detection - no .git found
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - no workspace files found
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{}, nil).Times(1)

	result, _, err := wtm.ListWorktrees()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no Git repository or workspace found")
	assert.Nil(t, result)
}

func TestWTM_ListWorktrees_ProjectDetectionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	wtm := NewWTM(createTestConfig())
	c := wtm.(*realWTM)
	c.fs = mockFS

	// Mock single repo detection - no .git found
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection error
	mockFS.EXPECT().Glob("*.code-workspace").Return(nil, assert.AnError).Times(1)

	result, _, err := wtm.ListWorktrees()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to detect project mode")
	assert.Nil(t, result)
}
