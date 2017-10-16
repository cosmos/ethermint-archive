package ethereum

import (
	"flag"
	"net"
	"testing"

	"gopkg.in/urfave/cli.v1"

	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/node"
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

	ncfg := NewNodeConfig(dummyContext)
	assert.NotNil(t, ncfg, "expecting a non-nil config")
}

func TestNewEthConfig(t *testing.T) {
	defer func() {
		err := recover()
		assert.Nil(t, err, "expecting no panics")
	}()

	ecfg := NewEthConfig(dummyContext, dummyNode)
	assert.NotNil(t, ecfg, "expecting a non-nil config")
}

func TestEnsureDisabledEthereumP2PStack(t *testing.T) {
	cfg := new(node.Config)
	*cfg = node.DefaultConfig
	cfg.P2P.ListenAddr = ":34555"
	node, err := New(cfg)
	if err != nil {
		t.Fatalf("cannot initialise new node from config: %v", err)
	}

	if err := node.Start(); err != nil {
		t.Fatalf("cannot start node: %v", err)
	}
	// Make a listener and ensure that ListenAddr can be bound to
	// i.e that no other service is listening on it
	addr := cfg.P2P.ListenAddr
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("failed to bind to %q: %v", addr, err)
	}
	if err := listener.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}
