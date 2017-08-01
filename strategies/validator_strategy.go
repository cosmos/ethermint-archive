package strategies

import (
	"github.com/ethereum/go-ethereum/common"

	abci "github.com/tendermint/abci/types"
)

type ValidatorStrategy struct {
	validators []*abci.Validator
}

func NewValidatorStrategy(validators []*abci.Validator) ValidatorStrategy {
	return ValidatorStrategy{validators: validators}
}

func (v ValidatorStrategy) SetValidators(validators []*abci.Validator) {
	v.validators = validators
}

func (v ValidatorStrategy) Validators() []*abci.Validator {
	return v.validators
}

func (v ValidatorStrategy) Beneficiary() common.Address {
	return common.HexToAddress("7ef5a6135f1fd6a02593eedc869c6d41d934aef8")
}
