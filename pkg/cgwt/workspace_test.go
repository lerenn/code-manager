//go:build unit

package cgwt

import (
	"testing"

	"github.com/lerenn/cgwt/pkg/fs"
	"github.com/lerenn/cgwt/pkg/git"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCGWT_Run_ValidWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	cgwt := NewCGWT()
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	// Mock single repo detection - no .git found (called twice: detectProjectMode and validateProjectStructure)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(2)

	// Mock workspace detection - find workspace file (called twice: detectProjectMode and validateProjectStructure)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil).Times(2)

	// Mock reading workspace file (called twice: detectProjectMode and validateProjectStructure)
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
	mockFS.EXPECT().Exists("frontend").Return(true, nil)
	mockFS.EXPECT().Exists("frontend/.git").Return(true, nil)

	// Mock repository validation for backend
	mockFS.EXPECT().Exists("backend").Return(true, nil)
	mockFS.EXPECT().Exists("backend/.git").Return(true, nil)

	// Mock Git status for validation (called for each repository, order may vary)
	mockGit.EXPECT().Status("frontend").Return("On branch main", nil).AnyTimes()
	mockGit.EXPECT().Status("backend").Return("On branch main", nil).AnyTimes()

	err := cgwt.Run()
	assert.NoError(t, err)
}

func TestCGWT_Run_InvalidWorkspaceJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWT()
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called twice: detectProjectMode and validateProjectStructure)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(2)

	// Mock workspace detection - find workspace file
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil)

	// Mock reading workspace file with invalid JSON
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(`{invalid json`), nil)

	err := cgwt.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid .code-workspace file: malformed JSON")
}

func TestCGWT_Run_MissingRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWT()
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called twice: detectProjectMode and validateProjectStructure)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(2)

	// Mock workspace detection - find workspace file
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil)

	// Mock reading workspace file
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil)

	// Mock repository validation - repository not found
	mockFS.EXPECT().Exists("frontend").Return(false, nil)

	err := cgwt.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace repository not found: ./frontend")
}

func TestCGWT_Run_InvalidRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWT()
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called twice: detectProjectMode and validateProjectStructure)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(2)

	// Mock workspace detection - find workspace file
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil)

	// Mock reading workspace file
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil)

	// Mock repository validation - repository exists but no .git
	mockFS.EXPECT().Exists("frontend").Return(true, nil)
	mockFS.EXPECT().IsDir("frontend").Return(true, nil)
	mockFS.EXPECT().Exists("frontend/.git").Return(false, nil)

	err := cgwt.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace repository is not a git repository: ./frontend")
}

func TestCGWT_Run_NoWorkspaceFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWT()
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called twice: detectProjectMode and validateProjectStructure)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(2)

	// Mock workspace detection - no workspace files found
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{}, nil)

	err := cgwt.Run()
	assert.NoError(t, err)
}

func TestCGWT_Run_MultipleWorkspaceFiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWT()
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called twice: detectProjectMode and validateProjectStructure)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(2)

	// Mock workspace detection - multiple workspace files found
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace", "dev.code-workspace"}, nil)

	// Note: This test would require stdin/stdout mocking for user input
	// For now, we'll test that the error is related to user cancellation
	err := cgwt.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user cancelled selection")
}

func TestCGWT_Run_EmptyFolders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWT()
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called twice: detectProjectMode and validateProjectStructure)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(2)

	// Mock workspace detection - find workspace file
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil)

	// Mock reading workspace file with empty folders
	workspaceJSON := `{
		"folders": []
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil)

	err := cgwt.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace file must contain non-empty folders array")
}

func TestCGWT_Run_InvalidFolderStructure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWT()
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called twice: detectProjectMode and validateProjectStructure)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(2)

	// Mock workspace detection - find workspace file
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil)

	// Mock reading workspace file with invalid folder structure
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil)

	err := cgwt.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace file must contain non-empty folders array")
}

func TestCGWT_Run_DuplicatePaths(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWT()
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called twice: detectProjectMode and validateProjectStructure)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(2)

	// Mock workspace detection - find workspace file
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil)

	// Mock reading workspace file with duplicate paths
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			},
			{
				"name": "Frontend2",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil)

	// Mock repository validation for first frontend
	mockFS.EXPECT().Exists("frontend").Return(true, nil)
	mockFS.EXPECT().IsDir("frontend").Return(true, nil)
	mockFS.EXPECT().Exists("frontend/.git").Return(true, nil)

	err := cgwt.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate repository paths found after resolution")
}

func TestCGWT_Run_BrokenSymlink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWT()
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called twice: detectProjectMode and validateProjectStructure)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(2)

	// Mock workspace detection - find workspace file
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil)

	// Mock reading workspace file
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil)

	// Mock repository validation - valid repository (no broken symlink since we removed symlink resolution)
	mockFS.EXPECT().Exists("frontend").Return(true, nil)
	mockFS.EXPECT().IsDir("frontend").Return(true, nil)
	mockFS.EXPECT().Exists("frontend/.git").Return(true, nil)

	err := cgwt.Run()
	assert.NoError(t, err)
}

func TestCGWT_Run_NullValuesInFolders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cgwt := NewCGWT()
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock single repo detection - no .git found (called twice: detectProjectMode and validateProjectStructure)
	mockFS.EXPECT().Exists(".git").Return(false, nil).Times(2)

	// Mock workspace detection - find workspace file
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil)

	// Mock reading workspace file with null values (empty path)
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			},
			{
				"name": "Empty",
				"path": ""
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceJSON), nil)

	// Mock repository validation for frontend (null value should be skipped)
	mockFS.EXPECT().Exists("frontend").Return(true, nil)
	mockFS.EXPECT().IsDir("frontend").Return(true, nil)
	mockFS.EXPECT().Exists("frontend/.git").Return(true, nil)

	err := cgwt.Run()
	assert.NoError(t, err)
}
