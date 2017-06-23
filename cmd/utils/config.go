package utils

import (
	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
	cli "gopkg.in/urfave/cli.v1"
	"math/big"

	"github.com/tendermint/ethermint/ethereum"

	rpcClient "github.com/tendermint/tendermint/rpc/lib/client"
)

const (
	// Client identifier to advertise over the network
	clientIdentifier = "ethermint"
)

var (
	// Gas limit of the Genesis block.
	GenesisGasLimit = big.NewInt(100000000)
)

type ethstatsConfig struct {
	URL string `toml:",omitempty"`
}

type gethConfig struct {
	Eth      eth.Config
	Node     node.Config
	Ethstats ethstatsConfig
}

// MakeFullNode creates a full go-ethereum node
func MakeFullNode(ctx *cli.Context) *node.Node {
	stack, cfg := makeConfigNode(ctx)

	tendermintLAddr := ctx.GlobalString(TendermintAddrFlag.Name)
	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return ethereum.NewBackend(ctx, &cfg.Eth, rpcClient.NewURIClient(tendermintLAddr))
	}); err != nil {
		ethUtils.Fatalf("Failed to register the ABCI application service: %v", err)
	}

	return stack
}

func makeConfigNode(ctx *cli.Context) (*node.Node, gethConfig) {
	cfg := gethConfig{
		Eth:  eth.DefaultConfig,
		Node: DefaultNodeConfig(),
	}

	ethUtils.SetNodeConfig(ctx, &cfg.Node)
	SetEthermintNodeConfig(&cfg.Node)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		ethUtils.Fatalf("Failed to create the protocol stack: %v", err)
	}

	ethUtils.SetEthConfig(ctx, stack, &cfg.Eth)
	SetEthermintEthConfig(&cfg.Eth)

	return stack, cfg
}

// DefaultNodeConfig returns the default configuration for a go-ethereum node
func DefaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.HTTPModules = append(cfg.HTTPModules, "eth")
	cfg.WSModules = append(cfg.WSModules, "eth")
	cfg.IPCPath = "geth.ipc"
	return cfg
}

// SetEthermintNodeConfig takes a node configuration and applies ethermint specific configuration
func SetEthermintNodeConfig(cfg *node.Config) {
	cfg.P2P.MaxPeers = 0
	cfg.P2P.NoDiscovery = true
}

// SetEthermintEthConfig takes a ethereum configuration and applies ethermint specific configuration
func SetEthermintEthConfig(cfg *eth.Config) {
	cfg.MaxPeers = 0
	cfg.PowFake = true
}
