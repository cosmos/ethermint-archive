# Changelog

## In-Progress
* add semver guarantees to all exported APIs
* add getting started documentation to the docs folder
* add readthedocs to the docs folder
* update to go-ethereum 1.6.7
* update dependencies
* add unsafe reset command to easily reset all the initialisation files
* run integration tests on CI
* add build targets for OSX
* add `--with-tendermint` flag to start tendermint in process
  * tendermint needs to be installed as a binary
  * tendermint core is not pulled in as a dependency
* rework the docs
* no networking is started by the ethereum node
* performance optimisation
* doc updates

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
