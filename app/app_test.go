package app

import (
	"io/ioutil"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/tendermint/ethermint/ethereum"
	"github.com/tendermint/ethermint/types"

	abciTypes "github.com/tendermint/abci/types"
)

func setupTestCase(t *testing.T, addresses []common.Address) (tearDown func(t *testing.T),
	app *EthermintApplication, backend *ethereum.Backend, mockClient *types.MockClient) {
	t.Log("Setup test case")

	// Setup the temporary directory for a test case
	temporaryDirectory, err := ioutil.TempDir("", "ethermint_test")
	if err != nil {
		t.Errorf("Unable to create the temporary directory for the tests: %v", err)
	}

	// Setup the app and backend for a test case
	mockClient = types.NewMockClient()
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

func TestStrictlyIncrementingNonces(t *testing.T) {
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
func TestBumpingNoncesWithRawTransaction(t *testing.T) {
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
	case <-mockClient.SentBroadcastTx:
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
