.PHONY: get_deps build all list_deps install

all: get_deps install test

TMROOT = $${TMROOT:-$$HOME/.tendermint}
define NEWLINE


endef
NOVENDOR = go list github.com/tendermint/ethermint/... | grep -v /vendor/

install: get_deps
	go install github.com/tendermint/ethermint/cmd/ethermint

test: build
	go test `${NOVENDOR}`
	
test_race: build
	go test -race `${NOVENDOR}`

get_deps:
	go get -d `${NOVENDOR}`
	go list -f '{{join .TestImports "\n"}}' github.com/tendermint/ethermint/... | \
		grep -v /vendor/ | sort | uniq | \
		xargs go get

get_vendor_deps:
	go get github.com/Masterminds/glide
	glide install --strip-vendor

docker_build:
	rm -rf ./docker/ethermint
	docker run -it --rm -v "$(PWD):/go/src/github.com/tendermint/ethermint" -w "/go/src/github.com/tendermint/ethermint" \
        golang:1.6 go build --ldflags '-extldflags "-static"' \
        -o /go/src/github.com/tendermint/ethermint/docker/ethermint ./cmd/ethermint/
	docker build -t "tendermint/ethermint" ./docker
