package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Ethermint error codes
const (
	// DefaultCodespace reserves a Codespace for Ethermint.
	DefaultCodespace sdk.CodespaceType = "ethermint"

	CodeInvalidValue         sdk.CodeType = 1
	CodeInvalidAccountNumber sdk.CodeType = 2
)

func codeToDefaultMsg(code sdk.CodeType) string {
	switch code {
	case CodeInvalidValue:
		return "invalid value"
	case CodeInvalidAccountNumber:
		return "invalid account number"
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

// ErrInvalidValue returns a standardized SDK error resulting from an invalid
// value.
func ErrInvalidValue(msg string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidValue, msg)
}

// ErrInvalidAccountNumber returns a standardized SDK error resulting from an
// invalid account number.
func ErrInvalidAccountNumber(msg string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidAccountNumber, msg)
}
