package utils

import (
	"io"
	"os"

	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/log/term"
	colorable "github.com/mattn/go-colorable"
)

var (
	// ----------------------------
	// ABCI Flags

	TendermintAddrFlag = cli.StringFlag{
		Name:  "tendermint_addr",
		Value: "tcp://localhost:46657",
		Usage: "This is the address that ethermint will use to connect to the tendermint core node. Please provide a port.",
	}

	ABCIAddrFlag = cli.StringFlag{
		Name:  "abci_laddr",
		Value: "tcp://0.0.0.0:46658",
		Usage: "This is the address that the ABCI server will use to listen to incoming connection from tendermint core.",
	}

	ABCIProtocolFlag = cli.StringFlag{
		Name:  "abci_protocol",
		Value: "socket",
		Usage: "socket | grpc",
	}

	VerbosityFlag = cli.IntFlag{
		Name:  "verbosity",
		Value: 3,
		Usage: "Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=core, 5=debug, 6=detail",
	}

	DebugFlag = cli.BoolFlag{
		Name:  "debug",
		Usage: "Prepends log messages with call-site location (file and line number)",
	}
)

var glogger *log.GlogHandler

func init() {
	usecolor := term.IsTty(os.Stderr.Fd()) && os.Getenv("TERM") != "dumb"
	output := io.Writer(os.Stderr)
	if usecolor {
		output = colorable.NewColorableStderr()
	}
	glogger = log.NewGlogHandler(log.StreamHandler(output, log.TerminalFormat(usecolor)))
}

func Setup(ctx *cli.Context) error {
	log.PrintOrigins(ctx.GlobalBool(DebugFlag.Name))
	glogger.Verbosity(log.Lvl(ctx.GlobalInt(VerbosityFlag.Name)))
	log.Root().SetHandler(glogger)

	return nil
}
