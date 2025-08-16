package forge

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v62/github"
	"github.com/lerenn/wtm/pkg/git"
)

const (
	// GitHubName is the name identifier for GitHub forge.
	GitHubName = "github"
	// GitHubDomain is the GitHub domain for URL validation.
	GitHubDomain = "github.com"
	// MaxTitleLength is the maximum length for sanitized issue titles in branch names.
	MaxTitleLength = 80
)

// GitHub represents the GitHub forge implementation.
type GitHub struct {
	client *github.Client
	git    git.Git
}

// NewGitHub creates a new GitHub forge instance.
func NewGitHub() *GitHub {
	var client *github.Client

	// Add authentication if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		client = github.NewTokenClient(context.Background(), token)
	} else {
		client = github.NewClient(nil)
	}

	return &GitHub{
		client: client,
		git:    git.NewGit(),
	}
}

// Name returns the name of the forge.
func (g *GitHub) Name() string {
	return GitHubName
}

// GetIssueInfo fetches issue information from GitHub API.
func (g *GitHub) GetIssueInfo(issueRef string) (*IssueInfo, error) {
	// Parse the issue reference to get repository and issue number
	ref, err := g.parseIssueReference(issueRef)
	if err != nil {
		return nil, err
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fetch the issue using the GitHub client
	issue, resp, err := g.client.Issues.Get(ctx, ref.Owner, ref.Repository, ref.IssueNumber)
	if err != nil {
		return nil, g.handleGitHubError(err, resp, ref.IssueNumber)
	}

	// Validate issue state
	if issue.GetState() != "open" {
		return nil, fmt.Errorf("%w: issue #%d", ErrIssueClosed, issue.GetNumber())
	}

	return &IssueInfo{
		Number:      issue.GetNumber(),
		Title:       issue.GetTitle(),
		Description: issue.GetBody(),
		State:       issue.GetState(),
		URL:         issue.GetHTMLURL(),
		Repository:  ref.Repository,
		Owner:       ref.Owner,
	}, nil
}

// parseIssueReference parses the issue reference and handles context extraction.
func (g *GitHub) parseIssueReference(issueRef string) (*IssueReference, error) {
	ref, err := g.ParseIssueReference(issueRef)
	if err != nil {
		// If it's an issue number format error, try to extract repository info from current repo
		if strings.Contains(err.Error(), "issue number format requires repository context") {
			ref, err = g.parseIssueNumberWithContext(issueRef)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", ErrInvalidIssueRef, err)
			}
		} else {
			return nil, fmt.Errorf("%w: %w", ErrInvalidIssueRef, err)
		}
	}
	return ref, nil
}

// handleGitHubError handles GitHub API errors and returns appropriate error messages.
func (g *GitHub) handleGitHubError(err error, resp *github.Response, issueNumber int) error {
	if resp != nil {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("%w: issue #%d", ErrIssueNotFound, issueNumber)
		case http.StatusUnauthorized:
			return fmt.Errorf("%w: check GITHUB_TOKEN environment variable", ErrUnauthorized)
		case http.StatusForbidden:
			// Check if it's rate limiting
			if resp.Header.Get("X-RateLimit-Remaining") == "0" {
				return fmt.Errorf("%w: GitHub API rate limit exceeded", ErrRateLimited)
			}
			return fmt.Errorf("%w: access forbidden", ErrUnauthorized)
		}
	}
	return fmt.Errorf("failed to fetch issue: %w", err)
}

// ValidateForgeRepository validates that repository has GitHub remote origin.
func (g *GitHub) ValidateForgeRepository(repoPath string) error {
	// Get the remote origin URL
	originURL, err := g.git.GetRemoteURL(repoPath, "origin")
	if err != nil {
		return fmt.Errorf("failed to get remote origin: %w", err)
	}

	// Check if it's a GitHub repository by looking for github.com in the URL
	// This handles both HTTPS (https://github.com/owner/repo.git) and SSH (git@github.com:owner/repo.git) URLs
	if !strings.Contains(originURL, GitHubDomain) {
		return fmt.Errorf("repository does not have GitHub as remote origin")
	}

	return nil
}

// ParseIssueReference parses various issue reference formats.
func (g *GitHub) ParseIssueReference(issueRef string) (*IssueReference, error) {
	// Try different formats

	// 1. GitHub issue URL: https://github.com/owner/repo/issues/123
	if strings.Contains(issueRef, "github.com") && strings.Contains(issueRef, "/issues/") {
		return g.parseGitHubURL(issueRef)
	}

	// 2. Owner/repo#issue format: owner/repo#123
	if strings.Contains(issueRef, "#") {
		return g.parseOwnerRepoFormat(issueRef)
	}

	// 3. Issue number only: 123 (requires current repository to be GitHub)
	// This will be handled by the caller who needs to provide the repository context
	if matched, _ := regexp.MatchString(`^\d+$`, issueRef); matched {
		return nil, fmt.Errorf("issue number format requires repository context")
	}

	return nil, fmt.Errorf("unsupported issue reference format: %s", issueRef)
}

