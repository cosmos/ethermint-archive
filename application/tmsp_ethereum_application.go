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
	backend                       *backend.TMSPEthereumBackend
	commitMutex                   *sync.Mutex
	currentHeader                 *types.Header
	currentBlockHash              []byte
	currentBlockError             error
	currentBlockProcessingDetails *BlockProcessingDetails
	currentTxPool                 *core.TxPool
	rpcClient                     rpc.Client

	minerRewardStrategy tmspEthTypes.MinerRewardStrategy
	validatorsStrategy  tmspEthTypes.ValidatorsStrategy
}

type BlockProcessingDetails struct {
	receipts     ethTypes.Receipts
	totalUsedGas *big.Int
	err          error
	header       *ethTypes.Header
	allLogs      vm.Logs
	gp           *core.GasPool

	state        *state.StateDB
	transactions []*ethTypes.Transaction
	txIndex      int
}

// NewTMSPEthereumApplication creates the tmsp application for ethermint
func NewTMSPEthereumApplication(
	backend *backend.TMSPEthereumBackend,
	client rpc.Client,
	minerRewardStrategy tmspEthTypes.MinerRewardStrategy,
	validatorsStrategy tmspEthTypes.ValidatorsStrategy,
) *TMSPEthereumApplication {
	app := &TMSPEthereumApplication{
		backend:     backend,
		commitMutex: &sync.Mutex{},
		rpcClient:   client,
	}
	app.currentTxPool = app.createNewTxPool()
	app.minerRewardStrategy = minerRewardStrategy
	app.validatorsStrategy = validatorsStrategy
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
	app.currentBlockProcessingDetails.state.StartRecord(tx.Hash(), blockHash, app.currentBlockProcessingDetails.txIndex)
	receipt, logs, _, err := core.ApplyTransaction(
		app.backend.Config().ChainConfig,
		app.backend.Ethereum().BlockChain(),
		app.currentBlockProcessingDetails.gp,
		app.currentBlockProcessingDetails.state,
		app.currentBlockProcessingDetails.header,
		tx,
		app.currentBlockProcessingDetails.totalUsedGas,
		app.backend.Config().ChainConfig.VmConfig,
	)
	if err != nil {
		glog.V(logger.Debug).Infof("AppendTx error: %v", err)
		return types.ErrInternalError
	}

	app.currentBlockProcessingDetails.transactions = append(app.currentBlockProcessingDetails.transactions, tx)
	app.currentBlockProcessingDetails.receipts = append(app.currentBlockProcessingDetails.receipts, receipt)
	app.currentBlockProcessingDetails.allLogs = append(app.currentBlockProcessingDetails.allLogs, logs...)

	if app.validatorsStrategy != nil {
		app.validatorsStrategy.CollectTx(tx)
	}
	return types.OK
}

// CheckTx checks a transaction is valid but does not mutate the state
func (app *TMSPEthereumApplication) CheckTx(txBytes []byte) types.Result {
	glog.V(logger.Debug).Infof("Check tx")

	tx, err := decodeTx(txBytes)
	if err != nil {
		return types.ErrEncodingError
	}
	txpool := app.currentTxPool
	txpool.SetLocal(tx)
	if err := txpool.Add(tx); err != nil {
		return types.ErrInternalError
	}
	return types.OK
}

// Commit returns a hash of the current state
func (app *TMSPEthereumApplication) Commit() types.Result {
	app.commitMutex.Lock()
	defer app.commitMutex.Unlock()

	if app.currentBlockError != nil {
		glog.V(logger.Debug).Infof("Commit error %v", app.currentBlockError)
		return types.ErrInternalError
	}

	glog.V(logger.Debug).Infof("Commit")
	hash := app.backend.Ethereum().BlockChain().CurrentBlock().Hash()
	glog.V(logger.Debug).Infof("Committing %v", hash)

	app.currentTxPool.Stop()
	app.currentTxPool = app.createNewTxPool()
	return types.NewResultOK(hash[:], "")
}

