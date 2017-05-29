package minerRewardStrategies

import (
	"github.com/ethereum/go-ethereum/common"
)

type RewardConstant struct {
}

func (r *RewardConstant) Receiver() common.Address {
	return common.HexToAddress("7ef5a6135f1fd6a02593eedc869c6d41d934aef8")
}
