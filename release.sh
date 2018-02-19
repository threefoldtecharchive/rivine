#!/bin/bash
set -e

# TODO:
#  + add validation...
#  + integrate gitian (https://gitian.org/)?

# version is supplied as argument
package="github.com/rivine/rivine"
full_version=$(git describe | cut -d '-' -f 1,3)
version=$(echo "$full_version" | cut -d '-' -f 1)

for os in darwin linux windows; do
	echo Packaging ${os}...
	# create workspace
	folder="release/rivine-${version}-${os}-amd64"
	rm -rf "$folder"
	mkdir -p "$folder"
	# compile and sign binaries
	for pkg in cmd/rivinec cmd/rivined; do
		bin=$pkg
		if [ "$os" == "windows" ]; then
			bin=${pkg}.exe
		fi
		GOOS=${os} go build -a -tags 'netgo' \
			-ldflags="-X ${package}/build.rawVersion=${full_version} -s -w" \
			-o "${folder}/${bin}" "./${pkg}"

	done
	# add other artifacts
	cp -r doc LICENSE README.md "$folder"
	# zip
	(
		zip -rq "release/rivine-${version}-${os}-amd64.zip" \
			"release/rivine-${version}-${os}-amd64"
	)
done