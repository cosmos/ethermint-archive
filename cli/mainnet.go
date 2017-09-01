package cli

import (
	"github.com/spf13/cobra"
)

// MainNetworkCommand is the command to connect to the main
// network.
var MainNetworkCommand = &cobra.Command{
	Use:   "mainnet",
	Short: "Connect to the main network.",
	Run:   mainNetwork,
}

func mainNetwork(cmd *cobra.Command, args []string) {
	// TODO: Run main network
}
