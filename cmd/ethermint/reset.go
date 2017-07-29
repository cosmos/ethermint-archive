package main

import (
	"github.com/ethereum/go-ethereum/log"
	emtUtils "github.com/tendermint/ethermint/cmd/utils"
	"gopkg.in/urfave/cli.v1"
	"os"
	"path/filepath"
)

func resetCmd(ctx *cli.Context) error {
	dbDir := filepath.Join(emtUtils.MakeDataDir(ctx), "ethermint")
	os.RemoveAll(dbDir)

	log.Info("Successfully removed all data", "dir", dbDir)

	return nil
}
