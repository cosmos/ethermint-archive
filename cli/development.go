package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ethereum/go-ethereum/core"
)

func init() {
	DevelopmentCommand.Flags().String("gaslimit", "", `The maximum amount of gas per block.`)
}

// DevelopmentCommand is the entry point to create a private network.
var DevelopmentCommand = &cobra.Command{
	Use:   "development",
	Short: "Start a local network for development.",
	Run:   development,
}

func development(cmd *cobra.Command, args []string) {
	// TODO: Launch ethermint and tendermint. Copy private keys.

	// TODO: If a genesis.json file is supplied then try to initialise.
	// If it isn't supplied check if initialisation already happened.
	// If none of the above is true throw an error and warn the user.
	// Do the same for Tendermint.

	// TODO: Start Ethermint and Tendermint.
	if len(args) == 0 {
		// TODO: Check that ethermint has been initialised,
		// if it hasn't throw an error
	}

	genesis, err := parseGenesisOrElse(args, developmentGenesisConfig)
	if err != nil {
		// TODO: Throw an error
	}

	home := viper.GetString("home")
	fmt.Println(home)
	fmt.Println(genesis.Difficulty)
}

var developmentGenesisConfig = []byte(`
{
    "config": {
        "chainId": 15,
        "homesteadBlock": 0,
        "eip155Block": 0,
        "eip158Block": 0
    },
    "nonce": "0xdeadbeefdeadbeef",
    "timestamp": "0x00",
    "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "difficulty": "0x40",
    "gasLimit": "0x8000000",
    "alloc": {
        "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc": { "balance": "10000000000000000000000000000000000" },
	"0xc6713982649D9284ff56c32655a9ECcCDA78422A": { "balance": "10000000000000000000000000000000000" }
    }
}`)

// -----------------------------------------------------------------------------

// TODO: Move this into the ethereum package. The cli package should no have
// any dependencies on go-ethereum. This function will be used from all
// commands that start a network.
func parseGenesisOrElse(genesisPath []string, defaultConfig []byte) (*core.Genesis, error) {
	var genesisConfig = defaultConfig[:]
	if len(genesisPath) > 0 {
		config, err := ioutil.ReadFile(genesisPath[0])
		if err != nil {
			return nil, err
		}
		genesisConfig = config
	}

	genesis := new(core.Genesis)
	if err := json.Unmarshal(genesisConfig, genesis); err != nil {
		return nil, err
	}

	return genesis, nil
}
