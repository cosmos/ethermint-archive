package main

import (
	"github.com/ethereum/go-ethereum/log"
	"gopkg.in/urfave/cli.v1"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	emtUtils "github.com/tendermint/ethermint/cmd/utils"
)

// nolint: vetshadow
func initCmd(ctx *cli.Context) error {

	// ethereum genesis.json
	genesisPath := ctx.Args().First()
	if len(genesisPath) == 0 {
		ethUtils.Fatalf("must supply path to genesis JSON file")
	}

	_, _, hash := emtUtils.SetupGenesisBlock(genesisPath, ctx)
	log.Info("successfully wrote genesis block and/or chain rule set", "hash", hash)

	return nil
}
