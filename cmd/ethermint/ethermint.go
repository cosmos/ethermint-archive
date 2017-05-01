package main

import (
	"fmt"
	"os"

	"gopkg.in/urfave/cli.v1"

	ethereumUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/logger/glog"

	"github.com/tendermint/ethermint/app"
	"github.com/tendermint/ethermint/ethereum"
	"github.com/tendermint/ethermint/version"

	"github.com/tendermint/abci/server"
	tendermintNode "github.com/tendermint/tendermint/node"
	ethermintUtils "github.com/tendermint/ethermint/cmd/utils"
)

func ethermintCmd(ctx *cli.Context) error {
	stack := ethereum.MakeSystemNode(clientIdentifier, version.Version, ctx)
	ethereumUtils.StartNode(stack)
	addr := ctx.GlobalString("addr")
	abci := ctx.GlobalString("abci")

	//set verbosity level for go-ethereum
	glog.SetToStderr(true)
	glog.SetV(ctx.GlobalInt(ethermintUtils.VerbosityFlag.Name))

	var backend *ethereum.Backend
	if err := stack.Service(&backend); err != nil {
		ethereumUtils.Fatalf("backend service not running: %v", err)
	}
	client, err := stack.Attach()
	if err != nil {
		ethereumUtils.Fatalf("Failed to attach to the inproc geth: %v", err)
	}
	ethApp, err := app.NewEthermintApplication(backend, client, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)

	}
	_, err = server.NewServer(addr, abci, ethApp)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("tm node")
	tendermintNode.RunNode(config)
	return nil
}
