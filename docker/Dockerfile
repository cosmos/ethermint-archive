FROM alpine:latest

MAINTAINER ethan@tendermint.com
# Where to install binaries
ENV INSTALL_BASE /usr/local/bin

# User Creation
# Choose the user id number 1000 to work well with Boot2Docker volumes.  See:
# https://github.com/boot2docker/boot2docker/issues/581#issuecomment-62491280
ENV USER ethermint
RUN addgroup -S $USER && adduser -s /bin/bash -S -D -u 1000 $USER $USER
RUN chown $USER /var/run


RUN mkdir /ethermint && \
  mkdir /ethermint/setup && \
  chown --recursive $USER:$USER /ethermint


COPY ./ethermint $INSTALL_BASE/ethermint
RUN chmod +x $INSTALL_BASE/ethermint

USER $USER

COPY genesis.json /ethermint/setup/
RUN ethermint -datadir /ethermint/data init /ethermint/setup/genesis.json
COPY keystore /ethermint/data/keystore

EXPOSE 8545
EXPOSE 8546

EXPOSE 46656
EXPOSE 46657
EXPOSE 46658

EXPOSE 30303

ENTRYPOINT "$INSTALL_BASE/ethermint"

CMD ["--datadir=/ethermint/data", "--rpc", "--rpcaddr=0.0.0.0",  "--ws", "--wsaddr=0.0.0.0", "--rpcapi", "eth,net,web3,personal,admin"]
