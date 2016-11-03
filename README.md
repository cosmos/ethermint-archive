## TMSP-Ethereum

### Overview
Ethereum as a TMSP application. 

This means running an Ethereum EVM-based network that uses Tendermint consesnsus instead of proof-of-work.
The way it's built makes it easy to use existing Ethereum tools (geth attach, web3) to interact with the node.

### Docker image
The easiest way to get started is to use the docker image:
```
docker run kobigurk/tmsp-ethereum
```

If you prefer instead to build locally, go to the `docker` directory and run `docker built -t tmsp-ethereum .`.

After running the container, you can attach to it:
* First, find its ID using `docker ps` and use it to find its IP using `docker inspect CONTAINER_ID | grep IPAddress`. 
* Use the IP address to attach with `geth attach rpc:http://CONTAINER_IP:8545`.

### Development and building locally
tmsp-ethereum uses glide for package management. After running `glide install`, issue the following commands to build the appropriate development version of tendermint and geth locally:
* `pushd vendor/github.com/tendermint/tendermint/cmd/tendermint && go get . && go build . && popd`
* `pushd vendor/github.com/ethereum/go-ethereum/cmd/geth && go get . && go build . && popd`

Then, you need to run:
* `vendor/github.com/tendermint/tendermint/cmd/tendermint/tendermint init` - to initialize tendermint
* `vendor/github.com/ethereum/go-ethereum/cmd/geth --datadir geth_data init dev/genesis.json` - to initialize a geth data directory with a genesis block with balance for a pre-made account. 

tmsp-ethereum is based on the following development branches of tendermint and tmsp:
* tendermint: handshake
* tmsp: info\_and\_header

### TODO
* Validator management
* Some more :-)
