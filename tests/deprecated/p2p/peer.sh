#! /bin/bash
set -eu

DOCKER_IMAGE=$1
NETWORK_NAME=$2
ID=$3

set +u
SEEDS=$4
set -u
if [[ "$SEEDS" != "" ]]; then
	SEEDS=" --seeds $SEEDS "
fi

echo "starting tendermint peer ID=$ID"
# start ethermint container on the network
docker run -d \
	--net=$NETWORK_NAME \
	--ip=$(test/p2p/ip.sh $ID) \
	--name local_testnet_$ID \
	--entrypoint ethermint \
	-e TMROOT=/go/src/github.com/tendermint/ethermint/test/p2p/data/mach$ID \
	$DOCKER_IMAGE --datadir \$TMROOT $SEEDS --log_level=notice
