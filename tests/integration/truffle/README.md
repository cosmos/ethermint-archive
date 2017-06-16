Truffle Tests Runner
===

## Dependencies
 - Node.JS/NPM
 - truffle
 - ethermint
 - geth

Usage:
  `npm i` to install dependencies then  
  `./compare.sh file` and `node diff.js`

`file` format: Each file contains repo and the directory in that repo, which should contain truffle tests and configs.
Checkout `file-example`.

`compare.sh` will setup directory in /tmp/testnet folder, run ethermint and tests against it. Then it will create ethermint folder, run it and test against it. Test repositories are cloned into directory /tmp/truffle-tests and test outputs are saved. `./diff.sh` will return for outputs, empty response means both are same(success).

## Test
  [Truffle Initial Repo](https://github.com/trufflesuite/truffle-init-default) tests were run against both networks. It contains simple MetaCoin implementation with it's test.  
  Diff tool checks whether the results are the same for `ethereum` and `ethermint`, if both networks fail certain test then it's expected behaviour.

## Gas limit test

Run `npm test`