// parseIssueNumberWithContext parses an issue number and extracts repository info from current repo.
func (g *GitHub) parseIssueNumberWithContext(issueRef string) (*IssueReference, error) {
	// Validate that it's a number
	if matched, _ := regexp.MatchString(`^\d+$`, issueRef); !matched {
		return nil, fmt.Errorf("invalid issue number format: %s", issueRef)
	}

	// Get the remote origin URL to extract owner and repository
	originURL, err := g.git.GetRemoteURL(".", "origin")
	if err != nil {
		return nil, fmt.Errorf("failed to get remote origin: %w", err)
	}

	// Extract owner and repository from the remote URL
	// Handle both HTTPS and SSH formats
	var owner, repo string

	// Try HTTPS format: https://github.com/owner/repo.git
	if strings.Contains(originURL, "https://github.com/") {
		re := regexp.MustCompile(`github\.com/([^/]+)/([^/]+?)(?:\.git)?$`)
		matches := re.FindStringSubmatch(originURL)
		if len(matches) == 3 {
			owner = matches[1]
			repo = matches[2]
		}
	} else if strings.Contains(originURL, "git@github.com:") {
		// SSH format: git@github.com:owner/repo.git
		re := regexp.MustCompile(`github\.com:([^/]+)/([^/]+?)(?:\.git)?$`)
		matches := re.FindStringSubmatch(originURL)
		if len(matches) == 3 {
			owner = matches[1]
			repo = matches[2]
		}
	}

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("failed to extract owner and repository from remote origin: %s", originURL)
	}

	// Convert issue number to int
	var issueNumber int
	if _, err := fmt.Sscanf(issueRef, "%d", &issueNumber); err != nil {
		return nil, fmt.Errorf("invalid issue number: %s", issueRef)
	}

	// Build the URL
	url := fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, repo, issueNumber)

	return &IssueReference{
		Owner:       owner,
		Repository:  repo,
		IssueNumber: issueNumber,
		URL:         url,
	}, nil
}

// parseGitHubURL parses GitHub issue URLs.
func (g *GitHub) parseGitHubURL(urlStr string) (*IssueReference, error) {
	// Extract owner, repo, and issue number from URL
	// https://github.com/owner/repo/issues/123
	re := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/issues/(\d+)`)
	matches := re.FindStringSubmatch(urlStr)
	if len(matches) != 4 {
		return nil, fmt.Errorf("invalid GitHub issue URL format")
	}

	owner := matches[1]
	repo := matches[2]
	issueNum := matches[3]

	// Convert issue number to int
	var issueNumber int
			if _, err := fmt.Sscanf(issueNum, "%d", &issueNumber); err != nil {
			return nil, fmt.Errorf("invalid issue number: %s", issueNum)
		}

	return &IssueReference{
		Owner:       owner,
		Repository:  repo,
		IssueNumber: issueNumber,
		URL:         urlStr,
	}, nil
}

// parseOwnerRepoFormat parses owner/repo#issue format.
func (g *GitHub) parseOwnerRepoFormat(ref string) (*IssueReference, error) {
	// owner/repo#123
	parts := strings.Split(ref, "#")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid owner/repo#issue format")
	}

	ownerRepo := parts[0]
	issueNum := parts[1]

	// Split owner/repo
	ownerRepoParts := strings.Split(ownerRepo, "/")
	if len(ownerRepoParts) != 2 {
		return nil, fmt.Errorf("invalid owner/repo format")
	}

	owner := ownerRepoParts[0]
	repo := ownerRepoParts[1]

	// Convert issue number to int
	var issueNumber int
	if _, err := fmt.Sscanf(issueNum, "%d", &issueNumber); err != nil {
		return nil, fmt.Errorf("invalid issue number: %s", issueNum)
	}

	// Build the URL
	url := fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, repo, issueNumber)

	return &IssueReference{
		Owner:       owner,
		Repository:  repo,
		IssueNumber: issueNumber,
		URL:         url,
	}, nil
}

// GenerateBranchName generates branch name from issue information.
func (g *GitHub) GenerateBranchName(issueInfo *IssueInfo) string {
	// Format: <issue-nb>-<sanitized-issue-title>

	// Sanitize the title
	sanitizedTitle := g.sanitizeTitle(issueInfo.Title)

	// Limit length to MaxTitleLength
	if len(sanitizedTitle) > MaxTitleLength {
		sanitizedTitle = sanitizedTitle[:MaxTitleLength]
	}

	// Ensure no trailing hyphens
	sanitizedTitle = strings.Trim(sanitizedTitle, "-")

	return fmt.Sprintf("%d-%s", issueInfo.Number, sanitizedTitle)
}

// sanitizeTitle sanitizes the issue title for use in branch names.
func (g *GitHub) sanitizeTitle(title string) string {
	// Convert to lowercase
	title = strings.ToLower(title)

	// Replace spaces with hyphens
	title = strings.ReplaceAll(title, " ", "-")

	// Replace all non-alphanumeric characters with hyphens
	re := regexp.MustCompile(`[^a-z0-9-]`)
	title = re.ReplaceAllString(title, "-")

	// Replace multiple consecutive hyphens with single hyphen
	re = regexp.MustCompile(`-+`)
	title = re.ReplaceAllString(title, "-")

	return title
}
