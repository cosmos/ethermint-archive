#! /bin/bash

docker build --no-cache -t ethermint_js_test -f ./tests/integration/truffle/Dockerfile ./tests/integration/truffle
