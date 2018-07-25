BRANCH=`git rev-parse --abbrev-ref HEAD`
COMMIT=`git rev-parse --short HEAD`
GOLDFLAGS="-X main.branch $(BRANCH) -X main.commit $(COMMIT)"

default: install

install:
	@go install -v github.com/rivine/bbolt/cmd/bolt

race:
	@go test -v -race -test.run="TestSimulate_(100op|1000op)"

fmt:
	gofmt -s -l -w .

# go get honnef.co/go/tools/cmd/gosimple
gosimple:
	gosimple ./...

# go get honnef.co/go/tools/cmd/unused
unused:
	unused ./...

# go get github.com/kisielk/errcheck
errcheck:
	@errcheck -ignorepkg=bytes -ignore=os:Remove github.com/rivine/bbolt

test:
	go test -timeout 20m -v
	# Note: gets "program not an importable package" in out of path builds
	go test -v ./cmd/bolt

.PHONY: install race fmt errcheck test gosimple unused
