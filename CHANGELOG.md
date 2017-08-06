# Changelog

## In-Progress
* implements correct syncing status as a custom RPC endpoint  
  * we override the geth RPC endpoint and give it a custom one that uses the tendermint rpc
* rewrite tests into separate directory
* stop P2P layer from starting
* unsafe_reset_all command
* allow multiple transactions per block
* cleanup strategy to prepare for IBC implementation
* refactor Ethereum backend structure
* test improvements
* benchmark improvements

## 0.4.0
### Breaking
* pending.commit(...) takes an ethereum object instead of a blockchain object
    * the change was necessary due to a change in the API exposed by geth
* syncing RPC endpoint always returns true for go-ethereum
    * once tendermint is patched it will return the accurate sync status of the
      underlying tendermint core instance

### Improvements
* update to geth v1.6.6
* add more money to the default account
* increase the wait times in the contract tests to guard against false positives
* added EMHOME variable
* binaries now correctly unzip to `ethermint`
* added benchmarking suite
* general improvements to the build process
* better documentation
* proper linting and error detection
  * can be accessed with `make metalinter` and `make megacheck`


## 0.3.0 (June 23, 2017)
### Improvements
* Gas Limit improvements
* Added multinode deployment possibilities with kubernetes and minikube
* Add tests for deploying of smart contracts and transactions
* Add initial work on the peg zone to the main ethereum network
* Remove all deprecated files
* Cleaned up the repo


## 0.2.0 (June 02, 2017)

BREAKING CHANGES:

FEATURES:

- Update to tmlibs
- Update to abci
- Update to go-ethereum 1.6.1

IMPROVEMENTS:

- Solve numerous bugs
- Add test suite
