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

type CodeManager struct{}

// Publish a new release.
func (ci *CodeManager) PublishTag(
	ctx context.Context,
	sourceDir *dagger.Directory,
	user *string,
	token *dagger.Secret,
) error {
	// Create Git repo access
	repo, err := NewGit(ctx, NewGitOptions{
		SrcDir: sourceDir,
		User:   user,
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
	// Get the latest tag
	repo, err := NewGit(ctx, NewGitOptions{
		SrcDir: sourceDir,
		User:   user,
		Token:  token,
	})
	if err != nil {
		return err
	}

	latestTag, err := repo.GetLastTag(ctx)
	if err != nil {
		return err
	}

	// GitHub Packages registry URL
	registry := "ghcr.io"
	imageName := fmt.Sprintf("%s/code-manager", *user)
	fullImageName := fmt.Sprintf("%s/%s:%s", registry, imageName, latestTag)

	// Get all platforms
	platforms := AvailablePlatforms()

	// Build and push for each platform
	for _, platform := range platforms {
		runnerInfo := GoRunnersInfo[platform]

		// Build the image for this platform using the existing Runner function
		image := Runner(sourceDir, runnerInfo)

		// Push the image to GitHub Packages using Dagger's registry operations
		_, err = image.
			WithRegistryAuth(registry, *user, token).
			Publish(ctx, fullImageName)

		if err != nil {
			return fmt.Errorf("failed to push image for %s: %w", platform, err)
		}
	}

	return nil
}

// CreateGitHubRelease creates a GitHub release with binaries for all supported platforms.
func (ci *CodeManager) CreateGitHubRelease(
	ctx context.Context,
	sourceDir *dagger.Directory,
	user *string,
	token *dagger.Secret,
) error {
	// Get the latest tag
	repo, err := NewGit(ctx, NewGitOptions{
		SrcDir: sourceDir,
		User:   user,
		Token:  token,
	})
	if err != nil {
		return err
	}

	latestTag, err := repo.GetLastTag(ctx)
	if err != nil {
		return err
	}

	// Get release notes from the last commit
	releaseNotes, err := repo.GetLastCommitTitle(ctx)
	if err != nil {
		return err
	}

	// Create the release first
	_, err = dag.Container().
		From("alpine/curl").
		WithSecretVariable("GITHUB_TOKEN", token).
		WithExec([]string{"sh", "-c", fmt.Sprintf(
			"curl -X POST -H \"Authorization: token $GITHUB_TOKEN\" "+
				"-H \"Accept: application/vnd.github.v3+json\" "+
				"https://api.github.com/repos/%s/code-manager/releases "+
				"-d '{\"tag_name\":\"%s\",\"name\":\"Release %s\",\"body\":\"%s\"}'",
			*user, latestTag, latestTag, releaseNotes,
		)}).
		Sync(ctx)

	if err != nil {
		return fmt.Errorf("failed to create GitHub release: %w", err)
	}

	// Build binaries for each platform
	platforms := AvailablePlatforms()

	for _, platform := range platforms {
		runnerInfo := GoRunnersInfo[platform]

		// Build binary for this platform
		binaryName := fmt.Sprintf("code-manager-%s-%s", runnerInfo.OS, runnerInfo.Arch)
		if runnerInfo.OS == "windows" {
			binaryName += ".exe"
		}

		// Build the binary
		container := dag.Container().
			From(runnerInfo.BuildBaseImage).
			WithMountedDirectory("/src", sourceDir).
			WithWorkdir("/src").
			WithMountedCache("/root/.cache/go-build", dag.CacheVolume("gobuild")).
			WithMountedCache("/go/pkg/mod", dag.CacheVolume("gocache")).
			WithExec([]string{"sh", "-c", fmt.Sprintf(
				"CGO_ENABLED=0 GOOS=%s GOARCH=%s go build -o %s ./cmd/cm",
				runnerInfo.OS, runnerInfo.Arch, binaryName,
			)})

		// Upload the binary asset to the release
		_, err = container.
			WithSecretVariable("GITHUB_TOKEN", token).
			WithExec([]string{"sh", "-c", fmt.Sprintf(
				"curl -X POST -H \"Authorization: token $GITHUB_TOKEN\" "+
					"-H \"Content-Type: application/octet-stream\" "+
					"https://uploads.github.com/repos/%s/code-manager/releases/latest/assets?name=%s "+
					"--data-binary @%s",
				*user, binaryName, binaryName,
			)}).
			Sync(ctx)

		if err != nil {
			return fmt.Errorf("failed to upload binary for %s: %w", platform, err)
		}
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
