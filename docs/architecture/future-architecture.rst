.. _future-architecture:

Architecture
============

Motivation
----------

The current version of ethermint works and works well. It can already sustain more than 200 tx/s, offers
immediate finality and uses proof-of-stake instead of proof-of-work. However it is currently designed
to fit around go-ethereum which was never meant to be used as a library. Due to this the current
implementation is clunky and does not feel like the rest of the tendermint ecosystem. This starts with
the CLI and ends with the logging facilities. The problem with using go-ethereum stems from our choice
to use the highest level objects possible. Those objects where never intended to be consumed as a library
and hence are awkward to work with. However, the lower level packages such as RPC and State, were and are
used as libraries.

The goal of the following architecture is to use the lower level components from go-ethereum and tendermint
to build a unique and coherent application that can also be used as a library.

Goals
^^^^^

* a user can easily connect to the testnets, the live networks as well as run a private network locally

* usability as a library - a developer can import the top level ethermint objects and quickly assemble a web3 compatible, EVM enabled and PoS based cryptocurrency that optionally supports IBC

* full compatibility with all existing web3 tooling while enabling us to change endpoints based on the context of ethermint and to add new methods for IBC

* developers can use ethermint as a library to develop beautiful light clients that take advantage of all the benefits of tendermint while also interacting with ethermint and smart contracts securely

*NOTE: Ethermint supports a superset of Web3, since we are extending it with methods for IBC.*

Implementation Process
----------------------

This is not a rewrite from scratch. It is designed to be a second version that fixes some issues from the
past. A lot of the ideas and current code can be used and packages will be replaced one by one.

1. Implement the new CLI and logging in order to provide a similar experience as Tendermint.

2. Implement the RPC package in order to provide a better user experience.

3. Implement the accounts and rewards in order.

4. Implement the ethereum object that reworks the internals.

5. Implement ibc.

6. Implement light.


The reason that we live ethereum towards the middle is that we cannot change the ethereum object without
having the previous packages.

User Experience
---------------

The first entry point is installing ethermint. This is as easy as ``brew install ethermint``. It
installs the binaries for ethermint and tendermint in one step. Other platforms and package
managers such as apt-get or choco are also supported.

Supported commands
^^^^^^^^^^^^^^^^^^

* ``ethermint version`` - prints the version of ethermint and tendermint and is used to verify the installation

* ``ethermint`` - initialises and runs ethermint and tendermint. It is used to connect to the live ethermint network.

  * the initialisation files are included in the binary
  * a reasonable home directory is assumed
  * flags can be used to configure options such as RPC

* ``ethermint testnet`` - initialises and runs ethermint and tendermint. It is used to connect to the test ethermint network.

  * the bullet points from above apply
  * flags can be used to configure which testnet to connect to

* ``ethermint development`` - initialises and runs ethermint and tendermint. It is used to setup a local network.

  * the bullet points from above apply
  * private keys are included for the default account
  * flags can be used to work with multiple local networks

Supported flags
^^^^^^^^^^^^^^^

* ``--home`` - defines the data directory for all files

* ``--gasprice`` - sets to the minimal gasprice for a validating node to include a transaction. This flag is not consensus relevant and only applies to validating nodes.

* ``--coinbase`` - defines the ethereum address which can receive the transaction fees and block rewards depending on the chosen reward strategy. This flag is not consensus relevant and only applies to validating nodes.

* ``--gasfloor`` - sets the minimum number of gas that has to be expended in every block. This prevents empty blocks. This flag is not consensus relevant and only applies to validating nodes.

* ``--gaslimit`` - sets the maximum amount of gas that can be in one ethereum block. This flag is consensus relevant and needs to be uniform in the same network.

* all flags that Tendermint exposes

Ethereum Developer Experience
-----------------------------

A developer can easily switch from using Ethereum to using Ethermint for all his smart contracting
needs. Switching between the two platform is as easy as changing one RPC endpoint. We support all
modern Ethereum tooling such as Truffle as well as all UIs such as Mist.

A developer has access to a superset of Web3. Our implementation also offers the ability to use IBC.

Go Developer Experience
-----------------------

A developer chooses Ethermint as a library to build their own PoS backed, EVM enabled cryptocurrency.
He can choose his own reward strategy to distribute transaction fees and block rewards as well as his
own IBC strategy to define how his network interacts with the Cosmos hub.

When a developer uses the Ethermint library he does not have to redefine all the commands and flags
and can rather just create a default CLI object that has to standard commands and flags already set.
He can then modify those and add more in order to fit his own needs.

Every public API will have properly formatted ``godocs`` which enable new developers to easily use
Ethermint as a library.

The documentation contains a subsection for developers that are using Ethermint as a library in Go.

Architecture Design
-------------------

Ethermint is developed by the fine folks of the Cosmos team. As such it feels like any of our other
applications, such as Tendermint Core or the cosmos-sdk. It provides the functionality of the EVM
and Web3 and pairs it with PoS and IBC. The code base and architecture follows the Tendermint way
instead of the go-ethereum way.

High Level Overview
^^^^^^^^^^^^^^^^^^^

We start by describing the high level packages that Ethermint has. The all live under
``github.com/tendermint/ethermint/``.

* cmd - does not export anything. It only pulls in other packages to setup the ethermint node.

* cli - bundles all commands and flags to provide a cli interface for an ethermint node.

* ethermint - the highest level package. It implements ABCI, coordinates the starting and shutting down of a node and wires together all the independent components.

