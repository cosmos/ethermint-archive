package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethdb"
)

// 1. set data dir via EMHOME env variable
// 2. init genesis
// 3. reset all
// 4. check dir is empty
func TestResetAll(t *testing.T) {
	// setup temp data dir
	tempDatadir, err := ioutil.TempDir("", "ethermint_test")
	if err != nil {
		t.Error("unable to create temporary datadir")
	}
	defer os.RemoveAll(tempDatadir) // nolint: errcheck

	// set EMHOME env variable
	if err = os.Setenv(emHome, tempDatadir); err != nil {
		t.Errorf("could not set env: %v", err)
	}
	defer func() {
		if err = os.Unsetenv(emHome); err != nil {
			t.Errorf("could not unset env: %v", err)
		}
	}()

	// context with empty flag set
	context := getContextNoFlag()

	dataDir := filepath.Join(MakeDataDir(context), "ethermint/chaindata")

	chainDb, err := ethdb.NewLDBDatabase(dataDir, 0, 0)
	if err != nil {
		t.Errorf("could not open database: %v", err)
	}

	// setup genesis
	_, _, err = core.SetupGenesisBlock(chainDb, core.DefaultTestnetGenesisBlock())
	if err != nil {
		t.Errorf("failed to write genesis block: %v", err)
	}

	// check dir exists
	if _, err = os.Stat(dataDir); err != nil {
		t.Errorf("database doesn't exist: %v", err)

	}

	// clear
	if err = ResetAll(context); err != nil {
		t.Errorf("Failed to remove ethermint home directory: %+v", err)
	}
}
