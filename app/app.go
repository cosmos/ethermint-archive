package app

import (
	"bytes"
	"encoding/json"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/tendermint/ethermint/ethereum"
	emtTypes "github.com/tendermint/ethermint/types"

	tmspTypes "github.com/tendermint/tmsp/types"
)

// EthermintApplication implements a TMSP application
type EthermintApplication struct {
	backend *ethereum.Backend

	// block processing
	commitMutex  *sync.Mutex
	blockResults *BlockResults

	// for CheckTx. reset on Commit
	txPool *core.TxPool

	// for queries
	rpcClient rpc.Client

	// economics
	strategy *emtTypes.Strategy
}

// Intermediate state of a block, updated with each AppendTx
// and reset on Commit
type BlockResults struct {
	header *ethTypes.Header
	parent *ethTypes.Block
	state  *state.StateDB

	txIndex      int
	transactions []*ethTypes.Transaction
	receipts     ethTypes.Receipts
	allLogs      vm.Logs

	totalUsedGas *big.Int
	gp           *core.GasPool
}

func (app *EthermintApplication) Backend() *ethereum.Backend {
	return app.backend
}

func (app *EthermintApplication) Strategy() *emtTypes.Strategy {
	return app.strategy
}

// NewEthermintApplication creates the tmsp application for ethermint
func NewEthermintApplication(backend *ethereum.Backend,
	client rpc.Client, strategy *emtTypes.Strategy) (*EthermintApplication, error) {
	app := &EthermintApplication{
		backend:     backend,
		commitMutex: &sync.Mutex{},
		rpcClient:   client,
		txPool:      createNewTxPool(backend),
		strategy:    strategy,
	}
	state, err := app.backend.Ethereum().BlockChain().State()
	if err != nil {
		return nil, err
	}
	app.resetBlockResults(state) // init the block results
	return app, nil
}

// Info returns information about EthermintApplication to the tendermint engine
func (app *EthermintApplication) Info() (string, *tmspTypes.TMSPInfo, *tmspTypes.LastBlockInfo, *tmspTypes.ConfigInfo) {
	blockchain := app.backend.Ethereum().BlockChain()
	currentBlock := blockchain.CurrentBlock()
	height := currentBlock.Number()
	hash := currentBlock.Hash()
	return "TMSPEthereum", nil, &tmspTypes.LastBlockInfo{
		BlockHeight: height.Uint64(),
		AppHash:     hash[:],
	}, nil
}

// SetOption sets a configuration option for EthermintApplication
func (app *EthermintApplication) SetOption(key string, value string) (log string) {
	return ""
}

// InitChain does nothing
func (app *EthermintApplication) InitChain(validators []*tmspTypes.Validator) {
	glog.V(logger.Debug).Infof("InitChain")
	if app.strategy != nil {
		app.strategy.SetValidators(validators)
	}
}

// CheckTx checks a transaction is valid but does not mutate the state
func (app *EthermintApplication) CheckTx(txBytes []byte) tmspTypes.Result {
	glog.V(logger.Debug).Infof("Check tx")

	tx, err := decodeTx(txBytes)
	if err != nil {
		return tmspTypes.ErrEncodingError
	}
	txpool := app.txPool
	txpool.SetLocal(tx)
	if err := txpool.Add(tx); err != nil {
		return tmspTypes.ErrInternalError
	}
	return tmspTypes.OK
}

// BeginBlock starts a new Ethereum block
func (app *EthermintApplication) BeginBlock(hash []byte, tmHeader *tmspTypes.Header) {
	app.commitMutex.Lock()
	defer app.commitMutex.Unlock()

	glog.V(logger.Debug).Infof("Begin block")

	// update the eth header with the tendermint header
	app.updateHeaderWithTimeInfo(tmHeader)
}

// AppendTx processes a transaction in the EthermintApplication state
func (app *EthermintApplication) AppendTx(txBytes []byte) tmspTypes.Result {
	app.commitMutex.Lock()
	defer app.commitMutex.Unlock()

	glog.V(logger.Debug).Infof("Got AppendTx: %X", txBytes)
	tx, err := decodeTx(txBytes)
	if err != nil {
		return tmspTypes.ErrEncodingError
	}
	glog.V(logger.Debug).Infof("Got AppendTx (tx): %v", tx)

	blockHash := common.Hash{}
	app.blockResults.state.StartRecord(tx.Hash(), blockHash, app.blockResults.txIndex)
	receipt, logs, _, err := core.ApplyTransaction(
		app.backend.Config().ChainConfig,
		app.backend.Ethereum().BlockChain(),
		app.blockResults.gp,
		app.blockResults.state,
		app.blockResults.header,
		tx,
		app.blockResults.totalUsedGas,
		app.backend.Config().ChainConfig.VmConfig,
	)
	if err != nil {
		glog.V(logger.Debug).Infof("AppendTx error: %v", err)
		return tmspTypes.ErrInternalError
	}

	app.blockResults.txIndex += 1

	app.blockResults.transactions = append(app.blockResults.transactions, tx)
	app.blockResults.receipts = append(app.blockResults.receipts, receipt)
	app.blockResults.allLogs = append(app.blockResults.allLogs, logs...)

	if app.strategy != nil {
		app.strategy.CollectTx(tx)
	}
	return tmspTypes.OK
}

