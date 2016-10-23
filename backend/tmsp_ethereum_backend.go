package backend

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"unsafe"

	"github.com/ethereum/go-ethereum/core"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/pow"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/kobigurk/tmsp-ethereum/processor"
	client "github.com/tendermint/go-rpc/client"
	core_types "github.com/tendermint/tendermint/rpc/core/types"
)

// TMSPEthereumBackend handles the chain database and VM
type TMSPEthereumBackend struct {
	ethereum *eth.Ethereum
	txSub    event.Subscription
	client   *client.ClientURI
	config   *eth.Config
}

func setFakePow(ethereum *eth.Ethereum) {
	powToSet := pow.PoW(core.FakePow{})
	pointerVal := reflect.ValueOf(ethereum.BlockChain())
	val := reflect.Indirect(pointerVal)
	member := val.FieldByName("pow")
	ptrToPow := unsafe.Pointer(member.UnsafeAddr())
	realPtrToPow := (*pow.PoW)(ptrToPow)
	*realPtrToPow = powToSet
}

func (s *TMSPEthereumBackend) setFakeMuxTxPool(txPoolAPI *eth.PublicTransactionPoolAPI) {
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

// New creates a new TMSPEthereumBackend
func New(ctx *node.ServiceContext, config *eth.Config) (*TMSPEthereumBackend, error) {
	ethereum, err := eth.New(ctx, config)
	if err != nil {
		return nil, err
	}
	setFakePow(ethereum)
	ethereum.BlockChain().SetValidator(processor.NullBlockProcessor{})
	tmspBackend := &TMSPEthereumBackend{
		ethereum: ethereum,
		client:   client.NewClientURI("tcp://localhost:46657"),
		config:   config,
		//		client: client.NewClientURI(fmt.Sprintf("http://%s", ctx.String(TendermintCoreHostFlag.Name))),
	}

	return tmspBackend, nil
}

// APIs returns the collection of RPC services the ethereum package offers.
func (s *TMSPEthereumBackend) APIs() []rpc.API {
	apis := s.Ethereum().APIs()
	retApis := []rpc.API{}
	for _, v := range apis {
		if v.Namespace == "net" {
			continue
		}
		if txPoolAPI, ok := v.Service.(*eth.PublicTransactionPoolAPI); ok {
			s.setFakeMuxTxPool(txPoolAPI)
			go s.txBroadcastLoop()
		}
		retApis = append(retApis, v)
	}
	return retApis
}

// Start implements node.Service, starting all internal goroutines needed by the
// Ethereum protocol implementation.
func (s *TMSPEthereumBackend) Start(srvr *p2p.Server) error {
	//	s.netRPCService = NewPublicNetAPI(srvr, s.NetVersion())
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Ethereum protocol.
func (s *TMSPEthereumBackend) Stop() error {
	s.ethereum.Stop()
	return nil
}

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *TMSPEthereumBackend) Protocols() []p2p.Protocol {
	return nil
}

// Ethereum returns the underlying the ethereum object
func (s *TMSPEthereumBackend) Ethereum() *eth.Ethereum {
	return s.ethereum
}

func (s *TMSPEthereumBackend) txBroadcastLoop() {
	for obj := range s.txSub.Chan() {
		event := obj.Data.(core.TxPreEvent)
		err := s.BroadcastTx(event.Tx)
		glog.V(logger.Info).Infof("Broadcast, err=%s", err)
	}
}

// BroadcastTx broadcasts a transaction to tendermint core
func (s *TMSPEthereumBackend) BroadcastTx(tx *ethTypes.Transaction) error {
	var result core_types.TMResult
	buf := new(bytes.Buffer)
	if err := tx.EncodeRLP(buf); err != nil {
		return err
	}
	params := map[string]interface{}{
		"tx": hex.EncodeToString(buf.Bytes()),
	}
	_, err := s.client.Call("broadcast_tx_sync", params, &result)
	return err
}

// Config returns the eth.Config
func (s *TMSPEthereumBackend) Config() *eth.Config {
	return s.config
}
