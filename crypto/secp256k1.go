package crypto

import (
	"bytes"
	"crypto/ecdsa"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	ethsecp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"

	tmcrypto "github.com/tendermint/tendermint/crypto"
)

// ----------------------------------------------------------------------------
// secp256k1 Private Key

var _ tmcrypto.PrivKey = PrivKeySecp256k1{}

// PrivKeySecp256k1 defines a type alias for an ecdsa.PrivateKey that implements
// Tendermint's PrivateKey interface.
type PrivKeySecp256k1 ecdsa.PrivateKey

// GenerateKey generates a new random private key. It returns an error upon
// failure.
func GenerateKey() (PrivKeySecp256k1, error) {
	priv, err := ethcrypto.GenerateKey()
	if err != nil {
		return PrivKeySecp256k1{}, err
	}

	return PrivKeySecp256k1(*priv), nil
}

// PubKey returns the ECDSA private key's public key.
func (privkey PrivKeySecp256k1) PubKey() tmcrypto.PubKey {
	return PubKeySecp256k1{privkey.PublicKey}
}

// Bytes returns the raw ECDSA private key bytes.
func (privkey PrivKeySecp256k1) Bytes() []byte {
	return ethcrypto.FromECDSA(privkey.ToECDSA())
}

// Sign creates a recoverable ECDSA signature on the secp256k1 curve over the
// Keccak256 hash of the provided message. The produced signature is 65 bytes
// where the last byte contains the recovery ID.
func (privkey PrivKeySecp256k1) Sign(msg []byte) ([]byte, error) {
	return ethcrypto.Sign(ethcrypto.Keccak256Hash(msg).Bytes(), privkey.ToECDSA())
}

// Equals returns true if two ECDSA private keys are equal and false otherwise.
func (privkey PrivKeySecp256k1) Equals(other tmcrypto.PrivKey) bool {
	if other, ok := other.(PrivKeySecp256k1); ok {
		return bytes.Equal(privkey.Bytes(), other.Bytes())
	}

	return false
}

// ToECDSA returns the ECDSA private key as a reference to ecdsa.PrivateKey type.
func (privkey PrivKeySecp256k1) ToECDSA() *ecdsa.PrivateKey {
	return (*ecdsa.PrivateKey)(&privkey)
}

// ----------------------------------------------------------------------------
// secp256k1 Public Key

var _ tmcrypto.PubKey = (*PubKeySecp256k1)(nil)

// PubKeySecp256k1 defines a type alias for an ecdsa.PublicKey that implements
// Tendermint's PubKey interface.
type PubKeySecp256k1 struct {
	pubkey ecdsa.PublicKey
}

// Address returns the address of the ECDSA public key.
func (key PubKeySecp256k1) Address() tmcrypto.Address {
	return tmcrypto.Address(ethcrypto.PubkeyToAddress(key.pubkey).Bytes())
}

// Bytes returns the raw bytes of the ECDSA public key.
func (key PubKeySecp256k1) Bytes() []byte {
	return ethcrypto.FromECDSAPub(&key.pubkey)
}

// VerifyBytes verifies that the ECDSA public key created a given signature over
// the provided message. It will calculate the Keccak256 hash of the message
// prior to verification.
func (key PubKeySecp256k1) VerifyBytes(msg []byte, sig []byte) bool {
	if len(sig) == 65 {
		// remove recovery ID if contained in the signature
		sig = sig[:len(sig)-1]
	}

	// the signature needs to be in [R || S] format when provided to VerifySignature
	return ethsecp256k1.VerifySignature(key.Bytes(), ethcrypto.Keccak256Hash(msg).Bytes(), sig)
}

// Equals returns true if two ECDSA public keys are equal and false otherwise.
func (key PubKeySecp256k1) Equals(other tmcrypto.PubKey) bool {
	if other, ok := other.(PubKeySecp256k1); ok {
		return bytes.Equal(key.Bytes(), other.Bytes())
	}

	return false
}
