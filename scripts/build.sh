#!/bin/bash

export CGO_ENABLED=0
export GOOS=linux

BUILD_ARCHITECTURES="amd64/x86_64 386/i686 arm64/aarch64"

go get

for build_arch in ${BUILD_ARCHITECTURES}; do
    export GOARCH=$(echo "${build_arch}" | cut -d '/' -f 1)
    void_arch=$(echo "${build_arch}" | cut -d '/' -f 2)

    go build -ldflags="-extldflags=-static -w -s" -o bin/xdeb-install-linux-${void_arch}
done