func (app *TMSPEthereumApplication) createNewTxPool() *core.TxPool {
	return core.NewTxPool(app.backend.Config().ChainConfig, app.backend.Ethereum().EventMux(), app.backend.Ethereum().BlockChain().State, app.backend.Ethereum().BlockChain().GasLimit)
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

func (app *TMSPEthereumApplication) resetBlockProcessingDetails() error {
	ethHeader, _ := app.createIntermediateBlockHeader()
	state, _ := app.backend.Ethereum().BlockChain().State()
	app.currentBlockProcessingDetails = &BlockProcessingDetails{
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
	if app.minerRewardStrategy != nil {
		receiver = app.minerRewardStrategy.Receiver(app)
	}
	blockchain := app.backend.Ethereum().BlockChain()
	currentBlock := blockchain.CurrentBlock()
	var tstamp big.Int
	tstamp.SetUint64(app.currentHeader.Time)
	header := &ethTypes.Header{
		Number:     currentBlock.Number().Add(currentBlock.Number(), big.NewInt(1)),
		ParentHash: currentBlock.Hash(),
		GasLimit:   core.CalcGasLimit(currentBlock),
		Difficulty: core.CalcDifficulty(app.backend.Config().ChainConfig, app.currentHeader.Time, currentBlock.Time().Uint64(), currentBlock.Number(), currentBlock.Difficulty()),
		Time:       &tstamp,
		Coinbase:   receiver,
	}
	return header, nil
}

// BeginBlock starts a new Ethereum block
func (app *TMSPEthereumApplication) BeginBlock(hash []byte, header *types.Header) {
	app.commitMutex.Lock()
	defer app.commitMutex.Unlock()

	glog.V(logger.Debug).Infof("Begin block")

	app.currentHeader = header
	app.currentBlockHash = hash

	app.resetBlockProcessingDetails()
}

// EndBlock adds the block to chain db
func (app *TMSPEthereumApplication) EndBlock(height uint64) (diffs []*types.Validator) {

	glog.V(logger.Debug).Infof("End block")

	core.AccumulateRewards(app.currentBlockProcessingDetails.state, app.currentBlockProcessingDetails.header, []*ethTypes.Header{})
	app.currentBlockProcessingDetails.header.GasUsed = app.currentBlockProcessingDetails.totalUsedGas
	hashArray, err := app.currentBlockProcessingDetails.state.Commit()
	hash := hashArray[:]
	if err != nil {
		app.currentBlockError = types.ErrInternalError
		return
	}

	for _, log := range app.currentBlockProcessingDetails.allLogs {
		log.BlockHash = hashArray
	}
	glog.V(logger.Debug).Infof("Block transactions: %v", app.currentBlockProcessingDetails.transactions)
	app.currentBlockProcessingDetails.header.Root = hashArray
	block := ethTypes.NewBlock(app.currentBlockProcessingDetails.header, app.currentBlockProcessingDetails.transactions, nil, app.currentBlockProcessingDetails.receipts)
	blockHash := block.Hash()

	glog.V(logger.Debug).Infof("Writing block: %s", hex.EncodeToString(blockHash[:]))
	_, err = app.backend.Ethereum().BlockChain().InsertChain([]*ethTypes.Block{block})
	if err != nil {
		app.currentBlockError = types.ErrInternalError
		return
	}

	app.resetBlockProcessingDetails()
	glog.V(logger.Debug).Infof("Committing %s", hex.EncodeToString(hash))

	//	app.backend.Ethereum().TxPool().RemoveBatch(app.currentTransactions)
	if app.validatorsStrategy != nil {
		return app.validatorsStrategy.GetUpdatedValidators()
	}
	return nil
}

// InitChain does nothing
func (app *TMSPEthereumApplication) InitChain(validators []*types.Validator) {
	glog.V(logger.Debug).Infof("InitChain")
	if app.validatorsStrategy != nil {
		app.validatorsStrategy.SetValidators(validators)
	}
}

func (app *TMSPEthereumApplication) ValidatorsStrategy() tmspEthTypes.ValidatorsStrategy {
	return app.validatorsStrategy
}

func (app *TMSPEthereumApplication) MinerRewardStrategy() tmspEthTypes.MinerRewardStrategy {
	return app.minerRewardStrategy
}

func (app *TMSPEthereumApplication) Backend() *backend.TMSPEthereumBackend {
	return app.backend
}
