package utils

import (
	"encoding/json"
	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/params"
	"gopkg.in/urfave/cli.v1"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// HomeDir returns the user's home most likely home directory
func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

// DefaultDataDir tries to guess the default directory for ethermint data
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := HomeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "Ethermint")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "Ethermint")
		} else {
			return filepath.Join(home, ".ethermint")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

// MakeChain creates a chain manager from set command line flags.
func MakeChain(genesisPath string, ctx *cli.Context) (chain *core.BlockChain, chainDb ethdb.Database) {
	var err error

	chainDb, config, _ := SetupGenesisBlock(genesisPath, ctx)

	engine := ethash.NewFaker()

	vmcfg := vm.Config{EnablePreimageRecording: ctx.GlobalBool(ethUtils.VMEnableDebugFlag.Name)}
	chain, err = core.NewBlockChain(chainDb, config, engine, new(event.TypeMux), vmcfg)
	if err != nil {
		ethUtils.Fatalf("Can't create BlockChain: %v", err)
	}

	return chain, chainDb
}

// SetupGenesisBlock creates chain database and write genesis to it
func SetupGenesisBlock(genesisPath string, ctx *cli.Context) (ethdb.Database, *params.ChainConfig, common.Hash) {
	file, err := os.Open(genesisPath)
	if err != nil {
		ethUtils.Fatalf("Failed to read genesis file: %v", err)
	}
	defer file.Close() // nolint: errcheck

	genesis := new(core.Genesis)
	if err := json.NewDecoder(file).Decode(genesis); err != nil {
		ethUtils.Fatalf("invalid genesis file: %v", err)
	}

	chainDb, err := ethdb.NewLDBDatabase(filepath.Join(ethUtils.MakeDataDir(ctx), "ethermint/chaindata"), 0, 0)
	if err != nil {
		ethUtils.Fatalf("could not open database: %v", err)
	}

	config, hash, err := core.SetupGenesisBlock(chainDb, genesis)
	if err != nil {
		ethUtils.Fatalf("failed to write genesis block: %v", err)
	}

	return chainDb, config, hash
}
