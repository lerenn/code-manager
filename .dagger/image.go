package main

import (
	"code-manager/dagger/internal/dagger"
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
	ExportImage     bool
}

var (
	// GoImageInfo represents the different OS/Arch platform wanted for binaries.
	GoImageInfo = map[string]ImageInfo{
		"linux/386": {
			OS:              "linux",
			Arch:            "386",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
			ExportImage:     true,
		},
		"linux/amd64": {
			OS:              "linux",
			Arch:            "amd64",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
			ExportImage:     true,
		},
		"linux/arm/v6": {
			OS:              "linux",
			Arch:            "arm/v6",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
			ExportImage:     true,
		},
		"linux/arm/v7": {
			OS:              "linux",
			Arch:            "arm/v7",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
			ExportImage:     true,
		},
		"linux/arm64/v8": {
			OS:              "linux",
			Arch:            "arm64/v8",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
			ExportImage:     true,
		},
		"linux/ppc64le": {
			OS:              "linux",
			Arch:            "ppc64le",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
			ExportImage:     true,
		},
		"linux/riscv64": {
			OS:              "linux",
			Arch:            "riscv64",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
			ExportImage:     true,
		},
		"linux/s390x": {
			OS:              "linux",
			Arch:            "s390x",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
			ExportImage:     true,
		},
		"darwin/amd64": {
			OS:              "darwin",
			Arch:            "amd64",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
			ExportImage:     false,
		},
		"darwin/386": {
			OS:              "darwin",
			Arch:            "386",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
			ExportImage:     false,
		},
		"darwin/arm64": {
			OS:              "darwin",
			Arch:            "arm64",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
			ExportImage:     false,
		},
		"windows/386": {
			OS:              "windows",
			Arch:            "386",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
			ExportImage:     false,
		},
		"windows/amd64": {
			OS:              "windows",
			Arch:            "amd64",
			BuildBaseImage:  "golang:alpine",
			TargetBaseImage: "alpine",
			ExportImage:     false,
		},
	}
)

func AvailablePlatforms() []string {
	return slices.Collect(maps.Keys(GoImageInfo))
}

// Image returns a container running the code-manager.
func Image(
	sourceDir *dagger.Directory,
	runnerInfo ImageInfo,
) *dagger.Container {
	// Determine target platform
	targetPlatform := "linux/amd64" // default local platform
	if runnerInfo.ExportImage {
		targetPlatform = runnerInfo.OS + "/" + runnerInfo.Arch
	}

	buildOpts := dagger.DirectoryDockerBuildOpts{
		BuildArgs: []dagger.BuildArg{
			{Name: "BUILDPLATFORM", Value: runtime.GOOS + "/" + runtime.GOARCH},
			{Name: "TARGETOS", Value: runnerInfo.OS},
			{Name: "TARGETARCH", Value: runnerInfo.Arch},
			{Name: "BUILDBASEIMAGE", Value: runnerInfo.BuildBaseImage},
			{Name: "TARGETBASEIMAGE", Value: runnerInfo.TargetBaseImage},
			{Name: "TARGETPLATFORM", Value: targetPlatform},
		},
		Dockerfile: "build/container/Dockerfile",
	}

	return sourceDir.DockerBuild(buildOpts)
}
