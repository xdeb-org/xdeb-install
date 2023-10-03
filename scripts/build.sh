#!/bin/bash

export CGO_ENABLED=0
export GOOS=linux

export GIT_TAG=$(git describe --tags || git rev-parse HEAD)
export GIT_AUTHOR=$(git show -s --pretty=format:'%an <%ae>')

export GO_PACKAGE_VERSION="main.VersionString=${GIT_TAG}"
export GO_PACKAGE_DATETIME="main.VersionDate=$(date -u)"
export GO_PACKAGE_AUTHOR="main.VersionAuthor=${GIT_AUTHOR}"

export GO_BUILD_FLAGS="-extldflags=-static -w -s"
export GO_PACKAGE_FLAGS="-X '${GO_PACKAGE_VERSION}' -X '${GO_PACKAGE_DATETIME}' -X '${GO_PACKAGE_AUTHOR}'"

BUILD_ARCHITECTURES="amd64/x86_64 386/i686 arm64/aarch64"

go get

for build_arch in ${BUILD_ARCHITECTURES}; do
    export GOARCH=$(echo "${build_arch}" | cut -d '/' -f 1)
    void_arch=$(echo "${build_arch}" | cut -d '/' -f 2)

    go build -ldflags="${GO_BUILD_FLAGS} ${GO_PACKAGE_FLAGS}" -o bin/xdeb-install-linux-${void_arch}
done
