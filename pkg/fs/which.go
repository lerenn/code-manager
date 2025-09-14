package fs

import "os/exec"

// Which finds the executable path for a command using the system's PATH.
func (f *realFS) Which(command string) (string, error) {
	// Use exec.LookPath to find the executable in PATH
	path, err := exec.LookPath(command)
	if err != nil {
		return "", err
	}
	return path, nil
}
