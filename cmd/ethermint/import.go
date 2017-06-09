package main

import (
	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	emtUtils "github.com/tendermint/ethermint/cmd/utils"
	"gopkg.in/urfave/cli.v1"
	"github.com/ethereum/go-ethereum/log"

)

func importCmd(ctx *cli.Context) error {
	if len(ctx.Args()) < 2 {
		ethUtils.Fatalf("This command requires an arguments.")
	}

	chain, chainDb := emtUtils.MakeChain(ctx.Args().First(), ctx)
	defer chainDb.Close()

	if len(ctx.Args()) == 2 {
		if err := ethUtils.ImportChain(chain, ctx.Args().Get(1)); err != nil {
			ethUtils.Fatalf("Import error: %v", err)
		}
	} else {
		for i, arg := range ctx.Args() {
			//skip genesis location
			if i == 0 {
				continue
			}

			if err := ethUtils.ImportChain(chain, arg); err != nil {
				log.Error("Import error", "file", arg, "err", err)
			}
		}
	}

	return nil
}
