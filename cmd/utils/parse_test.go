package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
)

var defaultGenesis = func() *core.Genesis {
	g := new(core.Genesis)
	if err := json.Unmarshal(defaultGenesisBlob, g); err != nil {
		log.Fatalf("parsing defaultGenesis: %v", err)
	}
	return g
}()

func bigString(s string) *big.Int { // nolint: unparam
	b, _ := big.NewInt(0).SetString(s, 10)
	return b
}

var genesis1 = &core.Genesis{
	Difficulty: big.NewInt(0x40),
	GasLimit:   0x8000000,
	Alloc: core.GenesisAlloc{
		ethCommon.HexToAddress("0x7eff122b94897ea5b0e2a9abf47b86337fafebdc"): {
			Balance: bigString("10000000000000000000000000000000000"),
		},
		ethCommon.HexToAddress("0xc6713982649D9284ff56c32655a9ECcCDA78422A"): {
			Balance: bigString("10000000000000000000000000000000000"),
		},
	},
}

func TestParseGenesisOrDefault(t *testing.T) {
	tests := [...]struct {
		path    string
		want    *core.Genesis
		wantErr bool
	}{
		0: {path: "", want: defaultGenesis},
		1: {want: defaultGenesis},
		2: {path: fmt.Sprintf("non-existent-%d", rand.Int()), want: defaultGenesis},
		3: {path: "./testdata/blank-genesis.json", want: defaultGenesis},
		4: {path: "./testdata/genesis1.json", want: genesis1},
		5: {path: "./testdata/non-genesis.json", wantErr: true},
	}

	for i, tt := range tests {
		gen, err := ParseGenesisOrDefault(tt.path)
		if tt.wantErr {
			assert.NotNil(t, err, "#%d: cannot be nil", i)
			continue
		}

		if err != nil {
			t.Errorf("#%d: path=%q unexpected error: %v", i, tt.path, err)
			continue
		}

		assert.NotEqual(t, blankGenesis, gen, true, "#%d: expecting a non-blank", i)
		assert.Equal(t, gen, tt.want, "#%d: expected them to be the same", i)
	}
}
