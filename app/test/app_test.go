package test

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"

	abciTypes "github.com/tendermint/abci/types"
	tmLog "github.com/tendermint/tmlibs/log"

	"github.com/tendermint/ethermint/types"
)

func TestBumpingNonces(t *testing.T) {
	// generate key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Errorf("Error generating key %v", err)
	}

	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	ctx := context.Background()

	// used to intercept rpc calls to tendermint
	mockclient := types.NewMockClient()

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
	app.SetLogger(tmLog.TestingLogger())

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
	assert.Equal(t, core.ErrNonceTooLow, backend.Ethereum().ApiBackend.SendTx(ctx, tx1))

	// second transaction is sent via geth RPC, or at least pretending to be so
	// with a correct nonce this time, it should pass
	nonce2 := uint64(1)
	tx2, _ := createTransaction(privateKey, nonce2)

	assert.Equal(t, backend.Ethereum().ApiBackend.SendTx(ctx, tx2), nil)

	ticker := time.NewTicker(5 * time.Second)
	select {
	case <-ticker.C:
		assert.Fail(t, "Timeout waiting for transaction on the tendermint rpc")
	case <-mockclient.SentBroadcastTx:
	}

	stack.Stop() // nolint: errcheck
}

// TestMultipleTxOneAcc sends multiple TXs from the same account in the same block
func TestMultipleTxOneAcc(t *testing.T) {
	// generate key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Errorf("Error generating key %v", err)
	}
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)

	// used to intercept rpc calls to tendermint
	mockclient := types.NewMockClient()

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
	app.SetLogger(tmLog.TestingLogger())

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

// TestMultipleTxTwoAcc sends multiple TXs from two different accounts
func TestMultipleTxTwoAcc(t *testing.T) {
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
	mockclient := types.NewMockClient()

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
	app.SetLogger(tmLog.TestingLogger())

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
