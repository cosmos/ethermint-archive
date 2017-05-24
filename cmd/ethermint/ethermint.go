package main

import (
	"fmt"
	"os"

	"gopkg.in/urfave/cli.v1"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"

	abciApp "github.com/tendermint/ethermint/app"
	emtUtils "github.com/tendermint/ethermint/cmd/utils"
	"github.com/tendermint/ethermint/ethereum"
	"github.com/tendermint/ethermint/version"

	"github.com/tendermint/abci/server"
	rpcClient "github.com/tendermint/tendermint/rpc/lib/client"
	cmn "github.com/tendermint/tmlibs/common"
	tmlog "github.com/tendermint/tmlibs/log"
)

type ethstatsConfig struct {
	URL string `toml:",omitempty"`
}

type gethConfig struct {
	Eth      eth.Config
	Node     node.Config
	Ethstats ethstatsConfig
}

func defaultNodeConfig(ctx *cli.Context) node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = version.Version
	cfg.HTTPModules = append(cfg.HTTPModules, "eth")
	cfg.WSModules = append(cfg.WSModules, "eth")
	cfg.IPCPath = "geth.ipc"
	return cfg
}

func makeConfigNode(ctx *cli.Context) (*node.Node, gethConfig) {
	// Load defaults
	cfg := gethConfig{
		Eth:  eth.DefaultConfig,
		Node: defaultNodeConfig(ctx),
	}

	// Apply flags
	ethUtils.SetNodeConfig(ctx, &cfg.Node)
	cfg.Node.P2P.MaxPeers = 0
	cfg.Node.P2P.NoDiscovery = true
	stack, err := node.New(&cfg.Node)
	if err != nil {
		ethUtils.Fatalf("Failed to create the protocol stack: %v", err)
	}

	ethUtils.SetEthConfig(ctx, stack, &cfg.Eth)
	cfg.Eth.PowFake = true

	return stack, cfg
}

func makeFullNode(ctx *cli.Context) *node.Node {
	stack, cfg := makeConfigNode(ctx)

	tendermintURI := ctx.GlobalString(emtUtils.TendermintAddrFlag.Name)

	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return ethereum.NewBackend(ctx, &cfg.Eth, rpcClient.NewURIClient(tendermintURI))
	}); err != nil {
		ethUtils.Fatalf("Failed to register the ABCI application service: %v", err)
	}
	return stack
}

func ethermintCmd(ctx *cli.Context) error {
	// Setup the go-ethereum node and start it
	node := makeFullNode(ctx)
	ethUtils.StartNode(node)

	// Setup the ABCI server and start it
	addr := ctx.GlobalString(emtUtils.ABCIAddrFlag.Name)
	abci := ctx.GlobalString(emtUtils.ABCIProtocolFlag.Name)

	// Fetch the registered service of this type
	var backend *ethereum.Backend
	if err := node.Service(&backend); err != nil {
		ethUtils.Fatalf("backend service not running: %v", err)
	}

	// In-proc RPC connection so ABCI.Query can be forwarded over the ethereum rpc
	rpcClient, err := node.Attach()
	if err != nil {
		ethUtils.Fatalf("Failed to attach to the inproc geth: %v", err)
	}

	// Create the ABCI app
	ethApp, err := abciApp.NewEthermintApplication(backend, rpcClient, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start the app on the ABCI server
	srv, err := server.NewServer(addr, abci, ethApp)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logger := tmlog.NewTMLogger(tmlog.NewSyncWriter(os.Stdout))
	srv.SetLogger(logger.With("module", "abci-server"))

	if _, err := srv.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cmn.TrapSignal(func() {
		srv.Stop()
	})

	return nil
}
