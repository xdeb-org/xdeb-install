#!/bin/bash

go get
go build -ldflags="-extldflags=-static" -o bin/xdeb-install-linux-$(uname -m)

echo "# What's Changed" > bin/checksums.md
echo >> bin/checksums.md

echo "## SHA256 Checksums" >> bin/checksums.md
echo '```' >> bin/checksums.md

for binary in $(ls -d -1 bin/xdeb-*); do
    sha256sum ${binary} | sed 's/bin\///' >> bin/checksums.md
done

echo '```' >> bin/checksums.md
