FROM golang:1.9.2

RUN apt-get update && apt-get install -y --no-install-recommends \
		zip \
	&& rm -rf /var/lib/apt/lists/*

RUN mkdir -p $GOPATH/src/github.com/tendermint/ethermint
WORKDIR $GOPATH/src/github.com/tendermint/ethermint
