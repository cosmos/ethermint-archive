package types

import (
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tendermint/ethermint/backend"
	"github.com/tendermint/tmsp/types"
)

type MinerRewardStrategy interface {
	Receiver(app TMSPEthereumApplicationInterface) common.Address
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
type TMSPEthereumApplicationInterface interface {
	Backend() *backend.TMSPEthereumBackend
	Strategy() *Strategy
}
