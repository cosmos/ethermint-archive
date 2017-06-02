FROM alpine:latest
MAINTAINER ethan@tendermint.com

RUN mkdir /ethermint && \
    chmod 777 /ethermint

RUN mkdir -p /app
WORKDIR /app
COPY ./build/ethermint /app/ethermint
RUN ls -la /app

RUN mkdir -p /ethermint/setup
COPY docker/genesis.json /ethermint/setup/
RUN /app/ethermint -datadir /ethermint/data init /ethermint/setup/genesis.json
COPY docker/keystore /ethermint/data/keystore

EXPOSE 8545
EXPOSE 8546

EXPOSE 46656
EXPOSE 46657
EXPOSE 46658

EXPOSE 30303

ENTRYPOINT ["./ethermint"]

CMD ["--datadir=/ethermint/data", "--rpc", "--rpcaddr=0.0.0.0", "--ws", "--wsaddr=0.0.0.0", "--rpcapi", "eth,net,web3,personal,admin"]
