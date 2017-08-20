package ethereum

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/tendermint/ethermint/types"
)

func TestNoSyncing(t *testing.T) {
	// rpc client for tendermint without syncing
	client := types.NewMockClient()

	api := NewPublicEthereumAPI(nil, client)

	syncing, err := api.Syncing()
	if err != nil {
		t.Fatalf("Error calling Syncing: %+v", err)
	}

	assert.Equal(t, false, syncing)
}

func TestSyncing(t *testing.T) {
	// rpc client for tendermint with syncing
	client := types.NewMockSyncingClient()

	api := NewPublicEthereumAPI(nil, client)

	syncing, err := api.Syncing()
	if err != nil {
		t.Fatalf("Error calling Syncing: %+v", err)
	}

	assert.Equal(t, true, syncing)
}
