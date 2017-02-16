package types

import (
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/tendermint/abci/types"
)

type MinerRewardStrategy interface {
	Receiver() common.Address
}

type ValidatorsStrategy interface {
	SetValidators(validators []*types.Validator)
	CollectTx(tx *ethTypes.Transaction)
	GetUpdatedValidators() []*types.Validator
}

type Strategy struct {
	MinerRewardStrategy
	ValidatorsStrategy
}
