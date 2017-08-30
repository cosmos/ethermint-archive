package cli

import (
	"github.com/spf13/cobra"
)

func init() {
	// TODO: Define the default values for the flags through a config file.
	RootCommand.PersistentFlags().String("home", "~/.ethermint/", "Home directory for all data.")
	RootCommand.PersistentFlags().String("gasprice", "", "The minimal gasprice to accept for a transaction.")
	RootCommand.PersistentFlags().String("coinbase", "", "The address which receives the fees and block rewards.")
	RootCommand.PersistentFlags().String("gasfloor", "", "The minimum amount of gas per block.")
}

// RootCommand is the entry point for the Ethermint cli.
var RootCommand = &cobra.Command{
	Use:   "ethermint",
	Short: "Ethermint proof-of-stake and EVM",
}
