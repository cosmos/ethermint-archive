.. _inter-blockchain-communication:

IBC Support in Ethermint
========================

Goals
-----

* Ethermint can send and receive native and ERC20 tokens via IBC
* No changes to native Ethereum transaction format
* Use cosmos-sdk implementation of IBC

Design
------

* A native contract at address 0x00...000494243 (ie. "IBC")
* Contains its own independent merkle IAVL tree
* Hash of the IAVL tree is appended to ethereum block hash in return value of ABCI Commit message
* Contract has functions for:

	* registering a blockchain (stored in iavl tree)
	* updating blockchain headers (stored in iavl tree)
	* sending outgoing IBC (stored in iavl tree)
	* receiving incoming IBC (validated against data in iavl tree)

* Contract has the ability to inflate token supply and send to ethereum accounts in the trie
* Receiving:

	* serialized IBC packet received in transaction data
	* deserialize, validate against known state of chain
	* if valid, new tokens created and sent to destination address
        * NOTE ^ we need to be able to inflate both native token and ERC20 tokens!
* Sending: 

	* send function is called by some ethereum account with tokens
	* IBC packet is formed and serialized and stored in the iavl tree

Additional Notes
----------------

We want to support more general extensions to Ethermint; for instance, a native contract that handles validators.
If we reuse cosmos-sdk libs, we may want to use just one IAVL tree for all extensions.
Thus, all native contracts would be accounts in the one exstension IAVL tree, and this one trees' root would be appended to the Tendermint AppHash.
