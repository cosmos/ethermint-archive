#!/bin/bash

# -x for debug
set -eux

# count of tendermint/ethermint node
N=4

# Docker version and info
docker version
docker info

# Get the directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"

# Directory where we keep tendermint genesis and private key
DATA_DIR="$DIR/tendermint_data"

# Build docker image for ethermint from current source code
echo
echo "* [$(date +"%T")] building ethermint docker image"
bash "$DIR/docker/build.sh"

# Build docker image for web3 js tests
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

# Generate seeds parameter to connect all tendermint nodes in one cluster
SEEDS=$(bash $DIR/p2p/seeds.sh $N)
if [[ "$SEEDS" != "" ]]; then
	SEEDS="--p2p.seeds $SEEDS"
fi

# Start N nodes of tendermint and N node of ethermint.
for i in $(seq 1 "$N"); do
	echo
	echo "* [$(date +"%T")] run tendermint $i container"

	# We have N as input parameter and we need to generate N*2 IP addresses
	# for tendermint and ethermint.
	# So, here we calculate offset for IP
	index=$(($i*2))
	nextIndex=$(($i*2+1))

	TENDERMINT_IP=$($DIR/p2p/ip.sh $index)
    ETHERMINT_IP=$($DIR/p2p/ip.sh $nextIndex)

	# Start tendermint container. Pass ethermint IP and seeds
    docker run -d \
        --net=ethermint_net \
        --ip "$TENDERMINT_IP" \
        --name tendermint_$i \
        -v "$DATA_DIR/tendermint_$i:/tendermint" \
        tendermint/tendermint node --proxy_app tcp://$ETHERMINT_IP:46658 $SEEDS

	# Start ethermint container. Pass tendermint IP
	echo
    echo "* [$(date +"%T")] run ethermint $i container"
    docker run -d \
        --net=ethermint_net \
        --ip $ETHERMINT_IP \
        --name ethermint_$i \
        ethermint_tester ethermint --datadir=/ethermint/data --rpc --rpcaddr=0.0.0.0 --ws --wsaddr=0.0.0.0 --rpcapi eth,net,web3,personal,admin --tendermint_addr tcp://$TENDERMINT_IP:46657

done


# Wait for tendermint & ethermint start
sleep 60

# Run container with web3 js tests
# Pass IP address of last ethermint node
echo
echo "* [$(date +"%T")] run tests"
docker run --net=ethermint_net \
    --rm -it \
    -e NODE_ENV=test \
    -e WEB3_HOST=$ETHERMINT_IP \
    -e WEB3_PORT=8545 \
    ethermint_js_test npm test

# Stop and remove containers. Remove network
echo
echo "* [$(date +"%T")] stop containers"
bash "$DIR/p2p/stop_tests.sh" $N

echo
echo "* [$(date +"%T")] done"
