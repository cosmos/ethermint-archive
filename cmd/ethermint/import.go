package main

import (
	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/log"
	emtUtils "github.com/tendermint/ethermint/cmd/utils"
	rpcClient "github.com/tendermint/tendermint/rpc/lib/client"
	"gopkg.in/urfave/cli.v1"
	"github.com/tendermint/ethermint/ethereum"
)

func importCmd(ctx *cli.Context) error {
	if len(ctx.Args()) < 2 {
		ethUtils.Fatalf("This command requires an arguments.")
	}

	// connect to tendermint
	tendermintLAddr := ctx.GlobalString(emtUtils.TendermintAddrFlag.Name)
	client := rpcClient.NewURIClient(tendermintLAddr)

	ethereum.WaitForServer(client)

	chain, chainDb, eventMux := emtUtils.MakeChain(ctx.Args().First(), ctx)
	defer chainDb.Close()

	// Start a subscriber.
	done := make(chan struct{})
	sub := eventMux.Subscribe(core.ChainEvent{})
	go func() {
		for obj := range sub.Chan() {
			event := obj.Data.(core.ChainEvent)
			log.Info("Get block", "block", event.Block)
			for _, transaction := range event.Block.Transactions() {
				if err := ethereum.BroadcastTransaction(client, transaction); err != nil {
					log.Error("Broadcast error", "err", err)
				}
			}

		}

		log.Info("Subscriber done")

		close(done)
	}()

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

	//eventMux.Stop()

	// Wait for subscriber.
	<-done

	return nil
}
