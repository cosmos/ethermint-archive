package importer

import (
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/signal"
	"runtime/pprof"
	"sort"
	"syscall"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/ethermint/core"
	"github.com/cosmos/ethermint/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	ethmisc "github.com/ethereum/go-ethereum/consensus/misc"
	ethcore "github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethvm "github.com/ethereum/go-ethereum/core/vm"
	ethparams "github.com/ethereum/go-ethereum/params"
	ethrlp "github.com/ethereum/go-ethereum/rlp"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmlog "github.com/tendermint/tendermint/libs/log"
)

var (
	flagDataDir    string
	flagBlockchain string
	flagCPUProfile string

	miner501    = ethcmn.HexToAddress("0x35e8e5dC5FBd97c5b421A80B596C030a2Be2A04D")
	genInvestor = ethcmn.HexToAddress("0x756F45E3FA69347A9A973A725E3C98bC4db0b5a0")

	paramsKey  = sdk.NewKVStoreKey("params")
	tParamsKey = sdk.NewTransientStoreKey("transient_params")
	accKey     = sdk.NewKVStoreKey("acc")
	storageKey = sdk.NewKVStoreKey("storage")
	codeKey    = sdk.NewKVStoreKey("code")

	logger = tmlog.NewNopLogger()

	rewardBig8  = big.NewInt(8)
	rewardBig32 = big.NewInt(32)
)

func init() {
	flag.StringVar(&flagCPUProfile, "cpu-profile", "", "write CPU profile")
	flag.StringVar(&flagDataDir, "datadir", "", "test data directory for state storage")
	flag.StringVar(&flagBlockchain, "blockchain", "blockchain", "ethereum block export file (blocks to import)")
	flag.Parse()
}

func newTestCodec() *codec.Codec {
	cdc := codec.New()

	evmtypes.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc
}

func cleanup() {
	fmt.Println("cleaning up test execution...")
	os.RemoveAll(flagDataDir)

	if flagCPUProfile != "" {
		pprof.StopCPUProfile()
	}
}

func trapSignals() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cleanup()
		os.Exit(1)
	}()
}

func createAndTestGenesis(t *testing.T, cms sdk.CommitMultiStore, ak auth.AccountKeeper) {
	genBlock := ethcore.DefaultGenesisBlock()
	ms := cms.CacheMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{}, false, logger)

	stateDB, err := evmtypes.NewCommitStateDB(ctx, ak, storageKey, codeKey)
	require.NoError(t, err, "failed to create a StateDB instance")

	// sort the addresses and insertion of key/value pairs matters
	genAddrs := make([]string, len(genBlock.Alloc))
	i := 0
	for addr := range genBlock.Alloc {
		genAddrs[i] = addr.String()
		i++
	}

	sort.Strings(genAddrs)

	for _, addrStr := range genAddrs {
		addr := ethcmn.HexToAddress(addrStr)
		acc := genBlock.Alloc[addr]

		stateDB.AddBalance(addr, acc.Balance)
		stateDB.SetCode(addr, acc.Code)
		stateDB.SetNonce(addr, acc.Nonce)

		for key, value := range acc.Storage {
			stateDB.SetState(addr, key, value)
		}
	}

	// get balance of one of the genesis account having 200 ETH
	b := stateDB.GetBalance(genInvestor)
	require.Equal(t, "200000000000000000000", b.String())

	// commit the stateDB with 'false' to delete empty objects
	//
	// NOTE: Commit does not yet return the intra merkle root (version)
	_, err = stateDB.Commit(false)
	require.NoError(t, err)

	// persist multi-store cache state
	ms.Write()

	// persist multi-store root state
	cms.Commit()

	// verify account mapper state
	genAcc := ak.GetAccount(ctx, sdk.AccAddress(genInvestor.Bytes()))
	require.NotNil(t, genAcc)
	require.Equal(t, sdk.NewIntFromBigInt(b), genAcc.GetCoins().AmountOf(types.DenomDefault))
}

