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
		Usage: "Verbosity of go-ethereum",
	}

	// ----------------------------
	// Tendermint Flags

	MonikerFlag = cli.StringFlag{
		Name:  "moniker",
		Value: "",
		Usage: "Node's moniker",
	}

	NodeLaddrFlag = cli.StringFlag{
		Name:  "node_laddr",
		Value: "tcp://0.0.0.0:46656",
		Usage: "Node listen address. (0.0.0.0:0 means any interface, any port)",
	}

	LogLevelFlag = cli.StringFlag{
		Name:  "log_level",
		Value: "info",
		Usage: "Tendermint Log level",
	}

	SeedsFlag = cli.StringFlag{
		Name:  "seeds",
		Value: "",
		Usage: "Comma delimited host:port seed nodes",
	}

	FastSyncFlag = cli.BoolFlag{
		Name:  "fast_sync",
		Usage: "Fast blockchain syncing",
	}

	SkipUpnpFlag = cli.BoolFlag{
		Name:  "skip_upnp",
		Usage: "Skip UPNP configuration",
	}

	RpcLaddrFlag = cli.StringFlag{
		Name:  "rpc_laddr",
		Value: "tcp://0.0.0.0:46657",
		Usage: "RPC listen address. Port required",
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
