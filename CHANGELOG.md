# Changelog

## 0.5.0
### Breaking Changes
* none

### Improvements
* temporary state is applied during CheckTx calls
  * allows multiple transactions per block from the same account
* strict enforcement of nonce increases
  * a nonce has to be strictly increasing instead of just being greater
* semver guarantees to all exported APIs
* documentation for readthedocs in docs folder
* new getting started documentation for newcomers
* update to go-ethereum 1.6.7
* general dependency update
* new `--with-tendermint` flag to start Tendermint in process
  * Tendermint binary needs to be installed in $PATH
  * NOTE: no library-level dependency on Tendermint core
* networking is turned off for ethereum node
  * NOTE: allows running of go-ethereum and ethermint side-by-side
* new `unsafe_reset_all` command to reset all initialisation files
* incremental performance optimisations
* rework test suite for readability and completeness
* add metalinter to build process for code quality
* add maintainer and OWNER file


## 0.4.0
### Breaking
* pending.commit(...) takes an ethereum object instead of a blockchain object
the change was necessary due to a change in the API exposed by geth

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
* ethereum node p2p server does not get started anymore


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
