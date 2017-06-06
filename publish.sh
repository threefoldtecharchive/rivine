#!/usr/bin/env bash
set -e

VERSION="$(git describe)"

echo "Building version $VERSION"

docker build -t rivinebuilder .
docker run --rm -v "$PWD":/go/src/github.com/rivine/rivine --entrypoint sh rivinebuilder -c "GOOS=linux GOARCH=amd64 go build -ldflags '-s' -v -o dist/linux/rivined github.com/rivine/rivine/rivined && GOOS=linux GOARCH=amd64 go build -ldflags '-s' -v -o dist/linux/rivinec github.com/rivine/rivine/rivinec"
docker run --rm -v "$PWD":/go/src/github.com/rivine/rivine --entrypoint sh rivinebuilder -c "GOOS=windows GOARCH=amd64 go build -ldflags '-s' -v -o dist/windows/rivined.exe github.com/rivine/rivine/rivined && GOOS=windows GOARCH=amd64 go build -ldflags '-s' -v -o dist/windows/rivinec.exe github.com/rivine/rivine/rivinec"
docker run --rm -v "$PWD":/go/src/github.com/rivine/rivine --entrypoint sh rivinebuilder -c "GOOS=darwin GOARCH=amd64 go build -ldflags '-s' -v -o dist/darwin/rivined github.com/rivine/rivine/rivined && GOOS=darwin GOARCH=amd64 go build -ldflags '-s' -v -o dist/darwin/rivinec github.com/rivine/rivine/rivinec"
docker build -t rivine/rivine:"$VERSION" -f DockerfileMinimal .

docker push rivine/rivine:"$VERSION"

pushd dist

zip -D rivine_"$VERSION"_windows.zip windows/*
tar -c -z -f rivine_"$VERSION"_darwin.tar.gz darwin/*
tar -c -z -f rivine_"$VERSION"_linux.tar.gz linux/*

popd
