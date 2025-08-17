// Package issue provides data structures and error types for handling forge issues.
package issue

import "errors"

// Info represents information about a forge issue.
type Info struct {
	Number      int    `yaml:"number"`
	Title       string `yaml:"title"`
	Description string `yaml:"description,omitempty"`
	State       string `yaml:"state,omitempty"`
	URL         string `yaml:"url,omitempty"`
	Repository  string `yaml:"repository,omitempty"`
	Owner       string `yaml:"owner,omitempty"`
}

// Reference represents a parsed issue reference.
type Reference struct {
	Owner       string
	Repository  string
	IssueNumber int
	URL         string
}

// Issue-specific error types.
var (
	ErrIssueNotFound              = errors.New("issue not found")
	ErrIssueClosed                = errors.New("issue is closed, only open issues are supported")
	ErrInvalidIssueReference      = errors.New("invalid issue reference format")
	ErrIssueNumberRequiresContext = errors.New("issue number format requires repository context")
)
