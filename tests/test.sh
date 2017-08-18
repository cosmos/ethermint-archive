#!/bin/bash

# -x for debug
set -eux

# count of tendermint/ethermint node
N=1

# Docker version and info
docker version
docker info

# Get the directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"

DATA_DIR="$DIR/tendermint_data"

echo
echo "* [$(date +"%T")] building ethermint docker image"
bash "$DIR/docker/build.sh"

echo
echo "* [$(date +"%T")] building nodejs docker image"
bash "$DIR/integration/truffle/build.sh"

# stop existing container and remove network
set +e
bash "$DIR/p2p/stop_tests.sh" $N
set -e

echo
echo "* [$(date +"%T")] create docker network"
docker network create --driver bridge --subnet 172.58.0.0/16 ethermint_net

echo
echo "* [$(date +"%T")] run tendermint container"

#TODO add loop.
TENDERMINT_IP=$($DIR/p2p/ip.sh 1)
ETHERMINT_IP=$($DIR/p2p/ip.sh 2)

docker run -d \
    --net=ethermint_net \
    --ip "$TENDERMINT_IP" \
    --name tendermint_1 \
    -v "$DATA_DIR/tendermint_1:/tendermint" \
    tendermint/tendermint node --proxy_app tcp://$ETHERMINT_IP:46658

echo
echo "* [$(date +"%T")] run ethermint container"
docker run -d \
    --net=ethermint_net \
    --ip $ETHERMINT_IP \
    --name ethermint_1 \
    ethermint_tester ethermint --datadir=/ethermint/data --rpc --rpcaddr=0.0.0.0 --ws --wsaddr=0.0.0.0 --rpcapi eth,net,web3,personal,admin --tendermint_addr tcp://$TENDERMINT_IP:46657

#wait for tendermint & ethermint start
sleep 60

echo
echo "* [$(date +"%T")] run tests"
docker run --net=ethermint_net \
    --rm -it \
    -e NODE_ENV=test \
    -e WEB3_HOST=$ETHERMINT_IP \
    -e WEB3_PORT=8545 \
    ethermint_js_test npm test

echo
echo "* [$(date +"%T")] stop containers"
bash "$DIR/p2p/stop_tests.sh" $N

echo
echo "* [$(date +"%T")] done"
