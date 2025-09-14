//go:build unit

package ide

import (
	"errors"
	"testing"

	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
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
			name:        "existing IDE - VS Code",
			ideName:     VSCodeName,
			expectError: false,
		},
		{
			name:        "existing IDE - Cursor",
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

			mockFS := fsmocks.NewMockFS(ctrl)

			manager := NewManager(NewManagerParams{
				FS:     mockFS,
				Logger: logger.NewNoopLogger(),
			})

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
			name:        "successful IDE opening - VS Code",
			ideName:     VSCodeName,
			path:        "/path/to/repo",
			verbose:     false,
			expectError: false,
		},
		{
			name:        "successful IDE opening - Cursor",
			ideName:     CursorName,
			path:        "/path/to/repo",
			verbose:     false,
			expectError: false,
		},
		{
			name:        "IDE not installed - VS Code",
			ideName:     VSCodeName,
			path:        "/path/to/repo",
			verbose:     false,
			expectError: true,
			errorType:   ErrIDENotInstalled,
		},
		{
			name:        "IDE not installed - Cursor",
			ideName:     CursorName,
			path:        "/path/to/repo",
			verbose:     false,
			expectError: true,
			errorType:   ErrIDENotInstalled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := fsmocks.NewMockFS(ctrl)

			manager := NewManager(NewManagerParams{
				FS:     mockFS,
				Logger: logger.NewNoopLogger(),
			})

			// Mock IDE installation check
			if tt.expectError && tt.errorType == ErrIDENotInstalled {
				mockFS.EXPECT().Which(gomock.Any()).Return("", errors.New("not found"))
			} else {
				mockFS.EXPECT().Which(gomock.Any()).Return("/usr/bin/ide", nil)
				// All IDEs expect path with trailing slash
				expectedPath := tt.path
				if expectedPath[len(expectedPath)-1] != '/' {
					expectedPath += "/"
				}
				mockFS.EXPECT().ExecuteCommand(gomock.Any(), expectedPath).Return(nil)
			}

			err := manager.OpenIDE(tt.ideName, tt.path)

			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errorType)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVSCode_IsInstalled(t *testing.T) {
	tests := []struct {
		name           string
		whichReturns   string
		whichError     error
		expectedResult bool
	}{
		{
			name:           "VS Code is installed",
			whichReturns:   "/usr/bin/code",
			whichError:     nil,
			expectedResult: true,
		},
		{
			name:           "VS Code is not installed",
			whichReturns:   "",
			whichError:     errors.New("not found"),
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := fsmocks.NewMockFS(ctrl)
			mockFS.EXPECT().Which(VSCodeCommand).Return(tt.whichReturns, tt.whichError)

			vscode := NewVSCode(mockFS)
			result := vscode.IsInstalled()

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestVSCode_OpenRepository(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "successful VS Code opening",
			path:        "/path/to/repo",
			expectError: false,
		},
		{
			name:        "VS Code command fails",
			path:        "/path/to/repo",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := fsmocks.NewMockFS(ctrl)

			// VS Code expects path with trailing slash
			expectedPath := tt.path
			if expectedPath[len(expectedPath)-1] != '/' {
				expectedPath += "/"
			}

			if tt.expectError {
				mockFS.EXPECT().ExecuteCommand(VSCodeCommand, expectedPath).Return(errors.New("command failed"))
			} else {
				mockFS.EXPECT().ExecuteCommand(VSCodeCommand, expectedPath).Return(nil)
			}

			vscode := NewVSCode(mockFS)
			err := vscode.OpenRepository(tt.path)

			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrIDEExecutionFailed)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCursor_IsInstalled(t *testing.T) {
	tests := []struct {
		name           string
		whichReturns   string
		whichError     error
		expectedResult bool
	}{
		{
			name:           "Cursor is installed",
			whichReturns:   "/usr/bin/cursor",
			whichError:     nil,
			expectedResult: true,
		},
		{
			name:           "Cursor is not installed",
			whichReturns:   "",
			whichError:     errors.New("not found"),
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := fsmocks.NewMockFS(ctrl)
			mockFS.EXPECT().Which(CursorCommand).Return(tt.whichReturns, tt.whichError)

			cursor := NewCursor(mockFS)
			result := cursor.IsInstalled()

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestCursor_OpenRepository(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "successful Cursor opening",
			path:        "/path/to/repo",
			expectError: false,
		},
		{
			name:        "Cursor command fails",
			path:        "/path/to/repo",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := fsmocks.NewMockFS(ctrl)

			// Cursor expects path with trailing slash
			expectedPath := tt.path
			if expectedPath[len(expectedPath)-1] != '/' {
				expectedPath += "/"
			}

			if tt.expectError {
				mockFS.EXPECT().ExecuteCommand(CursorCommand, expectedPath).Return(errors.New("command failed"))
			} else {
				mockFS.EXPECT().ExecuteCommand(CursorCommand, expectedPath).Return(nil)
			}

			cursor := NewCursor(mockFS)
			err := cursor.OpenRepository(tt.path)

			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrIDEExecutionFailed)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDummy_IsInstalled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	dummy := NewDummy(mockFS)

	// Dummy IDE should always be installed
	result := dummy.IsInstalled()
	assert.True(t, result)
}

func TestDummy_OpenRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	dummy := NewDummy(mockFS)

	// Dummy IDE should always succeed
	err := dummy.OpenRepository("/path/to/repo")
	assert.NoError(t, err)
}
