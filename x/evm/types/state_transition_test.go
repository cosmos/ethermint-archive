package types_test

import (
	"math/big"

	"github.com/cosmos/ethermint/crypto"
	"github.com/cosmos/ethermint/x/evm/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func (suite *StateDBTestSuite) TestTransitionDb() {
	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)
	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)
	nonce := uint64(123)
	balance := new(big.Int).SetUint64(5000)
	suite.stateDB.SetNonce(addr, nonce)
	suite.stateDB.AddBalance(addr, balance)

	priv, err = crypto.GenerateKey()
	suite.Require().NoError(err)
	recipient := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)

	testCase := []struct {
		name     string
		malleate func(state types.StateTransition)
		state    types.StateTransition
	}{
		{
			"passing test",
			func(state types.StateTransition) {
				_, err = state.TransitionDb(suite.ctx)
				suite.Require().NoError(err)
				fromBalance := suite.stateDB.GetBalance(addr)
				toBalance := suite.stateDB.GetBalance(recipient)
				suite.Require().Equal(fromBalance, new(big.Int).SetUint64(4950))
				suite.Require().Equal(toBalance, new(big.Int).SetUint64(50))
			},
			types.StateTransition{
				AccountNonce: nonce,
				Price:        new(big.Int).SetUint64(10),
				GasLimit:     11,
				Recipient:    &recipient,
				Amount:       new(big.Int).SetUint64(50),
				Payload:      []byte("data"),
				ChainID:      new(big.Int).SetUint64(1),
				Csdb:         suite.stateDB,
				TxHash:       &ethcmn.Hash{},
				Sender:       addr,
				Simulate:     suite.ctx.IsCheckTx(),
			},
		},
		{
			"send more than balance",
			func(state types.StateTransition) {
				_, err = state.TransitionDb(suite.ctx)
				suite.Require().Error(err)

			},
			types.StateTransition{
				AccountNonce: nonce,
				Price:        new(big.Int).SetUint64(10),
				GasLimit:     11,
				Recipient:    &recipient,
				Amount:       new(big.Int).SetUint64(4951),
				Payload:      []byte("data"),
				ChainID:      new(big.Int).SetUint64(1),
				Csdb:         suite.stateDB,
				TxHash:       &ethcmn.Hash{},
				Sender:       addr,
				Simulate:     suite.ctx.IsCheckTx(),
			},
		},
	}

	for _, tc := range testCase {
		tc.malleate(tc.state)
	}
}
