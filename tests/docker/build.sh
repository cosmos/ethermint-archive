#! /bin/bash

docker build --no-cache -t ethermint_tester -f ./tests/docker/Dockerfile .
