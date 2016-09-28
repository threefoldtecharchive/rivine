#!/usr/bin/env bash
set -e

VERSION="$(git describe)"

echo "Building version $VERSION"

docker build -t rivinebuilder .
docker run --rm -v "$PWD":/go/src/github.com/rivine/rivine --entrypoint sh rivinebuilder -c "go build -ldflags '-s' -v -o dist/rivined github.com/rivine/rivine/rivined && go build -ldflags '-s' -v -o dist/rivinec github.com/rivine/rivine/rivinec"
docker build -t rivine/rivine:"$VERSION" -f DockerfileMinimal .

docker push rivine/rivine:"$VERSION"
