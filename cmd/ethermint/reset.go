package main

import (
	emtUtils "github.com/tendermint/ethermint/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

func resetCmd(ctx *cli.Context) error {
	emtUtils.ResetAll(ctx)

	return nil
}
