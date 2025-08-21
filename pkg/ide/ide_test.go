//go:build unit

package ide

import (
	"errors"
	"testing"

	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestManager_GetIDE(t *testing.T) {
	tests := []struct {
		name        string
		ideName     string
		expectError bool
		errorType   error
	}{
		{
			name:        "existing IDE",
			ideName:     CursorName,
			expectError: false,
		},
		{
			name:        "non-existing IDE",
			ideName:     "unknown-ide",
			expectError: true,
			errorType:   ErrUnsupportedIDE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := fs.NewMockFS(ctrl)
			mockLogger := logger.NewNoopLogger()
			manager := NewManager(mockFS, mockLogger)

			ide, err := manager.GetIDE(tt.ideName)

			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errorType)
				assert.Nil(t, ide)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ide)
				assert.Equal(t, tt.ideName, ide.Name())
			}
		})
	}
}

func TestManager_OpenIDE(t *testing.T) {
	tests := []struct {
		name        string
		ideName     string
		path        string
		verbose     bool
		expectError bool
		errorType   error
	}{
		{
			name:        "successful IDE opening",
			ideName:     CursorName,
			path:        "/path/to/repo",
			verbose:     false,
			expectError: false,
		},
		{
			name:        "IDE not installed",
			ideName:     CursorName,
			path:        "/path/to/repo",
			verbose:     false,
			expectError: true,
			errorType:   ErrIDENotInstalled,
		},
		{
			name:        "IDE execution failed",
			ideName:     CursorName,
			path:        "/path/to/repo",
			verbose:     false,
			expectError: true,
			errorType:   ErrIDEExecutionFailed,
		},
		{
			name:        "unsupported IDE",
			ideName:     "unknown-ide",
			path:        "/path/to/repo",
			verbose:     false,
			expectError: true,
			errorType:   ErrUnsupportedIDE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := fs.NewMockFS(ctrl)
			mockLogger := logger.NewDefaultLogger()
			manager := NewManager(mockFS, mockLogger)

			// Setup mock expectations based on test case
			switch tt.name {
			case "successful IDE opening":
				mockFS.EXPECT().Which(CursorCommand).Return("/usr/local/bin/cursor", nil)
				mockFS.EXPECT().ExecuteCommand(CursorCommand, "/path/to/repo").Return(nil)
			case "IDE not installed":
				mockFS.EXPECT().Which(CursorCommand).Return("", errors.New("command not found"))
			case "IDE execution failed":
				mockFS.EXPECT().Which(CursorCommand).Return("/usr/local/bin/cursor", nil)
				mockFS.EXPECT().ExecuteCommand(CursorCommand, "/path/to/repo").Return(errors.New("execution failed"))
			case "unsupported IDE":
				// No mock setup needed for unsupported IDE
			}

			// Execute test
			err := manager.OpenIDE(tt.ideName, tt.path, tt.verbose)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errorType)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCursor_IsInstalled(t *testing.T) {
	tests := []struct {
		name            string
		whichReturn     string
		whichError      error
		expectInstalled bool
	}{
		{
			name:            "cursor installed",
			whichReturn:     "/usr/local/bin/cursor",
			whichError:      nil,
			expectInstalled: true,
		},
		{
			name:            "cursor not installed",
			whichReturn:     "",
			whichError:      errors.New("command not found"),
			expectInstalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := fs.NewMockFS(ctrl)
			cursor := NewCursor(mockFS)

			// Setup mock expectations
			mockFS.EXPECT().Which(CursorCommand).Return(tt.whichReturn, tt.whichError)

			// Execute test
			installed := cursor.IsInstalled()

			// Assertions
			assert.Equal(t, tt.expectInstalled, installed)
		})
	}
}

func TestCursor_OpenRepository(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		execError   error
		expectError bool
	}{
		{
			name:        "successful opening",
			path:        "/path/to/repo",
			execError:   nil,
			expectError: false,
		},
		{
			name:        "execution failed",
			path:        "/path/to/repo",
			execError:   errors.New("execution failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := fs.NewMockFS(ctrl)
			cursor := NewCursor(mockFS)

			// Setup mock expectations
			mockFS.EXPECT().ExecuteCommand(CursorCommand, tt.path).Return(tt.execError)

			// Execute test
			err := cursor.OpenRepository(tt.path)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrIDEExecutionFailed)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
