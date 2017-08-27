Architecture
============

High Level Overview
-------------------

User Experience
^^^^^^^^^^^^^^^

This describes what using ethermint from a user experience is like.

The first entry point is installing ethermint and the underlying tendermint binary. This is as easy as
``brew install ethermint``. It pulls both binaries, since tendermint is a dependency of ethermint and
places them in the correct folders. On Linux distros this process should be the same using ``apt-get``.
On Windows this should be done with chocolate.

To check whether the binaries are installed correctly a user can use the version command
``ethermint version``. This prints the versions of the installed binaries of ethermint and tendermint.

Once the software is installed a user wants to run it. There are three possible scenarios.
1. The user wants to connect to the live network.
2. The user wants to connect to the test network.
3. The user wants to run a private network.

To connect to a live network the command simply is ``ethermint``. This initialises tendermint and
ethermint to with the correct values and starts both processes, which then start syncing.

To connect to a test network the command simply is ``ethermint testnet``. This does the same
initialisation as above but uses the testnet configuration.

To run a private network the command simply is ``ethermint development``. This does the same
initialisation as above but uses the pre-configured private testnet values. The extra step is
that it also copies a correct private key which has a very high amount of money.


Optionally a user can configure these global flags:
Non-consensus - these can be different between all nodes
* ``--gasprice`` sets the minimal gasprice for this node to include a transaction
* ``--coinbase`` sets the address which receives the rewards (this depends on implementation of
official networks)

Consensus - these need to be the same between all nodes
* ``--gaslimit``
* ``--home`` specifies the home directory
* ``--config`` specifies a TOML config file where all these parameters are read from, the cli flags
always take precedence
* ``--rpc``
* **TODO: Spec out the remaining flags like RPC, IPC, WS, USB**

* all flags from tendermint


Developer Experience
^^^^^^^^^^^^^^^^^^^^

Ethermint is designed to be a library. That is why we offer different reward strategies. Everything
in the ``cmd/`` folder is just wiring up the parts. It shouldn't introduce anything new. It should
be an example of how to create an application using the ethermint library. For example it shouldn't
declare flags. Everything should be unexported and it should be possible to create exactly the
same version of ethermint using only the exported packages without having to redefine anything yourself.

The web3 endpoints offered by ethermint are a superset of normal web3. It also allows to send IBC
transactions.


Light Client
^^^^^^^^^^^^

Since we are implementing our own RPC package (which wraps go-ethereum RPC) to expose the correct
web3 endpoints that are needed for ethermint, we can implement a very efficient tendermint light
client. The LC connects to the underlying tendermint instance to keep up with the validator set
changes as well as with recent block hashes. This part is exactly the same as in basecoin. When
a light client wants to query the state though, it uses the Web3 endpoints of the full node and
does the data verification by looking at tendermint block which contains the relevant app hash.
It checks that the block is validly signed by a majority of the current validators. Then it checks
that the information it received from web3 is valid as well and is backed by the app_hash that is
within the tendermint block.

This way we developers can write fully secure ethermint wallets that build on top of our RPC
package so that it offers exactly the same web3 endpoints that they would normally work with.
For example, you can write a phone wallet, which uses our light client package to securely
keep up with the state of the ethermint chain.

We need to write a light-client package that unifies the tendermint and web3 connections and
does the proving for you. It should expose a web3 RPC interface or C functions so that other
languages can easily build on top of it.

**TODO: Ask Frey for help on verifying the data from web3.**


Implementation
^^^^^^^^^^^^^^

The CLI library is Viper. The logger is a tmLogger.


