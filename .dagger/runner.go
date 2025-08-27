package main

import (
	"code-manager/dagger/internal/dagger"
	"maps"
	"slices"
)

// RunnerInfo represents a Docker runner.
type RunnerInfo struct {
	OS              string
	Arch            string
	BuildBaseImage  string
	TargetBaseImage string
}

var (
	// GoRunnersInfo represents the different OS/Arch platform wanted for docker hub in Go service.
	GoRunnersInfo = map[string]RunnerInfo{
		"linux/386":      {OS: "linux", Arch: "386", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/amd64":    {OS: "linux", Arch: "amd64", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/arm/v6":   {OS: "linux", Arch: "arm/v6", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/arm/v7":   {OS: "linux", Arch: "arm/v7", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/arm64/v8": {OS: "linux", Arch: "arm64/v8", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/ppc64le":  {OS: "linux", Arch: "ppc64le", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/riscv64":  {OS: "linux", Arch: "riscv64", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/s390x":    {OS: "linux", Arch: "s390x", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
	}
)

func AvailablePlatforms() []string {
	return slices.Collect(maps.Keys(GoRunnersInfo))
}

// Runner returns a container running the code-manager
func Runner(
	sourceDir *dagger.Directory,
	runnerInfo RunnerInfo,
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
