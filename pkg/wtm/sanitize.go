package wtm

import (
	"fmt"
	"regexp"
	"strings"
)

// sanitizeBranchName sanitizes a branch name according to Git's branch naming rules.
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
func (c *realWTM) sanitizeBranchName(branchName string) (string, error) {
	if branchName == "" {
		return "", ErrBranchNameEmpty
	}

	// Check for single @ character (not allowed)
	if branchName == "@" {
		return "", fmt.Errorf("branch name cannot be the single character @")
	}

	// Check for @{ sequence (not allowed)
	if strings.Contains(branchName, "@{") {
		return "", fmt.Errorf("branch name cannot contain the sequence @{")
	}

	// Check for backslash (not allowed)
	if strings.Contains(branchName, "\\") {
		return "", fmt.Errorf("branch name cannot contain backslash")
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
