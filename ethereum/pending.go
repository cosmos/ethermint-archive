package ethereum

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth"
	//"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"

	emtTypes "github.com/tendermint/ethermint/types"

	abciTypes "github.com/tendermint/abci/types"
)

//----------------------------------------------------------------------
// Pending manages concurrent access to the intermediate work object
// The ethereum tx pool fires TxPreEvent in a go-routine,
// and the miner subscribes to this in another go-routine and processes the tx onto
// an intermediate state. We used to use `unsafe` to overwrite the miner, but this
// didn't work because it didn't affect the already launched go-routines.
// So instead we introduce the Pending API in a small commit in go-ethereum
// so we don't even start the miner there, and instead manage the intermediate state from here.
// In the same commit we also fire the TxPreEvent synchronously so the order is preserved,
// instead of using a go-routine.

type pending struct {
	mtx  *sync.Mutex
	work *work
}

func newPending() *pending {
	return &pending{mtx: &sync.Mutex{}}
}

func (p *pending) checkTx(tx *ethTypes.Transaction) abciTypes.Result {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	currentState := p.work.intermediaryState

	// TODO: Check whether the tx has a higher than minimum gasprice

	var signer ethTypes.Signer = ethTypes.FrontierSigner{}
	if tx.Protected() {
		signer = ethTypes.NewEIP155Signer(tx.ChainId())
	}

	from, err := ethTypes.Sender(signer, tx)
	if err != nil {
		return abciTypes.ErrBaseInvalidSignature.
			AppendLog(core.ErrInvalidSender.Error())
	}

	if currentState.GetNonce(from) > tx.Nonce() {
		return abciTypes.ErrBadNonce.
			AppendLog(core.ErrNonceTooLow.Error())
	}

	// Check the transaction doesn't exceed the current block limit gas.
	gasLimit := p.gasLimit()
	if gasLimit.Cmp(tx.Gas()) < 0 {
		return abciTypes.ErrInternalError.
			AppendLog(core.ErrGasLimitReached.Error())
	}

	// Transactions can't be negative. This may never happen
	// using RLP decoded transactions but may occur if you create
	// a transaction using the RPC for example.
	if tx.Value().Sign() < 0 {
		return abciTypes.ErrBaseInvalidInput.
			AppendLog(core.ErrNegativeValue.Error())
	}

	// Transactor should have enough funds to cover the costs
	// cost == V + GP * GL
	if currentState.GetBalance(from).Cmp(tx.Cost()) < 0 {
		return abciTypes.ErrInsufficientFunds.
			AppendLog(core.ErrInsufficientFunds.Error())
	}

	// homestead == true
	intrGas := core.IntrinsicGas(tx.Data(), tx.To() == nil, true)
	if tx.Gas().Cmp(intrGas) < 0 {
		return abciTypes.ErrBaseInsufficientFees.
			SetLog(core.ErrIntrinsicGas.Error())
	}

	if tx.Size() > 32*1024 {
		return abciTypes.ErrBaseInvalidInput.
			AppendLog(core.ErrOversizedData.Error())
	}

	return abciTypes.OK
}

// execute the transaction
func (p *pending) deliverTx(blockchain *core.BlockChain, config *eth.Config, chainConfig *params.ChainConfig, tx *ethTypes.Transaction) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	blockHash := common.Hash{}
	return p.work.deliverTx(blockchain, config, chainConfig, blockHash, tx)
}

// accumulate validator rewards
func (p *pending) accumulateRewards(strategy *emtTypes.Strategy) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.work.accumulateRewards(strategy)
}

// commit and reset the work
func (p *pending) commit(ethereum *eth.Ethereum, receiver common.Address) (common.Hash, error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	blockHash, err := p.work.commit(ethereum.BlockChain(), ethereum.ChainDb())
	if err != nil {
		return common.Hash{}, err
	}

	work, err := p.resetWork(ethereum.BlockChain(), receiver)
	if err != nil {
		return common.Hash{}, err
	}

	p.work = work
	return blockHash, err
}

// return a new work object with the latest block and state from the chain
func (p *pending) resetWork(blockchain *core.BlockChain, receiver common.Address) (*work, error) {
	state, err := blockchain.State()
	if err != nil {
		return nil, err
	}

	currentBlock := blockchain.CurrentBlock()
	ethHeader := newBlockHeader(receiver, currentBlock)

	iState, err := blockchain.State()
	if err != nil {
		return nil, err
	}
	intermediaryState := iState.Copy()

	return &work{
		header:            ethHeader,
		parent:            currentBlock,
		state:             state,
		intermediaryState: intermediaryState,
		txIndex:           0,
		totalUsedGas:      big.NewInt(0),
		gp:                new(core.GasPool).AddGas(ethHeader.GasLimit),
	}, nil
}

