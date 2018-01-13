// TODO: Replace this with the mockClient in
// github.com/tendermint/tendermint/rpc/client/mock/abci.go

package types

import (
	data "github.com/tendermint/go-wire/data"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	ttypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tmlibs/events"
	"github.com/tendermint/tmlibs/log"
)

// MockClient is a mock implementation of a tendermint rpc client
type MockClient struct {
	SentBroadcastTx chan struct{} // fires when we call broadcast_tx_sync
	syncing         bool
}

// NewMockClient returns a pointer to a new non-syncing mock tendermint rpc client
func NewMockClient(syncing bool) *MockClient {
	return &MockClient{
		make(chan struct{}),
		syncing,
	}
}

// ---------------------
// ABCIClient implementation

// ABCIInfo ...
func (mc *MockClient) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	return &ctypes.ResultABCIInfo{}, nil
}

// ABCIQuery ...
func (mc *MockClient) ABCIQuery(path string, data data.Bytes,
	prove bool) (*ctypes.ResultABCIQuery, error) {

	return &ctypes.ResultABCIQuery{}, nil
}

// BroadcastTxCommit ...
func (mc *MockClient) BroadcastTxCommit(tx ttypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	return &ctypes.ResultBroadcastTxCommit{}, nil
}

// BroadcastTxAsync ...
func (mc *MockClient) BroadcastTxAsync(tx ttypes.Tx) (*ctypes.ResultBroadcastTx, error) {
	close(mc.SentBroadcastTx)

	return &ctypes.ResultBroadcastTx{}, nil
}

// BroadcastTxSync ...
func (mc *MockClient) BroadcastTxSync(tx ttypes.Tx) (*ctypes.ResultBroadcastTx, error) {
	mc.SentBroadcastTx <- struct{}{}

	return &ctypes.ResultBroadcastTx{}, nil
}

// ----------------------
// SignClient implementation

// Block ...
func (mc *MockClient) Block(height int) (*ctypes.ResultBlock, error) {
	return &ctypes.ResultBlock{}, nil
}

// Commit ...
func (mc *MockClient) Commit(height int) (*ctypes.ResultCommit, error) {
	return &ctypes.ResultCommit{}, nil
}

// Validators ...
func (mc *MockClient) Validators() (*ctypes.ResultValidators, error) {
	return &ctypes.ResultValidators{}, nil
}

// Tx ...
func (mc *MockClient) Tx(hash []byte, prove bool) (*ctypes.ResultTx, error) {
	return &ctypes.ResultTx{}, nil
}

// -------------------
// HistoryClient implementation

// Genesis ...
func (mc *MockClient) Genesis() (*ctypes.ResultGenesis, error) {
	return &ctypes.ResultGenesis{}, nil
}

// BlockchainInfo ...
func (mc *MockClient) BlockchainInfo(minHeight,
	maxHeight int) (*ctypes.ResultBlockchainInfo, error) {

	return &ctypes.ResultBlockchainInfo{}, nil
}

// -----------------------
// StatusClient implementation

// Status ...
func (mc *MockClient) Status() (*ctypes.ResultStatus, error) {
	return &ctypes.ResultStatus{Syncing: mc.syncing}, nil
}

// -----------------------
// Service implementation

// Start ...
func (mc *MockClient) Start() (bool, error) {
	return true, nil
}

// OnStart ...
func (mc *MockClient) OnStart() error {
	return nil
}

// Stop ...
func (mc *MockClient) Stop() bool {
	return true
}

// OnStop ...
func (mc *MockClient) OnStop() {
	close(mc.SentBroadcastTx)
}

// Reset ...
func (mc *MockClient) Reset() (bool, error) {
	return true, nil
}

// OnReset ...
func (mc *MockClient) OnReset() error {
	return nil
}

// IsRunning ...
func (mc *MockClient) IsRunning() bool {
	return true
}

// String ...
func (mc *MockClient) String() string {
	return "MockClient"
}

// SetLogger ...
func (mc *MockClient) SetLogger(log.Logger) {

}

// -----------------------
// types.EventSwitch implementation

// AddListenerForEvent ...
func (mc *MockClient) AddListenerForEvent(listenerID, event string, cb events.EventCallback) {
	// nop
}

// FireEvent ...
func (mc *MockClient) FireEvent(event string, data events.EventData) {
	// nop
}

// RemoveListenerForEvent ...
func (mc *MockClient) RemoveListenerForEvent(event string, listenerID string) {
	// nop
}

// RemoveListener ...
func (mc *MockClient) RemoveListener(listenerID string) {
	// nop
}
