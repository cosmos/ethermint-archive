[![CircleCI](https://circleci.com/gh/cosmos/ethermint.svg?style=svg)](https://circleci.com/gh/cosmos/ethermint)
[![](https://godoc.org/github.com/cosmos/ethermint?status.svg)](http://godoc.org/github.com/cosmos/ethermint) [![Go Report Card](https://goreportcard.com/badge/github.com/cosmos/ethermint)](https://goreportcard.com/report/github.com/cosmos/ethermint)

# Ethermint

### Ethermint is currently being worked on by [Chainsafe](https://chainsafe.io/), you can follow the progress [here](https://github.com/ChainSafe/ethermint)

\***\*WARNING:\*\*** Ethermint is under VERY ACTIVE DEVELOPMENT and should be treated as pre-alpha software. This means it is not meant to be run in production, its APIs are subject to change without warning and should not be relied upon, and it should not be used to hold any value. We will remove this warning when we have a release that is stable, secure, and properly tested.

## What is it?

`ethermint` will be an implementation of the EVM that runs on top of [`tendermint`](https://github.com/tendermint/tendermint) consensus, a Proof of Stake system. This project has as its primary goals:

- [Hard Spoon](https://blog.cosmos.network/introducing-the-hard-spoon-4a9288d3f0df) enablement: This is the ability to take a token from the Ethereum mainnet and "spoon" (shift) the balances over to another network. This feature is intended to make it easy for applications that require more transactions than the Ethereum main chain can provide to move their code over to a compatible chain with much more capacity.
- Web3 Compatibility: In order enable applications to be moved over to an ethermint chain existing tooling (i.e. web3 compatible clients) need to be able to interact with `ethermint`.

### Implementation

#### Completed

- Have a working implementation that can parse and validate the existing ETH Chain and persist it in a Tendermint store
- Implement Ethereum transactions in the CosmosSDK

#### Current Work

- Implement web3 compatible API layer
- Implement the EVM as a CosmosSDK module
- Allow the Ethermint EVM to interact with other [Cosmos SDK modules](https://github.com/cosmos/cosmos-sdk/blob/master/docs/core/app3.md)

#### Next Steps

- Hard spoon enablement: The ability to export state from `geth` and import token balances into Ethermint
- Ethermint is a functioning Cosmos SDK application and can be deployed as its own zone
- Full web3 compatibility will enable existing Ethereum applications to use Ethermint

### Building Ethermint

To build, execute the following commands:

```bash
# To build the binary and put the resulting binary in ./build
$ make tools deps build

# To build the project and install it in $GOBIN
$ make tools deps install
```

### Tests

Integration tests are invoked via:

```bash
$ make test
```

To run CLI tests, execute:

```bash
$ make test-cli
```

#### Ethereum Mainnet Import

There is an included Ethereum mainnet exported blockchain file in `importer/blockchain`
that includes blocks up to height `97638`. To execute and test a full import of
these blocks using the EVM module, execute:

```bash
$ make test-import
```

You may also provide a custom blockchain export file to test importing more blocks
via the `--blockchain` flag. See `TestImportBlocks` for further documentation.

### Community

The following chat channels and forums are a great spot to ask questions about Ethermint:

- [Cosmos Riot Chat Channel](https://riot.im/app/#/group/+cosmos:matrix.org)
- Cosmos Forum [![Discourse status](https://img.shields.io/discourse/https/forum.cosmos.network/status.svg)](https://forum.cosmos.network)
