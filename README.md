## Ethermint

[![Build Status](https://circleci.com/gh/tendermint/ethermint/tree/master.svg?style=shield)](https://circleci.com/gh/tendermint/ethermint/tree/master)

### Overview

Ethereum as a [ABCI](https://github.com/tendermint/abci) application.

This means running an Ethereum EVM-based network that uses Tendermint consesnsus instead of proof-of-work.
The way it's built makes it easy to use existing Ethereum tools (geth attach, web3) to interact with the node.

### Docker image

An easy way to get started is to use the docker image:

```
docker run -d --restart=unless-stopped -p 8545:8545 tendermint/ethermint
```

If you prefer instead to build locally, run `make build-docker`.

After running the container, you can attach to it: `geth attach http://localhost:8545`.

### Install
You can build the `ethermint` executable by running `go install ./cmd/ethermint`. For vendored packages, Ethermint uses glide.

### Initialization

To get started, you need to initialize the genesis block for tendermint and geth.
This is still a work in progress, but it's possible to do:

For example you want to store all file inside `/tmp/eth/` dir.

```
tendermint init --home /tmp/eth/tendermint
```

Set `app_hash=D4E56740F876AEF8C010B86A40D5F56745A118D0906A34E69AEC8C0DB1CB8FA3` in `/tmp/eth/tendermint/genesis.json`

Run tendermint:
```
tendermint node --moniker node1 --proxy_app tcp://127.0.0.1:46658 --home /tmp/eth/tendermint
```

Then run ethermint:
```
ethermint -datadir /tmp/eth/ethermint init dev/genesis.json
cp -r dev/keystore /tmp/eth/ethermint/keystore
ethermint --datadir /tmp/eth/ethermint --tendermint_addr tcp://localhost:46657 --rpc --rpcaddr=0.0.0.0 --ws --wsaddr=0.0.0.0 --rpcapi eth,net,web3,personal,admin
```

The `dev/genesis.json` file specifies some initial ethereum account with funds,
and the corresponding private key is copied over from `dev/keystore`.
The password to this key is `1234`

### Run

In another window, run `geth attach http://localhost:8545` or `geth attach /tmp/eth/ethermint/geth.ipc` to drop into a web3 console for
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
