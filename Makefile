GOTOOLS = \
					github.com/karalabe/xgo \
					github.com/Masterminds/glide
PACKAGES=$(shell go list ./... | grep -v '/vendor/')

TMROOT = $${TMROOT:-$$HOME/.tendermint}

all: install test

build:
	go build --ldflags "-extldflags '-static' \
		-X github.com/tendermint/ethermint/version.GitCommit=`git rev-parse HEAD`"  -o $(GOPATH)/bin/ethermint ./cmd/ethermint/

# dist builds binaries for all platforms and packages them for distribution
dist: tools get_vendor_deps clean_dist
	@$(CURDIR)/scripts/dist.sh

install: get_vendor_deps
	@go install ./cmd/ethermint

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

build-docker:
	rm -f ./ethermint
	docker run -it --rm -v "$(PWD):/go/src/github.com/tendermint/ethermint" -w "/go/src/github.com/tendermint/ethermint" golang:latest go build \
	    --ldflags "-extldflags '-static' -X github.com/tendermint/ethermint/version.GitCommit=`git rev-parse HEAD`" \
	    ./cmd/ethermint
	docker build -t "tendermint/ethermint" -f docker/Dockerfile .

clean:
	-rm -f ./build/ethermint

clean_dist:
	-rm -rf ./build/pkg
	-rm -rf ./build/dist
  
.PHONY: all build install test test_race get_deps get_vendor_deps tools build-docker clean clean_dist
