package app

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	abciTypes "github.com/tendermint/abci/types"

	"github.com/tendermint/ethermint/ethereum"
	emtTypes "github.com/tendermint/ethermint/types"
)

// EthermintApplication implements an ABCI application
type EthermintApplication struct {

	// backend handles the ethereum state machine
	// and wrangles other services started by an ethereum node (eg. tx pool)
	backend *ethereum.Backend // backend ethereum struct

	// a closure to return the latest current state from the ethereum blockchain
	currentState func() (*state.StateDB, error)

	// an ethereum rpc client we can forward queries to
	rpcClient *rpc.Client

	// strategy for validator compensation
	strategy *emtTypes.Strategy
}

// NewEthermintApplication creates the abci application for ethermint
func NewEthermintApplication(backend *ethereum.Backend,
	client *rpc.Client, strategy *emtTypes.Strategy) (*EthermintApplication, error) {
	app := &EthermintApplication{
		backend:      backend,
		rpcClient:    client,
		currentState: backend.Ethereum().BlockChain().State,
		strategy:     strategy,
	}

	err := app.backend.ResetWork(app.Receiver()) // init the block results
	return app, err
}

// Info returns information about the last height and app_hash to the tendermint engine
func (app *EthermintApplication) Info() abciTypes.ResponseInfo {
	log.Info("Info")
	blockchain := app.backend.Ethereum().BlockChain()
	currentBlock := blockchain.CurrentBlock()
	height := currentBlock.Number()
	hash := currentBlock.Hash()

	// This check determines whether it is the first time ethermint gets started.
	// If it is the first time, then we have to respond with an empty hash, since
	// that is what tendermint expects.
	if height.Cmp(big.NewInt(0)) == 0 {
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
func (app *EthermintApplication) SetOption(key string, value string) (log string) {
	//log.Info("SetOption")
	return ""
}

// InitChain initializes the validator set
func (app *EthermintApplication) InitChain(validators []*abciTypes.Validator) {
	log.Info("InitChain")
	app.SetValidators(validators)
}

// CheckTx checks a transaction is valid but does not mutate the state
func (app *EthermintApplication) CheckTx(txBytes []byte) abciTypes.Result {
	tx, err := decodeTx(txBytes)
	log.Info("Received CheckTx", "tx", tx)
	if err != nil {
		return abciTypes.ErrEncodingError.AppendLog(err.Error())
	}

	return app.validateTx(tx)
}

// DeliverTx executes a transaction against the latest state
func (app *EthermintApplication) DeliverTx(txBytes []byte) abciTypes.Result {
	tx, err := decodeTx(txBytes)
	if err != nil {
		return abciTypes.ErrEncodingError.AppendLog(err.Error())
	}

	log.Info("Got DeliverTx", "tx", tx)
	err = app.backend.DeliverTx(tx)
	if err != nil {
		log.Warn("DeliverTx error", "err", err)

		return abciTypes.ErrInternalError.AppendLog(err.Error())
	}
	app.CollectTx(tx)

	return abciTypes.OK
}

// BeginBlock starts a new Ethereum block
func (app *EthermintApplication) BeginBlock(hash []byte, tmHeader *abciTypes.Header) {
	log.Info("BeginBlock")

	// update the eth header with the tendermint header
	app.backend.UpdateHeaderWithTimeInfo(tmHeader)
}

// EndBlock accumulates rewards for the validators and updates them
func (app *EthermintApplication) EndBlock(height uint64) abciTypes.ResponseEndBlock {
	log.Info("EndBlock")
	app.backend.AccumulateRewards(app.strategy)
	return app.GetUpdatedValidators()
}

// Commit commits the block and returns a hash of the current state
func (app *EthermintApplication) Commit() abciTypes.Result {
	log.Info("Commit")
	blockHash, err := app.backend.Commit(app.Receiver())
	if err != nil {
		log.Warn("Error getting latest ethereum state", "err", err)
		return abciTypes.ErrInternalError.AppendLog(err.Error())
	}
	return abciTypes.NewResultOK(blockHash[:], "")
}

// Query queries the state of EthermintApplication
func (app *EthermintApplication) Query(query abciTypes.RequestQuery) abciTypes.ResponseQuery {
	log.Info("Query")
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
	currentState, err := app.currentState()
	if err != nil {
		return abciTypes.ErrInternalError.AppendLog(err.Error())
	}

	var signer ethTypes.Signer = ethTypes.FrontierSigner{}
	if tx.Protected() {
		signer = ethTypes.NewEIP155Signer(tx.ChainId())
	}

	from, err := ethTypes.Sender(signer, tx)
	if err != nil {
		return abciTypes.ErrBaseInvalidSignature.
			AppendLog(core.ErrInvalidSender.Error())
	}

	// Make sure the account exist. Non existent accounts
	// haven't got funds and well therefor never pass.
	if !currentState.Exist(from) {
		return abciTypes.ErrBaseUnknownAddress.
			AppendLog(core.ErrInvalidSender.Error())
	}

	// Check for nonce errors
	currentNonce := currentState.GetNonce(from)
	if currentNonce > tx.Nonce() {
		return abciTypes.ErrBadNonce.
			AppendLog(fmt.Sprintf("Got: %d, Current: %d", tx.Nonce(), currentNonce))
	}

	// Check the transaction doesn't exceed the current block limit gas.
	gasLimit := app.backend.GasLimit()
	if gasLimit.Cmp(tx.Gas()) < 0 {
		return abciTypes.ErrInternalError.AppendLog(core.ErrGasLimitReached.Error())
	}

	// Transactions can't be negative. This may never happen
	// using RLP decoded transactions but may occur if you create
	// a transaction using the RPC for example.
	if tx.Value().Cmp(common.Big0) < 0 {
		return abciTypes.ErrBaseInvalidInput.
			SetLog(core.ErrNegativeValue.Error())
	}

	// Transactor should have enough funds to cover the costs
	// cost == V + GP * GL
	currentBalance := currentState.GetBalance(from)
	if currentBalance.Cmp(tx.Cost()) < 0 {
		return abciTypes.ErrInsufficientFunds.
			AppendLog(fmt.Sprintf("Current balance: %s, tx cost: %s", currentBalance, tx.Cost()))

	}

	intrGas := core.IntrinsicGas(tx.Data(), tx.To() == nil, true) // homestead == true
	if tx.Gas().Cmp(intrGas) < 0 {
		return abciTypes.ErrBaseInsufficientFees.
			SetLog(core.ErrIntrinsicGas.Error())
	}

	return abciTypes.OK
}
