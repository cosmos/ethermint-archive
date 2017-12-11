FROM golang:1.8.3

RUN mkdir /ethermint && \
    chmod 777 /ethermint

# Setup ethermint repo
ENV REPO $GOPATH/src/github.com/tendermint/ethermint
WORKDIR $REPO

# Install the vendored dependencies before copying code
# docker caching prevents reinstall on code change!
ADD glide.yaml glide.yaml
ADD glide.lock glide.lock
ADD Makefile Makefile
RUN make get_vendor_deps

# Now copy in the code
COPY . $REPO

RUN make install

RUN mkdir -p /ethermint/setup
COPY setup/genesis.json /ethermint/setup/
RUN ethermint -datadir /ethermint/data init /ethermint/setup/genesis.json
COPY setup/keystore /ethermint/data/keystore

# expose the volume for debugging
VOLUME $REPO

EXPOSE 46658
EXPOSE 8545
EXPOSE 8546
