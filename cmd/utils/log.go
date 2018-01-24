package utils

import (
	"io"
	"os"

	"gopkg.in/urfave/cli.v1"

	colorable "github.com/mattn/go-colorable"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/log/term"

	tmlog "github.com/tendermint/tmlibs/log"
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

// Setup sets up the logging infrastructure
// #unstable
func Setup(ctx *cli.Context) error {
	glogger.Verbosity(log.Lvl(ctx.GlobalInt(VerbosityFlag.Name)))
	log.Root().SetHandler(glogger)

	return nil
}

// ---------------------------
// EthermintLogger - wraps the logger in tmlibs

// Interface assertions
var _ tmlog.Logger = (*ethermintLogger)(nil)

type ethermintLogger struct {
	keyvals []interface{}
}

// EthermintLogger returns a new instance of an ethermint logger. With() should
// be called upon the returned instance to set default keys
// #unstable
func EthermintLogger() tmlog.Logger {
	logger := ethermintLogger{keyvals: make([]interface{}, 0)}
	return logger
}

// Debug proxies everything to the go-ethereum logging facilities
// #unstable
func (l ethermintLogger) Debug(msg string, ctx ...interface{}) {
	ctx = append(l.keyvals, ctx...)
	log.Debug(msg, ctx...)
}

// Info proxies everything to the go-ethereum logging facilities
// #unstable
func (l ethermintLogger) Info(msg string, ctx ...interface{}) {
	ctx = append(l.keyvals, ctx...)
	log.Info(msg, ctx...)
}

// Error proxies everything to the go-ethereum logging facilities
// #unstable
func (l ethermintLogger) Error(msg string, ctx ...interface{}) {
	ctx = append(l.keyvals, ctx...)
	log.Error(msg, ctx...)
}

// With proxies everything to the go-ethereum logging facilities
// #unstable
func (l ethermintLogger) With(ctx ...interface{}) tmlog.Logger {
	l.keyvals = append(l.keyvals, ctx...)

	return l
}
