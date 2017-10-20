#!/bin/bash

set -u

N=$1

for i in $(seq 1 "$N"); do
  docker stop "tendermint_$i"
  docker stop "ethermint_$i"

  docker rm -fv "tendermint_$i"
  docker rm -fv "ethermint_$i"
done

#delete network
docker network rm ethermint_net

