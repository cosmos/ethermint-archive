package types

import (
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"

	sdkcodec "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/ethermint/codec"
	"github.com/cosmos/ethermint/crypto"
	ethermint "github.com/cosmos/ethermint/types"
)

type StateTransitionTestSuite struct {
	suite.Suite

	address ethcmn.Address
	ctx     sdk.Context
	stateDB *CommitStateDB
}

func newTestCodec() *codec.Codec {
	cdc := sdkcodec.New()

	RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	crypto.RegisterCodec(cdc)
	sdkcodec.RegisterCrypto(cdc)
	ethermint.RegisterCodec(cdc)

	appCodec := codec.NewAppCodec(cdc)

	return appCodec
}

func (suite *StateTransitionTestSuite) TestTransitionDb(t *testing.T) {
	nonce := uint64(123)
	suite.stateDB.SetNonce(suite.address, nonce)

	priv, err := crypto.GenerateKey()
	recipient := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)

	state := StateTransition{
		AccountNonce: nonce,
		Price:        new(big.Int).SetUint64(1),
		GasLimit:     5,
		Recipient:    &recipient,
		Amount:       new(big.Int).SetUint64(25),
		Payload:      []byte("data"),
		ChainID:      new(big.Int).SetUint64(1),
		Csdb:         suite.stateDB,
		TxHash:       &ethcmn.Hash{},
		Sender:       suite.address,
		Simulate:     suite.ctx.IsCheckTx(),
	}

	result, err := state.TransitionDb(suite.ctx)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}

func (suite *StateTransitionTestSuite) SetupTest() {
	authKey := sdk.NewKVStoreKey(auth.StoreKey)
	bankKey := sdk.NewKVStoreKey(bank.StoreKey)
	storeKey := sdk.NewKVStoreKey(StoreKey)

	db := tmdb.NewDB("state", tmdb.GoLevelDBBackend, "temp")
	defer func() {
		os.RemoveAll("temp")
	}()

	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(authKey, sdk.StoreTypeIAVL, db)
	cms.MountStoreWithDB(bankKey, sdk.StoreTypeIAVL, db)
	cms.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)

	err := cms.LoadLatestVersion()
	suite.Require().NoError(err)

	appCodec := newTestCodec()

	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	paramsKeeper := params.NewKeeper(appCodec, keyParams, tkeyParams)

	authSubspace := paramsKeeper.Subspace(auth.DefaultParamspace)
	bankSubspace := paramsKeeper.Subspace(bank.DefaultParamspace)

	ak := auth.NewAccountKeeper(appCodec, authKey, authSubspace, ethermint.ProtoAccount)
	bk := bank.NewBaseKeeper(appCodec, bankKey, ak, bankSubspace, nil)

	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)

	suite.ctx = sdk.NewContext(cms, abci.Header{ChainID: "8"}, false, tmlog.NewNopLogger())
	suite.stateDB = NewCommitStateDB(suite.ctx, storeKey, ak, bk).WithContext(suite.ctx)
	suite.address = ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)
}
