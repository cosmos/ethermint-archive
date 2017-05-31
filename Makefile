GOTOOLS = \
					github.com/mitchellh/gox \
					github.com/Masterminds/glide
PACKAGES=$(shell go list ./... | grep -v '/vendor/')
BUILD_TAGS?=ethermint

all: install test

install: get_vendor_deps
	@go install --ldflags '-extldflags "-static"' \
		--ldflags "-X github.com/tendermint/tendermint/version.GitCommit=`git rev-parse HEAD`" \
		./cmd/ethermint

build:
	@go build \
		--ldflags "-X github.com/tendermint/tendermint/version.GitCommit=`git rev-parse HEAD`" \
		-o ./build/ethermint ./cmd/ethermint

build_race:
	@go build -race -o build/ethermint ./cmd/ethermint

# dist builds binaries for all platforms and packages them for distribution
dist:
	@BUILD_TAGS='$(BUILD_TAGS)' sh -c "'$(CURDIR)/scripts/dist.sh'"

publish:
	@sh -c "'$(CURDIR)/scripts/publish.sh'"

test:
	@echo "--> Running go test"
	@go test $(PACKAGES)

test_race:
	@echo "--> Running go test --race"
	@go test -race $(PACKAGES)

draw_deps:
# requires brew install graphviz or apt-get install graphviz
	@go get github.com/RobotsAndPencils/goviz
	@goviz -i github.com/tendermint/ethermint/cmd/ethermint -d 2 | dot -Tpng -o dependency-graph.png

list_deps:
	@go list -f '{{join .Deps "\n"}}' ./... | \
		grep -v /vendor/ | sort | uniq | \
		xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}'

get_deps:
	@echo "--> Running go get"
	@go get -v -d $(PACKAGES)
	@go list -f '{{join .TestImports "\n"}}' ./... | \
		grep -v /vendor/ | sort | uniq | \
		xargs go get -v -d

get_vendor_deps: ensure_tools
	@rm -rf vendor/
	@echo "--> Running glide install"
	@glide install --strip-vendor

tools:
	go get -u -v $(GOTOOLS)

ensure_tools:
	go get $(GOTOOLS)

.PHONY: all install build build_race dist test test_race draw_deps list_deps get_deps get_vendor_deps tools ensure_tools
