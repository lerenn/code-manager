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

func (ci *CodeManager) getActualUser(user *string) string {
	if user != nil {
		return *user
	}
	return defaultUser
}

// BuildForArchitecture builds a binary for a specific architecture without releasing it.
func (ci *CodeManager) BuildForArchitecture(
	sourceDir *dagger.Directory,
	architecture string,
) (*dagger.Container, error) {
	// Validate architecture
	if _, exists := GoImageInfo[architecture]; !exists {
		return nil, fmt.Errorf("unsupported architecture: %s", architecture)
	}

	// Build Docker image for this architecture
	runnerInfo := GoImageInfo[architecture]
	container := Image(sourceDir, runnerInfo)

	return container, nil
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

	// Check if binary for this architecture already exists in the release
	binaryExists, err := gh.checkBinaryExists(ctx, architecture, actualUser, releaseID, token)
	if err != nil {
		return err
	}

	if binaryExists {
		fmt.Printf("Binary for architecture %s already exists in release %s, "+
			"skipping build and upload\n", architecture, latestTag)
		return nil
	}

	// Build binary for this architecture
	container, err := ci.BuildForArchitecture(sourceDir, architecture)
	if err != nil {
		return err
	}

	// Only push the Docker image if TargetEnabled is true
	runnerInfo := GoImageInfo[architecture]
	if runnerInfo.TargetEnabled {
		// Push the Docker image to GitHub Container Registry with version tag
		imageName := fmt.Sprintf("ghcr.io/%s/code-manager:%s", actualUser, latestTag)

		// Push the image to GitHub Container Registry
		// Note: In GitHub Actions, Docker authentication is handled automatically
		// when using the GITHUB_TOKEN with appropriate permissions
		_, err = container.Publish(ctx, imageName)
		if err != nil {
			return fmt.Errorf("failed to push Docker image for %s: %w", architecture, err)
		}
	}

	// Upload binary for this architecture
	err = gh.uploadBinary(ctx, container, architecture, actualUser, releaseID, token)
	if err != nil {
		return err
	}

	return nil
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
