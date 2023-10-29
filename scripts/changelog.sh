#!/bin/bash

first_tag=$(git tag --sort=-version:refname | head -n 2 | tail -1)
second_tag=$(git tag --sort=-version:refname | head -n 1)

mkdir -p bin

echo "## What's Changed" > bin/changelog.md

echo >> bin/changelog.md
git log ${first_tag}..${second_tag} --pretty=format:"- %s: %h" >> bin/changelog.md
echo >> bin/changelog.md
echo >> bin/changelog.md

go_module=$(go list -m | sed 's/\/v.*//')

echo "Test report: https://xdeb-org.github.io/xdeb-install" >> bin/changelog.md
echo >> bin/changelog.md

echo "**Full Changelog**: [\`${first_tag}..${second_tag}\`](https://${go_module}/compare/${first_tag}..${second_tag})" >> bin/changelog.md
echo >> bin/changelog.md

echo "## SHA256 Checksums" >> bin/changelog.md
echo '```' >> bin/changelog.md

for binary in $(ls -d -1 bin/xdeb-*); do
    sha256sum ${binary} | sed 's/bin\///' >> bin/changelog.md
done

echo '```' >> bin/changelog.md
