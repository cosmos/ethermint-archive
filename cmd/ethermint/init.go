package main

import (
	"os"

	. "github.com/tendermint/go-common"
	"github.com/tendermint/tendermint/types"
)

func init_files() {
	// if no priv val, make it
	privValFile := config.GetString("priv_validator_file")
	if _, err := os.Stat(privValFile); err != nil {
		privValidator := types.GenPrivValidator()
		privValidator.SetFile(privValFile)
		privValidator.Save()

		// if no genesis, make it using the priv val
		genFile := config.GetString("genesis_file")
		if _, err := os.Stat(genFile); err != nil {
			genDoc := types.GenesisDoc{
				ChainID: Fmt("test-chain-%v", RandStr(6)),
			}
			genDoc.Validators = []types.GenesisValidator{types.GenesisValidator{
				PubKey: privValidator.PubKey,
				Amount: 10,
			}}
			genDoc.SaveAs(genFile)
		}
	}

	// TODO: if there is a priv val but no genesis
}
