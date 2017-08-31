.. _venus.rst:

Connect to a testnet
====================

Join the ethermint testnet with a few commands. This guide is based on the `original Medium article <https://blog.cosmos.network/join-venus-the-new-ethermint-testnet-part-3-8e30c7d5abcf>`_.

Pre-requisites
--------------

You'll need the ``ethermint`` and ``tendermint`` binaries installed with the correct version. Ensure you are using compatible versions (how?). See their respective repos for install instructions.

Check that everything is installed correctly:

::

        tendermint version
        v0.10.3

        ethermint version
        v0.3.0

First, we need some directories for our files:

::

        mkdir --parents ~/.venus/tendermint
        mkdir ~/.venus/ethermint

Second, we're going to need the required initialization files. These consist of a ``genesis.json`` for each ethermint and tendermint and a ``config.toml`` for tendermint. These can be got using the ``--testnet`` flag on each respective program or by cloning the testnets repo. The ``genesis.json`` for tendermint will look like:

::

        {
          "genesis_time":"2017-06-28T13:39:18Z",
          "chain_id":"venus",
          "validators":
          [
            
            {
              "pub_key": {
                "data": "E24396CFDCBAF3BD0F2C5CF510551F70B4634E4E5EBF9655B1FB57F451ABB344",
                "type": "ed25519"
              },
              "amount":10,
              "name":"venus-node3"
            }
            ,
            {
              "pub_key": {
                "data": "2C97BEA0B5A4D9EAE04C242C7AD2A6D2BA989E3C4A7B276AB137300C37EB22F7",
                "type": "ed25519"
              },
              "amount":10,
              "name":"venus-node4"
            }
            ,
            {
              "pub_key": {
                "data": "CA53D568F6EDC245D80D887F068F8FE7E03D540B2F7D2212CA436FC962394EA3",
                "type": "ed25519"
              },
              "amount":10,
              "name":"venus-node0"
            }
            ,
            {
              "pub_key": {
                "data": "CE4877B1E25EDF845EC8F13FAC3B74E0B3A863EBA5B24D10FCDA4A239065E86A",
                "type": "ed25519"
              },
              "amount":10,
              "name":"venus-node5"
            }
            ,
            {
              "pub_key": {
                "data": "8A825277C6A71C89B1F3E9AE5C3853E282F73B6F77A4798C03291EDB1F1F5CA5",
                "type": "ed25519"
              },
              "amount":10,
              "name":"venus-node2"
            }
            ,
            {
              "pub_key": {
                "data": "AF8D2B55E6FAD5DCF6000752E2A05A19B3F42E3072B75BBB2217C43B574ACE99",
                "type": "ed25519"
              },
              "amount":10,
              "name":"venus-node1"
            }
            ,
            {
              "pub_key": {
                "data": "B127402B86C673807AF407B15E324C4565B1A1DC77C85881E7D8D8640C17EBBB",
                "type": "ed25519"
              },
              "amount":10,
              "name":"venus-node6"
            }
          ],
          "app_hash":"",
          "app_options": {}
        }

which shows the validators each with 10 bonded tokens. The name of each validator can be used to view the node's information at, for example: http://venus-node0.testnets.interblock.io/

Let's take a look at the ``config.toml`` for tendermint:

::

        # This is a TOML config file.
        # For more information, see https://github.com/toml-lang/toml
        
        proxy_app = "tcp://127.0.0.1:46658"
        moniker = "bob_the_farmer"
        fast_sync = true
        db_backend = "leveldb"
        log_level = "debug"
        
        [rpc]
        laddr = "tcp://0.0.0.0:46657"
        
        [p2p]
        laddr = "tcp://0.0.0.0:46656"
        seeds = "138.197.113.220:46656,138.68.12.252:46656,128.199.179.178:46656,139.59.184.2:46656,207.154.246.77:46656,138.197.175.237:46656"

The main relevant part is the ``seeds =`` field which has the peers to we'll be dialing to join the network. These IPs should match the URL of each node. The ``moniker =`` can be anything you'd like to name your node.

Finally, we have a ``genesis.json`` for ``ethermint``. It looks pretty much like a ``genesis.json`` for ethereum:

