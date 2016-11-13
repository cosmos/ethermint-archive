package minerRewardStrategies

import (
	"github.com/ethereum/go-ethereum/common"
	tmspEthTypes "github.com/kobigurk/tmsp-ethereum/types"
)

type RewardConstant struct {
}

func (rewardStrategy *RewardConstant) Receiver(app tmspEthTypes.TMSPEthereumApplicationInterface) common.Address {
	return common.HexToAddress("7ef5a6135f1fd6a02593eedc869c6d41d934aef8")
}
