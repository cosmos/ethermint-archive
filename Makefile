GOTOOLS = \
					github.com/Masterminds/glide
PACKAGES=$(shell go list ./... | grep -v '/vendor/')

TMROOT = $${TMROOT:-$$HOME/.tendermint}

all: get_deps install test

build:
	go build \
		--ldflags "-X github.com/tendermint/ethermint/version.GitCommit=`git rev-parse HEAD`"  -o build/ethermint ./cmd/ethermint/

install: get_vendor_deps get_deps
	@go install \
		--ldflags "-X github.com/tendermint/ethermint/version.GitCommit=`git rev-parse HEAD`" ./cmd/ethermint

test:
	@echo "--> Running go test"
	@go test $(PACKAGES)

test_race:
	@echo "--> Running go test --race"
	@go test -race $(PACKAGES)

get_deps:
	@echo "--> Running go get"
	@go get -v -d $(PACKAGES)
	@go list -f '{{join .TestImports "\n"}}' ./... | \
		grep -v /vendor/ | sort | uniq | \
		xargs go get -v -d

tools:
	go get -v $(GOTOOLS)

get_vendor_deps: tools
	@echo "--> Running glide install"
	@glide install --strip-vendor

clean:
	rm -rf build/ethermint
  
.PHONY: all install test test_race get_deps get_vendor_deps tools build-docker clean
