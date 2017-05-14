package main

import (
	"fmt"
	"os"

	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/logger/glog"

	"github.com/tendermint/ethermint/app"
	"github.com/tendermint/ethermint/ethereum"
	"github.com/tendermint/ethermint/version"
	//	minerRewardStrategies "github.com/tendermint/ethermint/strategies/miner"
	//	validatorsStrategy "github.com/tendermint/ethermint/strategies/validators"

	"github.com/tendermint/abci/server"
	tendermintNode "github.com/tendermint/tendermint/node"
)

func ethermintCmd(ctx *cli.Context) error {
	stack := ethereum.MakeSystemNode(clientIdentifier, version.Version, ctx)
	utils.StartNode(stack)

	addr := ctx.GlobalString("addr")
	abci := ctx.GlobalString("abci")

	//set verbosity level for go-ethereum
	glog.SetToStderr(true)
	glog.SetV(ctx.GlobalInt(VerbosityFlag.Name))

	// Fetch the registered service of this type
	var backend *ethereum.Backend
	if err := stack.Service(&backend); err != nil {
		utils.Fatalf("backend service not running: %v", err)
	}

	// In-proc RPC connection so ABCI.Query can be forwarded over the ethereum rpc
	client, err := stack.Attach()
	if err != nil {
		utils.Fatalf("Failed to attach to the inproc geth: %v", err)
	}

	// Create the ABCI app
	ethApp, err := app.NewEthermintApplication(backend, client, nil)
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

	// Start tendermint
	fmt.Println("tm node")
	tendermintNode.RunNode(config)
	return nil
}
