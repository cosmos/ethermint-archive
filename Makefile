GOTOOLS := \
					 github.com/karalabe/xgo \
					 github.com/alecthomas/gometalinter

PACKAGES := $(shell glide novendor)

BUILD_TAGS? := ethermint

VERSION_TAG := 0.5.3

BUILD_FLAGS = -ldflags "-X github.com/tendermint/ethermint/version.GitCommit=`git rev-parse --short HEAD`"


### Development ###
all: get_vendor_deps install test

install:
	CGO_ENABLED=1 go install $(BUILD_FLAGS) ./cmd/ethermint

build:
	CGO_ENABLED=1 go build $(BUILD_FLAGS) -o ./build/ethermint ./cmd/ethermint

test:
	@echo "--> Running go test"
	@go test $(PACKAGES)

test_race:
	@echo "--> Running go test --race"
	@go test -v -race $(PACKAGES)

test_integrations:
	@echo "--> Running integration tests"
	@bash ./tests/test.sh

test_coverage:
	@echo "--> Running go test with coverage"
	bash ./tests/scripts/test_coverage.sh

linter:
	@echo "--> Running metalinter"
	gometalinter --install
	gometalinter --vendor --tests --deadline=120s --disable-all \
		--enable=unused \
		--enable=lll --line-length=100 \
		./...

clean:
	@echo "--> Cleaning the build and dependency files"
	rm -rf build/
	rm -rf vendor/
	rm -rf ethstats/


### Tooling ###
# requires brew install graphviz or apt-get install graphviz
draw_deps:
	@echo "--> Drawing dependencies"
	go get github.com/RobotsAndPencils/goviz
	goviz -i github.com/tendermint/ethermint/cmd/ethermint -d 2 | dot -Tpng -o dependency-graph.png

get_vendor_deps:
	@hash glide 2>/dev/null || go get github.com/Masterminds/glide
	@rm -rf vendor/
	@echo "--> Running glide install"
	@glide install
	@# ethereum/node.go:53:23: cannot use ctx (type *"github.com/tendermint/ethermint/vendor/gopkg.in/urfave/cli.v1".Context) as type *"github.com/tendermint/ethermint/vendor/github.com/ethereum/go-ethereum/vendor/gopkg.in/urfave/cli.v1".Context in argument to utils.SetEthConfig
	@rm -rf vendor/github.com/ethereum/go-ethereum/vendor

tools:
	@echo "--> Installing tools"
	go get $(GOTOOLS)

update_tools:
	@echo "--> Updating tools"
	@go get -u $(GOTOOLS)

### Building and Publishing ###
# dist builds binaries for all platforms and packages them for distribution
dist:
	@echo "--> Building binaries"
	@BUILD_TAGS='$(BUILD_TAGS)' sh -c "'$(CURDIR)/scripts/dist.sh'"

publish:
	@echo "--> Publishing binaries"
	sh -c "'$(CURDIR)/scripts/publish.sh'"


### Docker ###
docker_build_develop:
	docker build -t "tendermint/ethermint:develop" -t "adrianbrink/ethermint:develop" \
		-f scripts/docker/Dockerfile.develop .

docker_push_develop:
	docker push "tendermint/ethermint:develop"
	docker push "adrianbrink/ethermint:develop"

docker_build:
	docker build -t "tendermint/ethermint" -t "tendermint/ethermint:$(VERSION_TAG)" \
		-t "adrianbrink/ethermint" -t "adrianbrink/ethermint:$(VERSION_TAG)" -f scripts/docker/Dockerfile .

docker_push:
	docker push "tendermint/ethermint:latest"
	docker push "tendermint/ethermint:$(VERSION_TAG)"
	docker push "adrianbrink/ethermint:latest"
	docker push "adrianbrink/ethermint:$(VERSION_TAG)"


### Ethstats ###
ethstats:
	@git clone https://github.com/tendermint/eth-net-intelligence-api $(CURDIR)/ethstats

ethstats_setup: ethstats
	@cd $(CURDIR)/ethstats && npm install && node scripts/configure.js

ethstats_start:
	@cd $(CURDIR)/ethstats && pm2 start ./app.json

ethstats_stop:
	@cd $(CURDIR)/ethstats && pm2 stop ./app.json

.PHONY: all install build test test_race test_integrations test_coverage linter
	clean draw_deps get_vendor_deps tools update_tools dist publish
	docker_build_develop docker_push_develop docker_build docker_push ethstats
	ethstats_setup ethstats_start ethstats_stop
