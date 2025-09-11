// Package mode provides constants for detecting and handling different project modes.
package mode

// Mode represents the type of project detected.
type Mode int

// Mode constants.
const (
	ModeNone Mode = iota
	ModeSingleRepo
	ModeWorkspace
)