* rpc - contains all RPC endpoints. It re-exposes a lot of the go-ethereum RPC endpoints, but also adds our own whenever necessary, such as for syncing. It does not have some endpoints such as mining but also adds new ones for IBC.

  * leans heavily on ``github.com/ethereum/go-ethereum/rpc``

* account - provides key management and key storage. It also provides the code to use hardware wallets.

* reward - implements different types of strategies to reward validators.

* ibc - provides the functionality to handle IBC packets.

* light - bundles all functionality (also by re-exporting) to write secure ethermint light clients for mobile phones

  * exposes a C API in order to be as language agnostic as possible

* logging - unifies the logging for go-ethereum and tendermint.

Low Level Detail
^^^^^^^^^^^^^^^^

This section provides a package level description of the architecture. It, where applicable, also
includes description of the actual APIs.

cmd
"""

The cmd package does not define new types, functions or methods. It is a user of the ethermint library.
It integrates all other packages into a coherent application.

cli
"""

The cli package holds all the commands and flags. It allows a developer to create a new cli without
having to write his own flags. It exposes the features described above.

ethermint
"""""""""

The ethermint package is responsible for wiring up all the different components into a working
application. It is the top level object. It takes a configuration struct, a logger, a reward object,
and an ibc object. Everything else is instantiated within the constructor. The goal is to keep the
outwards facing API very simple in the beginning. The ethermint object implements ``ABCIApplication``
and ``BaseService``. It is responsible for mediating the communication with Tendermint Core and is
the entry point that starts all needed goroutines.

Internally it instantiates an RPC server that serves the Web3 endpoints. It also instantiates an RPC
client that the RPC server uses to forward transaction from Web3 to Tendermint Core. It instantiates
an ethereum object which handles state management and execution. Lastly, it instantiates an account
object which handles private keys either on disk or on an HSM.

Ethermint:

* Logger
* Reward
* IBC

* rpcServer - serves the web3 rpc server, depends on the config options
* rpcClient - sends transaction that where created over web3 to tendermint
* ethereum - manages and executes state
* account - an account manager that manages private keys stored under this ethermint node

To start an ethermint instance Start() has to be called. This in turn starts all other goroutines,
such as for the rpcServer, the rpcClient and the account.

ethereum
""""""""

The ethereum package resides in the internal package. It is responsible for all state mangement and
execution.

Ethereum:

* state for persistence and actual state
* checkState for ephemeral state
* logger
* reward strategy

The ethereum object is responsible for validating ethereum transactions and running them against a state.
All VM, state and state transition logic is imported from go-ethereum. It handles tendermint messages
such as BeginBlock and EndBlock. An important function is be able to respond to Commit.
Ideally, ethereum should not build its own blockchain but should rather just provide a database layer and
leave the blockchain to tendermint. However it seems that in the current implementation of go-ethereum
the state is tightly coupled to it being a blockchain state. This logic is not too different from
what we currently have.

The ethereum object implements ``BaseService`` and can be started and stopped properly.

rpc
"""

The rpc package has an RPC server as its main type. It is a wrapper around the RPC package from
go-ethereum. It provides the same functionality as ethereum, but adapts it to fit the needs of
ethermint.

The RPC server implements ``BaseService`` which is an interface provided by tendermint.

It takes a reference to an ethereum object in order to answer questions about the state.
It takes a reference to a rpcClient object in order to proxy ethereum transaction to Tendermint core.
It takes a reference to an account object in order to provide functionality related to accounts.

account
"""""""

Accounts wraps a go-ethereum account manager and provides that functionality. Accounts cannot be unlocked
by default when starting ethermint as that is a security risk. They have to be unlocked through some GUI.
It stores the keys the same way that go-ethereum deals with it inside the ethermint directory.

reward
""""""

The reward strategy defines how to distribute rewards. It holds the address that should receive the
rewards (``coinbase``) and decides how much and when that address should be rewarded. It is passed in by
the user of the library.

ibc
"""

See `inter-blockchain-communication <./inter-blockchain-communication.html>`__ for details on IBC.

light
"""""

Since we are implementing our own RPC package (which wraps go-ethereum RPC) to expose the correct
web3 endpoints that are needed for ethermint, we can implement a very efficient tendermint light
client. The LC connects to the underlying tendermint instance to keep up with the validator set
changes as well as with recent block hashes. This part is exactly the same as in basecoin. When
a light client wants to query the state though, it uses the Web3 endpoints of the full node and
does the data verification by looking at tendermint block which contains the relevant app hash.
It checks that the block is validly signed by a majority of the current validators. Then it checks
that the information it received from web3 is valid as well and is backed by the app_hash that is
within the tendermint block.

This way developers can write fully secure ethermint wallets that build on top of our RPC
package so that it offers exactly the same web3 endpoints that they would normally work with.
For example, you can write a phone wallet, which uses our light client package to securely
keep up with the state of the ethermint chain.

We need to write a light-client package that unifies the tendermint and web3 connections and
does the proving for you. It should expose a web3 RPC interface or C functions so that other
languages can easily build on top of it.

logging
"""""""

Implemented like the current logging package.

Tests
^^^^^

Every file has an associated test file that verifies the assumptions and invariants that are implicit
to the program and are not expressed by the type system.

Every package has an associated test suite that uses the public API like an ordinary developer would.
This package not only ensures that the exposed API is reasonable, but it also ensures that the
package works in its entirety.

The entire application has tests at the top level in order to ensure that all components work together
as expected.

Integration tests for all RPC endpoints are run against a live network that is setup with docker
containers.

Dependencies
^^^^^^^^^^^^

Dependencies are well encapsulated and do not span multiple packages.
