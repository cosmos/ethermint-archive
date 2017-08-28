package app

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/tendermint/ethermint/ethereum"
	emtTypes "github.com/tendermint/ethermint/types"

	abciTypes "github.com/tendermint/abci/types"
	tmLog "github.com/tendermint/tmlibs/log"
)

// EthermintApplication implements an ABCI application
// #stable - 0.4.0
type EthermintApplication struct {
	// backend handles the ethereum state machine
	// and wrangles other services started by an ethereum node (eg. tx pool)
	backend *ethereum.Backend // backend ethereum struct

	// a closure to return the latest current state from the ethereum blockchain
	getCurrentState func() (*state.StateDB, error)

	checkTxState *state.StateDB

	// an ethereum rpc client we can forward queries to
	rpcClient *rpc.Client

	// strategy for validator compensation
	strategy *emtTypes.Strategy

	logger tmLog.Logger
}

// NewEthermintApplication creates a fully initialised instance of EthermintApplication
// #stable - 0.4.0
func NewEthermintApplication(backend *ethereum.Backend,
	client *rpc.Client, strategy *emtTypes.Strategy) (*EthermintApplication, error) {

	state, err := backend.Ethereum().BlockChain().State()
	if err != nil {
		return nil, err
	}

	app := &EthermintApplication{
		backend:         backend,
		rpcClient:       client,
		getCurrentState: backend.Ethereum().BlockChain().State,
		checkTxState:    state.Copy(),
		strategy:        strategy,
	}

	if err := app.backend.ResetWork(app.Receiver()); err != nil {
		return nil, err
	}

	return app, nil
}

// SetLogger sets the logger for the ethermint application
// #unstable
func (app *EthermintApplication) SetLogger(log tmLog.Logger) {
	app.logger = log
}

var bigZero = big.NewInt(0)

// maxTransactionSize is 32KB in order to prevent DOS attacks
const maxTransactionSize = 32768

// Info returns information about the last height and app_hash to the tendermint engine
// #stable - 0.4.0
func (app *EthermintApplication) Info() abciTypes.ResponseInfo {
	blockchain := app.backend.Ethereum().BlockChain()
	currentBlock := blockchain.CurrentBlock()
	height := currentBlock.Number()
	hash := currentBlock.Hash()

	app.logger.Debug("Info", "height", height) // nolint: errcheck

	// This check determines whether it is the first time ethermint gets started.
	// If it is the first time, then we have to respond with an empty hash, since
	// that is what tendermint expects.
	if height.Cmp(bigZero) == 0 {
		return abciTypes.ResponseInfo{
			Data:             "ABCIEthereum",
			LastBlockHeight:  height.Uint64(),
			LastBlockAppHash: []byte{},
		}
	}

	return abciTypes.ResponseInfo{
		Data:             "ABCIEthereum",
		LastBlockHeight:  height.Uint64(),
		LastBlockAppHash: hash[:],
	}
}

// SetOption sets a configuration option
// #stable - 0.4.0
func (app *EthermintApplication) SetOption(key string, value string) string {
	app.logger.Debug("SetOption", "key", key, "value", value) // nolint: errcheck
	return ""
}

// InitChain initializes the validator set
// #stable - 0.4.0
func (app *EthermintApplication) InitChain(validators []*abciTypes.Validator) {
	app.logger.Debug("InitChain") // nolint: errcheck
	app.SetValidators(validators)
}

// CheckTx checks a transaction is valid but does not mutate the state
// #stable - 0.4.0
func (app *EthermintApplication) CheckTx(txBytes []byte) abciTypes.Result {
	tx, err := decodeTx(txBytes)
	if err != nil {
		app.logger.Debug("CheckTx: Received invalid transaction", "tx", tx) // nolint: errcheck
		return abciTypes.ErrEncodingError.AppendLog(err.Error())
	}
	app.logger.Debug("CheckTx: Received valid transaction", "tx", tx) // nolint: errcheck

	return app.validateTx(tx)
}

// DeliverTx executes a transaction against the latest state
// #stable - 0.4.0
func (app *EthermintApplication) DeliverTx(txBytes []byte) abciTypes.Result {
	tx, err := decodeTx(txBytes)
	if err != nil {
		app.logger.Debug("DelivexTx: Received invalid transaction", "tx", tx, "err", err) // nolint: errcheck
		return abciTypes.ErrEncodingError.AppendLog(err.Error())
	}
	app.logger.Debug("DeliverTx: Received valid transaction", "tx", tx) // nolint: errcheck

	err = app.backend.DeliverTx(tx)
	if err != nil {
		app.logger.Error("DeliverTx: Error delivering tx to ethereum backend", "tx", tx, "err", err) // nolint: errcheck
		return abciTypes.ErrInternalError.AppendLog(err.Error())
	}
	app.CollectTx(tx)

	return abciTypes.OK
}