func TestImportBlocks(t *testing.T) {
	if flagDataDir == "" {
		flagDataDir = os.TempDir()
	}

	if flagCPUProfile != "" {
		f, err := os.Create(flagCPUProfile)
		require.NoError(t, err, "failed to create CPU profile")

		err = pprof.StartCPUProfile(f)
		require.NoError(t, err, "failed to start CPU profile")
	}

	db := dbm.NewDB("state", dbm.LevelDBBackend, flagDataDir)
	defer cleanup()
	trapSignals()

	// create logger, codec and root multi-store
	cdc := newTestCodec()
	cms := store.NewCommitMultiStore(db)
	ak := auth.NewAccountKeeper(cdc, accKey, types.ProtoBaseAccount)

	// mount stores
	keys := []*sdk.KVStoreKey{accKey, storageKey, codeKey}
	for _, key := range keys {
		cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, nil)
	}

	cms.SetPruning(sdk.PruneNothing)

	// load latest version (root)
	err := cms.LoadLatestVersion()
	require.NoError(t, err)

	// set and test genesis block
	createAndTestGenesis(t, cms, ak)

	// open blockchain export file
	blockchainInput, err := os.Open(flagBlockchain)
	require.Nil(t, err)

	defer blockchainInput.Close()

	// ethereum mainnet config
	chainContext := core.NewChainContext()
	vmConfig := ethvm.Config{}
	chainConfig := ethparams.MainnetChainConfig

	// create RLP stream for exported blocks
	stream := ethrlp.NewStream(blockchainInput, 0)
	startTime := time.Now()

	var block ethtypes.Block
	for {
		err = stream.Decode(&block)
		if err == io.EOF {
			break
		}

		require.NoError(t, err, "failed to decode block")

		var (
			usedGas = new(uint64)
			gp      = new(ethcore.GasPool).AddGas(block.GasLimit())
		)

		header := block.Header()
		chainContext.Coinbase = header.Coinbase

		chainContext.SetHeader(block.NumberU64(), header)

		// Create a cached-wrapped multi-store based on the commit multi-store and
		// create a new context based off of that.
		ms := cms.CacheMultiStore()
		ctx := sdk.NewContext(ms, abci.Header{}, false, logger)
		ctx = ctx.WithBlockHeight(int64(block.NumberU64()))

		stateDB := createStateDB(t, ctx, ak)

		if chainConfig.DAOForkSupport && chainConfig.DAOForkBlock != nil && chainConfig.DAOForkBlock.Cmp(block.Number()) == 0 {
			ethmisc.ApplyDAOHardFork(stateDB)
		}

		for i, tx := range block.Transactions() {
			stateDB.Prepare(tx.Hash(), block.Hash(), i)

			_, _, err = ethcore.ApplyTransaction(
				chainConfig, chainContext, nil, gp, stateDB, header, tx, usedGas, vmConfig,
			)
			require.NoError(t, err, "failed to apply tx at block %d; tx: %X", block.NumberU64(), tx.Hash())
		}

		// apply mining rewards
		accumulateRewards(chainConfig, stateDB, header, block.Uncles())

		// commit stateDB
		_, err := stateDB.Commit(chainConfig.IsEIP158(block.Number()))
		require.NoError(t, err, "failed to commit StateDB")

		// simulate BaseApp EndBlocker commitment
		ms.Write()
		cms.Commit()

		// block debugging output
		if block.NumberU64() > 0 && block.NumberU64()%1000 == 0 {
			fmt.Printf("processed block: %d (time so far: %v)\n", block.NumberU64(), time.Since(startTime))
		}
	}
}

func createStateDB(t *testing.T, ctx sdk.Context, ak auth.AccountKeeper) *evmtypes.CommitStateDB {
	stateDB, err := evmtypes.NewCommitStateDB(ctx, ak, storageKey, codeKey)
	require.NoError(t, err, "failed to create a StateDB instance")

	return stateDB
}

// accumulateRewards credits the coinbase of the given block with the mining
// reward. The total reward consists of the static block reward and rewards for
// included uncles. The coinbase of each uncle block is also rewarded.
func accumulateRewards(
	config *ethparams.ChainConfig, stateDB *evmtypes.CommitStateDB,
	header *ethtypes.Header, uncles []*ethtypes.Header,
) {

	// select the correct block reward based on chain progression
	blockReward := ethash.FrontierBlockReward
	if config.IsByzantium(header.Number) {
		blockReward = ethash.ByzantiumBlockReward
	}

	// accumulate the rewards for the miner and any included uncles
	reward := new(big.Int).Set(blockReward)
	r := new(big.Int)

	for _, uncle := range uncles {
		r.Add(uncle.Number, rewardBig8)
		r.Sub(r, header.Number)
		r.Mul(r, blockReward)
		r.Div(r, rewardBig8)
		stateDB.AddBalance(uncle.Coinbase, r)
		r.Div(blockReward, rewardBig32)
		reward.Add(reward, r)
	}

	stateDB.AddBalance(header.Coinbase, reward)
}
