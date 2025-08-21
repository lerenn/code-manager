// Package issue provides data structures and error types for handling forge issues.
package issue

import "errors"

// Issue-specific error types.
var (
	ErrIssueNotFound              = errors.New("issue not found")
	ErrIssueClosed                = errors.New("issue is closed, only open issues are supported")
	ErrInvalidIssueReference      = errors.New("invalid issue reference format")
	ErrIssueNumberRequiresContext = errors.New("issue number format requires repository context")
)
