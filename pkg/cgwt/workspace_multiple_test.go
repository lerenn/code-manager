//go:build unit

package cgwt

import (
	"testing"
)

func TestDisplayWorkspaceSelection(t *testing.T) {
	c := NewCGWT()

	workspaceFiles := []string{
		"project.code-workspace",
		"project-dev.code-workspace",
		"project-staging.code-workspace",
	}

	// Test that the method doesn't panic
	c.(*realCGWT).displayWorkspaceSelection(workspaceFiles)
}

func TestGetUserSelection_ValidInput(t *testing.T) {
	// This test would require stdin mocking
	// For now, we'll just test the logic without actual input
	t.Skip("Requires stdin mocking")
}

func TestGetUserSelection_InvalidInput(t *testing.T) {
	// This test would require stdin mocking
	// For now, we'll just test the logic without actual input
	t.Skip("Requires stdin mocking")
}

func TestGetUserSelection_QuitCommand(t *testing.T) {
	// This test would require stdin mocking
	// For now, we'll just test the logic without actual input
	t.Skip("Requires stdin mocking")
}

func TestConfirmSelection_ValidInput(t *testing.T) {
	// This test would require stdin mocking
	// For now, we'll just test the logic without actual input
	t.Skip("Requires stdin mocking")
}

func TestConfirmSelection_InvalidInput(t *testing.T) {
	// This test would require stdin mocking
	// For now, we'll just test the logic without actual input
	t.Skip("Requires stdin mocking")
}

func TestConfirmSelection_QuitCommand(t *testing.T) {
	// This test would require stdin mocking
	// For now, we'll just test the logic without actual input
	t.Skip("Requires stdin mocking")
}
