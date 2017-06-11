#!/usr/bin/env bash

N=$1 # Number of Nodes
V=$2 # Number of Validators
S=$3 # Number of Seeds

DIR=`dirname ${BASH_SOURCE[0]}`

usage() {
  echo -e '\nUsage: \c'
  echo './start.sh N V S'
  echo '  N - Number of Nodes, should be more than 3'
  echo '  V - Number of Validators, should be more than 3'
  echo '  S - Number of Seeds, should be more than 1'
  echo
  echo 'Example:'
  echo -e '\t./start.sh 10 5 2 \c'
  echo -e '-- This will start 10 nodes, 5 of them as validators and first 2 will be used as Seeds'
}

if [ -z $N ]; then
  usage
  exit 1
fi

if [ -z $V ] || [ $V -gt $N ]; then
  V=$N
fi

if [ $N -lt 4 ] || [ $V -lt 4 ]; then
  echo "Minimum number of nodes and validators should be 4"
  usage
  exit 2
fi

if [ -z $S ] || [ $S -lt 1 ] || [ $S -gt $N ]; then
  S=$V
fi

SEEDS="tm-0"
VALIDATORS="tm-0"

if [ $S -gt 1 ]; then
  for i in `seq 1 $(($S-1))`; do
    SEEDS="$SEEDS,tm-$i"
  done
fi

for i in `seq 1 $(($V-1))`; do
  VALIDATORS="$VALIDATORS,tm-$i"
done

echo "Nodes: $N, Validators: $V, SEEDS: $S"
SETUP_DIR=$(realpath `basename $DIR`/../../setup)

cat $DIR/ethermint.template.yaml \
  | sed "s/{SEEDS}/$SEEDS/; s/{VALIDATORS}/$VALIDATORS/; s/{REPLICAS}/$N/;" \
  | sed "s/{S}/$S/; s/{N}/$N/; s/{V}/$V/;" \
  | sed "s#{HOSTPATH}#$SETUP_DIR#;" \
  > $DIR/ethermint.yaml

kubectl create -f $DIR/ethermint.yaml
