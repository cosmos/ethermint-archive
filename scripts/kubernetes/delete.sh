#!/usr/bin/env bash

DIR=`dirname ${BASH_SOURCE[0]}`

echo "Deleting..."
kubectl delete -f $DIR/ethermint.yaml

echo "Deleteing Volume Claims"
kubectl delete persistentvolumeclaim --selector="app=tm"
kubectl delete service --selector "app=tm"
