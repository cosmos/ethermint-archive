package utils

import (
	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/cmd/utils"
)

var (
	// ----------------------------
	// go-ethereum flags

	// So we can control the DefaultDir
	DataDirFlag = utils.DirectoryFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases and keystore",
		Value: utils.DirectoryString{DefaultDataDir()},
	}

	// Not exposed by go-ethereum
	VerbosityFlag = cli.IntFlag{
		Name:  "verbosity",
		Usage: "Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=detail",
		Value: 3,
	}

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
