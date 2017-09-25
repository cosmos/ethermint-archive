package main

import (
	"fmt"

	"github.com/tendermint/ethermint/cli"
)

func main() {
	// TODO: Check the error here
	rootCommand := cli.NewRootCommand(cli.VersionCommand,
		cli.TestNetworkCommand, cli.DevelopmentCommand,
		cli.MainNetworkCommand)

	err := rootCommand.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
