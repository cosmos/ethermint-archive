package test

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"path/filepath"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"

	"github.com/tendermint/ethermint/app"
	emtUtils "github.com/tendermint/ethermint/cmd/utils"
	"github.com/tendermint/ethermint/ethereum"

	data "github.com/tendermint/go-wire/data"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	ttypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tmlibs/events"
	"github.com/tendermint/tmlibs/log"
)

var (
	receiverAddress = common.StringToAddress("0x1234123412341234123412341234123412341234")
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

// mimics abciEthereumAction from cmd/ethermint/main.go
func makeTestApp(tempDatadir string, addresses []common.Address, mockclient *MockClient) (*node.Node, *ethereum.Backend, *app.EthermintApplication, error) {
	stack, err := makeTestSystemNode(tempDatadir, addresses, mockclient)
	if err != nil {
		return nil, nil, nil, err
	}
	ethUtils.StartNode(stack)

	var backend *ethereum.Backend
	if err = stack.Service(&backend); err != nil {
		return nil, nil, nil, err
	}

	app, err := app.NewEthermintApplication(backend, nil, nil)

	return stack, backend, app, err
}

func makeTestGenesis(addresses []common.Address) (*core.Genesis, error) {
	gopath := os.Getenv("GOPATH")
	genesisPath := filepath.Join(gopath, "src/github.com/tendermint/ethermint/setup/genesis.json")

	file, err := os.Open(genesisPath)
	if err != nil {
		return nil, err
	}
	defer file.Close() // nolint: errcheck

	genesis := new(core.Genesis)
	if err := json.NewDecoder(file).Decode(genesis); err != nil {
		ethUtils.Fatalf("invalid genesis file: %v", err)
	}

	balance, result := new(big.Int).SetString("10000000000000000000000000000000000", 10)
	if !result {
		return nil, errors.New("BigInt convertation error")
	}

	for _, addr := range addresses {
		genesis.Alloc[addr] = core.GenesisAccount{Balance: balance}
	}

	return genesis, nil
}

// mimics MakeSystemNode from ethereum/node.go
func makeTestSystemNode(tempDatadir string, addresses []common.Address, mockclient *MockClient) (*node.Node, error) {
	// Configure the node's service container
	nodeConf := emtUtils.DefaultNodeConfig()
	emtUtils.SetEthermintNodeConfig(&nodeConf)
	nodeConf.DataDir = tempDatadir

	// Configure the Ethereum service
	ethConf := eth.DefaultConfig
	emtUtils.SetEthermintEthConfig(&ethConf)

	genesis, err := makeTestGenesis(addresses)
	if err != nil {
		return nil, err
	}

	ethConf.Genesis = genesis

	// Assemble and return the protocol stack
	stack, err := node.New(&nodeConf)
	if err != nil {
		return nil, err
	}
	return stack, stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return ethereum.NewBackend(ctx, &ethConf, mockclient)
	})
}

func createTransaction(key *ecdsa.PrivateKey, nonce uint64) (*types.Transaction, error) {
	signer := types.HomesteadSigner{}

	return types.SignTx(
		types.NewTransaction(nonce, receiverAddress, big.NewInt(10), big.NewInt(21000), big.NewInt(10),
			nil),
		signer,
		key,
	)
}