::

        {
            "config": {
                "chainId": 15,
                "homesteadBlock": 0,
                "eip155Block": 0,
                "eip158Block": 0
            },
            "nonce": "0xdeadbeefdeadbeef",
            "timestamp": "0x00",
            "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
            "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
            "difficulty": "0x40",
            "gasLimit": "0x8000000",
            "alloc": {
                "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc": { "balance": "100000000000000" },
        	"0xc6713982649D9284ff56c32655a9ECcCDA78422A": { "balance": "10000000000000000000000000000000000" }
            }
        }

At this point you should have a ``genesis.json`` and ``config.toml`` in ``~/.venus/tendermint`` and a ``genesis.json`` in ``~/.venus/ethermint``.

Initialize
----------

Next, we will initialize ethermint:

::

        ethermint --datadir ~/.venus/ethermint init ~/.venus/ethermint/genesis.json

where the ``--datadir`` specifies the correct directory and ``init`` takes a path to the ``genesis.json``. Look inside ``~/.venus/ethermint/ethermint`` to see the files that were created. 

Run Tendermint
--------------

Then we start up the tendermint node:

::

        tendermint --home ~/.venus/tendermint node

where ``--home`` is basically like the ``--datadir`` flag from running ethermint, and ``node`` is command that starts up the node. You'll see the following output:

::

        I[08-18|17:13:25.880] Generated PrivValidator                      module=node file=/home/zach/.venus/tendermint/priv_validator.json
        I[08-18|17:13:26.100] Starting multiAppConn                        module=proxy impl=multiAppConn
        I[08-18|17:13:26.101] Starting socketClient                        module=abci-client connection=query impl=socketClient
        E[08-18|17:13:26.102] abci.socketClient failed to connect to tcp://127.0.0.1:46658.  Retrying... module=abci-client connection=query
        E[08-18|17:13:29.102] abci.socketClient failed to connect to tcp://127.0.0.1:46658.  Retrying... module=abci-client connection=query

with the last two lines in red. You'll see a steady stream of that error message every three seconds. Notice the first line; you should now have a ``priv_validator.json`` written to disk.

Run Ethermint
-------------

Now you can start the ethermint process:

::

        ethermint --datadir ~/.venus/ethermint/  --rpc --rpcaddr=0.0.0.0 --ws --wsaddr=0.0.0.0 --rpcapi eth,net,web3,personal,admin

There will be about a dozen lines of initialization information, then the output will look similar to:

::

        INFO [08-18|13:35:40] Accepted a new connection                module=abci-server
        INFO [08-18|13:35:40] Waiting for new connection...            module=abci-server
        INFO [08-18|13:35:40] Info 
        INFO [08-18|13:35:40] BeginBlock 
        INFO [08-18|13:35:40] EndBlock 
        INFO [08-18|13:35:40] Commit 
        INFO [08-18|13:35:40] Committing block                         stateHash=fbccc1…f0e986 blockHash=3cddd3…97eb13
        INFO [08-18|13:35:40] Imported new chain segment               blocks=1 txs=0 mgas=0.000 elapsed=2.516ms mgasps=0.000 number=3404 hash=3cddd3…97eb13
        INFO [08-18|13:35:40] Mapped network port                      proto=tcp extport=30303 intport=30303 interface="UPNP IGDv1-PPP1"
        INFO [08-18|13:35:41] BeginBlock 
        INFO [08-18|13:35:41] EndBlock 
        INFO [08-18|13:35:41] Commit 
        INFO [08-18|13:35:41] Committing block                         stateHash=2eb09c…58f60f blockHash=df0411…8c7321
        INFO [08-18|13:35:41] Imported new chain segment               blocks=1 txs=0 mgas=0.000 elapsed=1.315ms mgasps=0.000 number=3405 hash=df0411…8c7321
        INFO [08-18|13:35:41] BeginBlock 

The above is output after the syncing had been stopped at block height 3403 (by terminating the process). Look at ``Imported new chain segment`` => ``number=3404``, which increases by one as your node syncs with the testnet. Your output will start from number 1 unless you have been starting and stopping the nodes.

Congratulation! You are currently syncing up with the testnet. Next, you'll need testnet coins, then try using ``geth`` to create contracts on the testnet.
