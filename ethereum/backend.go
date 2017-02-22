package ethereum

import (
	"bytes"
	"os"
	"time"
	"reflect"
	"unsafe"

	"github.com/ethereum/go-ethereum/core"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"

	abciTypes "github.com/tendermint/abci/types"
	client "github.com/tendermint/go-rpc/client"
	core_types "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Backend handles the chain database and VM
type Backend struct {
	ethereum *eth.Ethereum
	client   *client.ClientURI
	config   *eth.Config
}

const (
	maxWaitForServerRetries = 10
)

// New creates a new Backend
func NewBackend(ctx *node.ServiceContext, config *eth.Config) (*Backend, error) {
	ethereum, err := eth.New(ctx, config)
	if err != nil {
		return nil, err
	}
	// setFakePow(ethereum)
	ethereum.BlockChain().SetValidator(NullBlockProcessor{})
	ethBackend := &Backend{
		ethereum: ethereum,
		client:   client.NewClientURI("tcp://localhost:46657"),
		config:   config,
		//		client: client.NewClientURI(fmt.Sprintf("http://%s", ctx.String(TendermintCoreHostFlag.Name))),
	}

	return ethBackend, nil
}

func waitForServer(s *Backend) error {
	// wait for Tendermint to open the socket and run http endpoint
	var result core_types.TMResult
	retriesCount := 0
	for result == nil {
		_, err := s.client.Call("status", map[string]interface{}{}, &result)
		if err != nil {
			glog.V(logger.Info).Infof("Waiting for tendermint endpoint to start: %s", err)
		}
		if retriesCount += 1; retriesCount >= maxWaitForServerRetries {
			return abciTypes.ErrInternalError
		}
		time.Sleep(time.Second)
	}
	return nil
}

// APIs returns the collection of RPC services the ethereum package offers.
func (s *Backend) APIs() []rpc.API {
	apis := s.Ethereum().APIs()
	retApis := []rpc.API{}
	for _, v := range apis {
		if v.Namespace == "net" {
			continue
		}
		if v.Namespace == "miner" {
			continue
		}
		if _, ok := v.Service.(*eth.PublicMinerAPI); ok {
			continue
		}
		retApis = append(retApis, v)
	}
	go s.txBroadcastLoop()
	return retApis
}

// Start implements node.Service, starting all internal goroutines needed by the
// Ethereum protocol implementation.
func (s *Backend) Start(srvr *p2p.Server) error {
	//	s.netRPCService = NewPublicNetAPI(srvr, s.NetVersion())
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Ethereum protocol.
func (s *Backend) Stop() error {
	s.ethereum.Stop()
	return nil
}

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Backend) Protocols() []p2p.Protocol {
	return nil
}

// Ethereum returns the underlying the ethereum object
func (s *Backend) Ethereum() *eth.Ethereum {
	return s.ethereum
}

// Config returns the eth.Config
func (s *Backend) Config() *eth.Config {
	return s.config
}

//----------------------------------------------------------------------
// Transactions sent via the go-ethereum rpc need to be routed to tendermint

// listen for txs and forward to tendermint
// TODO: some way to exit this (it runs in a go-routine)
func (s *Backend) txBroadcastLoop() {
	// minerWorkerUnsubscribe(s.ethereum)
	txSub := s.ethereum.EventMux().Subscribe(core.TxPreEvent{})

	if err := waitForServer(s); err != nil {
		// timeouted when waiting for tendermint communication failed
		glog.V(logger.Error).Infof("Failed to run tendermint HTTP endpoint, err=%s", err)
		os.Exit(1)
	}

	for obj := range txSub.Chan() {
		event := obj.Data.(core.TxPreEvent)
		if err := s.BroadcastTx(event.Tx); err != nil {
			glog.V(logger.Error).Infof("Broadcast, err=%s", err)
		}
	}
}

// BroadcastTx broadcasts a transaction to tendermint core
func (s *Backend) BroadcastTx(tx *ethTypes.Transaction) error {
	var result core_types.TMResult
	buf := new(bytes.Buffer)
	if err := tx.EncodeRLP(buf); err != nil {
		return err
	}
	params := map[string]interface{}{
		"tx": buf.Bytes(),
	}
	_, err := s.client.Call("broadcast_tx_sync", params, &result)
	return err
}

//----------------------------------------------------------------------
// NOTE: go-ethereum uses a monolithic Ethereum struct
// and does not expose many of the fields that we need to overwrite.
// So the quickest way forward is to use `unsafe` to overwrite those fields.

func minerWorkerUnsubscribe(ethereum *eth.Ethereum) {
	pointerVal := reflect.ValueOf(ethereum)
	val := reflect.Indirect(pointerVal)
	member := val.FieldByName("miner")
	val = reflect.Indirect(member)
	member = val.FieldByName("worker")

	ptrToWorker := unsafe.Pointer(member.UnsafeAddr())
	realPtrToWorker := (**interface{})(ptrToWorker)

	val = reflect.Indirect(member)
	member = val.FieldByName("events")

	ptrToEvents := unsafe.Pointer(member.UnsafeAddr())
	realPtrToEvents := (**event.TypeMuxSubscription)(ptrToEvents)

	(*realPtrToEvents).Unsubscribe()
	*realPtrToWorker = nil
}

/*
func setFakePow(ethereum *eth.Ethereum) {
	powToSet := pow.PoW(core.FakePow{})
	pointerVal := reflect.ValueOf(ethereum.BlockChain())
	val := reflect.Indirect(pointerVal)
	member := val.FieldByName("pow")
	ptrToPow := unsafe.Pointer(member.UnsafeAddr())
	realPtrToPow := (*pow.PoW)(ptrToPow)
	*realPtrToPow = powToSet
}
*/

/*
func (s *Backend) setFakeTxPool(txPoolAPI *ethapi.PublicTransactionPoolAPI) {
	mux := new(event.TypeMux)
	s.txSub = mux.Subscribe(core.TxPreEvent{})
	txPool := core.NewTxPool(s.Config().ChainConfig, mux, s.Ethereum().BlockChain().State, s.Ethereum().BlockChain().GasLimit)
	txPool.Pending()
	pointerVal := reflect.ValueOf(txPoolAPI)
	val := reflect.Indirect(pointerVal)
	member := val.FieldByName("txPool")
	ptrToTxPool := unsafe.Pointer(member.UnsafeAddr())
	realPtrToTxPool := (**core.TxPool)(ptrToTxPool)
	*realPtrToTxPool = txPool
}

func (s *Backend) setFakeMuxTxPool(txPoolAPI *ethapi.PublicTransactionPoolAPI) {
	mux := new(event.TypeMux)
	s.txSub = mux.Subscribe(core.TxPreEvent{})
	pointerVal := reflect.ValueOf(txPoolAPI)
	val := reflect.Indirect(pointerVal)
	member := val.FieldByName("txPool")
	ptrToTxPool := unsafe.Pointer(member.UnsafeAddr())
	realPtrToTxPool := (**core.TxPool)(ptrToTxPool)
	pointerVal = reflect.ValueOf(*realPtrToTxPool)
	val = reflect.Indirect(pointerVal)
	member = val.FieldByName("eventMux")
	ptrToEventMux := unsafe.Pointer(member.UnsafeAddr())
	realPtrToEventMux := (**event.TypeMux)(ptrToEventMux)
	*realPtrToEventMux = mux
}
*/
