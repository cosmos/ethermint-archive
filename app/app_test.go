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
	tmLog "github.com/tendermint/tmlibs/log"
)

var (
	receiverAddress = common.StringToAddress("0x1234123412341234123412341234123412341234")
)

func TestIncrementingNonces(t *testing.T) {
	privateKey, address := generateKeyPair(t)
	teardownTestCase, app, _, _ := setupTestCase(t, []common.Address{address})
	defer teardownTestCase(t)

	height := uint64(1)
	nonceOne := uint64(0)
	nonceTwo := uint64(1)
	nonceThree := uint64(2)

	encodedTransactionOne := createSignedTransaction(t, privateKey, nonceOne,
		receiverAddress, big.NewInt(10), big.NewInt(21000), big.NewInt(10), nil)
	encodedTransactionTwo := createSignedTransaction(t, privateKey, nonceTwo,
		receiverAddress, big.NewInt(10), big.NewInt(21000), big.NewInt(10), nil)
	encodedTransactionThree := createSignedTransaction(t, privateKey, nonceThree,
		receiverAddress, big.NewInt(10), big.NewInt(21000), big.NewInt(10), nil)

	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionOne))
	assert.Equal(t, abciTypes.ErrBadNonce, app.CheckTx(encodedTransactionThree))
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionTwo))

	app.BeginBlock([]byte{}, &abciTypes.Header{Height: height, Time: 1})

	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionOne))
	assert.Equal(t, abciTypes.ErrInternalError.Code, app.DeliverTx(encodedTransactionThree).Code)
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionTwo))

	app.EndBlock(height)

	assert.Equal(t, abciTypes.OK.Code, app.Commit().Code)
}

// first transaction is sent via ABCI by us pretending to be Tendermint, should pass
func TestBumpingNonces(t *testing.T) {
	ctx := context.Background()
	privateKey, address := generateKeyPair(t)
	teardownTestCase, app, backend, mockClient := setupTestCase(t, []common.Address{address})
	defer teardownTestCase(t)

	height := uint64(1)
	nonceOne := uint64(0)
	nonceTwo := uint64(1)

	rawTransactionOne := createRawTransaction(t, privateKey, nonceOne,
		receiverAddress, big.NewInt(10), big.NewInt(21000), big.NewInt(10), nil)
	encodedTransactionOne, err := rlp.EncodeToBytes(rawTransactionOne)
	if err != nil {
		t.Errorf("Error encoding the transaction: %v", err)
	}
	// second transaction is sent via geth RPC, or at least pretending to be so
	// with a correct nonce this time, it should pass
	rawTransactionTwo := createRawTransaction(t, privateKey, nonceTwo,
		receiverAddress, big.NewInt(10), big.NewInt(21000), big.NewInt(10), nil)

	// check transaction
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionOne))

	// set time greater than time of prev tx (zero)
	app.BeginBlock([]byte{}, &abciTypes.Header{Height: height, Time: 1})

	// check deliverTx
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionOne))

	app.EndBlock(height)

	// check commit
	assert.Equal(t, abciTypes.OK.Code, app.Commit().Code)

	// replays should fail - we're checking if the transaction got through earlier, by replaying the nonce
	assert.Equal(t, abciTypes.ErrBadNonce.Code, app.CheckTx(encodedTransactionOne).Code)

	// ...on both interfaces of the app
	assert.Equal(t, core.ErrNonceTooLow, backend.Ethereum().ApiBackend.SendTx(ctx, rawTransactionOne))

	assert.Equal(t, nil, backend.Ethereum().ApiBackend.SendTx(ctx, rawTransactionTwo))

	ticker := time.NewTicker(5 * time.Second)
	select {
	case <-ticker.C:
		assert.Fail(t, "Timeout waiting for transaction on the tendermint rpc")
	case <-mockClient.sentBroadcastTx:
	}
}

