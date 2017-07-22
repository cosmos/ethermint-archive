package utils

import (
	"flag"
	"os"
	"testing"

	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/node"
)

// we use EMHOME env variable if don't set datadir flag
func TestEMHome(t *testing.T) {
	// set env variable
	emHomedir := "/tmp/dir1"
	os.Setenv(emHome, emHomedir) // nolint: errcheck

	// context with empty flag set
	context := getContextNoFlag()

	_, config := makeConfigNode(context)

	if config.Node.DataDir != emHomedir {
		t.Errorf("DataDir is wrong: %s", config.Node.DataDir)
	}
}

func TestEmHomeDataDir(t *testing.T) {
	// set env variable
	emHomedir := "/tmp/dir1"
	os.Setenv(emHome, emHomedir) // nolint: errcheck

	// context with datadir flag
	dir := "/tmp/dir2"
	context := getContextDataDirFlag(dir)

	_, config := makeConfigNode(context)

	if config.Node.DataDir != dir {
		t.Errorf("DataDir is wrong: %s", config.Node.DataDir)
	}
}

// init cli.context with empty flag set
func getContextNoFlag() *cli.Context {
	set := flag.NewFlagSet("test", 0)
	globalSet := flag.NewFlagSet("test", 0)

	globalCtx := cli.NewContext(nil, globalSet, nil)
	ctx := cli.NewContext(nil, set, globalCtx)

	return ctx
}

// nolint: unparam
func getContextDataDirFlag(dir string) *cli.Context {
	set := flag.NewFlagSet("test", 0)
	globalSet := flag.NewFlagSet("test", 0)
	globalSet.String("datadir", node.DefaultDataDir(), "doc")

	globalCtx := cli.NewContext(nil, globalSet, nil)
	ctx := cli.NewContext(nil, set, globalCtx)

	globalSet.Parse([]string{"--datadir", dir}) // nolint: errcheck

	return ctx
}
