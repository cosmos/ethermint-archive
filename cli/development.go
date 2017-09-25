package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethdb"
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

	emHome := viper.GetString("home")

	// If the home directory does not exist then initialise everything.
	// If it does exist we start everything straight away.

	// TODO: This should be covered in five functions. TendermintInitialise(),
	// TendermintStart(), EthereumInitialise(), EthereumStart(), CreateAccounts()
	// The first two move into utils.go and the latter move into the ethereum
	// package.
	if _, err := os.Stat(emHome); os.IsNotExist(err) {
		genesis, err := parseEthereumGenesisOrElse(args, developmentGenesisConfig)
		if err != nil {
			// TODO: Handle error
		}

		withTendermint := viper.GetBool("with-tendermint")
		if withTendermint {
			tmHome := filepath.Join(emHome, "tendermint")
			tmArgs := []string{"init", "--home", tmHome}
			if _, err := InvokeTendermint(tmArgs...); err != nil {
				// TODO: Throw an error
			}
		}

		chainDB, err := ethdb.NewLDBDatabase(filepath.
			Join(emHome, "chaindata"), 0, 0)
		if err != nil {
			// TODO: Handle error
		}

		_, _, err = core.SetupGenesisBlock(chainDB, genesis)
		if err != nil {
			// TODO: Handle error
		}

		keystore := filepath.Join(emHome, "keystore")
		if err := os.MkdirAll(keystore, 0777); err != nil {
			// TODO: Handle error
		}

		for keyName, privKey := range developmentPrivateKeys {
			file := filepath.Join(keystore, keyName)
			f, err := os.Create(file)
			defer f.Close()
			if err != nil {
				// TODO: Shouldn't happen but handle it
				continue
			}
			if _, err := f.Write(privKey); err != nil {
				// TODO: Handle error
			}
		}

	}

	fmt.Println(emHome)
}

var developmentPrivateKeys = map[string][]byte{
	"UTC--2016-10-21T22-30-03.071787745Z--7eff122b94897ea5b0e2a9abf47b86337fafebdc": []byte(`
{
  "address":"7eff122b94897ea5b0e2a9abf47b86337fafebdc",
  "id":"f86a62b4-0621-4616-99af-c4b7f38fcc48","version":3,
  "crypto":{
    "cipher":"aes-128-ctr","ciphertext":"19de8a919e2f4cbdde2b7352ebd0be8ead2c87db35fc8e4c9acaf74aaaa57dad",
    "cipherparams":{"iv":"ba2bd370d6c9d5845e92fbc6f951c792"},
    "kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"c7cc2380a96adc9eb31d20bd8d8a7827199e8b16889582c0b9089da6a9f58e84"},
    "mac":"ff2c0caf051ca15d8c43b6f321ec10bd99bd654ddcf12dd1a28f730cc3c13730"
  }
}`),
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

// TODO: Move this into the ethereum package. The cli package should not have
// any dependencies on go-ethereum. This function will be used from all
// commands that start a network.
func parseEthereumGenesisOrElse(genesisPath []string, defaultConfig []byte) (*core.Genesis, error) {
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
