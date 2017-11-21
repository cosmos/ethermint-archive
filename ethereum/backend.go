package ethereum

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"

	abciTypes "github.com/tendermint/abci/types"

	rpcClient "github.com/tendermint/tendermint/rpc/lib/client"

	emtTypes "github.com/tendermint/ethermint/types"
)

//----------------------------------------------------------------------
// Backend manages the underlying ethereum state for storage and processing,
// and maintains the connection to Tendermint for forwarding txs

// Backend handles the chain database and VM
// #stable - 0.4.0
type Backend struct {
	// backing ethereum structures
	config   *eth.Config
	ethereum *eth.Ethereum

	// txBroadcastLoop subscription
	txSub *event.TypeMuxSubscription

	// pending ...
	pending *pending

	// client for forwarding txs to Tendermint
	client rpcClient.HTTPClient
}

// NewBackend creates a new Backend
// #stable - 0.4.0
func NewBackend(ctx *node.ServiceContext, config *eth.Config,
	client rpcClient.HTTPClient) (*Backend, error) {
	p := newPending()

	// eth.New takes a ServiceContext for the EventMux, the AccountManager,
	// and some basic functions around the DataDir.
	ethereum, err := eth.New(ctx, config, p)
	if err != nil {
		return nil, err
	}

	// send special event to go-ethereum to switch homestead=true
	currentBlock := ethereum.BlockChain().CurrentBlock()
	ethereum.EventMux().Post(core.ChainHeadEvent{currentBlock}) // nolint: vet, errcheck

	// We don't need PoW/Uncle validation
	ethereum.BlockChain().SetValidator(NullBlockProcessor{})

	ethBackend := &Backend{
		ethereum: ethereum,
		pending:  p,
		client:   client,
		config:   config,
	}
	return ethBackend, nil
}

// Ethereum returns the underlying the ethereum object
// #stable
func (b *Backend) Ethereum() *eth.Ethereum {
	return b.ethereum
}

// Config returns the eth.Config
// #stable
func (b *Backend) Config() *eth.Config {
	return b.config
}

//----------------------------------------------------------------------
// Handle block processing

// DeliverTx appends a transaction to the current block
// #stable
func (b *Backend) DeliverTx(tx *ethTypes.Transaction) abciTypes.Result {
	return b.pending.deliverTx(b.ethereum.BlockChain(), b.config,
		b.ethereum.ApiBackend.ChainConfig(), tx)
}

// AccumulateRewards accumulates the rewards based on the given strategy
// #unstable
func (b *Backend) AccumulateRewards(strategy *emtTypes.Strategy) {
	b.pending.accumulateRewards(strategy)
}

// Commit finalises the current block
// #unstable
func (b *Backend) Commit(receiver common.Address) (common.Hash, error) {
	return b.pending.commit(b.ethereum, receiver)
}

// ResetWork resets the current block to a fresh object
// #unstable
func (b *Backend) ResetWork(receiver common.Address) error {
	work, err := b.pending.resetWork(b.ethereum.BlockChain(), receiver)
	b.pending.work = work
	return err
}

// UpdateHeaderWithTimeInfo uses the tendermint header to update the ethereum header
// #unstable
func (b *Backend) UpdateHeaderWithTimeInfo(tmHeader *abciTypes.Header) {
	b.pending.updateHeaderWithTimeInfo(b.ethereum.ApiBackend.ChainConfig(), tmHeader.Time,
		tmHeader.GetNumTxs())
}

// GasLimit returns the maximum gas per block
// #unstable
func (b *Backend) GasLimit() big.Int {
	return b.pending.gasLimit()
}

//----------------------------------------------------------------------
// Implements: node.Service

// APIs returns the collection of RPC services the ethereum package offers.
// #stable - 0.4.0
func (b *Backend) APIs() []rpc.API {
	apis := b.Ethereum().APIs()
	retApis := []rpc.API{}
	for _, v := range apis {
		if v.Namespace == "net" {
			v.Service = NewNetRPCService(b.config.NetworkId)
		}
		if v.Namespace == "miner" {
			continue
		}
		if _, ok := v.Service.(*eth.PublicMinerAPI); ok {
			continue
		}
		retApis = append(retApis, v)
	}
	return retApis
}

// Start implements node.Service, starting all internal goroutines needed by the
// Ethereum protocol implementation.
// #stable
func (b *Backend) Start(_ *p2p.Server) error {
	go b.txBroadcastLoop()
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Ethereum protocol.
// #stable
func (b *Backend) Stop() error {
	b.txSub.Unsubscribe()
	b.ethereum.Stop() // nolint: errcheck
	return nil
}

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
// #stable
func (b *Backend) Protocols() []p2p.Protocol {
	return nil
}

//----------------------------------------------------------------------
// We need a block processor that just ignores PoW and uncles and so on

// NullBlockProcessor does not validate anything
// #unstable
type NullBlockProcessor struct{}

// ValidateBody does not validate anything
// #unstable
func (NullBlockProcessor) ValidateBody(*ethTypes.Block) error { return nil }

// ValidateState does not validate anything
// #unstable
func (NullBlockProcessor) ValidateState(block, parent *ethTypes.Block, state *state.StateDB,
	receipts ethTypes.Receipts, usedGas *big.Int) error {
	return nil
}