// BeginBlock starts a new Ethereum block
// #stable - 0.4.0
func (app *EthermintApplication) BeginBlock(hash []byte, tmHeader *abciTypes.Header) {
	app.logger.Debug("BeginBlock") // nolint: errcheck

	// update the eth header with the tendermint header
	app.backend.UpdateHeaderWithTimeInfo(tmHeader)
}

// EndBlock accumulates rewards for the validators and updates them
// #stable - 0.4.0
func (app *EthermintApplication) EndBlock(height uint64) abciTypes.ResponseEndBlock {
	app.logger.Debug("EndBlock", "height", height) // nolint: errcheck
	app.backend.AccumulateRewards(app.strategy)
	return app.GetUpdatedValidators()
}

// Commit commits the block and returns a hash of the current state
// #stable - 0.4.0
func (app *EthermintApplication) Commit() abciTypes.Result {
	app.logger.Debug("Commit") // nolint: errcheck
	blockHash, err := app.backend.Commit(app.Receiver())
	if err != nil {
		app.logger.Error("Error getting latest ethereum state", "err", err) // nolint: errcheck
		return abciTypes.ErrInternalError.AppendLog(err.Error())
	}
	state, err := app.getCurrentState()
	if err != nil {
		app.logger.Error("Error getting latest state", "err", err) // nolint: errcheck
	}

	app.checkTxState = state.Copy()
	return abciTypes.NewResultOK(blockHash[:], "")
}

// Query queries the state of the EthermintApplication
// #stable - 0.4.0
func (app *EthermintApplication) Query(query abciTypes.RequestQuery) abciTypes.ResponseQuery {
	app.logger.Debug("Query") // nolint: errcheck
	var in jsonRequest
	if err := json.Unmarshal(query.Data, &in); err != nil {
		return abciTypes.ResponseQuery{Code: abciTypes.ErrEncodingError.Code, Log: err.Error()}
	}
	var result interface{}
	if err := app.rpcClient.Call(&result, in.Method, in.Params...); err != nil {
		return abciTypes.ResponseQuery{Code: abciTypes.ErrInternalError.Code, Log: err.Error()}
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		return abciTypes.ResponseQuery{Code: abciTypes.ErrInternalError.Code, Log: err.Error()}
	}
	return abciTypes.ResponseQuery{Code: abciTypes.OK.Code, Value: bytes}
}

//-------------------------------------------------------

// validateTx checks the validity of a tx against the blockchain's current state.
// it duplicates the logic in ethereum's tx_pool
func (app *EthermintApplication) validateTx(tx *ethTypes.Transaction) abciTypes.Result {
	currentState := app.checkTxState

	var signer ethTypes.Signer = ethTypes.FrontierSigner{}
	if tx.Protected() {
		signer = ethTypes.NewEIP155Signer(tx.ChainId())
	}

	// Make sure the transaction is signed properl
	from, err := ethTypes.Sender(signer, tx)
	if err != nil {
		return abciTypes.ErrBaseInvalidSignature.
			AppendLog(core.ErrInvalidSender.Error())
	}

	// Heuristic limit, reject transactions over 32KB to prevent DOS attacks
	if tx.Size() > maxTransactionSize {
		return abciTypes.ErrInternalError.
			AppendLog(core.ErrOversizedData.Error())
	}

	// Transactions can't be negative. This may never happen using RLP decoded
	// transactions but may occur if you create a transaction using the RPC.
	if tx.Value().Sign() < 0 {
		return abciTypes.ErrBaseInvalidInput.
			AppendLog(core.ErrNegativeValue.Error())
	}

	// Make sure the account exist. Non existent accounts
	// haven't got funds and well therefor never pass.
	if !currentState.Exist(from) {
		return abciTypes.ErrBaseUnknownAddress.
			AppendLog(core.ErrInvalidSender.Error())
	}

	// Check the transaction doesn't exceed the current block limit gas.
	gasLimit := app.backend.GasLimit()
	if gasLimit.Cmp(tx.Gas()) < 0 {
		return abciTypes.ErrInternalError.
			AppendLog(core.ErrGasLimitReached.Error())
	}

	// Check if nonce is not strictly increasing
	if currentState.GetNonce(from) != tx.Nonce() {
		return abciTypes.ErrBadNonce
	}

	// Transactor should have enough funds to cover the costs
	// cost == V + GP * GL
	currentBalance := currentState.GetBalance(from)
	if currentBalance.Cmp(tx.Cost()) < 0 {
		return abciTypes.ErrInsufficientFunds.
			AppendLog(core.ErrInsufficientFunds.Error())
	}

	intrGas := core.IntrinsicGas(tx.Data(), tx.To() == nil, true) // homestead == true
	if tx.Gas().Cmp(intrGas) < 0 {
		return abciTypes.ErrBaseInsufficientFees.
			SetLog(core.ErrIntrinsicGas.Error())
	}

	// Update ether balances
	// amount + gasprice * gaslimit
	currentState.SubBalance(from, tx.Cost())
	currentState.AddBalance(*tx.To(), tx.Value())
	currentState.SetNonce(from, currentState.GetNonce(from)+1)

	return abciTypes.OK
}
