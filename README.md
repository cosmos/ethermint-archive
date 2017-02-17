## Ethermint 

[![Build Status](https://travis-ci.org/tendermint/ethermint.svg?branch=master)](https://travis-ci.org/tendermint/ethermint)

### Overview
Ethereum as a [ABCI](https://github.com/tendermint/abci) application. 

This means running an Ethereum EVM-based network that uses Tendermint consesnsus instead of proof-of-work.
The way it's built makes it easy to use existing Ethereum tools (geth attach, web3) to interact with the node.

### Docker image
An easy way to get started is to use the docker image:
```
docker run tendermint/ethermint
```

If you prefer instead to build locally, go to the `docker` directory and run `docker build -t ethermint .`.

After running the container, you can attach to it:
* First, find its ID using `docker ps` and use it to find its IP using `docker inspect CONTAINER_ID | grep IPAddress`. 
* Use the IP address to attach with `geth attach http://CONTAINER_IP:8545`.

### Install
You can build the `ethermint` executable by running `go install cmd/ethermint`. For vendored packages, Ethermint uses glide. 

### Initialization

To get started, you need to initialize the genesis block for tendermint and geth.
This is still a work in progress, but it's possible to do:

```
TMROOT=~/.ethermint tendermint init
ethermint -datadir ~/.ethermint init dev/genesis.json
cp -r dev/keystore ~/.ethermint/keystore
```

Note we're using `tendermint` to create the tendermint files,
and then `ethermint` to create the ethereum files, 
and everything is stored in `~/.ethermint`.
The `dev/genesis.json` file specifies some initial ethereum account with funds,
and the corresponding private key is copied over from `dev/keystore`.
The password to this key is `1234`

### Run

Now just run `ethermint`.  You should see Tendermint blocks start streaming by.
In another window, run `geth attach ~/.ethermint/geth.ipc` to drop into a web3 console for
your ethermint node:

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
