#!/usr/bin/env bash

DIR=$(dirname ${BASH_SOURCE[0]})
source $DIR/vars.sh

NETWORK=$1
NAME=$2
TMPFILES="/tmp/$TMPROOT/$NETWORK-$NAME"
NETWORK_DATADIR="$TMPFILES/$NETWORK"
TENDERMINT_DATADIR="$TMPFILES/tendermint"

TENDERMINT=$(which tendermint)
ETHERMINT=$(which ethermint)
EXEC=$(which geth)

KEYS=$(ls $KEYSTORE)


mkdir -p $TMPFILES $NETWORK_DATADIR
cp -rv keystore/ $NETWORK_DATADIR/keystore

if [ "$NETWORK" == "ethermint" ]; then
  mkdir $TENDERMINT_DATADIR
  cp -v ./priv_validator.json $TENDERMINT_DATADIR
  cp -v ./genesis.json $TENDERMINT_DATADIR
  cd $TENDERMINT_DATADIR
  $TENDERMINT init --home $TENDERMINT_DATADIR
  cd -
  EXEC=$ETHERMINT
fi

$EXEC --datadir $NETWORK_DATADIR init ./ethgen.json
