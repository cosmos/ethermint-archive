# Changelog

## 0.4.0
### Breaking
* pending.commit(...) takes an ethereum object instead of a blockchain object
    * the change was necessary due to a change in the API exposed by geth

### Improvements
* Update to geth v1.6.6
* Add more money to the default account
* Increase the wait times in the contract tests to guard against false positives


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
