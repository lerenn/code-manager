// Package prompt provides interactive prompt functionality for CM.
package prompt

import "errors"

// Error definitions for prompt package.
var (
	// User input errors.
	ErrInvalidInput  = errors.New("invalid user input")
	ErrUserCancelled = errors.New("user cancelled operation")
)
