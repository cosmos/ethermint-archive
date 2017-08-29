package ethereum

import (
	"gopkg.in/urfave/cli.v1"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
)

// Node is the main object.
type Node struct {
	node.Node
}

// New creates a new node.
func New(conf *node.Config) (*Node, error) {
	stack, err := node.New(conf)
	if err != nil {
		return nil, err
	}

	return &Node{*stack}, nil // nolint: vet
}

// Start starts base node and stop p2p server
func (n *Node) Start() error {
	// start p2p server
	err := n.Node.Start()
	if err != nil {
		return err
	}

	// stop it
	n.Node.Server().Stop()

	return nil
}

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
