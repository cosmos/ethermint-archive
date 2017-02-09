package validatorStrategies

import (
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/tendermint/abci/types"
)

type TxBasedValidatorsStrategy struct {
	currentValidators []*types.Validator
}

func (strategy *TxBasedValidatorsStrategy) SetValidators(validators []*types.Validator) {
	strategy.currentValidators = validators
}

func (strategy *TxBasedValidatorsStrategy) CollectTx(tx *ethTypes.Transaction) {
	if reflect.DeepEqual(tx.To(), common.HexToAddress("0000000000000000000000000000000000000001")) {
		glog.V(logger.Debug).Infof("Adding validator: %v", tx.Data())
		strategy.currentValidators = append(
			strategy.currentValidators,
			&types.Validator{
				PubKey: tx.Data(),
				Power:  tx.Value().Uint64(),
			},
		)
	}
}

func (strategy *TxBasedValidatorsStrategy) GetUpdatedValidators() []*types.Validator {
	return strategy.currentValidators
}
