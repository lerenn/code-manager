package forge

import "errors"

// Forge-specific errors
var (
	ErrUnsupportedForge = errors.New("unsupported forge")
	ErrIssueNotFound    = errors.New("issue not found")
	ErrIssueClosed      = errors.New("issue is closed, only open issues are supported")
	ErrInvalidIssueRef  = errors.New("invalid issue reference format")
	ErrRateLimited      = errors.New("rate limited by forge API")
	ErrUnauthorized     = errors.New("unauthorized access to forge API")
)
