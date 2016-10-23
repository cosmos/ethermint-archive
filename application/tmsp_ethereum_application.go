package application

import (
	"bytes"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/core"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/tendermint/tmsp/types"
	"math/big"
	"sync"

	"github.com/kobigurk/tmsp-ethereum/backend"
)

// TMSPEthereumApplication implements a TMSP application
type TMSPEthereumApplication struct {
	backend           *backend.TMSPEthereumBackend
	commitMutex       *sync.Mutex
	currentHeader     *types.Header
	currentBlockHash  []byte
	currentBlockError error
	currentTxPool     *core.TxPool
}

// NewTMSPEthereumApplication creates the tmsp application for tmsp-ethereum
func NewTMSPEthereumApplication(backend *backend.TMSPEthereumBackend) *TMSPEthereumApplication {
	app := &TMSPEthereumApplication{
		backend:     backend,
		commitMutex: &sync.Mutex{},
	}
	app.currentTxPool = app.createNewTxPool()
	return app
}

// Info returns information about TMSPEthereumApplication to the tendermint engine
func (app *TMSPEthereumApplication) Info() (string, *types.TMSPInfo, *types.LastBlockInfo, *types.ConfigInfo) {
	return "TMSPEthereum", nil, nil, nil
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
	txpool := app.backend.Ethereum().TxPool()
	txpool.SetLocal(tx)
	if err := txpool.Add(tx); err != nil {
		return types.ErrInternalError
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

	app.currentTxPool = app.createNewTxPool()
	return types.NewResultOK(hash[:], "")
}

func (app *TMSPEthereumApplication) createNewTxPool() *core.TxPool {
	return core.NewTxPool(app.backend.Config().ChainConfig, app.backend.Ethereum().EventMux(), app.backend.Ethereum().BlockChain().State, app.backend.Ethereum().BlockChain().GasLimit)
}

// Query queries the state of TMSPEthereumApplication
func (app *TMSPEthereumApplication) Query(query []byte) types.Result {
	return types.OK
}

func decodeTx(txBytes []byte) (*ethTypes.Transaction, error) {
	tx := new(ethTypes.Transaction)
	rlpStream := rlp.NewStream(bytes.NewBuffer(txBytes), 0)
	if err := tx.DecodeRLP(rlpStream); err != nil {
		return nil, err
	}
	return tx, nil
}

func (app *TMSPEthereumApplication) createIntermediateBlockHeader() (*ethTypes.Header, error) {
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
}

// EndBlock adds the block to chain db
func (app *TMSPEthereumApplication) EndBlock(height uint64) (diffs []*types.Validator) {

	glog.V(logger.Debug).Infof("End block")

	header, err := app.createIntermediateBlockHeader()
	if err != nil {
		app.currentBlockError = types.ErrInternalError
		return
	}
	pendingPerAddress := app.backend.Ethereum().TxPool().Pending()
	pending := []*ethTypes.Transaction{}
	for _, v := range pendingPerAddress {
		pending = append(pending, v...)
	}
	block := ethTypes.NewBlock(header, pending, nil, nil)
	blockchain := app.backend.Ethereum().BlockChain()
	state, err := blockchain.State()
	if err != nil {
		app.currentBlockError = types.ErrInternalError
		return
	}

	receipts, _, totalGasUsed, _ := blockchain.Processor().Process(block, state, app.backend.Config().ChainConfig.VmConfig)
	header.GasUsed = totalGasUsed
	header.Root = state.IntermediateRoot()
	block = ethTypes.NewBlock(header, pending, nil, receipts)
	blockHash := block.Hash()

	glog.V(logger.Debug).Infof("Writing block: %s", hex.EncodeToString(blockHash[:]))
	_, err = blockchain.InsertChain([]*ethTypes.Block{block})
	if err != nil {
		app.currentBlockError = types.ErrInternalError
		return
	}

	hashArray, err := state.Commit()
	hash := hashArray[:]
	if err != nil {
		app.currentBlockError = types.ErrInternalError
		return
	}
	glog.V(logger.Debug).Infof("Committing %s", hex.EncodeToString(hash))

	//	app.backend.Ethereum().TxPool().RemoveBatch(app.currentTransactions)
	return nil
}

// InitChain does nothing
func (app *TMSPEthereumApplication) InitChain(validators []*types.Validator) {
}
