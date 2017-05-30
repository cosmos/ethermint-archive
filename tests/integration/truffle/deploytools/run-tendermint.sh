#!/usr/bin/env bash

source ./vars.sh

NETWORK=$1
NAME=$2
DATADIR="/tmp/${TMPROOT}/$NETWORK-$NAME"

tendermint node --home $DATADIR/tendermint
