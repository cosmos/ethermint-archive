package main

import (
	"os"
	"path/filepath"

	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"

	cmn "github.com/tendermint/go-common"
	"github.com/tendermint/tendermint/types"
)

func initCmd(ctx *cli.Context) error {
	init_files()

	// ethereum genesis.json
	genesisPath := ctx.Args().First()
	if len(genesisPath) == 0 {
		utils.Fatalf("must supply path to genesis JSON file")
	}

	chainDb, err := ethdb.NewLDBDatabase(filepath.Join(utils.MakeDataDir(ctx), "chaindata"), 0, 0)
	if err != nil {
		utils.Fatalf("could not open database: %v", err)
	}

	genesisFile, err := os.Open(genesisPath)
	if err != nil {
		utils.Fatalf("failed to read genesis file: %v", err)
	}

	block, err := core.WriteGenesisBlock(chainDb, genesisFile)
	if err != nil {
		utils.Fatalf("failed to write genesis block: %v", err)
	}
	glog.V(logger.Info).Infof("successfully wrote genesis block and/or chain rule set: %x", block.Hash())
	return nil
}

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
