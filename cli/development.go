package cli

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCommand.AddCommand(DevelopmentCommand)
	DevelopmentCommand.PersistentFlags().String("gaslimit", "", "The maximum amount of gas per block.")
}

// DevelopmentCommand is the entry point to create a private network.
var DevelopmentCommand = &cobra.Command{
	Use:   "development",
	Short: "Start a local network for development.",
	Run:   development,
}

func development(cmd *cobra.Command, args []string) error {
	// TODO: Launch ethermint and tendermint. Copy private keys.
	return nil
}
