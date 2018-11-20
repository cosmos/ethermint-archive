package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
	secp256k1 "github.com/tendermint/btcd/btcec"
	tmsecp256k1 "github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestPrivKeyToSecp256k1(t *testing.T) {
	tmPrivKey := tmsecp256k1.GenPrivKey()
	convertedPriv := PrivKeyToSecp256k1(tmPrivKey)
	require.Equal(t, tmPrivKey[:], (*secp256k1.PrivateKey)(convertedPriv).Serialize())
}
