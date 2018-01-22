.. _getting-started: 

Getting Started
===============

Running Ethermint requires both ``ethermint`` and ``tendermint`` installed. Review the installation instructions below:

Setup Ethermint
---------------

To get started, you need to initialise the genesis block for tendermint core and go-ethereum. We provide initialisation
files with reasonable defaults and money allocated into a predefined account. If you installed from binary or docker
please download `the default files here <https://github.com/tendermint/ethermint/tree/develop/setup>`_.

You can choose where to store the ethermint files with ``--datadir``. For this guide, we will use ``~/.ethermint``, which is a reasonable default in most cases.

Before you can run ethermint you need to initialise tendermint and ethermint with their respective genesis states.
Please switch into the folder where you have the initialisation files. If you installed from source you can just follow
these instructions.

::

        tendermint init --home ~/.ethermint/tendermint

        ethermint --datadir ~/.ethermint init

Run Ethermint
-------------

To execute ethermint we need to start two processes. The first one is for tendermint, which handles the P2P
communication as well as the consensus process, while the second one is actually ethermint, which provides the
go-ethereum functionality.

::

        tendermint --home ~/.ethermint/tendermint node

        ethermint --datadir ~/.ethermint --rpc --rpcaddr=0.0.0.0 --ws --wsaddr=0.0.0.0 --rpcapi eth,net,web3,personal,admin

The **password** for the default account is *1234*, which we'll need below.

For development purposes, you can enable tendermint to only produce blocks when there are transactions. Add to the
``~/.ethermint/tendermint/config.toml`` file:

::

        [consensus]
        create_empty_blocks = false

        # create_empty_blocks_interval = 5

Notice the alternative feature which allows you to set an empty block interval. However, blocks will be produced without waiting
for the interval if there are transactions.

Connect to Geth
---------------

First, install `geth <https://github.com/ethereum/go-ethereum>`_, then:

::

        geth attach http://localhost:8545

which drops you into a web3 console. 

Now we have access to all the functions from web3 at our fingertips, try:

::
        
        > eth

to see all the options.

Let's start by looking at our default accounts with:

::
        
        > eth.accounts

There will be only one account and it should match the account given with:

::
        
        cat ~/.ethermint/keystore/UTC--2016-10-21T22-30-03.071787745Z--7eff122b94897ea5b0e2a9abf47b86337fafebdc

and note that the last portion of that file name is your account (except for ``0x``) as with the first field of the file itself.

We can also view the block number:

::
        
        > eth.blockNumber

which will be in sync with the logs of ``ethermint``:

::
        
        INFO [08-07|22:32:30] Imported new chain segment   blocks=1 txs=0 mgas=0.000 elapsed=1.705ms   mgasps=0.000  number=248 hash=7fbd05…a231a8
        INFO [08-07|22:32:31] BeginBlock 
        INFO [08-07|22:32:31] EndBlock 
        INFO [08-07|22:32:31] Commit 
        INFO [08-07|22:32:31] Committing block		   stateHash=c0d88d…3a474a blockHash=83b9df…5fc4cb

and of ``tendermint``:

::
        
        I[08-08|02:32:30.000] Executed block		module=state height=248 validTxs=0 invalidTxs=0
        I[08-08|02:32:30.042] Committed state		module=state height=248 txs=0 hash=A524F17E9E1EDE3438B2B8DB231B719BCA8A38B5872C48E43A6B29BB189FA749

note that the block height is increasing approximately once per second. Next, we can see the balance of our accounts with:

::
        
        > eth.getBalance("0x7eff122b94897ea5b0e2a9abf47b86337fafebdc")

which should ``1e+34`` if you haven't yet sent a transaction or deployed a contract

Before deploying a contract, we must unlock the key. First, note that it is locked:

::
        
        > web3.personal

and you'll see ``status: "Locked"`` a few lines down. But wait, why did we go from ``eth`` to ``web3``? We're not sure but that's how it works so follow along.

::
        
        > web3.personal.unlockAccount("0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "1234", 100000)

where the first argument is your account, the second your password (see above), and the third - the amount of time in seconds to keep key unlocked.


Now we can deploy a contract. Since ``eth.compile`` wasn't quite working for me, we can use `browser solidity <https://ethereum.github.io/browser-solidity>`_. Let's use a short contract like:

::
        
        pragma solidity ^0.4.0;
        
        contract Test { 
            function double(int a) constant returns(int) {
                return 2*a;
            } 
        }

then look for the ``Contract details (bytecode, interface etc.)`` on the right sidebar. Copy the code from the "Web3 deploy" section, which will be similar to:

::
        
        var browser_double_sol_testContract = web3.eth.contract([{"constant":true,"inputs":[{"name":"a","type":"int256"}],"name":"double","outputs":[{"name":"","type":"int256"}],"payable":false,"type":"function"}]);
        var browser_double_sol_test = browser_double_sol_testContract.new(
           {
             from: web3.eth.accounts[0], 
             data: '0x6060604052341561000f57600080fd5b5b60ab8061001e6000396000f30060606040526000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680636ffa1caa14603d575b600080fd5b3415604757600080fd5b605b60048080359060200190919050506071565b6040518082815260200191505060405180910390f35b60008160020290505b9190505600a165627a7a72305820c5fd101c8bd62761d1803c865fd4af5c57f3752e6212d7ccebd5b4a23fcd23180029', 
             gas: '4300000'
           }, function (e, contract){
            console.log(e, contract);
            if (typeof contract.address !== 'undefined') {
                 console.log('Contract mined! address: ' + contract.address + ' transactionHash: ' + contract.transactionHash);
            }
         })

and paste it directly in the ``geth`` console. A handful of dots will accrue on each line but the code should run and deploy the contract. You'll see something like:

::
        
        null [object Object]
        undefined
        Contract mined! address: 0xab119259ff325f845f8ce59de8ccf63e597a74cd transactionHash: 0xf3031c975ef55d14a0382df748b3e66a22c61922b80075ee244c493db5f80c5c

which has the information you need to call this contract on the chain.

From the ``ethermint`` logs we'll see a big stream of data while the ``tendermint`` logs will show the ``validTxs`` and ``txs`` fields increase from 0 to 1.

That's it, you've deployed a contract to ethermint! Next, we can call a contract or setup a testnet.
