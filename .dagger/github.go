package main

import (
	"context"
	"fmt"
	"strings"

	"code-manager/dagger/internal/dagger"
)

// GitHubReleaseManager handles GitHub release operations.
type GitHubReleaseManager struct{}

// NewGitHubReleaseManager creates a new GitHub release manager.
func NewGitHubReleaseManager() *GitHubReleaseManager {
	return &GitHubReleaseManager{}
}

// getReleaseInfo gets the latest tag and release notes from Git.
func (gh *GitHubReleaseManager) getReleaseInfo(
	ctx context.Context,
	sourceDir *dagger.Directory,
	actualUser string,
	token *dagger.Secret,
) (string, string, error) {
	repo, err := NewGit(ctx, NewGitOptions{
		SrcDir: sourceDir,
		User:   &actualUser,
		Token:  token,
	})
	if err != nil {
		return "", "", err
	}

	latestTag, err := repo.GetLastTag(ctx)
	if err != nil {
		return "", "", err
	}

	releaseNotes, err := repo.GetLastCommitTitle(ctx)
	if err != nil {
		return "", "", err
	}

	return latestTag, releaseNotes, nil
}

// createGitHubRelease creates a new GitHub release and returns the release ID.
func (gh *GitHubReleaseManager) createGitHubRelease(
	ctx context.Context,
	actualUser, latestTag, releaseNotes string,
	token *dagger.Secret,
) (string, error) {
	// Create the release and capture the response to get the release ID
	result, err := dag.Container().
		From("alpine/curl").
		WithSecretVariable("GITHUB_TOKEN", token).
		WithExec([]string{"sh", "-c", fmt.Sprintf(
			"curl -X POST -H \"Authorization: token $GITHUB_TOKEN\" "+
				"-H \"Accept: application/vnd.github.v3+json\" "+
				"https://api.github.com/repos/%s/code-manager/releases "+
				"-d '{\"tag_name\":\"%s\",\"name\":\"Release %s\",\"body\":\"%s\"}'",
			actualUser, latestTag, latestTag, strings.ReplaceAll(releaseNotes, "\"", "\\\""),
		)}).
		Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to create GitHub release: %w", err)
	}

	// Extract the release ID from the response
	// The response should contain "id": <number>
	releaseID := ""
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.Contains(line, "\"id\":") {
			parts := strings.Split(strings.TrimSpace(line), ":")
			if len(parts) >= 2 {
				releaseID = strings.Trim(strings.TrimSpace(parts[1]), ",\"")
				break
			}
		}
	}

	if releaseID == "" {
		return "", fmt.Errorf("failed to extract release ID from response: %s", result)
	}

	return releaseID, nil
}

// getExistingRelease gets an existing release by tag name.
func (gh *GitHubReleaseManager) getExistingRelease(
	ctx context.Context,
	actualUser, latestTag string,
	token *dagger.Secret,
) (string, error) {
	// Get the release by tag
	result, err := dag.Container().
		From("alpine/curl").
		WithSecretVariable("GITHUB_TOKEN", token).
		WithExec([]string{"sh", "-c", fmt.Sprintf(
			"curl -s -H \"Authorization: token $GITHUB_TOKEN\" "+
				"-H \"Accept: application/vnd.github.v3+json\" "+
				"https://api.github.com/repos/%s/code-manager/releases/tags/%s",
			actualUser, latestTag,
		)}).
		Stdout(ctx)

	if err != nil {
		return "", err
	}

	// Check if the response contains an error (release not found)
	if strings.Contains(result, "\"message\":\"Not Found\"") {
		return "", fmt.Errorf("release not found")
	}

	// Extract the release ID from the response
	releaseID := ""
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.Contains(line, "\"id\":") {
			parts := strings.Split(strings.TrimSpace(line), ":")
			if len(parts) >= 2 {
				releaseID = strings.Trim(strings.TrimSpace(parts[1]), ",\"")
				break
			}
		}
	}

	if releaseID == "" {
		return "", fmt.Errorf("failed to extract release ID from response: %s", result)
	}

	return releaseID, nil
}

// getOrCreateRelease gets an existing release or creates a new one.
func (gh *GitHubReleaseManager) getOrCreateRelease(
	ctx context.Context,
	actualUser, latestTag, releaseNotes string,
	token *dagger.Secret,
) (string, error) {
	// First, try to get the existing release
	existingReleaseID, err := gh.getExistingRelease(ctx, actualUser, latestTag, token)
	if err == nil && existingReleaseID != "" {
		return existingReleaseID, nil
	}

	// If release doesn't exist, create it
	return gh.createGitHubRelease(ctx, actualUser, latestTag, releaseNotes, token)
}

// uploadBinary uploads a binary file to a GitHub release.
func (gh *GitHubReleaseManager) uploadBinary(
	ctx context.Context,
	container *dagger.Container,
	platform, actualUser, releaseID string,
	token *dagger.Secret,
) error {
	runnerInfo := GoImageInfo[platform]
	binaryName := gh.buildBinaryName(runnerInfo)

	_, err := dag.Container().
		From("alpine/curl").
		WithSecretVariable("GITHUB_TOKEN", token).
		WithMountedFile("/binary", container.File("/usr/local/bin/cm")).
		WithExec([]string{"sh", "-c", fmt.Sprintf(
			"curl -X POST -H \"Authorization: token $GITHUB_TOKEN\" "+
				"-H \"Content-Type: application/octet-stream\" "+
				"https://uploads.github.com/repos/%s/code-manager/releases/%s/assets?name=%s "+
				"--data-binary @/binary",
			actualUser, releaseID, binaryName,
		)}).
		Sync(ctx)

	if err != nil {
		return fmt.Errorf("failed to upload binary for %s: %w", platform, err)
	}
	return nil
}

// buildBinaryName builds the binary filename for a platform.
func (gh *GitHubReleaseManager) buildBinaryName(runnerInfo ImageInfo) string {
	binaryName := fmt.Sprintf("code-manager-%s-%s", runnerInfo.OS, runnerInfo.Arch)
	if runnerInfo.OS == "windows" {
		binaryName += ".exe"
	}
	return binaryName
}
