GOTOOLS = \
					github.com/karalabe/xgo \
					github.com/Masterminds/glide \
					honnef.co/go/tools/cmd/megacheck \
					github.com/alecthomas/gometalinter
PACKAGES=$(shell go list ./... | grep -v '/vendor/')
BUILD_TAGS?=ethermint
VERSION_TAG=0.4.0

all: install test

install: get_vendor_deps
	@go install \
		--ldflags "-X github.com/tendermint/ethermint/version.GitCommit=`git rev-parse HEAD`" \
		./cmd/ethermint

build:
	@go build \
		--ldflags "-X github.com/tendermint/ethermint/version.GitCommit=`git rev-parse HEAD`" \
		-o ./build/ethermint ./cmd/ethermint

build_static:
	@go build \
		--ldflags "-extldflags '-static' -X github.com/tendermint/ethermint/version.GitCommit=`git rev-parse HEAD`" \
		-o ./build/ethermint ./cmd/ethermint

build_race:
	@go build -race -o build/ethermint ./cmd/ethermint

# dist builds binaries for all platforms and packages them for distribution
dist:
	@BUILD_TAGS='$(BUILD_TAGS)' sh -c "'$(CURDIR)/scripts/dist.sh'"

docker_build_develop:
	docker build -t "tendermint/ethermint:develop" -t "adrianbrink/ethermint:develop" -f scripts/docker/Dockerfile.develop .

docker_push_develop:
	docker push "tendermint/ethermint:develop"
	docker push "adrianbrink/ethermint:develop"

docker_build:
	docker build -t "tendermint/ethermint" -t "tendermint/ethermint:$(VERSION_TAG)" -t "adrianbrink/ethermint" -t "adrianbrink/ethermint:$(VERSION_TAG)" -f scripts/docker/Dockerfile .

docker_push:
	docker push "tendermint/ethermint:latest"
	docker push "tendermint/ethermint:$(VERSION_TAG)"
	docker push "adrianbrink/ethermint:latest"
	docker push "adrianbrink/ethermint:$(VERSION_TAG)"

clean:
	@rm -rf build/
	@rm -rf ethstats/

publish:
	@sh -c "'$(CURDIR)/scripts/publish.sh'"

test:
	@echo "--> Running go test"
	@go test $(PACKAGES)

test_coverage:
	@echo "--> Running go test with coverage"
	@bash ./tests/scripts/test_coverage.sh

test_race:
	@echo "--> Running go test --race"
	@go test -race $(PACKAGES)

test_integrations:
	@bash ./tests/test.sh

megacheck: ensure_tools
	@for pkg in ${PACKAGES}; do megacheck "$$pkg"; done

metalinter: ensure_tools
	@gometalinter --install
	gometalinter --vendor --deadline=600s --enable-all --disable=lll ./...

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

ethstats:
	@git clone https://github.com/tendermint/eth-net-intelligence-api $(CURDIR)/ethstats

ethstats_setup: ethstats
	@cd $(CURDIR)/ethstats && npm install && node scripts/configure.js

ethstats_start:
	@cd $(CURDIR)/ethstats && pm2 start ./app.json

ethstats_stop:
	@cd $(CURDIR)/ethstats && pm2 stop ./app.json

.PHONY: all install build build_race dist \
	test test_racetest_integrations \
	draw_deps list_deps get_deps get_vendor_deps tools ensure_tools \
	docker_build docker_build_develop docker_push docker_push_develop \
	ethstats_setup ethstats_run ethstats_stop
