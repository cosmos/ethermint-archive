package validatorStrategies

import (
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/tendermint/abci/types"
)

// TxBasedValidatorsStrategy represents a strategy to reward validators with CETH
type TxBasedValidatorsStrategy struct {
	currentValidators []*types.Validator
}

// SetValidators updates the current validators
func (strategy *TxBasedValidatorsStrategy) SetValidators(validators []*types.Validator) {
	strategy.currentValidators = validators
}

// CollectTx collects the rewards for a transaction
func (strategy *TxBasedValidatorsStrategy) CollectTx(tx *ethTypes.Transaction) {
	if reflect.DeepEqual(tx.To(), common.HexToAddress("0000000000000000000000000000000000000001")) {
		log.Info("Adding validator", "data", tx.Data())
		strategy.currentValidators = append(
			strategy.currentValidators,
			&types.Validator{
				PubKey: tx.Data(),
				Power:  tx.Value().Uint64(),
			},
		)
	}
}

// GetUpdatedValidators returns the current validators
func (strategy *TxBasedValidatorsStrategy) GetUpdatedValidators() []*types.Validator {
	return strategy.currentValidators
}
