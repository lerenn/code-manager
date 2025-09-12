// Package prompt provides interactive prompt functionality for CM.
package prompt

import "errors"

// Error definitions for prompt package.
var (
	ErrInvalidConfirmationInput = errors.New("invalid input: please enter 'y' or 'n'")
)
