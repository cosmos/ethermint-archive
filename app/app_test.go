package app_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"

	abciTypes "github.com/tendermint/abci/types"
	"github.com/tendermint/ethermint/app"
	"github.com/tendermint/ethermint/ethereum"
	core_types "github.com/tendermint/tendermint/rpc/core/types"
)

// implements backend.Client, used for rpc calls to tendermint
// TODO: consider using gomock or something
type MockClient struct {
	Broadcast_tx_sync_called bool
}

func (mockclient *MockClient) Call(method string, params map[string]interface{}, result interface{}) (interface{}, error) {
	tmresult := result.(*core_types.TMResult)
	switch method {
	case "status":
		*tmresult = &core_types.ResultStatus{}
		return tmresult, nil
	case "broadcast_tx_sync":
		mockclient.Broadcast_tx_sync_called = true
		*tmresult = &core_types.ResultBroadcastTx{}
		return tmresult, nil
	}
	return tmresult, abciTypes.ErrInternalError
}

func TestBumpingNonces(t *testing.T) {
	// setup test addresses and other useful stuff
	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)
	receiver_addr := common.StringToAddress("0x1234123412341234123412341234123412341234")
	signer := types.HomesteadSigner{}
	ctx := context.Background()

	// setup temp data dir and the app instance
	tempDatadir, err := ioutil.TempDir("", "ethermint_test")
	if err != nil {
		t.Error("unable to create temporary datadir")
	}
	defer os.RemoveAll(tempDatadir)

	// used to intercept rpc calls to tendermint
	mockclient := &MockClient{}

	backend, app, err := makeTestApp(tempDatadir, addr, mockclient)
	if err != nil {
		t.Errorf("Error making test EthermintApplication: %v", err)
	}

	// first transaction is sent via ABCI by us pretending to be Tendermint, should pass
	height := uint64(1)
	nonce1 := uint64(0)
	tx1, err := types.SignTx(
		types.NewTransaction(nonce1, receiver_addr, big.NewInt(10), big.NewInt(21000), big.NewInt(10), nil),
		signer,
		key,
	)
	encodedtx, err := rlp.EncodeToBytes(tx1)

	assert.Equal(t, app.CheckTx(encodedtx), abciTypes.OK)
	app.BeginBlock([]byte{}, &abciTypes.Header{Height: height})
	assert.Equal(t, app.DeliverTx(encodedtx), abciTypes.OK)
	app.EndBlock(height)
	app.Commit()

	// replays should fail - we're checking if the transaction got through earlier, by replaying the nonce
	assert.Equal(t, app.CheckTx(encodedtx), abciTypes.ErrInternalError)
	// ...on both interfaces of the app
	assert.Equal(t, backend.Ethereum().ApiBackend.SendTx(ctx, tx1), errors.New("Nonce too low"))

	// second transaction is sent via geth RPC, or at least pretending to be so
	// with a correct nonce this time, it should pass
	nonce2 := uint64(1)
	tx2, err := types.SignTx(
		types.NewTransaction(nonce2, receiver_addr, big.NewInt(10), big.NewInt(21000), big.NewInt(10), nil),
		signer,
		key,
	)

	assert.Equal(t, backend.Ethereum().ApiBackend.SendTx(ctx, tx2), nil)

	start := time.Now()
	for !mockclient.Broadcast_tx_sync_called {
		time.Sleep(200 * time.Millisecond)
		if time.Since(start) > 5*time.Second {
			assert.Fail(t, "Timeout waiting for transaction on the tendermint rpc")
			break
		}
	}
}

// mimics abciEthereumAction from cmd/ethermint/main.go
func makeTestApp(tempDatadir string, addr common.Address, mockclient *MockClient) (*ethereum.Backend, *app.EthermintApplication, error) {
	stack, err := makeTestSystemNode(tempDatadir, addr, mockclient)
	if err != nil {
		return nil, nil, err
	}
	utils.StartNode(stack)

	var backend *ethereum.Backend
	if err = stack.Service(&backend); err != nil {
		return nil, nil, err
	}

	app, err := app.NewEthermintApplication(backend, nil, nil)

	return backend, app, err
}

func makeTestGenesis(addr common.Address) (string, error) {
	gopath := os.Getenv("GOPATH")
	genesis, err := ioutil.ReadFile(filepath.Join(gopath, "src/github.com/tendermint/ethermint/dev/genesis.json"))
	if err != nil {
		return "", err
	}

	var genesisdict map[string]interface{}
	err = json.Unmarshal(genesis, &genesisdict)
	if err != nil {
		return "", err
	}

	// most important: update the alloc section to alloc to the test address
	genesisdict["alloc"] = map[string]map[string]string{addr.Hex(): {"balance": "10000000000000000000000000000000000"}}

	genesis, err = json.Marshal(genesisdict)
	if err != nil {
		return "", err
	}
	return string(genesis), nil
}

// mimics MakeSystemNode from ethereum/node.go
func makeTestSystemNode(tempDatadir string, addr common.Address, mockclient *MockClient) (*node.Node, error) {
	// Configure the node's service container
	stackConf := &node.Config{
		DataDir:     tempDatadir,
		NoDial:      true,
		NoDiscovery: true,
	}

	genesis, err := makeTestGenesis(addr)
	if err != nil {
		return nil, err
	}
	// Configure the Ethereum service
	ethConf := &eth.Config{
		Genesis: genesis,
		ChainConfig: &params.ChainConfig{
			ChainId: big.NewInt(0),
		},
		GasPrice:       common.String2Big("1000"),
		GpoMinGasPrice: common.String2Big("100"),
		GpoMaxGasPrice: common.String2Big("10000"),
		PowFake:        true,
	}

	// Assemble and return the protocol stack
	stack, err := node.New(stackConf)
	if err != nil {
		return nil, err
	}
	return stack, stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return ethereum.NewBackend(ctx, ethConf, mockclient)
	})
}
