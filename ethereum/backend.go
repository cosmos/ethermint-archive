package ethereum

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"

	abciTypes "github.com/tendermint/abci/types"
	emtTypes "github.com/tendermint/ethermint/types"
	rpcClient "github.com/tendermint/tendermint/rpc/lib/client"
)

//----------------------------------------------------------------------
// Backend manages the underlying ethereum state for storage and processing,
// and maintains the connection to Tendermint for forwarding txs

// Backend handles the chain database and VM
type Backend struct {
	// backing ethereum structures
	config   *eth.Config
	ethereum *eth.Ethereum

	// pending ...
	pending *pending

	// client for forwarding txs to tendermint
	client rpcClient.HTTPClient
}

// NewBackend creates a new Backend
func NewBackend(ctx *node.ServiceContext, config *eth.Config, client rpcClient.HTTPClient) (*Backend, error) {
	p := newPending()

	// eth.New takes a ServiceContext for the EventMux, the AccountManager,
	// and some basic functions around the DataDir.
	ethereum, err := eth.New(ctx, config, p)
	if err != nil {
		return nil, err
	}

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
func (b *Backend) Ethereum() *eth.Ethereum {
	return s.ethereum
}

// Config returns the eth.Config
func (b *Backend) Config() *eth.Config {
	return s.config
}

//----------------------------------------------------------------------
// Handle block processing

func (b *Backend) DeliverTx(tx *ethTypes.Transaction) error {
	return b.pending.deliverTx(b.ethereum.BlockChain(), b.config, b.ethereum.ApiBackend.ChainConfig(), tx)
}

func (b *Backend) AccumulateRewards(strategy *emtTypes.Strategy) {
	b.pending.accumulateRewards(strategy)
}

func (b *Backend) Commit(receiver common.Address) (common.Hash, error) {
	return b.pending.commit(b.ethereum.BlockChain(), receiver)
}

func (b *Backend) ResetWork(receiver common.Address) error {
	work, err := b.pending.resetWork(b.ethereum.BlockChain(), receiver)
	b.pending.work = work
	return err
}

func (b *Backend) UpdateHeaderWithTimeInfo(tmHeader *abciTypes.Header) {
	b.pending.updateHeaderWithTimeInfo(b.ethereum.ApiBackend.ChainConfig(), tmHeader.Time)
}

// GasLimit returns the maximum gas per block
func (b *Backend) GasLimit() big.Int {
	return b.pending.gasLimit()
}

//----------------------------------------------------------------------
// Implements: node.Service

// APIs returns the collection of RPC services the ethereum package offers.
func (s *Backend) APIs() []rpc.API {
	apis := s.Ethereum().APIs()
	retApis := []rpc.API{}
	for _, v := range apis {
		if v.Namespace == "net" {
			v.Service = NewNetRPCService(s.config.NetworkId)
		}
		if v.Namespace == "miner" {
			continue
		}
		if _, ok := v.Service.(*eth.PublicMinerAPI); ok {
			continue
		}
		retApis = append(retApis, v)
	}
	go s.txBroadcastLoop()
	return retApis
}

// Start implements node.Service, starting all internal goroutines needed by the
// Ethereum protocol implementation.
func (s *Backend) Start(srvr *p2p.Server) error {
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Ethereum protocol.
func (s *Backend) Stop() error {
	s.ethereum.Stop()
	return nil
}

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Backend) Protocols() []p2p.Protocol {
	return nil
}

//----------------------------------------------------------------------
// We need a block processor that just ignores PoW and uncles and so on

// NullBlockProcessor does not validate anything
type NullBlockProcessor struct{}

// ValidateBody does not validate anything
func (NullBlockProcessor) ValidateBody(*ethTypes.Block) error { return nil }

// ValidateState does not validate anything
func (NullBlockProcessor) ValidateState(block, parent *ethTypes.Block, state *state.StateDB, receipts ethTypes.Receipts, usedGas *big.Int) error {
	return nil
}
