GOTOOLS = \
					github.com/Masterminds/glide
PACKAGES=$(shell go list ./... | grep -v '/vendor/')

TMROOT = $${TMROOT:-$$HOME/.tendermint}

all: get_deps install test

build:
	rm -rf ./ethermint
	go build --ldflags '-extldflags "-static"' \
		--ldflags "-X main.gitCommit=`git rev-parse HEAD`" ./cmd/ethermint/

install: get_deps
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
	docker run -it --rm -v "$(PWD):/go/src/github.com/tendermint/ethermint" -w "/go/src/github.com/tendermint/ethermint" golang:latest go build ./cmd/ethermint
	docker build -t "tendermint/ethermint" -f docker/Dockerfile .

clean:
	rm -f ./ethermint

docker_build:
	rm -rf ./docker/ethermint
	docker run -it --rm -v "$(PWD):/go/src/github.com/tendermint/ethermint" -w "/go/src/github.com/tendermint/ethermint" \
        golang:1.6 go build --ldflags '-extldflags "-static"' \
        -o /go/src/github.com/tendermint/ethermint/docker/ethermint ./cmd/ethermint/
	docker build -t "tendermint/ethermint" ./docker
  
.PHONY: all install test test_race get_deps get_vendor_deps tools build-docker clean
