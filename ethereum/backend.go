package ethereum

import (
	"bytes"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	abciTypes "github.com/tendermint/abci/types"
	emtTypes "github.com/tendermint/ethermint/types"
	core_types "github.com/tendermint/tendermint/rpc/core/types"
)

// used by Backend to call tendermint rpc endpoints
// TODO: replace with HttpClient https://github.com/tendermint/go-rpc/issues/8
type Client interface {
	// see tendermint/go-rpc/client/http_client.go:115 func (c *ClientURI) Call(...)
	Call(method string, params map[string]interface{}, result interface{}) (interface{}, error)
}

// Intermediate state of a block, updated with each DeliverTx and reset on Commit
type work struct {
	header *ethTypes.Header
	parent *ethTypes.Block
	state  *state.StateDB

	txIndex      int
	transactions []*ethTypes.Transaction
	receipts     ethTypes.Receipts
	allLogs      []*ethTypes.Log

	totalUsedGas *big.Int
	gp           *core.GasPool
}

type pending struct {
	commitMutex *sync.Mutex
	work        *work
}

// Backend handles the chain database and VM
type Backend struct {
	ethereum *eth.Ethereum
	pending  *pending
	client   Client
	config   *eth.Config
}

const (
	maxWaitForServerRetries = 10
)

// New creates a new Backend
func NewBackend(ctx *node.ServiceContext, config *eth.Config, client Client) (*Backend, error) {
	p := &pending{commitMutex: &sync.Mutex{}}

	ethereum, err := eth.New(ctx, config, p)
	if err != nil {
		return nil, err
	}
	ethereum.BlockChain().SetValidator(NullBlockProcessor{})
	ethBackend := &Backend{
		ethereum: ethereum,
		pending:  p,
		client:   client,
		config:   config,
		//		client: client.NewClientURI(fmt.Sprintf("http://%s", ctx.String(TendermintCoreHostFlag.Name))),
	}

	return ethBackend, nil
}

func waitForServer(s *Backend) error {
	// wait for Tendermint to open the socket and run http endpoint
	var result core_types.TMResult
	retriesCount := 0
	for result == nil {
		_, err := s.client.Call("status", map[string]interface{}{}, &result)
		if err != nil {
			glog.V(logger.Info).Infof("Waiting for tendermint endpoint to start: %s", err)
		}
		if retriesCount += 1; retriesCount >= maxWaitForServerRetries {
			return abciTypes.ErrInternalError
		}
		time.Sleep(time.Second)
	}
	return nil
}

//----------------------------------------------------------------------

// we must implement our own net service since we don't have access to `internal/ethapi`
type NetRPCService struct {
	networkVersion int
}

func (n *NetRPCService) Version() string {
	return fmt.Sprintf("%d", n.networkVersion)
}

