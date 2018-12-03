package types

import (
	"github.com/stretchr/testify/suite"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/codec"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/cosmos/cosmos-sdk/store"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"testing"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"crypto/rand"
	"github.com/stretchr/testify/require"
)

type bkTestSuite struct {
	suite.Suite
	ctx    types.Context
	keeper *BlockKeeper
}

func (s *bkTestSuite) SetupSuite() {
	cdc := codec.New()
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	key := types.NewKVStoreKey("bk")
	ms.MountStoreWithDB(key, types.StoreTypeIAVL, nil)
	err := ms.LoadLatestVersion()
	require.NoError(s.T(), err)
	s.ctx = types.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	s.keeper = NewBlockKeeper(key, cdc)
}

func (s *bkTestSuite) TestGetBlockByNumber() {
	block := randomBlock(s.T())
	s.keeper.SetBlock(s.ctx, block)
	retBlock := s.keeper.GetBlockByNumber(s.ctx, block.NumberU64())
	require.Equal(s.T(), block.Hash(), retBlock.Hash())
	require.Equal(s.T(), block.NumberU64(), retBlock.NumberU64())
}

func (s *bkTestSuite) TestGetBlockByHash() {
	block := randomBlock(s.T())
	s.keeper.SetBlock(s.ctx, block)
	retBlock := s.keeper.GetBlockByHash(s.ctx, block.Hash())
	require.Equal(s.T(), block.Hash(), retBlock.Hash())
	require.Equal(s.T(), block.NumberU64(), retBlock.NumberU64())
}

func TestBlockKeeper(t *testing.T) {
	suite.Run(t, new(bkTestSuite))
}

func randomBlock(t *testing.T) *ethtypes.Block {
	var nonce ethtypes.BlockNonce
	copy(nonce[:], randBytes(t, 8))
	num, err := rand.Int(rand.Reader, big.NewInt(10000))
	require.NoError(t, err)
	return ethtypes.NewBlock(&ethtypes.Header{
		Number: num,
		Nonce:  nonce,
	}, nil, nil, nil)
}

func randBytes(t *testing.T, len int) []byte {
	buf := make([]byte, len, len)
	_, err := rand.Read(buf)
	require.NoError(t, err)
	return buf
}