Ethermint
"""""""""
At the top level there is the Ethermint application. An Ethermint object is instantiated within the
cmd package. It requires a reward strategy. It also takes a configuration struct with all parsed
options. Those values either come from the CLI or from the TOML file. The values on the config struct
will override the defaults. All other dependencies should be setup within the Ethermint object.
The big config struct is a nesting of smaller config structs for Reward, IBC, rpcServer and ethereum.

Ethermint:
* Config struct
* Reward strategy
* IBC strategy
* rpcServer - serves the web3 rpc server, depends on the config options
* rpcClient - sends transaction that where created over web3 to tendermint
* ethereum - is used to hold the state and execute transactions and answer questions about the state
* accounts - an account manager that manages private keys stored under this ethermint node
* logger - a tendermint logger

The Ethermint object is responsible for settinp up the ethereum object and starting the rpc server.
It implements ABCIApplication, however it proxies most requests to the ethereum object. It first
decides whether something is destined for IBC or Ethereum .
It does not implement ``Query`` for ethereum related transaction, but only to facilitate IBC
. It implements ``BaseService`` and is responsible for starting and stopping everything. It handles Info.


Accounts
""""""""
Accounts wraps a go-ethereum account manager and provides that functionality. Accounts cannot be unlocked
by default when starting ethermint as that is a security risk. They have to be unlocked through some GUI.
The RPC server can send a message to the accounts routine to ask for information or to sign a transaction.
It stores the keys the same way that go-ethereum deals with it inside the ethermint directory.


Ethereum
""""""""
The ethereum object is not exported. It handles state management/persistence and transaction processing.
It is a custom type from which we eventually will extract an interface. It handles checkTx, deliverTx
and commit. It takes a specific config struct with info such as gasprice, gaslimit and reward strategy.

Ethereum:
* stateDB for persistence and actual state
* checkTxState for ephemeral state
* logger
* reward strategy

The ethereum object is responsible for validating ethereum transactions and running them against a state.
All VM, state and state transition logic is imported from go-ethereum. It handles tendermint messages
such as BeginBlock and EndBlock. An important function is be able to respond to Commit.
Ideally, ethereum should not build its own blockchain but should rather just provide a databse layer and
leave the blockchain to tendermint. However it seems that in the current implementation of go-ethereum
the state is tightly coupled to it being a blockchain state. This logic is not too different from
what we currently have.
The ethereum object implements ``BaseService`` and can be started and stopped properly.

RPC
"""""""""
This is the RPC package.
The RPC server takes an ethereum object via an interface. The ethereum object needs to be able to answer
certain questions about the current state of ethereum, such as the syncing status. It is up to
ethereum to decide how to provide that information. The RPC server also needs to be able to submit
transactions via an rpcclient that is connected to tendermint. It also implements ``BaseService``.

The RPC package sets up all the required RPC endpoints to provide web3 compatability and overrides the
ones that don't make sense. It is a wrapper around the go-ethereum RPC package.

Same RPC methods need to be public and some private because the account methods might leave an account
unlocked and that should never be accessible to the public.

Possibly the RPC server should have a channel to communicate with the ethereum object.


IBC Strategy
""""""""""""
Ethermint decides where to route a transaction. If it is an ethereum transaction it routes it to the
ethereum object. If it is an IBC transaction it routes it to the IBCStrategy. IBCStrategy
understands how to deal with such a transaction. It can invoke transaction either directly on ethereum
or over an in-proc rpc over web3. It can also query the ethereum state over web3. It is probably
favourable to stick to a connection over web3 or through an ethereum interface. IBC should not depend
on the internals of ethereum. It is passed in by the user.
Receiving an IBC packet will work by intercepting the IBC packet, decoding it according to some rules
and creating an ethereum transaction from it that calls a special privileged smart contract.
Sending an IBC packet should be triggered by the web3 endpoints and involves providing a merkle proof
of some data, where the root hash matches the app hash.

Reward Strategy
"""""""""""""""
The reward strategy defines how to deal distribute rewards. If none is specified a default strategy
will be used. It holds the address that should receive the rewards (``coinbase``) and decides how
much and when that address should be rewarded. It is passed in by the user of the library.


Testing
"""""""
Every package should have close to full test coverage. Ideally we have generators that generate testcases.
For example for RPC in the tests it should spin up a live server and send it a combination of valid
and invalid requests in almost any order and the server should never crash.
For ethereum is should generate transactions and see if with any combination the object breaks. 
