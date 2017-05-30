#!/usr/bin/env bash
shopt -s huponexit

TMPDIR=/tmp/truffle-tests
TESTNAME=test1
TRUFFLE=$1

if [ "$TRUFFLE" == "" ] || [ ! -f $TRUFFLE ]
then
  echo "pass truffle repos file"
  exit 1
fi

ethermint version
geth version

# Start Ethereuum
cd deploytools
echo "Starting Ethereum"
./clean.sh && ./setup-ethereum.sh $TESTNAME
./run-ethereum.sh $TESTNAME &
sleep 20

# Run Tests
cd ..
echo "Running Tests"
./ether-test.sh $TRUFFLE $TMPDIR/$TESTNAME.output.ethereum
pkill -INT -P $(pgrep -P $!)
sleep 30

## Start Ethermint
cd deploytools
echo "Starting Ethermint"
./clean.sh && ./setup-ethermint.sh $TESTNAME
./run-ethermint.sh $TESTNAME &
sleep 20

## Run Tests
cd ..
echo "Running Tests"
./ether-test.sh $TRUFFLE $TMPDIR/$TESTNAME.output.ethermint
pkill -INT -P $(pgrep -P $!)

./deploytools/clean.sh
