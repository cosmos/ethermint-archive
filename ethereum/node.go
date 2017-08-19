package ethereum

import (
	"gopkg.in/urfave/cli.v1"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
)

// NewNodeConfig for p2p and network layer
// #unstable
func NewNodeConfig(ctx *cli.Context) *node.Config {
	nodeConfig := new(node.Config)
	ethUtils.SetNodeConfig(ctx, nodeConfig)

	return nodeConfig
}

// NewEthConfig for the ethereum services
// #unstable
func NewEthConfig(ctx *cli.Context, stack *node.Node) *eth.Config {
	ethConfig := new(eth.Config)
	ethUtils.SetEthConfig(ctx, stack, ethConfig)

	return ethConfig
}
