# all will build and install developer binaries, which have debugging enabled
# and much faster mining and block constants.
all: install

# pkgs changes which packages the makefile calls operate on. run changes which
# tests are run during testing.
run = Test
pkgs = ./build ./modules/gateway ./cmd/rivined ./cmd/rivinec
testpkgs = ./build ./crypto ./encoding ./modules ./modules/gateway ./modules/blockcreator ./modules/wallet ./persist ./cmd/rivinec ./cmd/rivined ./sync ./types

version = $(shell git describe | cut -d '-' -f 1)
versionTag = $(shell git describe | cut -d '-' -f 1| cut -d 'v' -f 2)

# fmt calls go fmt on all packages.
fmt:
	gofmt -s -l -w $(pkgs)

# vet calls go vet on all packages.
# NOTE: go vet requires packages to be built in order to obtain type info.
vet: release-std
	go vet $(pkgs)

# install builds and installs developer binaries.
install:
	go install -race -tags='dev debug profile' $(pkgs)

# release builds and installs release binaries.
release:
	go install -tags='debug profile' $(pkgs)
release-race:
	go install -race -tags='debug profile' $(pkgs)
release-std:
	go install $(pkgs)

# xc builds and packages release binaries
# for all windows, linux and mac, 64-bit only,
# using the standard Golang toolchain.
xc:
	docker build -t rivinebuilder -f DockerBuilder .
	docker run --rm -v $(shell pwd):/go/src/github.com/rivine/rivine rivinebuilder

# Release images builds and packages release binaries, and uses the linux based binary to create a minimal docker
release-images: get_hub_jwt xc
	docker build -t rivine/rivine:$(versionTag) -f DockerfileMinimal --build-arg binaries_location=release/rivine-$(version)-linux-amd64/cmd .
	docker push rivine/rivine:$(versionTag)
	# also create a latest tag
	docker tag rivine/rivine:$(versionTag) rivine/rivine
	docker push rivine/rivine:latest
	curl -b "active-user=rivine; caddyoauth=$(HUB_JWT)" -X POST --data "image=rivine/rivine:$(versionTag)" "https://hub.gig.tech/api/flist/me/docker"

test:
	go test -short -tags='debug testing' -timeout=10s $(testpkgs) -run=$(run)
test-v:
	go test -race -v -short -tags='debug testing' -timeout=20s $(testpkgs) -run=$(run)
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

.PHONY: all fmt install release release-std test test-v test-long cover cover-integration cover-unit ineffassign ensure_deps add_dep update_dep update_deps
