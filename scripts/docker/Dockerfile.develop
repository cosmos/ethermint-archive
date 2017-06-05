FROM alpine:3.6

ENV DATA_ROOT /ethermint
ENV TENDERMINT_ADDR tcp://0.0.0.0:46657

RUN addgroup emuser && \
    adduser -S -G emuser emuser

RUN mkdir -p $DATA_ROOT && \
    chown -R emuser:emuser $DATA_ROOT

RUN apk add --no-cache bash

ENV GOPATH /go
ENV PATH "$PATH:/go/bin"
RUN mkdir -p /go/src/github.com/tendermint/ethermint && \
    apk add --no-cache go build-base git linux-headers && \
    cd /go/src/github.com/tendermint/ethermint && \
    git clone https://github.com/tendermint/ethermint . && \
    git checkout develop && \
    make get_vendor_deps && \
    make install && \
    glide cc && \
    cd - && \
    rm -rf /go/src/github.com/tendermint/ethermint && \
    apk del go build-base git

VOLUME $DATA_ROOT

EXPOSE 46658

#ENTRYPOINT ["ethermint"]

CMD ethermint --datadir $DATA_ROOT --tendermint_addr $TENDERMINT_ADDR