// TestMultipleTxOneAcc sends multiple TXs from the same account in the same block
func TestMultipleTxOneAcc(t *testing.T) {
	privateKeyOne, address := generateKeyPair(t)
	teardownTestCase, app, _, _ := setupTestCase(t, []common.Address{address})
	defer teardownTestCase(t)

	// first transaction is sent via ABCI by us pretending to be Tendermint, should pass
	height := uint64(1)
	nonceOne := uint64(0)
	nonceTwo := uint64(1)

	encodedTransactionOne := createSignedTransaction(t, privateKeyOne, nonceOne,
		receiverAddress, big.NewInt(10), big.NewInt(21000), big.NewInt(10), nil)
	encodedTransactionTwo := createSignedTransaction(t, privateKeyOne, nonceTwo,
		receiverAddress, big.NewInt(10), big.NewInt(21000), big.NewInt(10), nil)

	// check transaction
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionOne))

	//check tx on 2nd tx should pass until we implement state in CheckTx
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionTwo))

	// set time greater than time of prev tx (zero)
	app.BeginBlock([]byte{}, &abciTypes.Header{Height: height, Time: 1})

	// check deliverTx for 1st tx
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionOne))
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionTwo))

	app.EndBlock(height)

	// check commit
	assert.Equal(t, abciTypes.OK.Code, app.Commit().Code)
}

func TestMultipleTxTwoAcc(t *testing.T) {
	privateKeyOne, addressOne := generateKeyPair(t)
	privateKeyTwo, addressTwo := generateKeyPair(t)
	teardownTestCase, app, _, _ := setupTestCase(t, []common.Address{addressOne, addressTwo})
	defer teardownTestCase(t)

	// first transaction is sent via ABCI by us pretending to be Tendermint, should pass
	height := uint64(1)
	nonceOne := uint64(0)
	nonceTwo := uint64(0)

	encodedTransactionOne := createSignedTransaction(t, privateKeyOne, nonceOne,
		receiverAddress, big.NewInt(10), big.NewInt(21000), big.NewInt(10), nil)
	encodedTransactionTwo := createSignedTransaction(t, privateKeyTwo, nonceTwo,
		receiverAddress, big.NewInt(10), big.NewInt(21000), big.NewInt(10), nil)

	// check transaction
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionOne))
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionTwo))

	// set time greater than time of prev tx (zero)
	app.BeginBlock([]byte{}, &abciTypes.Header{Height: height, Time: 1})

	// check deliverTx for 1st tx
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionOne))
	// and for 2nd tx
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionTwo))

	app.EndBlock(height)

	// check commit
	assert.Equal(t, abciTypes.OK.Code, app.Commit().Code)
}

// Test transaction from Acc1 to new Acc2 and then from Acc2 to another address
// in the same block
func TestFromAccToAcc(t *testing.T) {
	privateKeyOne, addressOne := generateKeyPair(t)
	privateKeyTwo, addressTwo := generateKeyPair(t)
	teardownTestCase, app, _, _ := setupTestCase(t, []common.Address{addressOne, addressTwo})
	defer teardownTestCase(t)

	// first transaction from Acc1 to Acc2 (which is not in genesis)
	height := uint64(1)
	nonceOne := uint64(0)
	nonceTwo := uint64(0)

	encodedTransactionOne := createSignedTransaction(t, privateKeyOne, nonceOne,
		addressTwo, big.NewInt(1000000), big.NewInt(21000), big.NewInt(10), nil)
	encodedTransactionTwo := createSignedTransaction(t, privateKeyTwo, nonceTwo,
		receiverAddress, big.NewInt(2), big.NewInt(21000), big.NewInt(10), nil)

	// check transaction
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionOne))

	// check tx2
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionTwo))

	// set time greater than time of prev tx (zero)
	app.BeginBlock([]byte{}, &abciTypes.Header{Height: height, Time: 1})

	// check deliverTx for 1st tx
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionOne))

	// and for 2nd tx
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionTwo))

	app.EndBlock(height)

	// check commit
	assert.Equal(t, abciTypes.OK.Code, app.Commit().Code)
}

