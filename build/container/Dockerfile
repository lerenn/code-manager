# Dockerfile arguments
ARG BUILDPLATFORM=linux/amd64
ARG TARGETPLATFORM=linux/amd64

ARG TARGETOS=linux
ARG TARGETARCH=amd64

ARG BUILDBASEIMAGE=golang:alpine
ARG TARGETBASEIMAGE=alpine:latest

# Building Golang image
FROM --platform=${BUILDPLATFORM:-linux/amd64} ${BUILDBASEIMAGE} AS build

# Get all remaining code
RUN mkdir -p /go/src/github.com/lerenn/code-manager
COPY ./ /go/src/github.com/lerenn/code-manager

# Set the workdir
WORKDIR /go/src/github.com/lerenn/code-manager

# Build everything in cmd/
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go install ./cmd/*

# Get final base image - use platform only if image should be pushed
FROM --platform=${TARGETPLATFORM} ${TARGETBASEIMAGE} AS final

# Get binary
COPY --from=build /go/bin/* /usr/local/bin

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/code-manager"]