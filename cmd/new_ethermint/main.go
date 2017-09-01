package main

import (
	"github.com/tendermint/ethermint/cli"
)

func main() {
	// TODO: Check the error here
	rootCommand, _ := cli.NewRootCommand(cli.VersionCommand,
		cli.TestNetworkCommand, cli.DevelopmentCommand,
		cli.MainNetworkCommand)

	rootCommand.Execute()
}
