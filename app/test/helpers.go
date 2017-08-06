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
	"github.com/tendermint/ethermint/ethereum/geth"
	"github.com/tendermint/ethermint/strategies"
	"github.com/tendermint/ethermint/utils"

	"github.com/tendermint/tmlibs/log"
)

var (
	receiverAddress = common.StringToAddress("0x1234123412341234123412341234123412341234")
)

// mimics abciEthereumAction from cmd/ethermint/main.go
func makeTestApp(tempDatadir string, addresses []common.Address, mockclient *utils.MockClient, strategy strategies.Strategy, logger log.Logger) (*node.Node, *geth.Backend, *app.EthermintApplication, error) {
	stack, err := makeTestSystemNode(tempDatadir, addresses, mockclient)
	if err != nil {
		return nil, nil, nil, err
	}
	ethUtils.StartNode(stack)

	var backend *geth.Backend
	if err = stack.Service(&backend); err != nil {
		return nil, nil, nil, err
	}

	app, err := app.NewEthermintApplication(backend, nil, strategy, logger)

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
func makeTestSystemNode(tempDatadir string, addresses []common.Address, mockclient *utils.MockClient) (*node.Node, error) {
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
		return geth.NewBackend(ctx, &ethConf, mockclient)
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
