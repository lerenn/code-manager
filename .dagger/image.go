package main

import (
	"code-manager/dagger/internal/dagger"
	"maps"
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
		"linux/386":      {OS: "linux", Arch: "386", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/amd64":    {OS: "linux", Arch: "amd64", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/arm/v6":   {OS: "linux", Arch: "arm/v6", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/arm/v7":   {OS: "linux", Arch: "arm/v7", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/arm64/v8": {OS: "linux", Arch: "arm64/v8", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/ppc64le":  {OS: "linux", Arch: "ppc64le", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/riscv64":  {OS: "linux", Arch: "riscv64", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/s390x":    {OS: "linux", Arch: "s390x", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"darwin/amd64":   {OS: "darwin", Arch: "amd64", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"darwin/386":     {OS: "darwin", Arch: "386", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"darwin/arm64":   {OS: "darwin", Arch: "arm64", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"windows/386":    {OS: "windows", Arch: "386", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"windows/amd64":  {OS: "windows", Arch: "amd64", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
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
	return sourceDir.DockerBuild(dagger.DirectoryDockerBuildOpts{
		BuildArgs: []dagger.BuildArg{
			{Name: "BUILDPLATFORM", Value: runnerInfo.OS + "/" + runnerInfo.Arch},
			{Name: "TARGETOS", Value: runnerInfo.OS},
			{Name: "TARGETARCH", Value: runnerInfo.Arch},
			{Name: "BUILDBASEIMAGE", Value: runnerInfo.BuildBaseImage},
			{Name: "TARGETBASEIMAGE", Value: runnerInfo.TargetBaseImage},
		},
		Platform:   dagger.Platform(runnerInfo.OS + "/" + runnerInfo.Arch),
		Dockerfile: "build/container/Dockerfile",
	})
}
