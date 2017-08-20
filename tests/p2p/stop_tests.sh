#!/bin/bash

set -u

N=$1

for i in $(seq 1 "$N"); do
  docker rm -f "tendermint_$i"
  docker rm -f "ethermint_$i"
done

#delete network
docker network rm ethermint_net