func (p *pending) updateHeaderWithTimeInfo(config *params.ChainConfig, parentTime uint64, numTx uint64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.work.updateHeaderWithTimeInfo(config, parentTime, numTx)
}

func (p *pending) gasLimit() big.Int {
	return big.Int(*p.work.gp)
}

//----------------------------------------------------------------------
// Implements: miner.Pending API (our custom patch to go-ethereum)

// Return a new block and a copy of the state from the latest work
// #unstable
func (p *pending) Pending() (*ethTypes.Block, *state.StateDB) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	return ethTypes.NewBlock(
		p.work.header,
		p.work.transactions,
		nil,
		p.work.receipts,
	), p.work.state.Copy()
}

//----------------------------------------------------------------------
//

// The work struct handles block processing.
// It's updated with each DeliverTx and reset on Commit
type work struct {
	header            *ethTypes.Header
	parent            *ethTypes.Block
	state             *state.StateDB
	intermediaryState *state.StateDB

	txIndex      int
	transactions []*ethTypes.Transaction
	receipts     ethTypes.Receipts
	allLogs      []*ethTypes.Log

	totalUsedGas *big.Int
	gp           *core.GasPool
}

// nolint: unparam
func (w *work) accumulateRewards(strategy *emtTypes.Strategy) {
	ethash.AccumulateRewards(w.state, w.header, []*ethTypes.Header{})
	w.header.GasUsed = w.totalUsedGas
}

// Runs ApplyTransaction against the ethereum blockchain, fetches any logs,
// and appends the tx, receipt, and logs
func (w *work) deliverTx(blockchain *core.BlockChain, config *eth.Config, chainConfig *params.ChainConfig, blockHash common.Hash, tx *ethTypes.Transaction) error {
	w.state.Prepare(tx.Hash(), blockHash, w.txIndex)
	receipt, _, err := core.ApplyTransaction(
		chainConfig,
		blockchain,
		nil, // defaults to address of the author of the header
		w.gp,
		w.state,
		w.header,
		tx,
		w.totalUsedGas,
		vm.Config{EnablePreimageRecording: config.EnablePreimageRecording},
	)
	if err != nil {
		return err
	}

	logs := w.state.GetLogs(tx.Hash())

	w.txIndex++

	// The slices are allocated in updateHeaderWithTimeInfo
	w.transactions = append(w.transactions, tx)
	w.receipts = append(w.receipts, receipt)
	w.allLogs = append(w.allLogs, logs...)

	return err
}

// Commit the ethereum state, update the header, make a new block and add it
// to the ethereum blockchain. The application root hash is the hash of the ethereum block.
func (w *work) commit(blockchain *core.BlockChain, db ethdb.Database) (common.Hash, error) {

	// commit ethereum state and update the header
	hashArray, err := w.state.CommitTo(db.NewBatch(), false) // XXX: ugh hardforks
	if err != nil {
		return common.Hash{}, err
	}
	w.header.Root = hashArray

	for _, log := range w.allLogs {
		log.BlockHash = hashArray
	}

	// create block object and compute final commit hash (hash of the ethereum block)
	block := ethTypes.NewBlock(w.header, w.transactions, nil, w.receipts)
	blockHash := block.Hash()

	// save the block to disk
	// log.Info("Committing block", "stateHash", hashArray, "blockHash", blockHash)
	_, err = blockchain.InsertChain([]*ethTypes.Block{block})
	if err != nil {
		// log.Info("Error inserting ethereum block in chain", "err", err)
		return common.Hash{}, err
	}
	return blockHash, err
}

func (w *work) updateHeaderWithTimeInfo(config *params.ChainConfig, parentTime uint64, numTx uint64) {
	lastBlock := w.parent
	parentHeader := &ethTypes.Header{
		Difficulty: lastBlock.Difficulty(),
		Number:     lastBlock.Number(),
		Time:       lastBlock.Time(),
	}
	w.header.Time = new(big.Int).SetUint64(parentTime)
	w.header.Difficulty = ethash.CalcDifficulty(config, parentTime, parentHeader)
	w.transactions = make([]*ethTypes.Transaction, 0, numTx)
	w.receipts = make([]*ethTypes.Receipt, 0, numTx)
	w.allLogs = make([]*ethTypes.Log, 0, numTx)
}

//----------------------------------------------------------------------

// Create a new block header from the previous block
func newBlockHeader(receiver common.Address, prevBlock *ethTypes.Block) *ethTypes.Header {
	return &ethTypes.Header{
		Number:     prevBlock.Number().Add(prevBlock.Number(), big.NewInt(1)),
		ParentHash: prevBlock.Hash(),
		GasLimit:   core.CalcGasLimit(prevBlock),
		Coinbase:   receiver,
	}
}
