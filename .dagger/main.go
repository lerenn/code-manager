// A generated module for CM functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"fmt"
	"runtime"

	"code-manager/dagger/internal/dagger"
)

const defaultUser = "lerenn"

type CodeManager struct{}

// Publish a new release.
func (ci *CodeManager) PublishTag(
	ctx context.Context,
	sourceDir *dagger.Directory,
	user *string,
	token *dagger.Secret,
) error {
	// Set default user if not provided
	actualUser := defaultUser
	if user != nil {
		actualUser = *user
	}
	// Create Git repo access
	repo, err := NewGit(ctx, NewGitOptions{
		SrcDir: sourceDir,
		User:   &actualUser,
		Token:  token,
	})
	if err != nil {
		return err
	}

	// Publish new tag
	return repo.PublishTagFromReleaseTitle(ctx)
}

// Lint runs golangci-lint on the main repo (./...) only.
func (ci *CodeManager) Lint(sourceDir *dagger.Directory) *dagger.Container {
	c := dag.Container().
		From("golangci/golangci-lint:v2.4.0").
		WithMountedCache("/root/.cache/golangci-lint", dag.CacheVolume("golangci-lint"))

	c = ci.withGoCodeAndCacheAsWorkDirectory(c, sourceDir)

	// Lint main repo only
	c = c.WithExec([]string{"golangci-lint", "run", "--timeout", "10m", "./..."})

	return c
}

// LintDagger runs golangci-lint on the .dagger directory only.
func (ci *CodeManager) LintDagger(sourceDir *dagger.Directory) *dagger.Container {
	c := dag.Container().
		From("golangci/golangci-lint:v2.4.0").
		WithMountedCache("/root/.cache/golangci-lint", dag.CacheVolume("golangci-lint"))

	c = ci.withGoCodeAndCacheAsWorkDirectory(c, sourceDir)

	// Lint .dagger directory using parent config and module context
	c = c.WithExec([]string{"sh", "-c", "cd .dagger && golangci-lint run --config ../.golangci.yml --timeout 10m ."})

	return c
}

// UnitTests returns a container that runs the unit tests.
func (ci *CodeManager) UnitTests(sourceDir *dagger.Directory) *dagger.Container {
	c := dag.Container().From("golang:" + goVersion() + "-alpine")
	return ci.withGoCodeAndCacheAsWorkDirectory(c, sourceDir).
		WithExec([]string{"sh", "-c",
			"go test -tags=unit ./...",
		})
}

// IntegrationTests returns a container that runs the integration tests.
func (ci *CodeManager) IntegrationTests(sourceDir *dagger.Directory) *dagger.Container {
	c := dag.Container().From("golang:" + goVersion() + "-alpine").
		// Install git for integration tests
		WithExec([]string{"apk", "add", "--no-cache", "git"})

	return ci.withGoCodeAndCacheAsWorkDirectory(c, sourceDir).
		WithExec([]string{"sh", "-c",
			"go test -tags=integration ./...",
		})
}

// EndToEndTests returns a container that runs the end-to-end tests.
func (ci *CodeManager) EndToEndTests(sourceDir *dagger.Directory) *dagger.Container {
	c := dag.Container().From("golang:" + goVersion() + "-alpine").
		// Install git for end-to-end tests
		WithExec([]string{"apk", "add", "--no-cache", "git"}).
		// Configure git for testing
		WithExec([]string{"git", "config", "--global", "user.name", "Test User"}).
		WithExec([]string{"git", "config", "--global", "user.email", "test@example.com"})

	return ci.withGoCodeAndCacheAsWorkDirectory(c, sourceDir).
		WithExec([]string{"sh", "-c",
			"go test -tags=e2e ./test/ -v",
		})
}

// BuildAndPushDockerImages builds and pushes Docker images for all supported platforms to GitHub Packages.
func (ci *CodeManager) BuildAndPushDockerImages(
	ctx context.Context,
	sourceDir *dagger.Directory,
	user *string,
	token *dagger.Secret,
) error {
	actualUser := ci.getActualUser(user)

	latestTag, err := ci.getLatestTag(ctx, sourceDir, actualUser, token)
	if err != nil {
		return err
	}

	fullImageName := buildImageName(actualUser, latestTag)
	platforms := AvailablePlatforms()

	// Build all images in parallel
	images := buildAllImages(sourceDir, platforms)

	// Push all images in parallel
	return pushAllImages(ctx, images, platforms, fullImageName, actualUser, token)
}

func (ci *CodeManager) getActualUser(user *string) string {
	if user != nil {
		return *user
	}
	return defaultUser
}

func (ci *CodeManager) getLatestTag(
	ctx context.Context,
	sourceDir *dagger.Directory,
	actualUser string,
	token *dagger.Secret,
) (string, error) {
	repo, err := NewGit(ctx, NewGitOptions{
		SrcDir: sourceDir,
		User:   &actualUser,
		Token:  token,
	})
	if err != nil {
		return "", err
	}

	return repo.GetLastTag(ctx)
}

// BuildAndReleaseForArchitecture builds a Docker image for a specific architecture,
// creates a GitHub release if it doesn't exist, and uploads the binary for that architecture.
func (ci *CodeManager) BuildAndReleaseForArchitecture(
	ctx context.Context,
	sourceDir *dagger.Directory,
	architecture string,
	user *string,
	token *dagger.Secret,
) error {
	// Validate architecture
	if _, exists := GoImageInfo[architecture]; !exists {
		return fmt.Errorf("unsupported architecture: %s", architecture)
	}

	actualUser := ci.getActualUser(user)
	gh := NewGitHubReleaseManager()

	latestTag, releaseNotes, err := gh.getReleaseInfo(ctx, sourceDir, actualUser, token)
	if err != nil {
		return err
	}

	// Try to get existing release ID, create if it doesn't exist
	releaseID, err := gh.getOrCreateRelease(ctx, actualUser, latestTag, releaseNotes, token)
	if err != nil {
		return err
	}

	// Build Docker image for this architecture
	runnerInfo := GoImageInfo[architecture]
	container := Image(sourceDir, runnerInfo)

	// Upload binary for this architecture
	return gh.uploadBinary(ctx, container, architecture, actualUser, releaseID, token)
}



func (ci *CodeManager) withGoCodeAndCacheAsWorkDirectory(
	c *dagger.Container,
	sourceDir *dagger.Directory,
) *dagger.Container {
	containerPath := "/go/src/github.com/lerenn/code-manager"
	return c.
		// Add Go caches
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("gobuild")).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("gocache")).

		// Add source code
		WithMountedDirectory(containerPath, sourceDir).

		// Add workdir
		WithWorkdir(containerPath)
}

func goVersion() string {
	return runtime.Version()[2:]
}
