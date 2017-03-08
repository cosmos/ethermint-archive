GOTOOLS = \
					github.com/mitchellh/gox \
					github.com/Masterminds/glide
PACKAGES=$(shell go list ./... | grep -v '/vendor/')
TMROOT = $${TMROOT:-$$HOME/.tendermint}

all: get_deps install test

install: get_deps
	@go install ./cmd/ethermint

test:
	@echo "--> Running go test"
	@go test $(PACKAGES)

test_race:
	@echo "--> Running go test --race"
	@go test -v -race $(PACKAGES)

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

.PHONY: all install test test_race get_deps get_vendor_deps tools