// APIs returns the collection of RPC services the ethereum package offers.
func (s *Backend) APIs() []rpc.API {
	apis := s.Ethereum().APIs()
	retApis := []rpc.API{}
	for _, v := range apis {
		if v.Namespace == "net" {
			networkVersion := 1
			v.Service = &NetRPCService{networkVersion}
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

// Ethereum returns the underlying the ethereum object
func (s *Backend) Ethereum() *eth.Ethereum {
	return s.ethereum
}

// Config returns the eth.Config
func (s *Backend) Config() *eth.Config {
	return s.config
}

//----------------------------------------------------------------------
// Transactions sent via the go-ethereum rpc need to be routed to tendermint

// listen for txs and forward to tendermint
// TODO: some way to exit this (it runs in a go-routine)
func (s *Backend) txBroadcastLoop() {
	txSub := s.ethereum.EventMux().Subscribe(core.TxPreEvent{})

	if err := waitForServer(s); err != nil {
		// timeouted when waiting for tendermint communication failed
		glog.V(logger.Error).Infof("Failed to run tendermint HTTP endpoint, err=%s", err)
		os.Exit(1)
	}

	for obj := range txSub.Chan() {
		event := obj.Data.(core.TxPreEvent)
		if err := s.BroadcastTx(event.Tx); err != nil {
			glog.V(logger.Error).Infof("Broadcast, err=%s", err)
		}
	}
}

// BroadcastTx broadcasts a transaction to tendermint core
func (s *Backend) BroadcastTx(tx *ethTypes.Transaction) error {
	var result core_types.TMResult
	buf := new(bytes.Buffer)
	if err := tx.EncodeRLP(buf); err != nil {
		return err
	}
	params := map[string]interface{}{
		"tx": buf.Bytes(),
	}
	_, err := s.client.Call("broadcast_tx_sync", params, &result)
	return err
}

//----------------------------------------------------------------------

func (s *pending) Pending() (*ethTypes.Block, *state.StateDB) {
	s.commitMutex.Lock()
	defer s.commitMutex.Unlock()

	return ethTypes.NewBlock(
		s.work.header,
		s.work.transactions,
		nil,
		s.work.receipts,
	), s.work.state.Copy()
}

func (s *pending) PendingBlock() *ethTypes.Block {
	s.commitMutex.Lock()
	defer s.commitMutex.Unlock()

	return ethTypes.NewBlock(
		s.work.header,
		s.work.transactions,
		nil,
		s.work.receipts,
	)
}

//----------------------------------------------------------------------

func (b *Backend) DeliverTx(tx *ethTypes.Transaction) error {
	return b.pending.deliverTx(b.ethereum.BlockChain(), b.config, tx)
}

func (p *pending) deliverTx(blockchain *core.BlockChain, config *eth.Config, tx *ethTypes.Transaction) error {
	p.commitMutex.Lock()
	defer p.commitMutex.Unlock()

	blockHash := common.Hash{}
	return p.work.deliverTx(blockchain, config, blockHash, tx)
}

func (w *work) deliverTx(blockchain *core.BlockChain, config *eth.Config, blockHash common.Hash, tx *ethTypes.Transaction) error {
	w.state.StartRecord(tx.Hash(), blockHash, w.txIndex)
	receipt, _, err := core.ApplyTransaction(
		config.ChainConfig,
		blockchain,
		w.gp,
		w.state,
		w.header,
		tx,
		w.totalUsedGas,
		vm.Config{EnablePreimageRecording: config.EnablePreimageRecording},
	)
	if err != nil {
		return err
		glog.V(logger.Debug).Infof("DeliverTx error: %v", err)
		return abciTypes.ErrInternalError
	}

	logs := w.state.GetLogs(tx.Hash())

	w.txIndex += 1

	w.transactions = append(w.transactions, tx)
	w.receipts = append(w.receipts, receipt)
	w.allLogs = append(w.allLogs, logs...)

	return err
}

//----------------------------------------------------------------------

func (b *Backend) AccumulateRewards(strategy *emtTypes.Strategy) {
	b.pending.accumulateRewards(strategy)
}

func (p *pending) accumulateRewards(strategy *emtTypes.Strategy) {
	p.commitMutex.Lock()
	defer p.commitMutex.Unlock()

	p.work.accumulateRewards(strategy)
}

func (w *work) accumulateRewards(strategy *emtTypes.Strategy) {
	core.AccumulateRewards(w.state, w.header, []*ethTypes.Header{})
	w.header.GasUsed = w.totalUsedGas
}

//----------------------------------------------------------------------

func (b *Backend) Commit(receiver common.Address) (common.Hash, error) {
	return b.pending.commit(b.ethereum.BlockChain(), receiver)
}

func (p *pending) commit(blockchain *core.BlockChain, receiver common.Address) (common.Hash, error) {
	p.commitMutex.Lock()
	defer p.commitMutex.Unlock()

	blockHash, err := p.work.commit(blockchain)
	if err != nil {
		return common.Hash{}, err
	}

	work, err := p.resetWork(blockchain, receiver)
	if err != nil {
		return common.Hash{}, err
	}

	p.work = work
	return blockHash, err
}

func (w *work) commit(blockchain *core.BlockChain) (common.Hash, error) {
	// commit ethereum state and update the header
	hashArray, err := w.state.Commit(false) // XXX: ugh hardforks
	if err != nil {
		return common.Hash{}, err
	}
	w.header.Root = hashArray

	// tag logs with state root
	// NOTE: BlockHash ?
	for _, log := range w.allLogs {
		log.BlockHash = hashArray
	}

	// create block object and compute final commit hash (hash of the ethereum block)
	block := ethTypes.NewBlock(w.header, w.transactions, nil, w.receipts)
	blockHash := block.Hash()

	// save the block to disk
	glog.V(logger.Debug).Infof("Committing block with state hash %X and root hash %X", hashArray, blockHash)
	_, err = blockchain.InsertChain([]*ethTypes.Block{block})
	if err != nil {
		glog.V(logger.Debug).Infof("Error inserting ethereum block in chain: %v", err)
		return common.Hash{}, err
	}
	return blockHash, err
}

//----------------------------------------------------------------------

func (b *Backend) ResetWork(receiver common.Address) error {
	work, err := b.pending.resetWork(b.ethereum.BlockChain(), receiver)
	b.pending.work = work
	return err
}

func (p *pending) resetWork(blockchain *core.BlockChain, receiver common.Address) (*work, error) {
	state, err := blockchain.State()
	if err != nil {
		return nil, err
	}

	currentBlock := blockchain.CurrentBlock()
	ethHeader := newBlockHeader(receiver, currentBlock)

	return &work{
		header:       ethHeader,
		parent:       currentBlock,
		state:        state,
		txIndex:      0,
		totalUsedGas: big.NewInt(0),
		gp:           new(core.GasPool).AddGas(ethHeader.GasLimit),
	}, nil
}

//----------------------------------------------------------------------

func (b *Backend) UpdateHeaderWithTimeInfo(tmHeader *abciTypes.Header) {
	b.pending.updateHeaderWithTimeInfo(b.Config().ChainConfig, tmHeader.Time)
}

func (p *pending) updateHeaderWithTimeInfo(config *params.ChainConfig, parentTime uint64) {
	p.commitMutex.Lock()
	defer p.commitMutex.Unlock()

	p.work.updateHeaderWithTimeInfo(config, parentTime)
}

func (w *work) updateHeaderWithTimeInfo(config *params.ChainConfig, parentTime uint64) {
	lastBlock := w.parent
	w.header.Time = new(big.Int).SetUint64(parentTime)
	w.header.Difficulty = core.CalcDifficulty(config, parentTime,
		lastBlock.Time().Uint64(), lastBlock.Number(), lastBlock.Difficulty())
}

//----------------------------------------------------------------------

func newBlockHeader(receiver common.Address, prevBlock *ethTypes.Block) *ethTypes.Header {
	return &ethTypes.Header{
		Number:     prevBlock.Number().Add(prevBlock.Number(), big.NewInt(1)),
		ParentHash: prevBlock.Hash(),
		GasLimit:   core.CalcGasLimit(prevBlock),
		Coinbase:   receiver,
	}
}
