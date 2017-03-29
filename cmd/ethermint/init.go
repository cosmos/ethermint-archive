package main

import (
	"os"

	cmn "github.com/tendermint/go-common"
	"github.com/tendermint/tendermint/types"
)

func init_files() {
	// if no priv val, make it
	privValFile := config.GetString("priv_validator_file")
	if _, err := os.Stat(privValFile); os.IsNotExist(err) {
		privValidator := types.GenPrivValidator()
		privValidator.SetFile(privValFile)
		privValidator.Save()

		// if no genesis, make it using the priv val
		genFile := config.GetString("genesis_file")
		if _, err := os.Stat(genFile); os.IsNotExist(err) {
			genDoc := types.GenesisDoc{
				ChainID: cmn.Fmt("test-chain-%v", cmn.RandStr(6)),
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
