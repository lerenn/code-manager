//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestValidateRepositoryName(t *testing.T) {
	tests := []struct {
		name           string
		repositoryName string
		expectedError  string
	}{
		{
			name:           "valid repository name",
			repositoryName: "my-repo",
			expectedError:  "",
		},
		{
			name:           "empty repository name",
			repositoryName: "",
			expectedError:  "repository name cannot be empty",
		},
		{
			name:           "repository name with forward slash (URL format)",
			repositoryName: "github.com/user/repo",
			expectedError:  "",
		},
		{
			name:           "repository name with backslash",
			repositoryName: "my\\repo",
			expectedError:  "repository name cannot contain backslashes",
		},
		{
			name:           "reserved name: dot",
			repositoryName: ".",
			expectedError:  "repository name '.' is reserved",
		},
		{
			name:           "reserved name: double dot",
			repositoryName: "..",
			expectedError:  "repository name '..' is reserved",
		},
		{
			name:           "reserved name: status.yaml",
			repositoryName: "status.yaml",
			expectedError:  "repository name 'status.yaml' is reserved",
		},
		{
			name:           "reserved name: config.yaml",
			repositoryName: "config.yaml",
			expectedError:  "repository name 'config.yaml' is reserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create CM instance with minimal setup
			cmInstance := &realCM{
				logger: logger.NewNoopLogger(),
			}

			// Execute test
			err := cmInstance.validateRepositoryName(tt.repositoryName)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRepositoryNotInWorkspace(t *testing.T) {
	tests := []struct {
		name           string
		repositoryName string
		workspaces     map[string]status.Workspace
		expectedError  string
	}{
		{
			name:           "repository not in any workspace",
			repositoryName: "my-repo",
			workspaces: map[string]status.Workspace{
				"workspace1": {
					Repositories: []string{"other-repo"},
				},
			},
			expectedError: "",
		},
		{
			name:           "repository in workspace",
			repositoryName: "my-repo",
			workspaces: map[string]status.Workspace{
				"workspace1": {
					Repositories: []string{"my-repo", "other-repo"},
				},
			},
			expectedError: "repository 'my-repo' is part of workspace 'workspace1'. Remove it from the workspace before deleting",
		},
		{
			name:           "repository in multiple workspaces",
			repositoryName: "my-repo",
			workspaces: map[string]status.Workspace{
				"workspace1": {
					Repositories: []string{"my-repo"},
				},
				"workspace2": {
					Repositories: []string{"my-repo", "other-repo"},
				},
			},
			expectedError: "is part of workspace",
		},
		{
			name:           "no workspaces exist",
			repositoryName: "my-repo",
			workspaces:     map[string]status.Workspace{},
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create status manager mock
			statusMock := statusmocks.NewMockManager(ctrl)
			statusMock.EXPECT().ListWorkspaces().Return(tt.workspaces, nil)

			// Create CM instance
			cmInstance := &realCM{
				statusManager: statusMock,
				logger:        logger.NewNoopLogger(),
			}

			// Execute test
			err := cmInstance.validateRepositoryNotInWorkspace(tt.repositoryName)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
