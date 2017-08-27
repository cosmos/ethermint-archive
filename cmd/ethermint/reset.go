package main

import (
	"gopkg.in/urfave/cli.v1"

	emtUtils "github.com/tendermint/ethermint/cmd/utils"
)

func resetCmd(ctx *cli.Context) error {
	return emtUtils.ResetAll(ctx)
}
