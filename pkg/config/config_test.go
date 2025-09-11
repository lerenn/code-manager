//go:build unit

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				RepositoriesDir: filepath.Join(t.TempDir(), "test", "path"),
				WorkspacesDir:   filepath.Join(t.TempDir(), "test", "workspaces"),
				StatusFile:      filepath.Join(t.TempDir(), "test", "status.yaml"),
			},
			wantErr: false,
		},
		{
			name: "valid config without worktrees_dir",
			config: Config{
				RepositoriesDir: filepath.Join(t.TempDir(), "test", "path"),
				WorkspacesDir:   filepath.Join(t.TempDir(), "test", "workspaces"),
				StatusFile:      filepath.Join(t.TempDir(), "test", "status.yaml"),
			},
			wantErr: false,
		},
		{
			name: "empty base path",
			config: Config{
				RepositoriesDir: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if tt.config.RepositoriesDir == "" {
					assert.ErrorIs(t, err, ErrRepositoriesDirEmpty)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRealManager_DefaultConfig(t *testing.T) {
	manager := NewManager()
	config := manager.DefaultConfig()

	assert.NotNil(t, config)
	assert.NotEmpty(t, config.RepositoriesDir)
	assert.Contains(t, config.RepositoriesDir, "Code")
}

func TestRealManager_LoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	// Write valid YAML config with a path that can be created
	validYAML := `repositories_dir: ` + filepath.Join(tempDir, "custom", "path", "to", "cm") + `
workspaces_dir: ` + filepath.Join(tempDir, "custom", "path", "to", "workspaces") + `
status_file: ` + filepath.Join(tempDir, "custom", "path", "to", "status.yaml") + `
`
	err := os.WriteFile(configPath, []byte(validYAML), 0644)
	assert.NoError(t, err)

	manager := NewManager()
	config, err := manager.LoadConfig(configPath)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, filepath.Join(tempDir, "custom", "path", "to", "cm"), config.RepositoriesDir)
	assert.Equal(t, filepath.Join(tempDir, "custom", "path", "to", "workspaces"), config.WorkspacesDir)
	assert.Equal(t, filepath.Join(tempDir, "custom", "path", "to", "status.yaml"), config.StatusFile)
}

func TestRealManager_LoadConfig_FileNotFound(t *testing.T) {
	manager := NewManager()
	config, err := manager.LoadConfig("/nonexistent/path/config.yaml")

	assert.Equal(t, Config{}, config)
	assert.ErrorIs(t, err, ErrConfigNotInitialized)
}

func TestRealManager_LoadConfig_InvalidYAML(t *testing.T) {
	// Create a temporary config file with invalid YAML
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid-config.yaml")

	// Write invalid YAML
	invalidYAML := `base_path: /custom/path/to/cm
invalid: yaml: structure: here`
	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	assert.NoError(t, err)

	manager := NewManager()
	config, err := manager.LoadConfig(configPath)

	assert.Equal(t, Config{}, config)
	assert.ErrorIs(t, err, ErrConfigFileParse)
}

func TestLoadConfigWithFallback_WithValidFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	// Write valid YAML config with a path that can be created
	validYAML := `repositories_dir: ` + filepath.Join(tempDir, "custom", "path", "to", "cm") + `
workspaces_dir: ` + filepath.Join(tempDir, "custom", "path", "to", "workspaces") + `
status_file: ` + filepath.Join(tempDir, "custom", "path", "to", "status.yaml") + `
`
	err := os.WriteFile(configPath, []byte(validYAML), 0644)
	assert.NoError(t, err)

	config, err := LoadConfigWithFallback(configPath)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, filepath.Join(tempDir, "custom", "path", "to", "cm"), config.RepositoriesDir)
	assert.Equal(t, filepath.Join(tempDir, "custom", "path", "to", "workspaces"), config.WorkspacesDir)
	assert.Equal(t, filepath.Join(tempDir, "custom", "path", "to", "status.yaml"), config.StatusFile)
}

func TestLoadConfigWithFallback_WithMissingFile(t *testing.T) {
	config, err := LoadConfigWithFallback("/nonexistent/path/config.yaml")

	assert.NoError(t, err) // Should not error, should fallback to default
	assert.NotNil(t, config)
	assert.Contains(t, config.RepositoriesDir, "Code")
}

func TestConfig_ExpandTildes(t *testing.T) {
	config := &Config{
		RepositoriesDir: "~/.cm-test",
		StatusFile:      "~/.cm-test/status.yaml",
	}

	err := config.expandTildes()
	assert.NoError(t, err)

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(homeDir, ".cm-test"), config.RepositoriesDir)
	assert.Equal(t, filepath.Join(homeDir, ".cm-test", "status.yaml"), config.StatusFile)
}

func TestConfig_ExpandTildes_NoTildes(t *testing.T) {
	originalRepositoriesDir := "/custom/path"
	originalStatusFile := "/custom/path/status.yaml"

	config := &Config{
		RepositoriesDir: originalRepositoriesDir,
		StatusFile:      originalStatusFile,
	}

	err := config.expandTildes()
	assert.NoError(t, err)

	// Paths should remain unchanged
	assert.Equal(t, originalRepositoriesDir, config.RepositoriesDir)
	assert.Equal(t, originalStatusFile, config.StatusFile)
}

func TestRealManager_LoadConfig_WithTildes(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	// Write YAML config with tildes
	validYAML := `repositories_dir: ~/.cm-test
workspaces_dir: ~/.cm-test/workspaces
status_file: ~/.cm-test/status.yaml
`
	err := os.WriteFile(configPath, []byte(validYAML), 0644)
	assert.NoError(t, err)

	manager := NewManager()
	config, err := manager.LoadConfig(configPath)

	assert.NoError(t, err)
	assert.NotNil(t, config)

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(homeDir, ".cm-test"), config.RepositoriesDir)
	assert.Equal(t, filepath.Join(homeDir, ".cm-test", "workspaces"), config.WorkspacesDir)
	assert.Equal(t, filepath.Join(homeDir, ".cm-test", "status.yaml"), config.StatusFile)
}
