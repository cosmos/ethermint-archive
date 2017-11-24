# Changelog

## Ongoing
### Breaking
* none so far

### Features
* no empty blocks is implemented
  * Ethermint will halt until there are transactions

### Fixes
* none so far


## v0.5.0 (2017-10-22)
### Breaking
* none

### Features
* temporary state is applied during CheckTx calls
  * allows multiple transactions per block from the same account
* strict enforcement of nonce increases
  * a nonce has to be strictly increasing instead of just being greater
 * new `--with-tendermint` flag to start Tendermint in process
  * Tendermint binary needs to be installed in $PATH
  * NOTE: no library-level dependency on Tendermint core
* networking is turned off for ethereum node
  * NOTE: allows running of go-ethereum and ethermint side-by-side
* new `unsafe_reset_all` command to reset all initialisation files 

### Improvements
* semver guarantees to all exported APIs
* documentation for readthedocs in docs folder
* new getting started documentation for newcomers
* update to go-ethereum 1.6.7
* general dependency update
* incremental performance optimisations
* rework test suite for readability and completeness
* add metalinter to build process for code quality
* add maintainer and OWNER file


## 0.4.0 (2017-07-19)
### Breaking
* pending.commit(...) takes an ethereum object instead of a blockchain object
the change was necessary due to a change in the API exposed by geth

### Features
* added benchmarking suite
* added EMHOME variable

### Improvements
* update to geth v1.6.6
* add more money to the default account
* increase the wait times in the contract tests to guard against false positives
* binaries now correctly unzip to `ethermint`
* general improvements to the build process
* better documentation
* proper linting and error detection
  * can be accessed with `make metalinter` and `make megacheck`


## v0.3.0 (2017-06-23)
### Breaking
* none

### Features
* Added multinode deployment possibilities with kubernetes and minikube
* Add tests for deploying of smart contracts and transactions

### Improvements
* Gas Limit improvements
* Add initial work on the peg zone to the main ethereum network
* Remove all deprecated files
* Cleaned up the repo


## v0.2.0 (2017-06-02)
### Breaking
* none

### Features
* Add test suite

### Improvements
* Update to tmlibs
* Update to abci
* Update to go-ethereum 1.6.1
* Solve numerous bugs
