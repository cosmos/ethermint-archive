#!/usr/bin/env bash
shopt -s huponexit

function join_by { local IFS="$1"; shift; echo "$*"; }

DIR=$(dirname ${BASH_SOURCE[0]})
source $DIR/vars.sh

NETWORK=$1
NAME=$2
PORT=$3
KEYSTORE=./keystore
DATADIR="/tmp/${TMPROOT}/$NETWORK-$NAME"

if [ "$DATADIR" == "" ]; then
  echo "Provide Datadir"
  exit 1
fi

if [ "$NETWORK" == "" ]; then
  echo "Provide Network"
  exit 1
fi

if [ "$PORT" == "" ]; then
  PORT="8545"
fi

KEYS=$(join_by , $(ls $KEYSTORE))
PASS=./passwords
FLAGS=(--datadir $DATADIR/$NETWORK \
  --rpc --rpcapi eth,net,web3,personal,miner,admin \
  --rpcaddr 127.0.0.1 --rpcport $PORT \
  --ws --wsapi eth,net,web3,personal,miner,admin \
  --wsaddr 127.0.0.1 --wsport 8546 \
  --unlock $KEYS --password $PASS \
)

if [ "$NETWORK" == "ethereum" ]; then
  geth version
  geth --mine --fakepow --nodiscover --maxpeers 0 ${FLAGS[@]} &
  wait
else
  echo tendermint node --home $DATADIR/tendermint
  tendermint node --home $DATADIR/tendermint &
  echo ethermint ${FLAGS[@]}
  ethermint ${FLAGS[@]} &
  wait
fi
