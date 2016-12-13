package application

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"reflect"
	"sync"
	"unsafe"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/tendermint/ethermint/backend"
	tmspEthTypes "github.com/tendermint/ethermint/types"
	"github.com/tendermint/tmsp/types"
)

// TMSPEthereumApplication implements a TMSP application
type TMSPEthereumApplication struct {
	backend *backend.TMSPEthereumBackend

	// block processing
	commitMutex  *sync.Mutex
	header       *types.Header
	blockError   error
	blockResults *BlockResults

	// for CheckTx
	txPool *core.TxPool

	// for queries
	rpcClient rpc.Client

	// economics
	strategy *tmspEthTypes.Strategy
}

// Intermediate state of a block, updated with each AppendTx
type BlockResults struct {
	receipts     ethTypes.Receipts
	totalUsedGas *big.Int
	gp           *core.GasPool
	allLogs      vm.Logs
	err          error

	header       *ethTypes.Header
	state        *state.StateDB
	transactions []*ethTypes.Transaction
	txIndex      int
}

func (app *TMSPEthereumApplication) Backend() *backend.TMSPEthereumBackend {
	return app.backend
}

func (app *TMSPEthereumApplication) Strategy() *tmspEthTypes.Strategy {
	return app.strategy
}

// NewTMSPEthereumApplication creates the tmsp application for ethermint
func NewTMSPEthereumApplication(backend *backend.TMSPEthereumBackend,
	client rpc.Client, strategy *tmspEthTypes.Strategy) *TMSPEthereumApplication {
	app := &TMSPEthereumApplication{
		backend:     backend,
		commitMutex: &sync.Mutex{},
		rpcClient:   client,
		strategy:    strategy,
	}
	app.txPool = app.createNewTxPool()
	return app
}

// Info returns information about TMSPEthereumApplication to the tendermint engine
func (app *TMSPEthereumApplication) Info() (string, *types.TMSPInfo, *types.LastBlockInfo, *types.ConfigInfo) {
	blockchain := app.backend.Ethereum().BlockChain()
	currentBlock := blockchain.CurrentBlock()
	height := currentBlock.Number()
	hash := currentBlock.Hash()
	return "TMSPEthereum", nil, &types.LastBlockInfo{
		BlockHeight: height.Uint64(),
		AppHash:     hash[:],
	}, nil
}

// SetOption sets a configuration option for TMSPEthereumApplication
func (app *TMSPEthereumApplication) SetOption(key string, value string) (log string) {
	return ""
}

// InitChain does nothing
func (app *TMSPEthereumApplication) InitChain(validators []*types.Validator) {
	glog.V(logger.Debug).Infof("InitChain")
	if app.strategy != nil {
		app.strategy.SetValidators(validators)
	}
}

// CheckTx checks a transaction is valid but does not mutate the state
func (app *TMSPEthereumApplication) CheckTx(txBytes []byte) types.Result {
	glog.V(logger.Debug).Infof("Check tx")

	tx, err := decodeTx(txBytes)
	if err != nil {
		return types.ErrEncodingError
	}
	txpool := app.txPool
	txpool.SetLocal(tx)
	if err := txpool.Add(tx); err != nil {
		return types.ErrInternalError
	}
	return types.OK
}

// BeginBlock starts a new Ethereum block
func (app *TMSPEthereumApplication) BeginBlock(hash []byte, header *types.Header) {
	app.commitMutex.Lock()
	defer app.commitMutex.Unlock()

	glog.V(logger.Debug).Infof("Begin block")

	app.header = header

	app.resetBlockResults()
}

// AppendTx processes a transaction in the TMSPEthereumApplication state
func (app *TMSPEthereumApplication) AppendTx(txBytes []byte) types.Result {
	app.commitMutex.Lock()
	defer app.commitMutex.Unlock()

	glog.V(logger.Debug).Infof("Got AppendTx: %s", hex.EncodeToString(txBytes))
	tx, err := decodeTx(txBytes)
	if err != nil {
		return types.ErrEncodingError
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
		return types.ErrInternalError
	}

	app.blockResults.txIndex += 1

	app.blockResults.transactions = append(app.blockResults.transactions, tx)
	app.blockResults.receipts = append(app.blockResults.receipts, receipt)
	app.blockResults.allLogs = append(app.blockResults.allLogs, logs...)

	app.strategy.CollectTx(tx)
	return types.OK
}

