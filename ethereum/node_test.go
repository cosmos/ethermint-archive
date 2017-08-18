package ethereum_test

import (
	"flag"
	"testing"

	"github.com/ethereum/go-ethereum/node"
	"github.com/stretchr/testify/assert"
	"gopkg.in/urfave/cli.v1"

	"github.com/tendermint/ethermint/ethereum"
)

var dummyApp = &cli.App{
	Name:   "test",
	Author: "Tendermint",
}

var dummyContext = cli.NewContext(dummyApp, flag.NewFlagSet("test", flag.ContinueOnError), nil)
var dummyNode, _ = node.New(&node.DefaultConfig)

func TestNewNodeConfig(t *testing.T) {
	defer func() {
		err := recover()
		assert.Nil(t, err, "expecting no panics")
	}()

	ncfg := ethereum.NewNodeConfig(dummyContext)
	assert.NotNil(t, ncfg, "expecting a non-nil config")
}

func TestNewEthConfig(t *testing.T) {
	defer func() {
		err := recover()
		assert.Nil(t, err, "expecting no panics")
	}()

	ecfg := ethereum.NewEthConfig(dummyContext, dummyNode)
	assert.NotNil(t, ecfg, "expecting a non-nil config")
}
