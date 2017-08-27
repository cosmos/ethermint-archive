package utils

import (
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"

	"gopkg.in/urfave/cli.v1"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/log"

	"github.com/tendermint/ethermint/ethereum"
)

func StartNode(stack *ethereum.Node) {
	if err := stack.Start(); err != nil {
		ethUtils.Fatalf("Error starting protocol stack: %v", err)
	}
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt)
		defer signal.Stop(sigc)
		<-sigc
		log.Info("Got interrupt, shutting down...")
		go stack.Stop()
		for i := 10; i > 0; i-- {
			<-sigc
			if i > 1 {
				log.Warn("Already shutting down, interrupt more to panic.", "times", i-1)
			}
		}
	}()
}

// HomeDir returns the user's home most likely home directory
// #unstable
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
// #unstable
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

func ResetAll(ctx *cli.Context) error {
	dbDir := filepath.Join(MakeDataDir(ctx), "ethermint")
	if err := os.RemoveAll(dbDir); err != nil {
		log.Debug("Could not reset ethermint. Failed to remove %+v", dbDir)
		return err
	}

	log.Info("Successfully removed all data", "dir", dbDir)
	return nil
}