// EndBlock adds the block to chain db
func (app *TMSPEthereumApplication) EndBlock(height uint64) (diffs []*types.Validator) {

	glog.V(logger.Debug).Infof("End block")

	core.AccumulateRewards(app.blockResults.state, app.blockResults.header, []*ethTypes.Header{})
	app.blockResults.header.GasUsed = app.blockResults.totalUsedGas
	hashArray, err := app.blockResults.state.Commit()
	hash := hashArray[:]
	if err != nil {
		app.blockError = types.ErrInternalError
		return
	}

	for _, log := range app.blockResults.allLogs {
		log.BlockHash = hashArray
	}
	glog.V(logger.Debug).Infof("Block transactions: %v", app.blockResults.transactions)
	app.blockResults.header.Root = hashArray
	block := ethTypes.NewBlock(app.blockResults.header, app.blockResults.transactions, nil, app.blockResults.receipts)
	blockHash := block.Hash()

	glog.V(logger.Debug).Infof("Writing block: %s", hex.EncodeToString(blockHash[:]))
	_, err = app.backend.Ethereum().BlockChain().InsertChain([]*ethTypes.Block{block})
	if err != nil {
		app.blockError = types.ErrInternalError
		return
	}

	app.resetBlockResults()
	glog.V(logger.Debug).Infof("Committing %s", hex.EncodeToString(hash))

	//	app.backend.Ethereum().TxPool().RemoveBatch(app.currentTransactions)
	if app.strategy != nil {
		return app.strategy.GetUpdatedValidators()
	}
	return nil
}

// Commit returns a hash of the current state
func (app *TMSPEthereumApplication) Commit() types.Result {
	app.commitMutex.Lock()
	defer app.commitMutex.Unlock()

	if app.blockError != nil {
		glog.V(logger.Debug).Infof("Commit error %v", app.blockError)
		return types.ErrInternalError
	}

	glog.V(logger.Debug).Infof("Commit")
	hash := app.backend.Ethereum().BlockChain().CurrentBlock().Hash()
	glog.V(logger.Debug).Infof("Committing %v", hash)

	// CheckTx and Commit should not run concurrently
	// so its safe to replace txpool here
	app.txPool.Stop()
	app.txPool = app.createNewTxPool()
	return types.NewResultOK(hash[:], "")
}

// Query queries the state of TMSPEthereumApplication
func (app *TMSPEthereumApplication) Query(query []byte) types.Result {
	var in rpc.JSONRequest
	if err := json.Unmarshal(query, &in); err != nil {
		return types.ErrInternalError
	}
	if err := app.rpcClient.Send(in); err != nil {
		return types.ErrInternalError
	}
	result := make(map[string]interface{})
	if err := app.rpcClient.Recv(&result); err != nil {
		return types.ErrInternalError
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		return types.ErrInternalError
	}
	return types.NewResultOK(bytes, "")
}

//----------------------------------------------------------------------------

func (app *TMSPEthereumApplication) createNewTxPool() *core.TxPool {
	eth, cfg := app.backend.Ethereum(), app.backend.Config()
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

func getStateCache(blockchain *core.BlockChain) *state.StateDB {
	pointerVal := reflect.ValueOf(blockchain)
	val := reflect.Indirect(pointerVal)
	member := val.FieldByName("stateCache")
	ptrToState := unsafe.Pointer(member.UnsafeAddr())
	realPtrToState := (**state.StateDB)(ptrToState)
	return *realPtrToState
}

func (app *TMSPEthereumApplication) resetBlockResults() error {
	ethHeader, _ := app.createIntermediateBlockHeader()
	state, _ := app.backend.Ethereum().BlockChain().State()
	app.blockResults = &BlockResults{
		totalUsedGas: big.NewInt(0),
		header:       ethHeader,
		gp:           new(core.GasPool).AddGas(ethHeader.GasLimit),
		state:        state,
		txIndex:      0,
	}
	return nil
}

func (app *TMSPEthereumApplication) createIntermediateBlockHeader() (*ethTypes.Header, error) {
	var receiver common.Address
	if app.strategy != nil {
		receiver = app.strategy.Receiver(app)
	}
	blockchain := app.backend.Ethereum().BlockChain()
	currentBlock := blockchain.CurrentBlock()
	var tstamp big.Int
	tstamp.SetUint64(app.header.Time)
	header := &ethTypes.Header{
		Number:     currentBlock.Number().Add(currentBlock.Number(), big.NewInt(1)),
		ParentHash: currentBlock.Hash(),
		GasLimit:   core.CalcGasLimit(currentBlock),
		Difficulty: core.CalcDifficulty(app.backend.Config().ChainConfig, app.header.Time, currentBlock.Time().Uint64(), currentBlock.Number(), currentBlock.Difficulty()),
		Time:       &tstamp,
		Coinbase:   receiver,
	}
	return header, nil
}
