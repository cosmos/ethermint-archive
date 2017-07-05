package app

import (
	abciTypes "github.com/tendermint/abci/types"
)

type ImportApp struct {
}

func NewImportApp() *ImportApp {
	return &ImportApp{}
}

func (app *ImportApp) Info() abciTypes.ResponseInfo {
	return abciTypes.ResponseInfo{}
}

func (app *ImportApp) SetOption(key string, value string) (log string) {
	return ""
}

// InitChain initializes the validator set
func (app *ImportApp) InitChain(validators []*abciTypes.Validator) {
}

// CheckTx checks a transaction is valid but does not mutate the state
func (app *ImportApp) CheckTx(txBytes []byte) abciTypes.Result {
	return abciTypes.OK
}

func (app *ImportApp) DeliverTx(txBytes []byte) abciTypes.Result {
	return abciTypes.OK
}

// BeginBlock starts a new Ethereum block
func (app *ImportApp) BeginBlock(hash []byte, tmHeader *abciTypes.Header) {
}

// EndBlock accumulates rewards for the validators and updates them
func (app *ImportApp) EndBlock(height uint64) abciTypes.ResponseEndBlock {
	return abciTypes.ResponseEndBlock{}
}

// Commit commits the block and returns a hash of the current state
func (app *ImportApp) Commit() abciTypes.Result {
	return abciTypes.OK
}
