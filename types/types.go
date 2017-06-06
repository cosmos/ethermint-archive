package types

import (
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/tendermint/abci/types"
)

// MinerRewardStrategy is a mining strategy
type MinerRewardStrategy interface {
	Receiver() common.Address
}

// ValidatorsStrategy is a validator strategy
type ValidatorsStrategy interface {
	SetValidators(validators []*types.Validator)
	CollectTx(tx *ethTypes.Transaction)
	GetUpdatedValidators() []*types.Validator
}

// Strategy encompasses all available strategies
type Strategy struct {
	MinerRewardStrategy
	ValidatorsStrategy
}
