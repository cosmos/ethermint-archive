package geth

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tendermint/ethermint/utils"
)

func TestNoSyncing(t *testing.T) {
	// rpc client for tendermint without syncing
	client := utils.NewMockClient()

	api := NewPublicEthereumAPI(nil, client)

	syncing, err := api.Syncing()
	if err != nil {
		t.Errorf("Error calling Syncing: %+v", err)
	}

	assert.Equal(t, false, syncing)
}

func TestSyncing(t *testing.T) {
	// rpc client for tendermint with syncing
	client := utils.NewMockSyncingClient()

	api := NewPublicEthereumAPI(nil, client)

	syncing, err := api.Syncing()
	if err != nil {
		t.Errorf("Error calling Syncing: %+v", err)
	}

	assert.Equal(t, true, syncing)
}
