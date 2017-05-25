package main

import (
	"fmt"
	"os"

	"gopkg.in/urfave/cli.v1"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/node"

	abciApp "github.com/tendermint/ethermint/app"
	emtUtils "github.com/tendermint/ethermint/cmd/utils"
	"github.com/tendermint/ethermint/ethereum"

	"github.com/tendermint/abci/server"
	cmn "github.com/tendermint/tmlibs/common"
	tmlog "github.com/tendermint/tmlibs/log"
)

func ethermintCmd(ctx *cli.Context) error {
	// Setup the go-ethereum node and start it
	node := makeFullNode(ctx)
	startNode(ctx, node)

	// Setup the ABCI server and start it
	addr := ctx.GlobalString(emtUtils.ABCIAddrFlag.Name)
	abci := ctx.GlobalString(emtUtils.ABCIProtocolFlag.Name)

	// Fetch the registered service of this type
	var backend *ethereum.Backend
	if err := node.Service(&backend); err != nil {
		ethUtils.Fatalf("ethereum backend service not running: %v", err)
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

func startNode(ctx *cli.Context, stack *node.Node) {
	ethUtils.StartNode(stack)
}
