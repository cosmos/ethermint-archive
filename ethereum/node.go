package ethereum

import (
	"gopkg.in/urfave/cli.v1"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
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
