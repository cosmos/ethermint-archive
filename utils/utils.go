package utils

import (
	data "github.com/tendermint/go-wire/data"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	ttypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tmlibs/events"
	"github.com/tendermint/tmlibs/log"
)

type MockClient struct {
	sentBroadcastTx chan struct{} // fires when we call broadcast_tx_sync
	syncing         bool
}

func NewMockClient() *MockClient {
	return &MockClient{
		make(chan struct{}),
		false,
	}
}

func NewMockSyncingClient() *MockClient {
	return &MockClient{
		make(chan struct{}),
		true,
	}
}

// ---------------------
// ABCIClient implementation

func (mc *MockClient) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	return &ctypes.ResultABCIInfo{}, nil
}

func (mc *MockClient) ABCIQuery(path string, data data.Bytes, prove bool) (*ctypes.ResultABCIQuery, error) {
	return &ctypes.ResultABCIQuery{}, nil
}

func (mc *MockClient) BroadcastTxCommit(tx ttypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	return &ctypes.ResultBroadcastTxCommit{}, nil
}

func (mc *MockClient) BroadcastTxAsync(tx ttypes.Tx) (*ctypes.ResultBroadcastTx, error) {
	return mc.BroadcastTxSync(tx)
}

func (mc *MockClient) BroadcastTxSync(tx ttypes.Tx) (*ctypes.ResultBroadcastTx, error) {
	close(mc.sentBroadcastTx)

	return &ctypes.ResultBroadcastTx{}, nil
}

// ----------------------
// SignClient implementation

func (mc *MockClient) Block(height int) (*ctypes.ResultBlock, error) {
	return &ctypes.ResultBlock{}, nil
}

func (mc *MockClient) Commit(height int) (*ctypes.ResultCommit, error) {
	return &ctypes.ResultCommit{}, nil
}

func (mc *MockClient) Validators() (*ctypes.ResultValidators, error) {
	return &ctypes.ResultValidators{}, nil
}

func (mc *MockClient) Tx(hash []byte, prove bool) (*ctypes.ResultTx, error) {
	return &ctypes.ResultTx{}, nil
}

// -------------------
// HistoryClient implementation

func (mc *MockClient) Genesis() (*ctypes.ResultGenesis, error) {
	return &ctypes.ResultGenesis{}, nil
}

func (mc *MockClient) BlockchainInfo(minHeight, maxHeight int) (*ctypes.ResultBlockchainInfo, error) {
	return &ctypes.ResultBlockchainInfo{}, nil
}

// -----------------------
// StatusClient implementation

func (mc *MockClient) Status() (*ctypes.ResultStatus, error) {
	return &ctypes.ResultStatus{Syncing: mc.syncing}, nil
}

// -----------------------
// Service implementation

func (mc *MockClient) Start() (bool, error) {
	return true, nil
}

func (mc *MockClient) OnStart() error {
	return nil
}

func (mc *MockClient) Stop() bool {
	return true
}

func (mc *MockClient) OnStop() {
	// nop
}

func (mc *MockClient) Reset() (bool, error) {
	return true, nil
}

func (mc *MockClient) OnReset() error {
	return nil
}

func (mc *MockClient) IsRunning() bool {
	return true
}

func (mc *MockClient) String() string {
	return "MockClient"
}

func (mc *MockClient) SetLogger(log.Logger) {

}

// -----------------------
// types.EventSwitch implementation

func (mc *MockClient) AddListenerForEvent(listenerID, event string, cb events.EventCallback) {
	// nop
}

func (mc *MockClient) FireEvent(event string, data events.EventData) {
	// nop
}

func (mc *MockClient) RemoveListenerForEvent(event string, listenerID string) {
	// nop
}

func (mc *MockClient) RemoveListener(listenerID string) {
	// nop
}
