NodeJS / Web3 based load runner for Ethereum / Ethermint

### Pre Requirements
ethermint node(s) started with Ethereum RPC and keystore setup for address `0x7eff122b94897ea5b0e2a9abf47b86337fafebdc`
```
ethermint --targetgaslimit ‭134217728‬ --node_laddr "tcp://127.0.0.1:46656" --rpc_laddr "tcp://127.0.0.1:46657" --addr "tcp://127.0.0.1:46658" --ipcdisable --rpc --rpcaddr "0.0.0.0" --rpccorsdomain "*" --rpcapi "db,eth,net,web3,personal,admin,txpool" --log_level error
```

Block size is limited by `700tx`
```
block_size = 700
```

### Prepare

setup dependencies
```
npm install --save
```

review config
```
config/default.json
```

config can be overwritten by `NODE_CONFIG` 


### Deploy contract
```
node deploy.js
```

### Run test
```
node main.js
```
