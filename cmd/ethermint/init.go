package main

import (
	"os"
	"path/filepath"

	"gopkg.in/urfave/cli.v1"

	"encoding/json"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
)

func initCmd(ctx *cli.Context) error {

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

	genesis := new(core.Genesis)
	if err := json.NewDecoder(genesisFile).Decode(genesis); err != nil {
		utils.Fatalf("invalid genesis file: %v", err)
	}
	_, hash, err := core.SetupGenesisBlock(chainDb, genesis)
	if err != nil {
		utils.Fatalf("failed to write genesis block: %v", err)
	}
	log.Info("successfully wrote genesis block and/or chain rule set: %x", hash)
	return nil
}
