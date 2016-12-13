package types

import (
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tendermint/tmsp/types"
)

type MinerRewardStrategy interface {
	Receiver(app EthermintApplicationInterface) common.Address
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
type EthermintApplicationInterface interface {
	Backend() *EthereumBackend
	Strategy() *Strategy
}
