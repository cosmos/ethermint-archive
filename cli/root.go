package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCommand.PersistentFlags().String("home", "$HOME/.ethermint", `Home directory for all data.`)
	rootCommand.PersistentFlags().String("gasprice", "", `The minimal gasprice to accept for a transaction.`)
	rootCommand.PersistentFlags().String("coinbase", "", `The address which receives the fees and block rewards.`)
	rootCommand.PersistentFlags().String("gasfloor", "", `The minimum amount of gas per block.`)
	rootCommand.PersistentFlags().Bool("with-tendermint", true, `Launch Ethermint without Tendermint.`)
}

// RootCommand is the entry point for the Ethermint cli.
var rootCommand = &cobra.Command{
	Use:   "ethermint",
	Short: "Ethermint proof-of-stake and EVM",
}

// NewRootCommand setups the root command with all other commands.
func NewRootCommand(cmds ...*cobra.Command) *cobra.Command {
	rootCommand.AddCommand(cmds...)
	rootCommand.PersistentPreRunE = bindFlags
	return rootCommand
}

func bindFlags(cmd *cobra.Command, args []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}
	return nil
}