// 1. put Acc1 and Acc2 to genesis with some amounts (X)
// 2. transfer 10 amount from Acc1 to Acc2
// 3. in the same block transfer from Acc2 to another Acc all his amounts (X+10)
func TestFromAccToAcc2(t *testing.T) {
	privateKeyOne, addressOne := generateKeyPair(t)
	privateKeyTwo, addressTwo := generateKeyPair(t)
	teardownTestCase, app, _, _ := setupTestCase(t, []common.Address{addressOne, addressTwo})
	defer teardownTestCase(t)

	height := uint64(1)
	nonceOne := uint64(0)
	nonceTwo := uint64(0)

	encodedTransactionOne := createSignedTransaction(t, privateKeyOne, nonceOne,
		addressTwo, big.NewInt(500000), big.NewInt(21000), big.NewInt(10), nil)
	encodedTransactionTwo := createSignedTransaction(t, privateKeyTwo, nonceTwo,
		receiverAddress, big.NewInt(1000000), big.NewInt(21000), big.NewInt(10), nil)

	// check transaction
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionOne))

	// check tx2
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionTwo))

	// set time greater than time of prev tx (zero)
	app.BeginBlock([]byte{}, &abciTypes.Header{Height: height, Time: 1})

	// check deliverTx for 1st tx
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionOne))

	// and for 2nd tx
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionTwo))

	app.EndBlock(height)

	// check commit
	assert.Equal(t, abciTypes.OK.Code, app.Commit().Code)
}

// implements: tendermint.rpc.client.HTTPClient
type MockClient struct {
	sentBroadcastTx chan struct{} // fires when we call broadcast_tx_sync
}

func NewMockClient() *MockClient { return &MockClient{make(chan struct{})} }

func (mc *MockClient) Call(method string, params map[string]interface{},
	result interface{}) (interface{}, error) {
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

func setupTestCase(t *testing.T, addresses []common.Address) (tearDown func(t *testing.T),
	app *app.EthermintApplication, backend *ethereum.Backend, mockClient *MockClient) {
	t.Log("Setup test case")

	// Setup the temporary directory for a test case
	temporaryDirectory, err := ioutil.TempDir("", "ethermint_test")
	if err != nil {
		t.Errorf("Unable to create the temporary directory for the tests: %v", err)
	}

	// Setup the app and backend for a test case
	mockClient = NewMockClient()
	node, backend, app, err := makeTestApp(temporaryDirectory, addresses, mockClient)
	if err != nil {
		t.Errorf("Error making test EthermintApplication: %v", err)
	}

	tearDown = func(t *testing.T) {
		t.Log("Tearing down test case")
		os.RemoveAll(temporaryDirectory)
		node.Stop()
	}

	return
}

func generateKeyPair(t *testing.T) (*ecdsa.PrivateKey, common.Address) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Errorf("Error generating a private key: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	return privateKey, address
}

func createRawTransaction(t *testing.T, key *ecdsa.PrivateKey, nonce uint64,
	to common.Address, amount, gasLimit, gasPrice *big.Int, data []byte) *types.Transaction {

	signer := types.HomesteadSigner{}

	transaction, err := types.SignTx(
		types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data),
		signer,
		key,
	)
	if err != nil {
		t.Errorf("Error creating the transaction: %v", err)
	}

	return transaction

}

func createSignedTransaction(t *testing.T, key *ecdsa.PrivateKey, nonce uint64,
	to common.Address, amount, gasLimit, gasPrice *big.Int, data []byte) []byte {

	transaction := createRawTransaction(t, key, nonce, to, amount, gasLimit, gasPrice, data)

	encodedTransaction, err := rlp.EncodeToBytes(transaction)
	if err != nil {
		t.Errorf("Error encoding the transaction: %v", err)
	}

	return encodedTransaction
}

// mimics abciEthereumAction from cmd/ethermint/main.go
func makeTestApp(tempDatadir string, addresses []common.Address,
	mockClient *MockClient) (*node.Node, *ethereum.Backend, *app.EthermintApplication, error) {
	stack, err := makeTestSystemNode(tempDatadir, addresses, mockClient)
	if err != nil {
		return nil, nil, nil, err
	}
	ethUtils.StartNode(stack)

	var backend *ethereum.Backend
	if err = stack.Service(&backend); err != nil {
		return nil, nil, nil, err
	}

	app, err := app.NewEthermintApplication(backend, nil, nil)
	app.SetLogger(tmLog.TestingLogger())

	return stack, backend, app, err
}

// mimics MakeSystemNode from ethereum/node.go
func makeTestSystemNode(tempDatadir string, addresses []common.Address,
	mockClient *MockClient) (*node.Node, error) {
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
		return ethereum.NewBackend(ctx, &ethConf, mockClient)
	})
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
