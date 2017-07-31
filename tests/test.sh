#!/bin/bash

set -eu

# Get the directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"

LOGS_DIR="$DIR/logs"
echo
echo "* [$(date +"%T")] cleaning up $LOGS_DIR"
rm -rf "$LOGS_DIR"
mkdir -p "$LOGS_DIR"

echo
echo "* [$(date +"%T")] building ethermint docker image"
bash "$DIR/docker/build.sh"

echo
echo "* [$(date +"%T")] building nodejs docker image"
bash "$DIR/integration/truffle/build.sh"

echo
echo "* [$(date +"%T")] run tendermint container"
docker pull tendermint/tendermint && \
     docker run -it --rm -v "$LOGS_DIR/tendermint:/tendermint" tendermint/tendermint init && \
     docker run -it --rm -v "$LOGS_DIR/tendermint:/tendermint" tendermint/tendermint node --moniker=node1 --proxy_app tcp://ethermint1:46658



#docker run --rm -it -e NODE_ENV=test -e WEB3_HOST=127.0.0.1 -e WEB3_PORT=8545 ethermint_js_test npm test