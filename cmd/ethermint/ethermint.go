package main

import (
	"fmt"
	"os"

	"gopkg.in/urfave/cli.v1"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"

	"github.com/tendermint/abci/server"
	abciApp "github.com/tendermint/ethermint/app"
	emtUtils "github.com/tendermint/ethermint/cmd/utils"
	"github.com/tendermint/ethermint/ethereum"
	"github.com/tendermint/ethermint/version"
	rpcClient "github.com/tendermint/tendermint/rpc/lib/client"
	cmn "github.com/tendermint/tmlibs/common"
)

type ethstatsConfig struct {
	URL string `toml:",omitempty"`
}

type gethConfig struct {
	Eth      eth.Config
	Node     node.Config
	Ethstats ethstatsConfig
}

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = version.Version
	cfg.HTTPModules = append(cfg.HTTPModules, "eth")
	cfg.WSModules = append(cfg.WSModules, "eth")
	cfg.IPCPath = "geth.ipc"
	return cfg
}

func makeConfigNode(ctx *cli.Context) (*node.Node, gethConfig) {
	cfg := gethConfig{
		Eth:  eth.DefaultConfig,
		Node: defaultNodeConfig(),
	}

	ethUtils.SetNodeConfig(ctx, &cfg.Node)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		ethUtils.Fatalf("Failed to create the protocol stack: %v", err)
	}
	ethUtils.SetEthConfig(ctx, stack, &cfg.Eth)

	if ctx.GlobalIsSet(ethUtils.EthStatsURLFlag.Name) {
		cfg.Ethstats.URL = ctx.GlobalString(ethUtils.EthStatsURLFlag.Name)
	}

	return stack, cfg
}

func ethermintCmd(ctx *cli.Context) error {

	stack, cfg := makeConfigNode(ctx)

	ethUtils.RegisterEthService(stack, &cfg.Eth)

	if cfg.Ethstats.URL != "" {
		ethUtils.RegisterEthStatsService(stack, cfg.Ethstats.URL)
	}

	//Remote tendermint RPC address
	tendermintURI := ctx.GlobalString(emtUtils.BroadcastTxAddrFlag.Name)

	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return ethereum.NewBackend(ctx, &cfg.Eth, rpcClient.NewURIClient(tendermintURI))
	}); err != nil {
		ethUtils.Fatalf("Failed to register the ABCI application service: %v", err)
	}

	ethUtils.StartNode(stack)

	addr := ctx.GlobalString("addr")
	abci := ctx.GlobalString("abci")

	// Fetch the registered service of this type
	var backend *ethereum.Backend
	if err := stack.Service(&backend); err != nil {
		ethUtils.Fatalf("backend service not running: %v", err)
	}

	// In-proc RPC connection so ABCI.Query can be forwarded over the ethereum rpc
	rpcClient, err := stack.Attach()
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
	_, err = server.NewServer(addr, abci, ethApp)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cmn.TrapSignal(func() {

	})

	return nil
}
