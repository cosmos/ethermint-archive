package app

import (
	"bytes"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	evmtypes "github.com/cosmos/ethermint/x/evm/types"
)

var dummySecp256k1Pubkey secp256k1.PubKeySecp256k1 // used for tx simulation

const (
	memoCostPerByte     = 1
	maxMemoCharacters   = 100
	secp256k1VerifyCost = 100
)

func init() {
	bz, _ := hex.DecodeString("035AD6810A47F073553FF30D2FCC7E0D3B1C0B74B61A1AAA2582344037151E143A")
	copy(dummySecp256k1Pubkey[:], bz)
}

// NewAnteHandler returns an ante handelr responsible for attempting to route an
// Ethereum or SDK transaction to an internal ante handler for performing
// transaction-level processing (e.g. fee payment, signature verification) before
// being passed onto it's respective handler.
//
// NOTE: The EVM will already consume (intrinsic) gas for signature verification
// and covering input size as well as handling nonce incrementing.
func NewAnteHandler(ak auth.AccountKeeper, _ auth.FeeCollectionKeeper) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {

		stdTx, ok := tx.(auth.StdTx)
		if !ok {
			return ctx, sdk.ErrInternal("transaction type invalid: must be StdTx").Result(), true
		}

		// TODO: Handle gas/fee checking and spam prevention. We may need two
		// different models for SDK and Ethereum txs. The SDK currently supports a
		// primitive model where a constant gas price is used.
		//
		// Ref: #473

		if ethTx, ok := isEthereumTx(stdTx); ethTx != nil && ok {
			return ethAnteHandler(ctx, ethTx, ak)
		}

		// NOTE: We could use the SDK's ante handler, but differences may exist in
		// gas costs (e.g. signature verification). So for now, implement a similar
		// ante handler flow.
		return defaultAnteHandler(ctx, stdTx, ak, sim)
	}
}

// ----------------------------------------------------------------------------
// Ethereum Ante Handler

func ethAnteHandler(
	ctx sdk.Context, ethTx *evmtypes.MsgEthereumTx, ak auth.AccountKeeper,
) (newCtx sdk.Context, res sdk.Result, abort bool) {

	// For now we simply pass the transaction on as the EVM shares common business
	// logic of an ante handler. Anything not handled by the EVM that should be
	// prior to transaction processing, should be done here.
	return ctx, sdk.Result{}, false
}

// ----------------------------------------------------------------------------
// Default Ante Handler

func defaultAnteHandler(
	ctx sdk.Context, stdTx auth.StdTx, ak auth.AccountKeeper, sim bool,
) (newCtx sdk.Context, res sdk.Result, abort bool) {

	newCtx = newCtxWithGasMeter(ctx, stdTx, sim)

	// Recover in case the transaction ran out of gas and can be reported to the
	// BaseApp.
	defer func() {
		if r := recover(); r != nil {
			switch rType := r.(type) {
			case sdk.ErrorOutOfGas:
				log := fmt.Sprintf("out of gas in location: %v", rType.Descriptor)
				res = sdk.ErrOutOfGas(log).Result()
				res.GasWanted = stdTx.Fee.Gas
				res.GasUsed = newCtx.GasMeter().GasConsumed()
				abort = true
			default:
				panic(r)
			}
		}
	}()

	err := validateBasic(stdTx)
	if err != nil {
		return newCtx, err.Result(), true
	}

	newCtx.GasMeter().ConsumeGas(
		memoCostPerByte*sdk.Gas(len(stdTx.GetMemo())),
		"memo",
	)

	stdSigs := stdTx.GetSignatures() // when simulating, the slice is empty
	signerAddrs := stdTx.GetSigners()

	signerAccs, res := getSignerAccounts(newCtx, ak, signerAddrs)
	if !res.IsOK() {
		return newCtx, res, true
	}

	res = validateAccount(ctx, signerAccs, stdSigs)
	if !res.IsOK() {
		return newCtx, res, true
	}

	// TODO: fee collection

	signBytesList := collectSignBytes(newCtx.ChainID(), stdTx, stdSigs)
	for i := 0; i < len(stdSigs); i++ {
		// check signature which returns an updated account with an incremented
		// sequence (nonce)
		signerAccs[i], res = processSig(newCtx, signerAccs[i], stdSigs[i], signBytesList[i], sim)
		if !res.IsOK() {
			return newCtx, res, true
		}

		ak.SetAccount(newCtx, signerAccs[i])
	}

	return newCtx, sdk.Result{GasWanted: stdTx.Fee.Gas}, false
}

// ----------------------------------------------------------------------------
// Auxiliary

// validateBasic performs basic validation with low overhead that depend on a
// context.
func validateBasic(tx auth.StdTx) (err sdk.Error) {
	// assert that there are signatures
	sigs := tx.GetSignatures()
	if len(sigs) == 0 {
		return sdk.ErrUnauthorized("no signers found")
	}

	// assert that number of signatures is correct
	var signerAddrs = tx.GetSigners()
	if len(sigs) != len(signerAddrs) {
		return sdk.ErrUnauthorized("invalid number of signers")
	}

	// validate memo does not exceed the maximum allowed number of characters
	memo := tx.GetMemo()
	if len(memo) > maxMemoCharacters {
		return sdk.ErrMemoTooLarge(
			fmt.Sprintf(
				"maximum number of characters exceed in memo; got: %d, max: %d",
				maxMemoCharacters, len(memo),
			),
		)
	}

	return nil
}

