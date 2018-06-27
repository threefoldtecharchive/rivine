# all will build and install developer binaries, which have debugging enabled
# and much faster mining and block constants.
all: install

# pkgs changes which packages the makefile calls operate on. run changes which
# tests are run during testing.
run = Test
daemonpkgs = ./cmd/rivined
clientpkgs = ./cmd/rivinec
pkgs = ./build ./modules/gateway $(daemonpkgs) $(clientpkgs)
testpkgs = ./build ./crypto ./encoding ./modules ./modules/gateway ./modules/blockcreator ./modules/wallet ./modules/explorer ./modules/consensus ./persist ./cmd/rivinec ./cmd/rivined ./sync ./types ./pkg/cli ./pkg/client ./pkg/daemon

version = $(shell git describe | cut -d '-' -f 1)
commit = $(shell git rev-parse --short HEAD)
ifeq ($(commit), $(shell git rev-list -n 1 $(version) | cut -c1-8))
fullversion = $(version)
else
fullversion = $(version)-$(commit)
endif

dockerVersion = $(shell git describe | cut -d '-' -f 1| cut -d 'v' -f 2)

ldflagsversion = -X github.com/rivine/rivine/build.rawVersion=$(fullversion)

stdoutput = $(GOPATH)/bin
daemonbin = $(stdoutput)/rivined
clientbin = $(stdoutput)/rivinec

# fmt calls go fmt on all packages.
fmt:
	gofmt -s -l -w $(pkgs)

# vet calls go vet on all packages.
# NOTE: go vet requires packages to be built in order to obtain type info.
vet: release-std
	go vet $(pkgs)

# installs developer binaries.
install:
	go build -race -tags='dev debug profile' -ldflags '$(ldflagsversion)' -o $(daemonbin) $(daemonpkgs)
	go build -race -tags='dev debug profile' -ldflags '$(ldflagsversion)' -o $(clientbin) $(clientpkgs)

# installs std (release) binaries
install-std:
	go build -ldflags '$(ldflagsversion)' -o $(daemonbin) $(daemonpkgs)
	go build -ldflags '$(ldflagsversion)' -o $(clientbin) $(clientpkgs)

# release builds and installs release binaries.
release:
	go build -tags='debug profile' -ldflags '$(ldflagsversion)' -o $(daemonbin) $(daemonpkgs)
	go build -tags='debug profile' -ldflags '$(ldflagsversion)' -o $(clientbin) $(clientpkgs)
release-race:
	go build -race -tags='debug profile' -ldflags '$(ldflagsversion)' -o $(daemonbin) $(daemonpkgs)
	go build -race -tags='debug profile' -ldflags '$(ldflagsversion)' -o $(clientbin) $(clientpkgs)
release-std:
	go build -ldflags '$(ldflagsversion)' -o $(daemonbin) $(daemonpkgs)
	go build -ldflags '$(ldflagsversion)' -o $(clientbin) $(clientpkgs)

# xc builds and packages release binaries
# for all windows, linux and mac, 64-bit only,
# using the standard Golang toolchain.
xc:
	docker build -t rivinebuilder -f DockerBuilder .
	docker run --rm -v $(shell pwd):/go/src/github.com/rivine/rivine rivinebuilder

# Release images builds and packages release binaries, and uses the linux based binary to create a minimal docker
release-images: get_hub_jwt xc
	docker build -t rivine/rivine:$(dockerVersion) -f DockerfileMinimal --build-arg binaries_location=release/rivine-$(version)-linux-amd64/cmd .
	docker push rivine/rivine:$(dockerVersion)
	# also create a latest tag
	docker tag rivine/rivine:$(dockerVersion) rivine/rivine
	docker push rivine/rivine:latest
	curl -b "active-user=rivine; caddyoauth=$(HUB_JWT)" -X POST --data "image=rivine/rivine:$(dockerVersion)" "https://hub.gig.tech/api/flist/me/docker"

test:
	go test -short -tags='debug testing' -timeout=30s $(testpkgs) -run=$(run)
test-v:
	go test -race -v -short -tags='debug testing' -timeout=60s $(testpkgs) -run=$(run)
test-long: fmt vet
	go test -v -race -tags='debug testing' -timeout=500s $(testpkgs) -run=$(run)
bench: fmt
	go test -tags='testing' -timeout=500s -run=XXX -bench=. $(testpkgs)
cover:
	@mkdir -p cover/modules
	@for package in $(testpkgs); do \
		go test -tags='testing debug' -timeout=500s -covermode=atomic -coverprofile=cover/$$package.out ./$$package \
		&& go tool cover -html=cover/$$package.out -o=cover/$$package.html \
		&& rm cover/$$package.out ; \
	done
cover-integration:
	@mkdir -p cover/modules
	@for package in $(testpkgs); do \
		go test -run=TestIntegration -tags='testing debug' -timeout=500s -covermode=atomic -coverprofile=cover/$$package.out ./$$package \
		&& go tool cover -html=cover/$$package.out -o=cover/$$package.html \
		&& rm cover/$$package.out ; \
	done
cover-unit:
	@mkdir -p cover/modules
	@for package in $(testpkgs); do \
		go test -run=TestUnit -tags='testing debug' -timeout=500s -covermode=atomic -coverprofile=cover/$$package.out ./$$package \
		&& go tool cover -html=cover/$$package.out -o=cover/$$package.html \
		&& rm cover/$$package.out ; \
	done

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

get_hub_jwt: check-HUB_APP_ID check-HUB_APP_SECRET
	$(eval HUB_JWT = $(shell curl -X POST "https://itsyou.online/v1/oauth/access_token?response_type=id_token&grant_type=client_credentials&client_id=$(HUB_APP_ID)&client_secret=$(HUB_APP_SECRET)&scope=user:memberof:rivine"))

check-%:
	@ if [ "${${*}}" = "" ]; then \
		echo "Required env var $* not present"; \
		exit 1; \
	fi

find-deadlock:
	find . -type d -name "vendor" -prune -o -name "*.go" -print | xargs -n 1 sed -i 's/sync.RWMutex/deadlock.RWMutex/'
	find . -type d -name "vendor" -prune -o -name "*.go" -print | xargs -n 1 sed -i 's/sync.Mutex/deadlock.Mutex/'
	find . -type d -name "vendor" -prune -o -name "*.go" -print | xargs -I {} goimports -w {}

.PHONY: all fmt install release release-std test test-v test-long cover cover-integration cover-unit ineffassign ensure_deps add_dep update_dep update_deps
