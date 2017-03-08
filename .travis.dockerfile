FROM golang:1.6
MAINTAINER ethan@tendermint.com

RUN curl https://glide.sh/get | sh
ADD . /go/src/github.com/tendermint/ethermint
RUN go install github.com/tendermint/ethermint/cmd/ethermint

RUN useradd -ms /bin/bash ethermint
RUN mkdir /data && chown ethermint /data

USER ethermint

VOLUME /data

EXPOSE 46656
EXPOSE 46657
EXPOSE 8545

ENTRYPOINT ["ethermint", "-datadir", "/data"]
