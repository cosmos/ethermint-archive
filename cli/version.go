package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tendermint/ethermint/version"
)

func init() {
}

// VersionCommand prints the version information of Ethermint and Tendermint
var VersionCommand = &cobra.Command{
	Use:   "version",
	Short: "Show the version information of Ethermint.",
	Run: func(cmd *cobra.Command, args []string) {
		// Prints the Ethermint version
		fmt.Println(version.Version)

		// TODO: Print the Tendermint version
	},
}
