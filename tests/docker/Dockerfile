FROM golang:1.8.3

# Add testing deps for curl
#RUN echo 'deb http://httpredir.debian.org/debian testing main non-free contrib' >> /etc/apt/sources.list

# Grab deps (jq, hexdump, xxd, killall)
#RUN apt-get update && \
#  apt-get install -y --no-install-recommends \
#  jq bsdmainutils vim-common psmisc netcat curl

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

# expose the volume for debugging
VOLUME $REPO

EXPOSE 46658
EXPOSE 8545
EXPOSE 8546