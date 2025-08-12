//go:build unit

package cgwt

import (
	"testing"

	"github.com/lerenn/cgwt/pkg/fs"
	"github.com/lerenn/cgwt/pkg/git"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCGWT_Run_WorkspaceMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	cgwt := NewCGWT(createTestConfig())
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	// Mock single repo detection - no .git found (called once: detectProjectType)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file (called once: detectProjectType)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(1)

	// Mock reading workspace file (called twice: once for display, once for validation)
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
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil).Times(2)

	// Mock repository validation for frontend
	mockFS.EXPECT().Exists("frontend").Return(true, nil).AnyTimes()
	mockFS.EXPECT().Exists("frontend/.git").Return(true, nil).AnyTimes()

	// Mock repository validation for backend
	mockFS.EXPECT().Exists("backend").Return(true, nil).AnyTimes()
	mockFS.EXPECT().Exists("backend/.git").Return(true, nil).AnyTimes()

	// Mock Git status for validation (called for each repository, order may vary)
	mockGit.EXPECT().Status("frontend").Return("On branch main", nil).AnyTimes()
	mockGit.EXPECT().Status("backend").Return("On branch main", nil).AnyTimes()

	err := cgwt.CreateWorkTree()
	assert.NoError(t, err)
}

func TestCGWT_Run_InvalidWorkspaceJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWT(createTestConfig())
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called once: detectProjectType)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file (called once: detectProjectType)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(1)

	// Mock reading workspace file with invalid JSON (called once: detectProjectType)
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(`{invalid json`), nil).Times(1)

	err := cgwt.CreateWorkTree()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid .code-workspace file: malformed JSON")
}

func TestCGWT_Run_MissingRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWT(createTestConfig())
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called once: detectProjectType)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file (called once: detectProjectType)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(1)

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

	err := cgwt.CreateWorkTree()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository not found in workspace: ./frontend")
}

func TestCGWT_Run_InvalidRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	cgwt := NewCGWT(createTestConfig())
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	// Mock single repo detection - no .git found (called once: detectProjectType)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file (called once: detectProjectType)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(1)

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

	err := cgwt.CreateWorkTree()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid repository in workspace: ./frontend - .git directory not found")
}

func TestCGWT_Run_GitStatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	cgwt := NewCGWT(createTestConfig())
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	// Mock single repo detection - no .git found (called once: detectProjectType)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file (called once: detectProjectType)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(1)

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

	err := cgwt.CreateWorkTree()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid repository in workspace: ./frontend - assert.AnError general error for testing")
}

func TestCGWT_Run_MultipleWorkspaceFiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWT(createTestConfig())
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called once: detectProjectType)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find multiple workspace files (called once: detectProjectType)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project1.code-workspace", "project2.code-workspace"}, nil).Times(1)

	err := cgwt.CreateWorkTree()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to handle multiple workspaces")
}

func TestCGWT_Run_WorkspaceFileReadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWT(createTestConfig())
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called once: detectProjectType)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file (called once: detectProjectType)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(1)

	// Mock reading workspace file error (called once: detectProjectType)
	mockFS.EXPECT().ReadFile("project.code-workspace").Return(nil, assert.AnError).Times(1)

	err := cgwt.CreateWorkTree()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse workspace file")
}

func TestCGWT_Run_WorkspaceGlobError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWT(createTestConfig())
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called once: detectProjectType)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection error (called once: detectProjectType)
	mockFS.EXPECT().Glob("*.code-workspace").Return(nil, assert.AnError).Times(1)

	err := cgwt.CreateWorkTree()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to detect workspace mode")
}

func TestCGWT_Run_WorkspaceVerboseMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	cgwt := NewCGWT(createTestConfig())
	cgwt.SetVerbose(true)
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	// Mock single repo detection - no .git found (called once: detectProjectType)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file (called once: detectProjectType)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(1)

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

	// Mock Git status for validation
	mockGit.EXPECT().Status("frontend").Return("On branch main", nil).AnyTimes()

	err := cgwt.CreateWorkTree()
	assert.NoError(t, err)
}

func TestCGWT_Run_EmptyWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWT(createTestConfig())
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called once: detectProjectType)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(1)

	// Mock workspace detection - find workspace file (called once: detectProjectType)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(1)

	// Mock reading workspace file with empty folders (called multiple times for display and validation)
	workspaceJSON := `{
		"folders": []
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil).AnyTimes()

	err := cgwt.CreateWorkTree()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace file must contain non-empty folders array")
}
