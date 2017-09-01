package cli

import (
	"github.com/spf13/cobra"
)

// TestNetworkCommand is the entry point to connect an ethermint instance to a
// test network
var TestNetworkCommand = &cobra.Command{
	Use:   "testnet",
	Short: "Connect to the test network.",
	Run:   testNetwork,
}

func testNetwork(cmd *cobra.Command, args []string) {
	//TODO: Launch ethermint, tendermint and connect to test network.
}
