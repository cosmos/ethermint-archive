package ethereum

import (
	"gopkg.in/urfave/cli.v1"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
)

// NewNodeConfig for p2p and network layer
func NewNodeConfig(ctx *cli.Context) *node.Config {
	var nodeConfig *node.Config
	ethUtils.SetNodeConfig(ctx, nodeConfig)

	return nodeConfig
}

// NewEthConfig for the ethereum services
func NewEthConfig(ctx *cli.Context, stack *node.Node) *eth.Config {
	var ethConfig *eth.Config
	ethUtils.SetEthConfig(ctx, stack, ethConfig)

	return ethConfig
}
