# all will build and install developer binaries, which have debugging enabled
# and much faster mining and block constants.
all: install

# pkgs changes which packages the makefile calls operate on. run changes which
# tests are run during testing.
run = Test
rivinecgpkgs = ./cmd/rivinecg
pkgs = ./build ./modules/gateway $(rivinecgpkgs)
testpkgs = ./build ./crypto ./pkg/encoding/siabin ./pkg/encoding/rivbin ./modules ./modules/gateway ./modules/blockcreator ./modules/wallet ./modules/explorer ./modules/consensus ./persist ./sync ./types ./pkg/cli ./pkg/client ./pkg/daemon ./cmd/rivinecg/cmd ./cmd/rivinecg/pkg/config

version = $(shell git describe | cut -d '-' -f 1)
commit = $(shell git rev-parse --short HEAD)
ifeq ($(commit), $(shell git rev-list -n 1 $(version) | cut -c1-8))
fullversion = $(version)
else
fullversion = $(version)-$(commit)
endif

dockerVersion = $(shell git describe | cut -d '-' -f 1| cut -d 'v' -f 2)

ldflagsversion = -X github.com/threefoldtech/rivine/build.rawVersion=$(fullversion)

stdoutput = $(GOPATH)/bin
rivinecgbin = $(stdoutput)/rivinecg

# fmt calls go fmt on all packages.
fmt:
	gofmt -s -l -w $(pkgs)
	cd examples/rivchain && make fmt

# vet calls go vet on all packages.
# NOTE: go vet requires packages to be built in order to obtain type info.
vet: release-std
	go vet $(pkgs)
	cd examples/rivchain && make vet

# installs developer binaries.
install:
	go build race -tags='dev debug profile' -ldflags '$(ldflagsversion)' -o $(rivinecgbin) $(rivinecgpkgs)
	cd examples/rivchain && make install

# installs std (release) binaries
install-std:
	go build -ldflags '$(ldflagsversion)' -o $(rivinecgbin) $(rivinecgpkgs)
	cd examples/rivchain && make install-std

ineffassign:
	ineffassign $(testpkgs)

ensure_deps:
	dep ensure -v

add_dep:
	dep ensure -v
	dep ensure -v -add $$DEP

update_dep:
	dep ensure -v
	dep ensure -v -update $$DEP

update_deps:
	dep ensure -v
	dep ensure -update -v

find-deadlock:
	find . -type d -name "vendor" -prune -o -name "*.go" -print | xargs -n 1 sed -i 's/sync.RWMutex/deadlock.RWMutex/'
	find . -type d -name "vendor" -prune -o -name "*.go" -print | xargs -n 1 sed -i 's/sync.Mutex/deadlock.Mutex/'
	find . -type d -name "vendor" -prune -o -name "*.go" -print | xargs -I {} goimports -w {}

.PHONY: all fmt install release release-std test test-v test-long cover cover-integration cover-unit ineffassign ensure_deps add_dep update_dep update_deps
