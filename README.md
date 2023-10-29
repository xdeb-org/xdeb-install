# xdeb-install testcontainer image

This branch aims to provide a container image which can be used as containers for use in CI pipelines to verify the integrity of `xdeb-install`.

It is based on the `x86_64` variant of [`ghcr.io/void-linux/void-glibc-full`](https://ghcr.io/void-linux/void-glibc-full).

## Schedule

The image is rebuilt on each push to the `testcontainer` branch and each sunday at 11:00.