// EndBlock computes the Ethereum state root and prepares the ethereum block
func (app *EthermintApplication) EndBlock(height uint64) (diffs []*tmspTypes.Validator) {

	glog.V(logger.Debug).Infof("End block with txs: %v", app.blockResults.transactions)

	// account for gas/rewards
	core.AccumulateRewards(app.blockResults.state, app.blockResults.header, []*ethTypes.Header{})
	app.blockResults.header.GasUsed = app.blockResults.totalUsedGas

	//	app.backend.Ethereum().TxPool().RemoveBatch(app.currentTransactions)

	// return validator updates
	if app.strategy != nil {
		return app.strategy.GetUpdatedValidators()
	}
	return nil
}

// Commit returns a hash of the current state
func (app *EthermintApplication) Commit() tmspTypes.Result {
	app.commitMutex.Lock()
	defer app.commitMutex.Unlock()

	// if there were nonces bumped by incomming txs, we need to bump them in the PublicTransactionPoolAPI too
	rememberedManagedState := app.txPool.State()
	if rememberedManagedState != nil {
		app.backend.UpdateNonces(rememberedManagedState)
	}

	// commit ethereum state and update the header
	hashArray, err := app.blockResults.state.Commit()
	if err != nil {
		glog.V(logger.Debug).Infof("Error committing ethereum state trie: %v", err)
		return tmspTypes.ErrInternalError
	}
	app.blockResults.header.Root = hashArray

	// tag logs with state root
	// NOTE: BlockHash ?
	for _, log := range app.blockResults.allLogs {
		log.BlockHash = hashArray
	}

	// create block object and compute final commit hash (hash of the ethereum block)
	block := ethTypes.NewBlock(app.blockResults.header, app.blockResults.transactions, nil, app.blockResults.receipts)
	blockHash := block.Hash()

	// save the block to disk
	glog.V(logger.Debug).Infof("Committing block with state hash %X and root hash %X", hashArray, blockHash)
	_, err = app.backend.Ethereum().BlockChain().InsertChain([]*ethTypes.Block{block})
	if err != nil {
		glog.V(logger.Debug).Infof("Error inserting ethereum block in chain: %v", err)
		return tmspTypes.ErrInternalError
	}

	// reset the block results for the next block
	// with a new eth header and the latest eth state
	state, err := app.backend.Ethereum().BlockChain().State()
	if err != nil {
		glog.V(logger.Debug).Infof("Error getting latest ethereum state: %v", err)
		return tmspTypes.ErrInternalError
	}
	app.resetBlockResults(state)

	// reset the tx pool for the next block
	// (CheckTx and Commit should not run concurrently, so its safe)
	app.txPool.Stop()
	app.txPool = createNewTxPool(app.backend)

	return tmspTypes.NewResultOK(blockHash[:], "")
}

// Query queries the state of EthermintApplication
func (app *EthermintApplication) Query(query []byte) tmspTypes.Result {
	var in rpc.JSONRequest
	if err := json.Unmarshal(query, &in); err != nil {
		return tmspTypes.ErrInternalError
	}
	if err := app.rpcClient.Send(in); err != nil {
		return tmspTypes.ErrInternalError
	}
	result := make(map[string]interface{})
	if err := app.rpcClient.Recv(&result); err != nil {
		return tmspTypes.ErrInternalError
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		return tmspTypes.ErrInternalError
	}
	return tmspTypes.NewResultOK(bytes, "")
}

//----------------------------------------------------------------------------

// runs in Commit once we have the new state
func (app *EthermintApplication) resetBlockResults(state *state.StateDB) {
	var receiver common.Address
	if app.strategy != nil {
		receiver = app.strategy.Receiver()
	}
	blockchain := app.backend.Ethereum().BlockChain()
	currentBlock := blockchain.CurrentBlock()

	ethHeader := newBlockHeader(receiver, currentBlock)

	app.blockResults = &BlockResults{
		header:       ethHeader,
		parent:       currentBlock,
		state:        state,
		txIndex:      0,
		totalUsedGas: big.NewInt(0),
		gp:           new(core.GasPool).AddGas(ethHeader.GasLimit),
	}
}

// update the eth header with info from tendermint header in BeginBlock
func (app *EthermintApplication) updateHeaderWithTimeInfo(tmHeader *tmspTypes.Header) {
	config := app.backend.Config().ChainConfig
	lastBlock := app.blockResults.parent
	app.blockResults.header.Time = new(big.Int).SetUint64(tmHeader.Time)
	app.blockResults.header.Difficulty = core.CalcDifficulty(config, tmHeader.Time, lastBlock.Time().Uint64(), lastBlock.Number(), lastBlock.Difficulty())
}

//----------------------------------------------------------------------------

// create new ethereum block header
func newBlockHeader(receiver common.Address, prevBlock *ethTypes.Block) *ethTypes.Header {
	return &ethTypes.Header{
		Number:     prevBlock.Number().Add(prevBlock.Number(), big.NewInt(1)),
		ParentHash: prevBlock.Hash(),
		GasLimit:   core.CalcGasLimit(prevBlock),
		Coinbase:   receiver,
	}
}

func createNewTxPool(backend *ethereum.Backend) *core.TxPool {
	eth, cfg := backend.Ethereum(), backend.Config()
	return core.NewTxPool(cfg.ChainConfig, eth.EventMux(), eth.BlockChain().State, eth.BlockChain().GasLimit)
}

func decodeTx(txBytes []byte) (*ethTypes.Transaction, error) {
	tx := new(ethTypes.Transaction)
	rlpStream := rlp.NewStream(bytes.NewBuffer(txBytes), 0)
	if err := tx.DecodeRLP(rlpStream); err != nil {
		return nil, err
	}
	return tx, nil
}
