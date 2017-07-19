package main

import (
	"os"
	"path/filepath"

	"gopkg.in/urfave/cli.v1"

	"encoding/json"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
)

// nolint: vetshadow
func initCmd(ctx *cli.Context) error {

	// ethereum genesis.json
	genesisPath := ctx.Args().First()
	if len(genesisPath) == 0 {
		ethUtils.Fatalf("must supply path to genesis JSON file")
	}

	file, err := os.Open(genesisPath)
	if err != nil {
		ethUtils.Fatalf("Failed to read genesis file: %v", err)
	}
	defer file.Close() // nolint: errcheck

	genesis := new(core.Genesis)
	if err := json.NewDecoder(file).Decode(genesis); err != nil {
		ethUtils.Fatalf("invalid genesis file: %v", err)
	}

	chainDb, err := ethdb.NewLDBDatabase(filepath.Join(ethUtils.MakeDataDir(ctx), "ethermint/chaindata"), 0, 0)
	if err != nil {
		ethUtils.Fatalf("could not open database: %v", err)
	}

	_, hash, err := core.SetupGenesisBlock(chainDb, genesis)
	if err != nil {
		ethUtils.Fatalf("failed to write genesis block: %v", err)
	}

	log.Info("successfully wrote genesis block and/or chain rule set", "hash", hash)
	return nil
}
