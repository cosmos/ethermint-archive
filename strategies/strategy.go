package strategies

import (
	"github.com/ethereum/go-ethereum/common"

	abci "github.com/tendermint/abci/types"
)

type Strategy interface {
	// SetValidators updates the underlying validator set.
	SetValidators(validators []*abci.Validator)
	// Validators returns the current validator set.
	Validators() []*abci.Validator
	// Beneficiary returns who the money should go to.
	Beneficiary() common.Address

	// TODO: Consider adding Author() to allow people to separate between
	// beneficiary and author
}
