TMROOT = $${TMROOT:-$$HOME/.tendermint}
PACKAGES=$(shell go list ./... | grep -v '/vendor/')

install: get_vendor_deps
	@go install ./cmd/ethermint

test:
	@echo "--> Running go test"
	@go test `${PACKAGES}`

test_race:
	@echo "--> Running go test --race"
	@go test -race `${PACKAGES}`

get_deps:
	@go get -d `${PACKAGES}`
	@go list -f '{{join .TestImports "\n"}}' ./... | \
		grep -v /vendor/ | sort | uniq | \
		xargs go get

get_vendor_deps:
	go get github.com/Masterminds/glide
	@echo "--> Running glide install"
	@glide install --strip-vendor

all: install test

.PHONY: get_deps get_vendor_deps install test test_race
