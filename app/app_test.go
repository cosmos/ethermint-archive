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

// TestStrictlyIncrementingNonces tests that nonces have to increment by 1
// instead of just being greater than the previous nonce.
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
	// expect a failure here since the nonce is not strictly increasing
	assert.Equal(t, abciTypes.ErrBadNonce, app.CheckTx(encodedTransactionThree))
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionTwo))

	app.BeginBlock([]byte{}, &abciTypes.Header{Height: height, Time: 1})

	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionOne))
	// expect a failure here since the nonce is not strictly increasing
	assert.Equal(t, abciTypes.ErrInternalError.Code, app.DeliverTx(encodedTransactionThree).Code)
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionTwo))

	app.EndBlock(height)

	assert.Equal(t, abciTypes.OK.Code, app.Commit().Code)
}

// TestBumpingNoncesWithRawTransaction sends a transaction over the RPC
// interface of Tendermint.
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

	rawTransactionTwo := createRawTransaction(t, privateKey, nonceTwo,
		receiverAddress, big.NewInt(10), big.NewInt(21000), big.NewInt(10), nil)

	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionOne))

	app.BeginBlock([]byte{}, &abciTypes.Header{Height: height, Time: 1})

	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionOne))

	app.EndBlock(height)

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

// TestMultipleTxOneAcc sends multiple txs from the same account in the same block
func TestMultipleTxOneAcc(t *testing.T) {
	privateKeyOne, address := generateKeyPair(t)
	teardownTestCase, app, _, _ := setupTestCase(t, []common.Address{address})
	defer teardownTestCase(t)

	height := uint64(1)
	nonceOne := uint64(0)
	nonceTwo := uint64(1)

	encodedTransactionOne := createSignedTransaction(t, privateKeyOne, nonceOne,
		receiverAddress, big.NewInt(10), big.NewInt(21000), big.NewInt(10), nil)
	encodedTransactionTwo := createSignedTransaction(t, privateKeyOne, nonceTwo,
		receiverAddress, big.NewInt(10), big.NewInt(21000), big.NewInt(10), nil)

	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionOne))
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionTwo))

	app.BeginBlock([]byte{}, &abciTypes.Header{Height: height, Time: 1})

	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionOne))
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionTwo))

	app.EndBlock(height)

	assert.Equal(t, abciTypes.OK.Code, app.Commit().Code)
}

// TestMultipleTxFromTwoAcc sends multiple txs from two different accounts
func TestMultipleTxFromTwoAcc(t *testing.T) {
	privateKeyOne, addressOne := generateKeyPair(t)
	privateKeyTwo, addressTwo := generateKeyPair(t)
	teardownTestCase, app, _, _ := setupTestCase(t, []common.Address{addressOne, addressTwo})
	defer teardownTestCase(t)

	height := uint64(1)
	nonceOne := uint64(0)
	nonceTwo := uint64(0)

	encodedTransactionOne := createSignedTransaction(t, privateKeyOne, nonceOne,
		receiverAddress, big.NewInt(10), big.NewInt(21000), big.NewInt(10), nil)
	encodedTransactionTwo := createSignedTransaction(t, privateKeyTwo, nonceTwo,
		receiverAddress, big.NewInt(10), big.NewInt(21000), big.NewInt(10), nil)

	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionOne))
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionTwo))

	app.BeginBlock([]byte{}, &abciTypes.Header{Height: height, Time: 1})

	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionOne))
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionTwo))

	app.EndBlock(height)

	assert.Equal(t, abciTypes.OK.Code, app.Commit().Code)
}

// TestFromAccToAcc sends a transaction from account A to account B and from
// account B to a third address
func TestFromAccToAcc(t *testing.T) {
	privateKeyOne, addressOne := generateKeyPair(t)
	privateKeyTwo, addressTwo := generateKeyPair(t)
	teardownTestCase, app, _, _ := setupTestCase(t, []common.Address{addressOne, addressTwo})
	defer teardownTestCase(t)

	height := uint64(1)
	nonceOne := uint64(0)
	nonceTwo := uint64(0)

	encodedTransactionOne := createSignedTransaction(t, privateKeyOne, nonceOne,
		addressTwo, big.NewInt(1000000), big.NewInt(21000), big.NewInt(10), nil)
	encodedTransactionTwo := createSignedTransaction(t, privateKeyTwo, nonceTwo,
		receiverAddress, big.NewInt(2), big.NewInt(21000), big.NewInt(10), nil)

	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionOne))
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionTwo))

	app.BeginBlock([]byte{}, &abciTypes.Header{Height: height, Time: 1})

	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionOne))
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionTwo))

	app.EndBlock(height)

	assert.Equal(t, abciTypes.OK.Code, app.Commit().Code)
}

// TestFromAccToAcc2 sends money from A to B and then all of B's money to C.
// This tests whether checkTx works correctly and allows multiple transactions
// per block.
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

	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionOne))
	assert.Equal(t, abciTypes.OK, app.CheckTx(encodedTransactionTwo))

	app.BeginBlock([]byte{}, &abciTypes.Header{Height: height, Time: 1})

	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionOne))
	assert.Equal(t, abciTypes.OK, app.DeliverTx(encodedTransactionTwo))

	app.EndBlock(height)

	assert.Equal(t, abciTypes.OK.Code, app.Commit().Code)
}
