all: install

daemonpkgs = ./cmd/rivined
clientpkgs = ./cmd/rivinec
pkgs = $(daemonpkgs) $(clientpkgs)

version = $(shell git describe --abbrev=0 || echo 'v0.1')
commit = $(shell git rev-parse --short HEAD)
ifeq ($(commit), $(shell (git rev-list -n 1 $(version) | cut -c1-7) || echo 'false' ))
	fullversion = $(version)
	fullversionpath = \/releases\/tag\/$(version)
else
	fullversion = $(version)-$(commit)
	fullversionpath = \/tree\/$(commit)
endif

configpkg = github.com/threefoldtech/rivine/examples/rivchain/pkg/config
ldflagsversion = -X $(configpkg).rawVersion=$(fullversion)

stdoutput = $(GOPATH)/bin
daemonbin = $(stdoutput)/rivined
clientbin = $(stdoutput)/rivinec

test: fmt vet

# fmt calls go fmt on all packages.
fmt:
	gofmt -s -l -w $(pkgs)

# vet calls go vet on all packages.
# NOTE: go vet requires packages to be built in order to obtain type info.
vet: install-std
	go vet $(pkgs)

# installs developer binaries.
install:
	go build -race -tags='dev debug profile' -ldflags '$(ldflagsversion)' -o $(daemonbin) $(daemonpkgs)
	go build -race -tags='dev debug profile' -ldflags '$(ldflagsversion)' -o $(clientbin) $(clientpkgs)

# installs std (release) binaries
install-std:
	go build -ldflags '$(ldflagsversion)' -o $(daemonbin) $(daemonpkgs)
	go build -ldflags '$(ldflagsversion)' -o $(clientbin) $(clientpkgs)

# installs std (release) binaries with profiling enabled on http on port 10501
install-profile-std:
	go build -tags='profile' -ldflags '$(ldflagsversion)' -o $(daemonbin) $(daemonpkgs)
	go build -ldflags '$(ldflagsversion)' -o $(clientbin) $(clientpkgs)

.PHONY: all test fmt vet install install-std install-profile-std
