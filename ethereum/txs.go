package ethereum

import (
	"bytes"
	"time"

	"github.com/ethereum/go-ethereum/core"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	rpcClient "github.com/tendermint/tendermint/rpc/lib/client"
)

//----------------------------------------------------------------------
// Transactions sent via the go-ethereum rpc need to be routed to tendermint
// TODO: Cleanup the relationship with Backend here ..

// listen for txs and forward to tendermint
// TODO: some way to exit this (it runs in a go-routine)
func (s *Backend) txBroadcastLoop() {
	txSub := s.ethereum.EventMux().Subscribe(core.TxPreEvent{})

	waitForServer(s.client)

	for obj := range txSub.Chan() {
		event := obj.Data.(core.TxPreEvent)
		if err := s.BroadcastTx(event.Tx); err != nil {
			log.Error("Broadcast error", "err", err)
		}
	}
}

// BroadcastTx broadcasts a transaction to tendermint core
func (s *Backend) BroadcastTx(tx *ethTypes.Transaction) error {
	var result interface{}
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
func waitForServer(c rpcClient.HTTPClient) {
	var result interface{}

	for {
		_, err := c.Call("status", map[string]interface{}{}, &result)
		if err == nil {
			break
		}

			log.Info("Waiting for tendermint endpoint to start", "err", err)
		time.Sleep(time.Second* 3)
	}
}
