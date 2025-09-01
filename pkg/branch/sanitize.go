// Package branch provides branch name sanitization functionality.
package branch

import (
	"regexp"
	"strings"
)

// ErrBranchNameEmpty is returned when the branch name is empty.
var ErrBranchNameEmpty = &Error{message: "branch name cannot be empty"}

// ErrBranchNameSingleAt is returned when the branch name is just a single @ character.
var ErrBranchNameSingleAt = &Error{message: "branch name cannot be the single character @"}

// ErrBranchNameContainsAtBrace is returned when the branch name contains the sequence @{.
var ErrBranchNameContainsAtBrace = &Error{message: "branch name cannot contain the sequence @{"}

// ErrBranchNameContainsBackslash is returned when the branch name contains a backslash.
var ErrBranchNameContainsBackslash = &Error{message: "branch name cannot contain backslash"}

// ErrBranchNameEmptyAfterSanitization is returned when the branch name becomes empty after sanitization.
var ErrBranchNameEmptyAfterSanitization = &Error{message: "branch name becomes empty after sanitization"}

// Error represents an error related to branch operations.
type Error struct {
	message string
}

func (e *Error) Error() string {
	return e.message
}

// SanitizeBranchName sanitizes a branch name according to Git's branch naming rules.
// Git imposes the following rules on how references are named:
//   - They can include slash / for hierarchical (directory) grouping, but no slash-separated component
//     can begin with a dot . or end with the sequence .lock.
//   - They cannot have two consecutive dots .. anywhere.
//   - They cannot have ASCII control characters (i.e. bytes whose values are lower than \040, or \177 DEL),
//     space, tilde ~, caret ^, or colon : anywhere.
//   - They cannot have question-mark ?, asterisk *, or open bracket [ anywhere.
//   - They cannot begin or end with a slash / or contain multiple consecutive slashes.
//   - They cannot end with a dot .
//   - They cannot contain a sequence @{.
//   - They cannot be the single character @.
//   - They cannot contain a \.
//   - Additional rule for branch names: They cannot start with a dash -.
func SanitizeBranchName(branchName string) (string, error) {
	if branchName == "" {
		return "", ErrBranchNameEmpty
	}

	// Check for single @ character (not allowed)
	if branchName == "@" {
		return "", ErrBranchNameSingleAt
	}

	// Check for @{ sequence (not allowed)
	if strings.Contains(branchName, "@{") {
		return "", ErrBranchNameContainsAtBrace
	}

	// Check for backslash (not allowed)
	if strings.Contains(branchName, "\\") {
		return "", ErrBranchNameContainsBackslash
	}

	// Replace invalid characters with underscores
	// Git rules: no ASCII control characters (< \040, \177 DEL), space, tilde ~, caret ^, colon :,
	// question-mark ?, asterisk *, open bracket [, close bracket ]
	invalidChars := regexp.MustCompile(`[\x00-\x1F\x7F ~^:?*\[\]#]`)
	sanitized := invalidChars.ReplaceAllString(branchName, "_")

	// Replace consecutive dots with single underscore
	consecutiveDots := regexp.MustCompile(`\.\.+`)
	sanitized = consecutiveDots.ReplaceAllString(sanitized, "_")

	// Replace consecutive slashes with single slash
	consecutiveSlashes := regexp.MustCompile(`/+`)
	sanitized = consecutiveSlashes.ReplaceAllString(sanitized, "/")

	// Remove leading/trailing slashes, dots, and underscores
	sanitized = strings.Trim(sanitized, "/._")

	// Remove leading dash (not allowed for branch names)
	sanitized = strings.TrimPrefix(sanitized, "-")

	// Limit length to 255 characters (filesystem limit)
	if len(sanitized) > 255 {
		sanitized = sanitized[:255]
		// Ensure we don't end with a dot, underscore, or slash
		sanitized = strings.TrimRight(sanitized, "._/")
	}

	if sanitized == "" {
		return "", ErrBranchNameEmptyAfterSanitization
	}

	return sanitized, nil
}
