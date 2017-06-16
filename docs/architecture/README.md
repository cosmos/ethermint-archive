# Current Design

There are a few pieces:

1. Setting up the ethereum stack
2. Forwarding txs received via ethereum rpc to tendermint rpc
3. Tendermint ABCI wrapper around ethereum stack


----


1. The general design is to bind enough of the ethereum flags that we can
use its highest level object, the `go-ethereum/node.Node`, which is just a generic container for services
that run a p2p protocol and expose APIs over the RPC.
We mostly follow the logic in `go-ethereum/cmd/utils`. Then we create a custom service to register on the Node, 
the `ethermint/ethereum.Backend`, which is inert on the p2p front but exposes only the APIs we want, or custom versions of them.
Many of the flags related to the node are irrelevant, since we don't use it for p2p.
But some, like for RPC, we need.
We should try to remove the dependence on the Node by setting up the Backend 
and starting the RPC servers ourselves. This should reduce the number of ethereum flags we have to wrestle with.
Need to make sure there's nothing else we need from the Node that we can't set up and manage with low overhead ourselves.

2. We need to listen for new txs received by the ethereum RPCs and forward them to
tendermint. Currently, we subscribe to the `TxPreEvent` in go-ethereum, and
forward any txs received to Tendermint's `/broadcast_tx_sync`.
One problem is the TxPreEvent is fired asynchronously by go-ethereum,
so we may receive txs out of order. We made a patch to fix this in tendermint/go-ethereum, 
but it would be good to either fix this upstream or use a more robust approach.
Technically, we don't need the ethereum TxPool, but it is started automatically 
with the Ethereum object and provides the simplest way to forward txs to tendermint.
It also buffers txs received out of order, waiting until they can be ordered properly 
(but still fires the event asynchronously!)

3. The simplest way to enable us to use all the web3 endpoints out of the box is to use the highest level
datastructures we can. So for the ABCI app, the application state consists of both the ethereum application state
and the ethereum blockchain. So even though we are building a chain of blocks of txs at the tendermint level,
for each tendermint block we also create and store an ethereum block to save in an ethereum blockchain managed
by the application.
For CheckTx, we replicate the code from `ethereum/core/tx_pool.go:validateTx`, but note we dont update any state
so this can only reliably process one tx per block right now! This is a big TODO!
For DeliverTx, we use our managed eth object, which provides a working state and env (config, header, gas) for applying transactions,
and slices for keeping track of all TXs, receipts and logs to be included in this block.

NOTE: The ethereum APIs expect to be able to get the most current "pending" block, ie. the block a miner is working on.
The equivalent for us is a block that is in the midst of being processed by the ABCI app.
Typically, the TxPreEvent is listened to by the miner, who then goes and updates its state.
Instead, we want to bypass the miner and listen for the event ourselves, forwarding it to tendermint
and centralizing the pending state in the ABCI app. This is why we needed the patch on go-ethereum which bypasses
the miner using the `Pending` interface. That said, in 1.6.0 there may be an alternative solution since they support 
a new consensus algorithm that doesnt use mining!
