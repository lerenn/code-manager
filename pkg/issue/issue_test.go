package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIssueInfo_Fields(t *testing.T) {
	info := &Info{
		Number:      123,
		Title:       "Test Issue",
		Description: "This is a test issue",
		State:       "open",
		URL:         "https://github.com/test/repo/issues/123",
		Repository:  "repo",
		Owner:       "test",
	}

	assert.Equal(t, 123, info.Number)
	assert.Equal(t, "Test Issue", info.Title)
	assert.Equal(t, "This is a test issue", info.Description)
	assert.Equal(t, "open", info.State)
	assert.Equal(t, "https://github.com/test/repo/issues/123", info.URL)
	assert.Equal(t, "repo", info.Repository)
	assert.Equal(t, "test", info.Owner)
}

func TestIssueReference_Fields(t *testing.T) {
	ref := &Reference{
		Owner:       "test",
		Repository:  "repo",
		IssueNumber: 123,
		URL:         "https://github.com/test/repo/issues/123",
	}

	assert.Equal(t, "test", ref.Owner)
	assert.Equal(t, "repo", ref.Repository)
	assert.Equal(t, 123, ref.IssueNumber)
	assert.Equal(t, "https://github.com/test/repo/issues/123", ref.URL)
}

func TestErrorTypes(t *testing.T) {
	assert.Equal(t, "issue not found", ErrIssueNotFound.Error())
	assert.Equal(t, "issue is closed, only open issues are supported", ErrIssueClosed.Error())
	assert.Equal(t, "invalid issue reference format", ErrInvalidIssueReference.Error())
	assert.Equal(t, "issue number format requires repository context", ErrIssueNumberRequiresContext.Error())
}