func validateAccount(ctx sdk.Context, accs []auth.Account, sigs []auth.StdSignature) sdk.Result {
	for i := 0; i < len(accs); i++ {
		// the account number must be zero on InitChain
		if ctx.BlockHeight() == 0 && sigs[i].AccountNumber != 0 {
			return sdk.ErrInvalidSequence(
				fmt.Sprintf(
					"invalid account number for block height 0; got %d, expected 0",
					sigs[i].AccountNumber,
				),
			).Result()
		}

		// validate the account number
		accnum := accs[i].GetAccountNumber()
		if ctx.BlockHeight() != 0 && accnum != sigs[i].AccountNumber {
			return sdk.ErrInvalidSequence(
				fmt.Sprintf(
					"invalid account number; got %d, expected %d",
					sigs[i].AccountNumber, accnum,
				),
			).Result()
		}

		// validate the sequence number
		seq := accs[i].GetSequence()
	
		if seq != sigs[i].Sequence {
			return sdk.ErrInvalidSequence(
				fmt.Sprintf(
					"invalid sequence; got %d, expected %d",
					sigs[i].Sequence, seq,
				),
			).Result()
		}
	}

	return sdk.Result{}
}

func newCtxWithGasMeter(ctx sdk.Context, stdTx auth.StdTx, sim bool) sdk.Context {
	if sim || ctx.BlockHeight() == 0 {
		return ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	}

	return ctx.WithGasMeter(sdk.NewGasMeter(stdTx.Fee.Gas))
}

func getSignerAccounts(
	ctx sdk.Context, ak auth.AccountKeeper, addrs []sdk.AccAddress,
) ([]auth.Account, sdk.Result) {

	accs := make([]auth.Account, len(addrs))
	for i := 0; i < len(accs); i++ {
		accs[i] = ak.GetAccount(ctx, addrs[i])
		if accs[i] == nil {
			return nil, sdk.ErrUnknownAddress(addrs[i].String()).Result()
		}
	}

	return accs, sdk.Result{}
}

func collectSignBytes(chainID string, stdTx auth.StdTx, stdSigs []auth.StdSignature) [][]byte {
	sigBytesList := make([][]byte, len(stdSigs))
	for i := 0; i < len(stdSigs); i++ {
		sigBytesList[i] = auth.StdSignBytes(
			chainID, stdSigs[i].AccountNumber, stdSigs[i].Sequence,
			stdTx.Fee, stdTx.Msgs, stdTx.Memo,
		)
	}

	return sigBytesList
}

// processSig verifies a transaction's signature for a given signer/account and
// increments the sequence (nonce). In addition, if the account will have it's
// public key set if has not already been set.
func processSig(
	ctx sdk.Context, acc auth.Account, sig auth.StdSignature, signBytes []byte, sim bool,
) (updatedAcc auth.Account, res sdk.Result) {

	pubKey, res := processPubKey(acc, sig, sim)
	if !res.IsOK() {
		return nil, res
	}

	if err := acc.SetPubKey(pubKey); err != nil {
		return nil, sdk.ErrInternal("failed setting public key on signer's account").Result()
	}

	consumeSigVerificationGas(ctx.GasMeter(), pubKey)
	if !sim && !pubKey.VerifyBytes(signBytes, sig.Signature) {
		return nil, sdk.ErrUnauthorized("failed to verify signature").Result()
	}

	if err := acc.SetSequence(acc.GetSequence() + 1); err != nil {
		return nil, sdk.ErrInternal("failed incrementing sequence on signer's account").Result()
	}

	return acc, res
}

// processPubKey verifies that the given account's public key matches the public
// key from either the given signature, if the account does not have it set, or
// directly from the account. During simulation, no verification is done and a
// dummy public key is returned.
func processPubKey(acc auth.Account, sig auth.StdSignature, sim bool) (crypto.PubKey, sdk.Result) {
	pubKey := acc.GetPubKey()
	if sim {
		// In simulate mode the transaction comes with no signatures, thus if the
		// account's public key is nil, both signature verification and
		// gasKVStore.Set() shall consume the largest amount.
		if pubKey == nil {
			return dummySecp256k1Pubkey, sdk.Result{}
		}

		return pubKey, sdk.Result{}
	}

	if pubKey == nil {
		pubKey = sig.PubKey
		if pubKey == nil {
			return nil, sdk.ErrInvalidPubKey("failed to find account public key").Result()
		}

		if !bytes.Equal(pubKey.Address(), acc.GetAddress()) {
			return nil, sdk.ErrInvalidPubKey(
				fmt.Sprintf("public key does not match signer address %s", acc.GetAddress())).Result()
		}
	}

	return pubKey, sdk.Result{}
}

// TODO:
// - Ethereum uses the secp256k1 elliptic curve. Should we support others?
// - Signature verification cost should be consistent with the EVM.
func consumeSigVerificationGas(meter sdk.GasMeter, pubkey crypto.PubKey) {
	switch pubkey.(type) {
	case secp256k1.PubKeySecp256k1:
		meter.ConsumeGas(secp256k1VerifyCost, "secp256k1 signature verification")
	default:
		panic(fmt.Sprintf("unrecognized signature type: %T", pubkey))
	}
}

// isEthereumTx returns a boolean if a given standard SDK transaction contains
// an Ethereum transaction. If so, the transaction is also returned. A standard
// SDK transaction contains an Ethereum transaction if it only has a single
// message and that embedded message if of type MsgEthereumTx.
func isEthereumTx(tx auth.StdTx) (*evmtypes.MsgEthereumTx, bool) {
	msgs := tx.GetMsgs()
	if len(msgs) == 1 {
		ethTx, ok := msgs[0].(*evmtypes.MsgEthereumTx)
		if ok {
			return ethTx, true
		}
	}

	return nil, false
}
