package app_test

import (
	"crypto/ecdsa"
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

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/tendermint/ethermint/app"
	emtUtils "github.com/tendermint/ethermint/cmd/utils"
	"github.com/tendermint/ethermint/ethereum"

	abciTypes "github.com/tendermint/abci/types"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var (
	receiverAddress = common.StringToAddress("0x1234123412341234123412341234123412341234")
)

// implements: tendermint.rpc.client.HTTPClient
type MockClient struct {
	sentBroadcastTx chan struct{} // fires when we call broadcast_tx_sync
}

func NewMockClient() *MockClient { return &MockClient{make(chan struct{})} }

func (mc *MockClient) Call(method string, params map[string]interface{}, result interface{}) (interface{}, error) {
	_ = result
	switch method {
	case "status":
		result = &ctypes.ResultStatus{}

		return result, nil
	case "broadcast_tx_sync":
		close(mc.sentBroadcastTx)
		result = &ctypes.ResultBroadcastTx{}

		return result, nil
	}

	return nil, abciTypes.ErrInternalError
}

func TestBumpingNonces(t *testing.T) {
	// generate key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Errorf("Error generating key %v", err)
	}
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	ctx := context.Background()

	// used to intercept rpc calls to tendermint
	mockclient := NewMockClient()

	// setup temp data dir and the app instance
	tempDatadir, err := ioutil.TempDir("", "ethermint_test")
	if err != nil {
		t.Error("unable to create temporary datadir")
	}
	defer os.RemoveAll(tempDatadir) // nolint: errcheck

	stack, backend, app, err := makeTestApp(tempDatadir, []common.Address{addr}, mockclient)
	if err != nil {
		t.Errorf("Error making test EthermintApplication: %v", err)
	}

	// first transaction is sent via ABCI by us pretending to be Tendermint, should pass
	height := uint64(1)
	nonce1 := uint64(0)
	tx1, err := createTransaction(privateKey, nonce1)
	if err != nil {
		t.Errorf("Error creating transaction: %v", err)

	}
	encodedtx, _ := rlp.EncodeToBytes(tx1)

	// check transaction
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedtx))

	// set time greater than time of prev tx (zero)
	app.BeginBlock([]byte{}, &abciTypes.Header{Height: height, Time: 1})

	// check deliverTx
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedtx))

	app.EndBlock(height)

	// check commit
	assert.Equal(t, abciTypes.OK.Code, app.Commit().Code)

	// replays should fail - we're checking if the transaction got through earlier, by replaying the nonce
	assert.Equal(t, abciTypes.ErrBadNonce.Code, app.CheckTx(encodedtx).Code)

	// ...on both interfaces of the app
	assert.Equal(t, core.ErrNonce, backend.Ethereum().ApiBackend.SendTx(ctx, tx1))

	// second transaction is sent via geth RPC, or at least pretending to be so
	// with a correct nonce this time, it should pass
	nonce2 := uint64(1)
	tx2, _ := createTransaction(privateKey, nonce2)

	assert.Equal(t, backend.Ethereum().ApiBackend.SendTx(ctx, tx2), nil)

	ticker := time.NewTicker(5 * time.Second)
	select {
	case <-ticker.C:
		assert.Fail(t, "Timeout waiting for transaction on the tendermint rpc")
	case <-mockclient.sentBroadcastTx:
	}

	stack.Stop() // nolint: errcheck
}

func TestMultipleTxOneAcc(t *testing.T) {
	// generate key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Errorf("Error generating key %v", err)
	}
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)

	// used to intercept rpc calls to tendermint
	mockclient := NewMockClient()

	// setup temp data dir and the app instance
	tempDatadir, err := ioutil.TempDir("", "ethermint_test")
	if err != nil {
		t.Error("unable to create temporary datadir")
	}
	defer os.RemoveAll(tempDatadir) // nolint: errcheck

	node, _, app, err := makeTestApp(tempDatadir, []common.Address{addr}, mockclient)
	if err != nil {
		t.Errorf("Error making test EthermintApplication: %v", err)
	}

	// first transaction is sent via ABCI by us pretending to be Tendermint, should pass
	height := uint64(1)

	nonce1 := uint64(0)
	tx1, err := createTransaction(privateKey, nonce1)
	if err != nil {
		t.Errorf("Error creating transaction: %v", err)

	}
	encodedTx1, _ := rlp.EncodeToBytes(tx1)

	//create 2-nd tx from the same account
	nonce2 := uint64(0)
	tx2, err := createTransaction(privateKey, nonce2)
	if err != nil {
		t.Errorf("Error creating transaction: %v", err)

	}
	encodedTx2, _ := rlp.EncodeToBytes(tx2)

	// check transaction
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTx1))

	//check tx on 2nd tx should pass until we implement state in CheckTx
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTx2))

	// set time greater than time of prev tx (zero)
	app.BeginBlock([]byte{}, &abciTypes.Header{Height: height, Time: 1})

	// check deliverTx for 1st tx
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTx1))

	// and for 2nd tx (should fail because of wrong nonce2)
	deliverTx2Result := app.DeliverTx(encodedTx2)

	assert.Equal(t, abciTypes.ErrInternalError.Code, deliverTx2Result.Code)

	app.EndBlock(height)

	// check commit
	assert.Equal(t, abciTypes.OK.Code, app.Commit().Code)

	node.Stop() // nolint: errcheck
}

func TestMultipleTxTwoAcc(t *testing.T) {
	// generate key
	//account 1
	privateKey1, err := crypto.GenerateKey()
	if err != nil {
		t.Errorf("Error generating key %v", err)
	}
	addr1 := crypto.PubkeyToAddress(privateKey1.PublicKey)

	//account 2
	privateKey2, err := crypto.GenerateKey()
	if err != nil {
		t.Errorf("Error generating key %v", err)
	}
	addr2 := crypto.PubkeyToAddress(privateKey2.PublicKey)

	// used to intercept rpc calls to tendermint
	mockclient := NewMockClient()

	// setup temp data dir and the app instance
	tempDatadir, err := ioutil.TempDir("", "ethermint_test")
	if err != nil {
		t.Error("unable to create temporary datadir")
	}
	defer os.RemoveAll(tempDatadir) // nolint: errcheck

	node, _, app, err := makeTestApp(tempDatadir, []common.Address{addr1, addr2}, mockclient)
	if err != nil {
		t.Errorf("Error making test EthermintApplication: %v", err)
	}

	// first transaction is sent via ABCI by us pretending to be Tendermint, should pass
	height := uint64(1)
	nonce1 := uint64(0)
	tx1, err := createTransaction(privateKey1, nonce1)
	if err != nil {
		t.Errorf("Error creating transaction: %v", err)

	}
	encodedtx1, _ := rlp.EncodeToBytes(tx1)

	//create 2-nd tx
	nonce2 := uint64(0)
	tx2, err := createTransaction(privateKey2, nonce2)
	if err != nil {
		t.Errorf("Error creating transaction: %v", err)

	}
	encodedTx2, _ := rlp.EncodeToBytes(tx2)

	// check transaction
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedtx1))

	// set time greater than time of prev tx (zero)
	app.BeginBlock([]byte{}, &abciTypes.Header{Height: height, Time: 1})

	// check deliverTx for 1st tx
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedtx1))
	// and for 2nd tx
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTx2))

	app.EndBlock(height)

	// check commit
	assert.Equal(t, abciTypes.OK.Code, app.Commit().Code)

	node.Stop() // nolint: errcheck
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
