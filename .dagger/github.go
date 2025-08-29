package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"code-manager/dagger/internal/dagger"
)

// GitHubReleaseManager handles GitHub release operations.
type GitHubReleaseManager struct{}

// NewGitHubReleaseManager creates a new GitHub release manager.
func NewGitHubReleaseManager() *GitHubReleaseManager {
	return &GitHubReleaseManager{}
}

// safeCloseResponse safely closes the response body and logs any errors.
func (gh *GitHubReleaseManager) safeCloseResponse(resp *http.Response) {
	if closeErr := resp.Body.Close(); closeErr != nil {
		// Log the error but don't fail the function
		fmt.Printf("warning: failed to close response body: %v\n", closeErr)
	}
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
	// Create the release payload
	releasePayload := map[string]string{
		"tag_name": latestTag,
		"name":     fmt.Sprintf("Release %s", latestTag),
		"body":     releaseNotes,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(releasePayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal release payload: %w", err)
	}

	// Get the token value
	tokenValue, err := token.Plaintext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://api.github.com/repos/%s/code-manager/releases", actualUser)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "token "+tokenValue)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create release: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response to get release ID
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	releaseID, ok := response["id"]
	if !ok {
		return "", fmt.Errorf("response does not contain release ID: %s", string(body))
	}

	// Convert to string
	releaseIDStr := fmt.Sprintf("%.0f", releaseID)
	return releaseIDStr, nil
}

// getExistingRelease gets an existing release by tag name.
func (gh *GitHubReleaseManager) getExistingRelease(
	ctx context.Context,
	actualUser, latestTag string,
	token *dagger.Secret,
) (string, error) {
	// Get the token value
	tokenValue, err := token.Plaintext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://api.github.com/repos/%s/code-manager/releases/tags/%s", actualUser, latestTag)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "token "+tokenValue)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for 404 (release not found)
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("release not found")
	}

	// Check for other errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get release: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response to get release ID
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	releaseID, ok := response["id"]
	if !ok {
		return "", fmt.Errorf("response does not contain release ID: %s", string(body))
	}

	// Convert to string
	releaseIDStr := fmt.Sprintf("%.0f", releaseID)
	return releaseIDStr, nil
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

	// Get the token value
	tokenValue, err := token.Plaintext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	// Get the binary file content
	binaryContent, err := container.File("/usr/local/bin/cm").Contents(ctx)
	if err != nil {
		return fmt.Errorf("failed to read binary file: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://uploads.github.com/repos/%s/code-manager/releases/%s/assets?name=%s", 
		actualUser, releaseID, binaryName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(binaryContent))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "token "+tokenValue)
	req.Header.Set("Content-Type", "application/octet-stream")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to upload binary for %s: status %d, body: %s", 
			platform, resp.StatusCode, string(body))
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
