package ethereum

import (
	"bytes"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/core"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"

	abciTypes "github.com/tendermint/abci/types"
	core_types "github.com/tendermint/tendermint/rpc/core/types"
)

const (
	// try this number of times to connect to tendermint
	maxWaitForServerRetries = 10
)

// used by Backend to call tendermint rpc endpoints
// TODO: replace with HttpClient https://github.com/tendermint/go-rpc/issues/8
type Client interface {
	// see tendermint/go-rpc/client/http_client.go:115 func (c *ClientURI) Call(...)
	Call(method string, params map[string]interface{}, result interface{}) (interface{}, error)
}

//----------------------------------------------------------------------
// Transactions sent via the go-ethereum rpc need to be routed to tendermint
// TODO: Cleanup the relationship with Backend here ..

// listen for txs and forward to tendermint
// TODO: some way to exit this (it runs in a go-routine)
func (s *Backend) txBroadcastLoop() {
	txSub := s.ethereum.EventMux().Subscribe(core.TxPreEvent{})

	if err := waitForServer(s.client); err != nil {
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
// wait for Tendermint to open the socket and run http endpoint
func waitForServer(c Client) error {
	var result core_types.TMResult
	retriesCount := 0
	for result == nil {
		_, err := c.Call("status", map[string]interface{}{}, &result)
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
