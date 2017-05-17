package ethereum

import (
	"gopkg.in/urfave/cli.v1"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"

	emtUtils "github.com/tendermint/ethermint/cmd/utils"
	"github.com/tendermint/go-rpc/client"
)

var clientIdentifier = "geth" // Client identifier to advertise over the network

// Config for p2p and network layer
func NewNodeConfig(ctx *cli.Context) *node.Config {
	var node_config *node.Config
	ethUtils.SetNodeConfig(ctx, node_config)

	return node_config
}

// Config for the ethereum services
// NOTE:(go-ethereum) stack.OpenDatabase could be moved off the stack
// and then we wouldnt need it as an arg
func NewEthConfig(ctx *cli.Context, stack *node.Node) *eth.Config {
	var eth_config *eth.Config
	ethUtils.SetEthConfig(ctx, stack, eth_config)

	return eth_config
}

// MakeSystemNode sets up a local node and configures the services to launch
func MakeSystemNode(name, version string, ctx *cli.Context) *node.Node {

	// Sets the target gas limit
	ethUtils.SetupNetwork(ctx)

	// Setup the node, a container for services
	// TODO: dont think we need a node.Node at all
	nodeConf := NewNodeConfig(ctx)
	stack, err := node.New(nodeConf)
	if err != nil {
		ethUtils.Fatalf("Failed to create the protocol stack: %v", err)
	}

	// Configure the eth
	ethConf := NewEthConfig(ctx, stack)

	//Remote tendermint RPC address
	tendermintURI := ctx.GlobalString(emtUtils.BroadcastTxAddrFlag.Name)

	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return NewBackend(ctx, ethConf, rpcclient.NewClientURI(tendermintURI))
	}); err != nil {
		ethUtils.Fatalf("Failed to register the ABCI application service: %v", err)
	}
	return stack
}
