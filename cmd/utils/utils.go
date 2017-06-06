package utils

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

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
