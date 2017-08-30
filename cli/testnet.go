package cli

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCommand.AddCommand(TestNetworkCommand)
}

// TestnetCommand is the entry point to connect an ethermint instance to a
// test network
var TestNetworkCommand = &cobra.Command{
	Use:   "testnet",
	Short: "Connect to the test network.",
	Run:   testNetwork,
}

func testNetwork(cmd *cobra.Command, args []string) error {
	//TODO: Launch ethermint, tendermint and connect to test network.
	return nil
}
