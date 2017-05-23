package utils

import (
	"gopkg.in/urfave/cli.v1"
)

var (
	// ----------------------------
	// ABCI Flags

	BroadcastTxAddrFlag = cli.StringFlag{
		Name:  "broadcast_tx_addr",
		Value: "tcp://localhost:46657",
		Usage: "Remote tendermint RPC address. Port required",
	}

	AddrFlag = cli.StringFlag{
		Name:  "addr",
		Value: "tcp://0.0.0.0:46658",
		Usage: "TMSP app listen address",
	}

	AbciFlag = cli.StringFlag{
		Name:  "abci",
		Value: "socket",
		Usage: "socket | grpc",
	}
)
