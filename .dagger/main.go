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
	"strings"
	"sync"

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

	fullImageName := ci.buildImageName(actualUser, latestTag)
	platforms := AvailablePlatforms()

	// Build all images in parallel
	images := ci.buildAllImages(sourceDir, platforms)

	// Push all images in parallel
	return ci.pushAllImages(ctx, images, platforms, fullImageName, actualUser, token)
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

func (ci *CodeManager) buildImageName(actualUser, latestTag string) string {
	registry := "ghcr.io"
	imageName := fmt.Sprintf("%s/code-manager", actualUser)
	return fmt.Sprintf("%s/%s:%s", registry, imageName, latestTag)
}

func (ci *CodeManager) buildAllImages(sourceDir *dagger.Directory, platforms []string) map[string]*dagger.Container {
	images := make(map[string]*dagger.Container)
	for _, platform := range platforms {
		runnerInfo := GoImageInfo[platform]
		images[platform] = Image(sourceDir, runnerInfo)
	}
	return images
}

func (ci *CodeManager) pushAllImages(
	ctx context.Context,
	images map[string]*dagger.Container,
	platforms []string,
	fullImageName, actualUser string,
	token *dagger.Secret,
) error {
	registry := "ghcr.io"
	errChan := make(chan error, len(platforms))
	var wg sync.WaitGroup

	for _, platform := range platforms {
		wg.Add(1)
		go func(platform string) {
			defer wg.Done()

			_, err := images[platform].
				WithRegistryAuth(registry, actualUser, token).
				Publish(ctx, fullImageName)

			if err != nil {
				errChan <- fmt.Errorf("failed to push image for %s: %w", platform, err)
			}
		}(platform)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		return err
	}

	return nil
}

// CreateGithubRelease creates a GitHub release with binaries for all supported platforms.
func (ci *CodeManager) CreateGithubRelease(
	ctx context.Context,
	sourceDir *dagger.Directory,
	user *string,
	token *dagger.Secret,
) error {
	actualUser := ci.getActualUser(user)
	
	latestTag, releaseNotes, err := ci.getReleaseInfo(ctx, sourceDir, actualUser, token)
	if err != nil {
		return err
	}

	if err := ci.createGitHubRelease(ctx, actualUser, latestTag, releaseNotes, token); err != nil {
		return err
	}

	return ci.uploadAllBinaries(ctx, sourceDir, actualUser, token)
}

func (ci *CodeManager) getReleaseInfo(
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

func (ci *CodeManager) createGitHubRelease(
	ctx context.Context,
	actualUser, latestTag, releaseNotes string,
	token *dagger.Secret,
) error {
	_, err := dag.Container().
		From("alpine/curl").
		WithSecretVariable("GITHUB_TOKEN", token).
		WithExec([]string{"sh", "-c", fmt.Sprintf(
			"curl -X POST -H \"Authorization: token $GITHUB_TOKEN\" "+
				"-H \"Accept: application/vnd.github.v3+json\" "+
				"https://api.github.com/repos/%s/code-manager/releases "+
				"-d '{\"tag_name\":\"%s\",\"name\":\"Release %s\",\"body\":\"%s\"}'",
			actualUser, latestTag, latestTag, strings.ReplaceAll(releaseNotes, "\"", "\\\""),
		)}).
		Sync(ctx)

	if err != nil {
		return fmt.Errorf("failed to create GitHub release: %w", err)
	}
	return nil
}

func (ci *CodeManager) uploadAllBinaries(
	ctx context.Context,
	sourceDir *dagger.Directory,
	actualUser string,
	token *dagger.Secret,
) error {
	platforms := AvailablePlatforms()
	containers := ci.buildAllImages(sourceDir, platforms)

	errChan := make(chan error, len(platforms))
	var wg sync.WaitGroup

	for _, platform := range platforms {
		wg.Add(1)
		go func(platform string) {
			defer wg.Done()

			if err := ci.uploadBinary(ctx, containers[platform], platform, actualUser, token); err != nil {
				errChan <- err
			}
		}(platform)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		return err
	}

	return nil
}

func (ci *CodeManager) uploadBinary(
	ctx context.Context,
	container *dagger.Container,
	platform, actualUser string,
	token *dagger.Secret,
) error {
	runnerInfo := GoImageInfo[platform]
	binaryName := ci.buildBinaryName(runnerInfo)

	_, err := dag.Container().
		From("alpine/curl").
		WithSecretVariable("GITHUB_TOKEN", token).
		WithMountedFile("/binary", container.File("/usr/local/bin/cm")).
		WithExec([]string{"sh", "-c", fmt.Sprintf(
			"curl -X POST -H \"Authorization: token $GITHUB_TOKEN\" "+
				"-H \"Content-Type: application/octet-stream\" "+
				"https://uploads.github.com/repos/%s/code-manager/releases/latest/assets?name=%s "+
				"--data-binary @/binary",
			actualUser, binaryName,
		)}).
		Sync(ctx)

	if err != nil {
		return fmt.Errorf("failed to upload binary for %s: %w", platform, err)
	}
	return nil
}

func (ci *CodeManager) buildBinaryName(runnerInfo ImageInfo) string {
	binaryName := fmt.Sprintf("code-manager-%s-%s", runnerInfo.OS, runnerInfo.Arch)
	if runnerInfo.OS == "windows" {
		binaryName += ".exe"
	}
	return binaryName
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
