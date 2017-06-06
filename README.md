## Ethermint

[![Build Status](https://circleci.com/gh/tendermint/ethermint/tree/master.svg?style=shield)](https://circleci.com/gh/tendermint/ethermint/tree/master)

[![](https://tokei.rs/b1/github/tendermint/ethermint)](https://github.com/tendermint/ethermint)

### Overview

Ethermint enables ethereum to run as an [ABCI](https://github.com/tendermint/abci) application on tendermint and the COSMOS hub. This application allows you to get all the benefits of ethereum without having to run your own miners.

This means running an Ethereum EVM-based network that uses Tendermint consesnsus instead of proof-of-work.
The way it's built makes it easy to use existing Ethereum tools (geth attach, web3) to interact with the node.

### Install
Currently, we are not shipping executable binaries and hence you have to build ethermint from source. To do so, please install go1.8. Once you have go installed you can build the `ethermint` executable by running `git clone https://github.com/tendermint/ethermint.git`. Afterwards, please switch into the ethermint directory and run `make install`, which will place the binary in your $GOPATH.

You will also need to have `tendermint` installed. Please follow this [guide](https://tendermint.com/docs/guides/install).

### Getting started
To get started, you need to initialise the genesis block for tendermint core and go-ethereum.
This is still a work in progress, but it's possible to do:

You can choose where to store the ethermint files with `--datadir`. For this guide, we will use `~/.ethermint`, which is a reasonable default in most cases.

#### Tendermint
First you need to initialise the tendermint engine

```
tendermint init --home ~/.ethermint/tendermint
```

and then run it with
```
tendermint node --home ~/.ethermint/tendermint
```

#### Ethermint
First you need to initialise ethermint.
For that please switch into the source code folder for ethermint.

```
ethermint --datadir ~/.ethermint init dev/genesis.json
cp -r dev/keystore ~/.ethermint/keystore
```

And the you can run ethermint with
```
ethermint --datadir ~/.ethermint --rpc --rpcaddr=0.0.0.0 --ws --wsaddr=0.0.0.0 --rpcapi eth,net,web3,personal,admin
```

The `dev/genesis.json` file specifies some initial ethereum account with funds,
and the corresponding private key is copied over from `dev/keystore`.
The password to this key is `1234`

#### Geth

Lastly, you can use all the cool ethereum tools that you are already used to. For example, you can use geth to interact with the ethereum instance.

```
geth attach http://localhost:8545
```
or
```
geth attach ~/.ethermint/geth.ipc
```
will drop you into a web3 console.

```
instance: Geth/linux/go1.7.4
coinbase: 0x7eff122b94897ea5b0e2a9abf47b86337fafebdc
at block: 71 (Thu, 16 Feb 2017 16:08:20 EST)
 datadir: /home/user/.ethermint
 modules: admin:1.0 debug:1.0 eth:1.0 personal:1.0 rpc:1.0 txpool:1.0 web3:1.0

> primary = eth.accounts[0]
"0x7eff122b94897ea5b0e2a9abf47b86337fafebdc"
> balance = web3.fromWei(eth.getBalance(primary), "ether");
10000000000000000
> personal.unlockAccount(primary, "1234", 100)
true
>
```

And you're off!


#### Mining reward and validator management
Ethermint implements hooks that can be customized for mining rewards and validator management. The default at this point is to not reward and not change the validator set. Example strategies can be found in the `strategies` folder.
