### Overview

Docker-compose allow to start 4 instances of tendermint and ethermint together in one cluster.

#### Build ethermint docker image

`make build-docker`

#### Run cluster

`docker-compose up -d`

#### Inspect IP address
`docker inspect --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $(docker ps -f name=ethermint* -q)`

### Check with geth

```
geth attach http://<IP_ADDRESS>:8545
Welcome to the Geth JavaScript console!

instance: ethermint/linux-amd64/go1.8.1
coinbase: 0x7eff122b94897ea5b0e2a9abf47b86337fafebdc
at block: 316 (Fri, 26 May 2017 11:12:03 MSK)
 datadir: /ethermint/data
 modules: admin:1.0 eth:1.0 net:1.0 personal:1.0 rpc:1.0 web3:1.0

> eth.blockNumber
317
```

### How setup docker-compose for any ABCI application

1. Initialize tendermint for each instance

```
tendermint init --home ./tendermint_1
tendermint init --home ./tendermint_2
tendermint init --home ./tendermint_3
tendermint init --home ./tendermint_4
```

2. Patch `genesis.json` in each dir
* set `chain_id` the same in all genesis
* compose validators into one and set appropriate name to it
* duplicate new `genesis.json` to each dir

3. In `docker-compose.yaml` define `--proxy-app` param to related ABCI application and `--p2p.seeds` to all another tendermint hosts