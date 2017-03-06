package app

import (
	"bytes"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"

	abciTypes "github.com/tendermint/abci/types"
	emtTypes "github.com/tendermint/ethermint/types"
	"github.com/tendermint/ethermint/ethereum"
)

// EthermintApplication implements a TMSP application
type EthermintApplication struct {
	backend *ethereum.Backend
	currentState func() (*state.StateDB, error)
	rpcClient *rpc.Client
	strategy *emtTypes.Strategy
}

func (app *EthermintApplication) Receiver() common.Address {
	if app.strategy != nil {
		return app.strategy.Receiver()
	}
	return common.Address{}
}

func (app *EthermintApplication) SetValidators(validators []*abciTypes.Validator) {
	if app.strategy != nil {
		app.strategy.SetValidators(validators)
	}
}

func (app *EthermintApplication) GetUpdatedValidators() abciTypes.ResponseEndBlock {
	if app.strategy != nil {
		return abciTypes.ResponseEndBlock{Diffs: app.strategy.GetUpdatedValidators()}
	}
	return abciTypes.ResponseEndBlock{}
}

func (app *EthermintApplication) CollectTx(tx *ethTypes.Transaction) {
	if app.strategy != nil {
		app.strategy.CollectTx(tx)
	}
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

// Info returns information about EthermintApplication to the tendermint engine
func (app *EthermintApplication) Info() abciTypes.ResponseInfo {
	blockchain := app.backend.Ethereum().BlockChain()
	currentBlock := blockchain.CurrentBlock()
	height := currentBlock.Number()
	hash := currentBlock.Hash()
	return abciTypes.ResponseInfo{
		Data:             "TMSPEthereum",
		LastBlockHeight:  height.Uint64(),
		LastBlockAppHash: hash[:],
	}
}

// SetOption sets a configuration option for EthermintApplication
func (app *EthermintApplication) SetOption(key string, value string) (log string) {
	return ""
}

// InitChain does nothing
func (app *EthermintApplication) InitChain(validators []*abciTypes.Validator) {
	glog.V(logger.Debug).Infof("InitChain")
	app.SetValidators(validators)
}

// CheckTx checks a transaction is valid but does not mutate the state
func (app *EthermintApplication) CheckTx(txBytes []byte) abciTypes.Result {
	glog.V(logger.Debug).Infof("Check tx")

	tx, err := decodeTx(txBytes)
	if err != nil {
		return abciTypes.ErrEncodingError
	}

	err = app.validateTx(tx)
	if err != nil {
		return abciTypes.ErrInternalError // TODO
	}

	return abciTypes.OK
}

func (app *EthermintApplication) validateTx(tx *ethTypes.Transaction) error {
	currentState, err := app.currentState()
	if err != nil {
		return err
	}

	var signer ethTypes.Signer = ethTypes.FrontierSigner{}
	if tx.Protected() {
		signer = ethTypes.NewEIP155Signer(tx.ChainId())
	}

	from, err := ethTypes.Sender(signer, tx)
	if err != nil {
		return core.ErrInvalidSender
	}

	// Make sure the account exist. Non existent accounts
	// haven't got funds and well therefor never pass.
	if !currentState.Exist(from) {
		return core.ErrInvalidSender
	}

	// Last but not least check for nonce errors
	if currentState.GetNonce(from) > tx.Nonce() {
		return core.ErrNonce
	}

	// Check the transaction doesn't exceed the current
	// block limit gas.
	// TODO
	/*if pool.gasLimit().Cmp(tx.Gas()) < 0 {
		return core.ErrGasLimit
	}*/

	// Transactions can't be negative. This may never happen
	// using RLP decoded transactions but may occur if you create
	// a transaction using the RPC for example.
	if tx.Value().Cmp(common.Big0) < 0 {
		return core.ErrNegativeValue
	}

	// Transactor should have enough funds to cover the costs
	// cost == V + GP * GL
	if currentState.GetBalance(from).Cmp(tx.Cost()) < 0 {
		return core.ErrInsufficientFunds
	}

	intrGas := core.IntrinsicGas(tx.Data(), tx.To() == nil, true) // homestead == true
	if tx.Gas().Cmp(intrGas) < 0 {
		return core.ErrIntrinsicGas
	}

	return nil
}

// BeginBlock starts a new Ethereum block
func (app *EthermintApplication) BeginBlock(hash []byte, tmHeader *abciTypes.Header) {
	glog.V(logger.Debug).Infof("Begin block")

	// update the eth header with the tendermint header
	app.backend.UpdateHeaderWithTimeInfo(tmHeader)
}

// DeliverTx processes a transaction in the EthermintApplication state
func (app *EthermintApplication) DeliverTx(txBytes []byte) abciTypes.Result {
	tx, err := decodeTx(txBytes)
	if err != nil {
		return abciTypes.ErrEncodingError
	}
	glog.V(logger.Debug).Infof("Got DeliverTx (tx): %v", tx)
	err = app.backend.DeliverTx(tx)
	if err != nil {
		glog.V(logger.Debug).Infof("DeliverTx error: %v", err)
		return abciTypes.ErrInternalError
	}
	app.CollectTx(tx)
	return abciTypes.OK
}

// EndBlock computes the Ethereum state root and prepares the ethereum block
func (app *EthermintApplication) EndBlock(height uint64) abciTypes.ResponseEndBlock {
	app.backend.AccumulateRewards(app.strategy)
	return app.GetUpdatedValidators()
}

// Commit returns a hash of the current state
func (app *EthermintApplication) Commit() abciTypes.Result {
	blockHash, err := app.backend.Commit(app.Receiver())
	if err != nil {
		glog.V(logger.Debug).Infof("Error getting latest ethereum state: %v", err)
		return abciTypes.ErrInternalError
	}
	return abciTypes.NewResultOK(blockHash[:], "")
}

type jsonRequest struct {
	Method  string          `json:"method"`
	Id      json.RawMessage `json:"id,omitempty"`
	Params []interface{}	`json:"params,omitempty"`
}

// Query queries the state of EthermintApplication
func (app *EthermintApplication) Query(query []byte) abciTypes.Result {
	var in jsonRequest
	if err := json.Unmarshal(query, &in); err != nil {
		return abciTypes.ErrInternalError
	}
	var result interface{}
	if err := app.rpcClient.Call(&result, in.Method, in.Params...); err != nil {
		return abciTypes.NewError(abciTypes.ErrInternalError.Code, err.Error())
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		return abciTypes.ErrInternalError
	}
	return abciTypes.NewResultOK(bytes, "")
}

func decodeTx(txBytes []byte) (*ethTypes.Transaction, error) {
	tx := new(ethTypes.Transaction)
	rlpStream := rlp.NewStream(bytes.NewBuffer(txBytes), 0)
	if err := tx.DecodeRLP(rlpStream); err != nil {
		return nil, err
	}
	return tx, nil
}
