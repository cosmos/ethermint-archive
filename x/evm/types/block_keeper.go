package types

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/rawdb"
)

// BlockKeeper persists blocks in a manner consistent with how
// the RPC latyer will query them.
type BlockKeeper struct {
	// The key used to access the store from the Context.
	key types.StoreKey

	// The codec for binary encoding/decoding of blocks.
	cdc *codec.Codec
}

func NewBlockKeeper(key types.StoreKey, cdc *codec.Codec) *BlockKeeper {
	return &BlockKeeper{
		key: key,
		cdc: cdc,
	}
}

// GetBlockByNumber returns a persisted Ethereum block by its number.
func (bk *BlockKeeper) GetBlockByNumber(ctx types.Context, blockNum uint64) *ethtypes.Block {
	wrapper := bk.wrapKVStore(ctx.KVStore(bk.key))
	hash := rawdb.ReadCanonicalHash(wrapper, blockNum)
	return bk.getBlock(wrapper, hash, blockNum)
}

// GetBlockByHash returns a persisted Ethereum block by its hash.
func (bk *BlockKeeper) GetBlockByHash(ctx types.Context, hash common.Hash) *ethtypes.Block {
	wrapper := bk.wrapKVStore(ctx.KVStore(bk.key))
	number := *rawdb.ReadHeaderNumber(wrapper, hash)
	return bk.getBlock(wrapper, hash, number)
}

// SetBlock persists an Ethereum block.
// TODO @mslipper: Transaction receipts need to be included separately.
func (bk *BlockKeeper) SetBlock(ctx types.Context, block *ethtypes.Block) {
	store := ctx.KVStore(bk.key)
	wrapper := bk.wrapKVStore(store)
	rawdb.WriteCanonicalHash(wrapper, block.Hash(), block.NumberU64())
	rawdb.WriteHeader(wrapper, block.Header())
	rawdb.WriteBlock(wrapper, block)
}

func (bk *BlockKeeper) getBlock(wrapper rawDBWrapper, hash common.Hash, number uint64) *ethtypes.Block {
	return rawdb.ReadBlock(wrapper, hash, number)
}

func (bk *BlockKeeper) wrapKVStore(store types.KVStore) rawDBWrapper {
	return &wrapper{store, bk}
}

// rawDBWrapper is a composite interface that wraps Geth's
// DatabaseReader/Writer/Deleter interfaces. It allows Keeper
// instances to work with Geth's rawdb package.
type rawDBWrapper interface {
	rawdb.DatabaseReader
	rawdb.DatabaseWriter
	rawdb.DatabaseDeleter
}

// wrapper takes a Keeper and a KVStore and creates a
// rawDBWrapper-compatible data store.
type wrapper struct {
	store  types.KVStore
	keeper *BlockKeeper
}

// Get gets a value for a given key from the backing store. Implements
// Geth's DatabaseReader interface for lookups.
func (w *wrapper) Get(key []byte) (value []byte, err error) {
	return w.store.Get(key), nil
}

// Has checks for a given key 's existence in the backing store. Implements
// Geth's DatabaseReader interface for lookups.
func (w *wrapper) Has(key []byte) (bool, error) {
	return w.store.Has(key), nil
}

// Put inserts a given key into the backing store. Implements Geth's
// DatabaseWriter interface for persistence.
func (w *wrapper) Put(key []byte, value []byte) error {
	w.store.Set(key, value)
	return nil
}

func (w *wrapper) Delete(key []byte) error {
	w.store.Delete(key)
	return nil
}
