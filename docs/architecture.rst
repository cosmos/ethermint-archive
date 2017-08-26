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
* logger - a tendermint logger

The Ethermint object is responsible for settinp up the ethereum object and starting the rpc server.
It implements ABCIApplication, however it proxies most requests to the ethereum object. It first
decides whether something is destined for IBC or Ethereum .
It does not implement ``Query`` since that's what web3 is for. It implements ``BaseService`` and is
responsible for starting and stopping everything. It handles Info.


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
