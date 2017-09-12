package ethereum

import (
	"bytes"
	"time"

	"github.com/ethereum/go-ethereum/core"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	rpcClient "github.com/tendermint/tendermint/rpc/client"
)

//----------------------------------------------------------------------
// Transactions sent via the go-ethereum rpc need to be routed to tendermint

// listen for txs and forward to tendermint
func (b *Backend) txBroadcastLoop() {
	b.txSub = b.ethereum.EventMux().Subscribe(core.TxPreEvent{})

	waitForServer(b.client)

	for obj := range b.txSub.Chan() {
		event := obj.Data.(core.TxPreEvent)
		if err := b.BroadcastTx(event.Tx); err != nil {
			log.Error("Broadcast error", "err", err)
		}
	}
}

// BroadcastTx broadcasts a transaction to tendermint core
// #unstable
func (b *Backend) BroadcastTx(tx *ethTypes.Transaction) error {
	buf := new(bytes.Buffer)
	if err := tx.EncodeRLP(buf); err != nil {
		return err
	}
	_, err := b.client.BroadcastTxSync(buf.Bytes())
	return err
}

//----------------------------------------------------------------------
// wait for Tendermint to open the socket and run http endpoint

func waitForServer(c rpcClient.Client) {
	for {
		_, err := c.Status()
		if err == nil {
			break
		}

		log.Info("Waiting for tendermint endpoint to start", "err", err)
		time.Sleep(time.Second * 3)
	}
}
