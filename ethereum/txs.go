package ethereum

import (
	"bytes"
	"github.com/ethereum/go-ethereum/core"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	rpcClient "github.com/tendermint/tendermint/rpc/lib/client"
	"time"
)

//----------------------------------------------------------------------
// Transactions sent via the go-ethereum rpc need to be routed to tendermint

// listen for txs and forward to tendermint
func (b *Backend) txBroadcastLoop() {
	b.txSub = b.ethereum.EventMux().Subscribe(core.TxPreEvent{})

	WaitForServer(b.client)

	for obj := range b.txSub.Chan() {
		event := obj.Data.(core.TxPreEvent)
		if err := b.BroadcastTx(event.Tx); err != nil {
			log.Error("Broadcast error", "err", err)
		}
	}
}

// BroadcastTx broadcasts a transaction to tendermint core
func (b *Backend) BroadcastTx(tx *ethTypes.Transaction) error {
	return BroadcastTransaction(b.client, tx)
}

// wait for Tendermint to open the socket and run http endpoint
func WaitForServer(c rpcClient.HTTPClient) {
	var result interface{}

	for {
		_, err := c.Call("status", map[string]interface{}{}, &result)
		if err == nil {
			break
		}

		log.Info("Waiting for tendermint endpoint to start", "err", err)
		time.Sleep(time.Second * 3)
	}
}

func BroadcastTransaction(client rpcClient.HTTPClient, tx *ethTypes.Transaction) error {
	var result interface{}
	buf := new(bytes.Buffer)
	if err := tx.EncodeRLP(buf); err != nil {
		return err
	}
	params := map[string]interface{}{
		"tx": buf.Bytes(),
	}
	_, err := client.Call("broadcast_tx_sync", params, &result)

	return err
}

func BroadcastBlock(client rpcClient.HTTPClient, block *ethTypes.Block) error {
	var result interface{}
	buf := new(bytes.Buffer)
	if err := block.EncodeRLP(buf); err != nil {
		return err
	}
	params := map[string]interface{}{
		"tx": buf.Bytes(),
	}
	_, err := client.Call("broadcast_tx_sync", params, &result)

	return err
}
