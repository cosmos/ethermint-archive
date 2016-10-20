package application

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/tendermint/tmsp/types"

	"github.com/kobigurk/tmsp-ethereum/backend"
)

// TMSPEthereumApplication implements a TMSP application
type TMSPEthereumApplication struct {
	backend             *backend.TMSPEthereumBackend
	currentTransactions []*ethTypes.Transaction
	currentState        *state.StateDB
	commitMutex         *sync.Mutex
}

// NewTMSPEthereumApplication creates the tmsp application for tmsp-ethereum
func NewTMSPEthereumApplication(backend *backend.TMSPEthereumBackend) *TMSPEthereumApplication {
	state, err := backend.Ethereum().BlockChain().State()
	if err != nil {
		panic(err)
	}
	return &TMSPEthereumApplication{
		backend:             backend,
		currentTransactions: []*ethTypes.Transaction{},
		commitMutex:         &sync.Mutex{},
		currentState:        state,
	}
}

// Info returns information about TMSPEthereumApplication to the tendermint engine
func (app *TMSPEthereumApplication) Info() string {
	return "TMSPEthereum"
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
	app.currentTransactions = append(app.currentTransactions, tx)
	return types.OK
}

// CheckTx checks a transaction is valid but does not mutate the state
func (app *TMSPEthereumApplication) CheckTx(txBytes []byte) types.Result {
	glog.V(logger.Debug).Infof("Check tx")
	totalUsedGas := big.NewInt(0)

	tx, err := decodeTx(txBytes)
	if err != nil {
		return types.ErrEncodingError
	}

	statedb := app.currentState
	blockchain := app.backend.Ethereum().BlockChain()
	header, err := app.createIntermediateBlockHeader()
	if err != nil {
		return types.ErrInternalError
	}
	chainConfig := app.backend.Config().ChainConfig
	_, _, _, err = core.ApplyTransaction(chainConfig, blockchain, new(core.GasPool).AddGas(header.GasLimit), statedb, header, tx, totalUsedGas, chainConfig.VmConfig)
	if err != nil {
		return types.ErrInternalError
	}
	return types.OK
}

// Commit returns a hash of the current state
func (app *TMSPEthereumApplication) Commit() types.Result {
	app.commitMutex.Lock()
	defer app.commitMutex.Unlock()

	header, err := app.createIntermediateBlockHeader()
	if err != nil {
		return types.ErrInternalError
	}
	block := ethTypes.NewBlock(header, app.currentTransactions, nil, nil)
	blockchain := app.backend.Ethereum().BlockChain()
	state, err := blockchain.State()
	if err != nil {
		return types.ErrInternalError
	}
	receipts, _, totalGasUsed, _ := blockchain.Processor().Process(block, state, app.backend.Config().ChainConfig.VmConfig)
	header.GasUsed = totalGasUsed
	header.Root = state.IntermediateRoot()
	block = ethTypes.NewBlock(header, app.currentTransactions, nil, receipts)
	blockHash := block.Hash()

	glog.V(logger.Debug).Infof("Writing block: %s", hex.EncodeToString(blockHash[:]))
	_, err = blockchain.InsertChain([]*ethTypes.Block{block})
	if err != nil {
		glog.V(logger.Error).Infof("Writing block, err: %s", err)
		return types.ErrInternalError
	}

	hashArray, err := state.Commit()
	hash := hashArray[:]
	if err != nil {
		return types.ErrInternalError
	}
	glog.V(logger.Debug).Infof("Committing %s", hex.EncodeToString(hash))

	app.backend.Ethereum().TxPool().RemoveBatch(app.currentTransactions)
	app.currentTransactions = []*ethTypes.Transaction{}
	app.currentState = state
	return types.OK
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
	tstart := time.Now()
	tstamp := tstart.Unix()
	header := &ethTypes.Header{
		Number:     currentBlock.Number().Add(currentBlock.Number(), big.NewInt(1)),
		ParentHash: currentBlock.Hash(),
		GasLimit:   core.CalcGasLimit(currentBlock),
		Difficulty: core.CalcDifficulty(app.backend.Config().ChainConfig, uint64(tstamp), currentBlock.Time().Uint64(), currentBlock.Number(), currentBlock.Difficulty()),
		Time:       big.NewInt(tstamp),
	}
	return header, nil
}
