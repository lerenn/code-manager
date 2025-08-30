package main

import (
	"code-manager/dagger/internal/dagger"
	"fmt"
	"maps"
	"runtime"
	"slices"
)

// ImageInfo represents a Docker image.
type ImageInfo struct {
	OS              string
	Arch            string
	BuildBaseImage  string
	TargetBaseImage string
}

var (
	// GoImageInfo represents the different OS/Arch platform wanted for binaries.
	GoImageInfo = map[string]ImageInfo{
		"linux/386": {
			OS:              "linux",
			Arch:            "386",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
		},
		"linux/amd64": {
			OS:              "linux",
			Arch:            "amd64",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
		},
		"linux/arm64": {
			OS:              "linux",
			Arch:            "arm64",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
		},
		"linux/arm/v6": {
			OS:              "linux",
			Arch:            "arm/v6",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
		},
		"linux/arm/v7": {
			OS:              "linux",
			Arch:            "arm/v7",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
		},
		"linux/ppc64le": {
			OS:              "linux",
			Arch:            "ppc64le",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
		},
		"linux/riscv64": {
			OS:              "linux",
			Arch:            "riscv64",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
		},
		"linux/s390x": {
			OS:              "linux",
			Arch:            "s390x",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
		},

		"darwin/amd64": {
			OS:              "darwin",
			Arch:            "amd64",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
		},
		"darwin/386": {
			OS:              "darwin",
			Arch:            "386",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
		},
		"darwin/arm64": {
			OS:              "darwin",
			Arch:            "arm64",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
		},

		"windows/386": {
			OS:              "windows",
			Arch:            "386",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
		},
		"windows/amd64": {
			OS:              "windows",
			Arch:            "amd64",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
		},
	}
)

func AvailablePlatforms() []string {
	return slices.Collect(maps.Keys(GoImageInfo))
}

// BuildImage returns a container for building binaries with cross-compilation.
// It returns an error if the current platform is not supported.
func BuildImage(
	sourceDir *dagger.Directory,
	runnerInfo ImageInfo,
) (*dagger.Container, error) {
	// Get the build base image for the current platform
	currentPlatform := runtime.GOOS + "/" + runtime.GOARCH
	currentRunnerInfo, exists := GoImageInfo[currentPlatform]
	if !exists {
		return nil, fmt.Errorf("unsupported build platform: %s", currentPlatform)
	}

	buildOpts := dagger.DirectoryDockerBuildOpts{
		BuildArgs: []dagger.BuildArg{
			{Name: "BUILDBASEIMAGE", Value: currentRunnerInfo.BuildBaseImage},
			{Name: "TARGETOS", Value: runnerInfo.OS},
			{Name: "TARGETARCH", Value: runnerInfo.Arch},
		},
		Dockerfile: "build/container/Dockerfile.build",
	}

	return sourceDir.DockerBuild(buildOpts), nil
}

// RuntimeImage returns a container for running (only for compatible platforms).
func RuntimeImage(
	buildContainer *dagger.Container,
	runnerInfo ImageInfo,
) *dagger.Container {
	// Only create runtime image if the target OS matches the current build environment
	// This ensures we only create Docker images for platforms we can actually run
	if runnerInfo.OS != runtime.GOOS {
		return nil
	}

	return dag.Container().
		From(runnerInfo.TargetBaseImage).
		WithFile("/usr/local/bin/cm", buildContainer.File(fmt.Sprintf("/go/bin/%s_%s/cm", runnerInfo.OS, runnerInfo.Arch))).
		WithEntrypoint([]string{"/usr/local/bin/cm"})
}
