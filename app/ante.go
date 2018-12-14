package app

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/ethermint/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcore "github.com/ethereum/go-ethereum/core"
)

// NewAnteHandler returns an ante handler responsible for attempting to route an
// Ethereum or SDK transaction to an internal ante handler for performing
// transaction-level processing (e.g. fee payment, signature verification) before
// being passed onto it's respective handler.
//
// NOTE: The EVM will already consume (intrinsic) gas for signature verification
// and covering input size as well as handling nonce incrementing.
func NewAnteHandler(ak auth.AccountKeeper, fck auth.FeeCollectionKeeper) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {

		switch castTx := tx.(type) {
		case auth.StdTx:
			return auth.NewAnteHandler(ak, fck)(ctx, castTx, sim)

		case *evmtypes.EthereumTxMsg:
			return ethAnteHandler(ctx, castTx, ak)

		default:
			return ctx, sdk.ErrInternal(fmt.Sprintf("transaction type invalid: %T", tx)).Result(), true
		}
	}
}

// ----------------------------------------------------------------------------
// Ethereum Ante Handler

// ethAnteHandler defines an internal ante handler for an Ethereum transaction
// ethTxMsg. During CheckTx, the transaction is passed through a series of
// pre-message execution validation checks such as signature and account
// verification in addition to minimum fees being checked. Otherwise, during
// DeliverTx, the transaction is simply passed to the EVM which will also
// perform the same series of checks. The distinction is made in CheckTx to
// prevent spam and DoS attacks.
func ethAnteHandler(
	ctx sdk.Context, ethTxMsg *evmtypes.EthereumTxMsg, ak auth.AccountKeeper,
) (newCtx sdk.Context, res sdk.Result, abort bool) {

	if ctx.IsCheckTx() {
		// Only perform pre-message (Ethereum transaction) execution validation
		// during CheckTx. Otherwise, during DeliverTx the EVM will handle them.
		if res := validateEthTxCheckTx(ctx, ak, ethTxMsg); !res.IsOK() {
			return newCtx, res, true
		}
	}

	return ctx, sdk.Result{}, false
}

// ----------------------------------------------------------------------------
// Auxiliary

func validateEthTxCheckTx(
	ctx sdk.Context, ak auth.AccountKeeper, ethTxMsg *evmtypes.EthereumTxMsg,
) sdk.Result {

	// parse the chainID from a string to a base-10 integer
	chainID, ok := new(big.Int).SetString(ctx.ChainID(), 10)
	if !ok {
		return types.ErrInvalidChainID(fmt.Sprintf("invalid chainID: %s", ctx.ChainID())).Result()
	}

	// Validate sufficient fees have been provided that meet a minimum threshold
	// defined by the proposer (for mempool purposes during CheckTx).
	if res := ensureSufficientMempoolFees(ctx, ethTxMsg); !res.IsOK() {
		return res
	}

	// validate enough intrinsic gas
	if res := validateIntrinsicGas(ethTxMsg); !res.IsOK() {
		return res
	}

	// validate sender/signature
	signer, err := ethTxMsg.VerifySig(chainID)
	if err != nil {
		return sdk.ErrUnauthorized("signature verification failed").Result()
	}

	// validate account (nonce and balance checks)
	if res := validateAccount(ctx, ak, ethTxMsg, signer); !res.IsOK() {
		return res
	}

	return sdk.Result{}
}

// validateIntrinsicGas validates that the Ethereum tx message has enough to
// cover intrinsic gas.
func validateIntrinsicGas(ethTxMsg *evmtypes.EthereumTxMsg) sdk.Result {
	gas, err := ethcore.IntrinsicGas(ethTxMsg.Data.Payload, ethTxMsg.To() == nil, false)
	if err != nil {
		return sdk.ErrInternal(fmt.Sprintf("failed to compute intrinsic gas cost: %s", err)).Result()
	}

	if ethTxMsg.Data.GasLimit < gas {
		return sdk.ErrInternal(
			fmt.Sprintf("intrinsic gas too low; %d < %d", ethTxMsg.Data.GasLimit, gas),
		).Result()
	}

	return sdk.Result{}
}

// validateAccount validates the account nonce and that the account has enough
// funds to cover the tx cost.
func validateAccount(
	ctx sdk.Context, ak auth.AccountKeeper, ethTxMsg *evmtypes.EthereumTxMsg, signer ethcmn.Address,
) sdk.Result {

	acc := ak.GetAccount(ctx, sdk.AccAddress(signer.Bytes()))

	// on InitChain make sure account number == 0
	if ctx.BlockHeight() == 0 && acc.GetAccountNumber() != 0 {
		return sdk.ErrInternal(
			fmt.Sprintf(
				"invalid account number for height zero; got %d, expected 0", acc.GetAccountNumber(),
			)).Result()
	}

	seq := acc.GetSequence()
	if ethTxMsg.Data.AccountNonce < seq {
		return sdk.ErrInvalidSequence(
			fmt.Sprintf("nonce too low; got %d, expected >= %d", ethTxMsg.Data.AccountNonce, seq)).Result()
	}

	// validate sender has enough funds
	balance := acc.GetCoins().AmountOf(types.DenomDefault)
	if balance.BigInt().Cmp(ethTxMsg.Cost()) < 0 {
		return sdk.ErrInsufficientFunds(
			fmt.Sprintf("insufficient funds: %s < %s", balance, ethTxMsg.Cost()),
		).Result()
	}

	return sdk.Result{}
}

// ensureSufficientMempoolFees verifies that enough fees have been provided by the
// Ethereum transaction that meet the minimum threshold set by the block
// proposer.
//
// NOTE: This should only be ran during a CheckTx mode.
func ensureSufficientMempoolFees(ctx sdk.Context, ethTxMsg *evmtypes.EthereumTxMsg) sdk.Result {
	// fee = GP * GL
	feeAmt := new(big.Int).Mul(ethTxMsg.Data.Price, new(big.Int).SetUint64(ethTxMsg.Data.GasLimit))
	fee := sdk.Coins{sdk.NewInt64Coin(types.DenomDefault, feeAmt.Int64())}

	if !ctx.MinimumFees().IsZero() && !fee.IsAllGTE(ctx.MinimumFees()) {
		// reject the transaction that does not meet the minimum fee
		return sdk.ErrInsufficientFee(
			fmt.Sprintf("insufficient fee, got: %q required: %q", fee, ctx.MinimumFees()),
		).Result()
	}

	return sdk.Result{}
}
