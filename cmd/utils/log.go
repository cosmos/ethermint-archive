package utils

import (
	"io"
	"os"

	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/log/term"
	colorable "github.com/mattn/go-colorable"
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
	glogger.Verbosity(log.Lvl(ctx.GlobalInt(VerbosityFlag.Name)))
	log.Root().SetHandler(glogger)

	return nil
}
