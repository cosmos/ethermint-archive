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
func Setup(ctx *cli.Context) error {
	glogger.Verbosity(log.Lvl(ctx.GlobalInt(VerbosityFlag.Name)))
	log.Root().SetHandler(glogger)

	return nil
}

func GetTMLogger() tmlog.Logger {
	return NewTMEthereumProxyLogger()
}

// Interface assertions
var _ tmlog.Logger = (*tmEthereumProxyLogger)(nil)

func NewTMEthereumProxyLogger() tmlog.Logger {
	logger := tmEthereumProxyLogger{}
	return logger
}

type tmEthereumProxyLogger struct {
	keyvals []interface{}
}

func (l tmEthereumProxyLogger) Debug(msg string, ctx ...interface{}) error {
	ctx = append(l.keyvals, ctx...)
	log.Debug(msg, ctx...)
	return nil
}

func (l tmEthereumProxyLogger) Info(msg string, ctx ...interface{}) error {
	ctx = append(l.keyvals, ctx...)
	log.Info(msg, ctx...)
	return nil
}

func (l tmEthereumProxyLogger) Error(msg string, ctx ...interface{}) error {
	ctx = append(l.keyvals, ctx...)
	log.Error(msg, ctx...)
	return nil
}

func (l tmEthereumProxyLogger) With(ctx ...interface{}) tmlog.Logger {
	logger := &tmEthereumProxyLogger{}
	logger.keyvals = make([]interface{}, 0, len(l.keyvals)+len(ctx))
	logger.keyvals = append(logger.keyvals, l.keyvals...)
	logger.keyvals = append(logger.keyvals, ctx...)

	return logger
}
