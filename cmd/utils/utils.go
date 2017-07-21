package utils

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethdb"
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

// NewMemoryCache returns new state db based on memory db
func NewMemoryCache() *state.StateDB {
	// Create an empty state
	db, _ := ethdb.NewMemDatabase()

	state, _ := state.New(common.Hash{}, state.NewDatabase(db))

	return state
}
