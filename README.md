## Ethermint 

[![Build Status](https://travis-ci.org/tendermint/ethermint.svg?branch=master)](https://travis-ci.org/tendermint/ethermint)

### Overview
Ethereum as a TMSP application. 

This means running an Ethereum EVM-based network that uses Tendermint consesnsus instead of proof-of-work.
The way it's built makes it easy to use existing Ethereum tools (geth attach, web3) to interact with the node.

### Docker image
An easy way to get started is to use the docker image:
```
docker run kobigurk/ethermint
```

If you prefer instead to build locally, go to the `docker` directory and run `docker build -t ethermint .`.

After running the container, you can attach to it:
* First, find its ID using `docker ps` and use it to find its IP using `docker inspect CONTAINER_ID | grep IPAddress`. 
* Use the IP address to attach with `geth attach rpc:http://CONTAINER_IP:8545`.

### Development and building locally
You can build the `ethermint` executable by running `go install cmd/ethermint`. For vendored packages, Ethermint uses glide. 

### Running
#### Ethermint
You need to init the genesis block for tendermint and geth. This part is still very much a work-in-progress, but it's possible to work with it now. 
By running `./ethermint -datadir data init genesis.json`, both the tendermint genesis and geth genesis will be generated. The tendermint genesis resides in `~/.tendermint`. I recommend using the `genesis.json` that exists in the `dev` directory as it gives an initial balance to an address whose private key is in the `geth_data` folder. Its password is `123`.

#### Mining reward and validator management
Ethermint implements hooks that can be customized for mining rewards and validator management. The default at this point is to not reward and not change the validator set. Example strategies can be found in the `strategies` folder.
