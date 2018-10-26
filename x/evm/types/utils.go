package types

import (
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethsha "github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/rlp"
)

func rlpHash(x interface{}) (hash ethcmn.Hash) {
	hasher := ethsha.NewKeccak256()

	rlp.Encode(hasher, x)
	hasher.Sum(hash[:0])

	return
}
