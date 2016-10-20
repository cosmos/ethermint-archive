package processor

import (
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// NullBlockProcessor does not validate anything
type NullBlockProcessor struct{}

// ValidateBlock does not validate anything
func (NullBlockProcessor) ValidateBlock(*ethTypes.Block) error { return nil }

// ValidateHeader does not validate anything
func (NullBlockProcessor) ValidateHeader(*ethTypes.Header, *ethTypes.Header, bool) error { return nil }

// ValidateState does not validate anything
func (NullBlockProcessor) ValidateState(block, parent *ethTypes.Block, state *state.StateDB, receipts ethTypes.Receipts, usedGas *big.Int) error {
	return nil
}
