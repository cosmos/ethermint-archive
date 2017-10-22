package app

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tendermint/tmlibs/log"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/tendermint/ethermint/cmd/utils"
	"github.com/tendermint/ethermint/ethereum"
)

var (
	receiverAddress = common.StringToAddress("0x1234123412341234123412341234123412341234")
)

// implements: tendermint.rpc.client.HTTPClient
type MockClient struct {
	SentBroadcastTx chan struct{} // fires when we call broadcast_tx_sync
}

func NewMockClient() *MockClient { return &MockClient{make(chan struct{})} }

func (mc *MockClient) Call(method string, params map[string]interface{}, result interface{}) (interface{}, error) {
	_ = result
	switch method {
	case "status":
		result = &ctypes.ResultStatus{}
		return result, nil
	case "broadcast_tx_sync":
		close(mc.SentBroadcastTx)
		result = &ctypes.ResultBroadcastTx{}
		return result, nil
	}

	return nil, errors.New("Shouldn't happen.")
}

func generateKeyPair(t *testing.T) (*ecdsa.PrivateKey, common.Address) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Errorf("Error generating a private key: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	return privateKey, address
}

func createTx(t *testing.T, key *ecdsa.PrivateKey, nonce uint64,
	to common.Address, amount, gasLimit, gasPrice *big.Int, data []byte) *ethTypes.Transaction {

	signer := ethTypes.HomesteadSigner{}

	transaction, err := ethTypes.SignTx(
		ethTypes.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data),
		signer,
		key,
	)
	if err != nil {
		t.Errorf("Error creating the transaction: %v", err)
	}

	return transaction

}

func createTxBytes(t *testing.T, key *ecdsa.PrivateKey, nonce uint64,
	to common.Address, amount, gasLimit, gasPrice *big.Int, data []byte) []byte {

	transaction := createTx(t, key, nonce, to, amount, gasLimit, gasPrice, data)

	encodedTransaction, err := rlp.EncodeToBytes(transaction)
	if err != nil {
		t.Errorf("Error encoding the transaction: %v", err)
	}

	return encodedTransaction
}

// TODO: [adrian] Change node.Node to use ethereum.Node, which is our own node
// without the networking stack. This should be held off until we decide on
// the new design.

// mimics abciEthereumAction from cmd/ethermint/main.go
func makeTestApp(tempDatadir string, addresses []common.Address,
	mockClient *MockClient) (*node.Node, *ethereum.Backend, *EthermintApplication, error) {
	stack, err := makeTestSystemNode(tempDatadir, addresses, mockClient)
	if err != nil {
		return nil, nil, nil, err
	}
	ethUtils.StartNode(stack)

	var backend *ethereum.Backend
	if err = stack.Service(&backend); err != nil {
		return nil, nil, nil, err
	}

	app, err := NewEthermintApplication(backend, nil, nil)
	app.SetLogger(log.TestingLogger())

	return stack, backend, app, err
}

// mimics MakeSystemNode from ethereum/node.go
func makeTestSystemNode(tempDatadir string, addresses []common.Address,
	mockClient *MockClient) (*node.Node, error) {
	// Configure the node's service container
	nodeConf := utils.DefaultNodeConfig()
	utils.SetEthermintNodeConfig(&nodeConf)
	nodeConf.DataDir = tempDatadir

	// Configure the Ethereum service
	ethConf := eth.DefaultConfig
	utils.SetEthermintEthConfig(&ethConf)

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
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	genesisPath := filepath.Join(strings.TrimSuffix(currentDir, "/app"), "setup/genesis.json")

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
		return nil, errors.New("BigInt conversion error")
	}

	for _, addr := range addresses {
		genesis.Alloc[addr] = core.GenesisAccount{Balance: balance}
	}

	return genesis, nil
}
