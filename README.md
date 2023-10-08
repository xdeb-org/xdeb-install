# xdeb-install devcontainer image

This branch aims to provide a container image which can be used both as [VS Code devcontainers](https://code.visualstudio.com/docs/devcontainers/containers) for local development/testing, as well as containers for use in CI pipelines to verify the integrity of `xdeb-install`.

It is based on the `x86_64` variant of https://github.com/thetredev/voidlinux-repository/tree/docker: [`ghcr.io/thetredev/voidlinux-repository:x86_64`](ghcr.io/thetredev/voidlinux-repository:x86_64).

## Schedule

The image is rebuilt on each push to the `devcontainer` branch and each sunday at 11:00.
