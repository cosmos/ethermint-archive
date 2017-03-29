FROM golang:latest

RUN mkdir -p /go/src/github.com/tendermint/ethermint
WORKDIR /go/src/github.com/tendermint/ethermint

COPY Makefile /go/src/github.com/tendermint/ethermint/
COPY glide.yaml /go/src/github.com/tendermint/ethermint/
COPY glide.lock /go/src/github.com/tendermint/ethermint/

RUN make get_vendor_deps

COPY . /go/src/github.com/tendermint/ethermint
