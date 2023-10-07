#!/bin/bash

export CGO_ENABLED=0
export GOOS=linux

export GIT_TAG=$(git describe --tags || git rev-parse HEAD)

GIT_AUTHORS=""
ga_all=$(git log --pretty="%an=%ae" | sort -u | paste -sd "," - | tr ' ' ':' | tr ',' ' ')

# hide author email if author didn't sign-off any commits
for ga in ${ga_all}
do
    ga_name=$(echo "${ga}" | tr ':' ' ' | cut -d '=' -f 1)
    ga_email=$(echo "${ga}" | cut -d '=' -f 2)

    ga_commit_count=$(git --no-pager log --grep="Signed-off-by: ${ga_name} <${ga_email}>" | sort -u | wc -l)

    if [ ! -z "${GIT_AUTHORS}" ]; then
        GIT_AUTHORS+=","
    fi

    if [ ${ga_commit_count} -gt 0 ]; then
        GIT_AUTHORS+="${ga_name} <${ga_email}>"
    else
        GIT_AUTHORS+="${ga_name}"
    fi
done

export GO_PACKAGE_VERSION="main.VersionString=${GIT_TAG}"
export GO_PACKAGE_COMPILED="main.VersionCompiled=$(date +%s)"
export GO_PACKAGE_AUTHORS="main.VersionAuthors=${GIT_AUTHORS}"

export GO_BUILD_FLAGS="-extldflags=-static -w -s"
export GO_PACKAGE_FLAGS="-X '${GO_PACKAGE_VERSION}' -X '${GO_PACKAGE_COMPILED}' -X '${GO_PACKAGE_AUTHORS}'"

BUILD_ARCHITECTURES="amd64/x86_64 386/i686 arm64/aarch64"

go get

for build_arch in ${BUILD_ARCHITECTURES}; do
    export GOARCH=$(echo "${build_arch}" | cut -d '/' -f 1)
    void_arch=$(echo "${build_arch}" | cut -d '/' -f 2)

    go build -ldflags="${GO_BUILD_FLAGS} ${GO_PACKAGE_FLAGS}" -o bin/xdeb-install-linux-${void_arch}
done
