#!/bin/bash
set -e

# TODO:
#  + add validation...
#  + integrate gitian (https://gitian.org/)?

# version is supplied as argument
version="$1"
if [ -z "$version" ]; then
    echo "no version specified"
    echo "usage: $0 version"
    exit 1
fi

for os in darwin linux windows; do
	echo Packaging ${os}...
	# create workspace
	folder="release/rivine-$version-$os-amd64"
	rm -rf "$folder"
	mkdir -p "$folder"
	# compile and sign binaries
	for pkg in rivinec rivined; do
		bin=$pkg
		if [ "$os" == "windows" ]; then
			bin=${pkg}.exe
		fi
		GOOS=${os} go build -a -tags 'netgo' -ldflags="-s -w" -o "$folder/$bin" "./$pkg"

	done
	# add other artifacts
	cp -r doc LICENSE README.md "$folder"
	# zip
	(
		zip -rq "release/rivine-$version-$os-amd64.zip" "release/rivine-$version-$os-amd64"
	)
done