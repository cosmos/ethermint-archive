#!/usr/bin/env bash

N=$1

if [ -z $N ]; then
  echo "Provide Number of nodes to expose"
  exit
fi

for i in `seq 0 $(($N-1))`; do
  kubectl expose pod tm-$i --type NodePort --port 8545,46657
done
