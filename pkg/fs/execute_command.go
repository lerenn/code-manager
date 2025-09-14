package fs

import "os/exec"

// ExecuteCommand executes a command with arguments in the background.
func (f *realFS) ExecuteCommand(command string, args ...string) error {
	// Create command
	cmd := exec.Command(command, args...)

	// Start command in background (don't wait for completion)
	if err := cmd.Start(); err != nil {
		return err
	}

	// Don't wait for the command to finish, let it run in background
	return nil
}
