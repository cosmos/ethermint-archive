package test

import (
	"testing"
	"github.com/tendermint/ethermint/app/test"
	"github.com/stretchr/testify/assert"

	"github.com/tendermint/ethermint/ethereum"
)

func TestNoSyncing(t *testing.T)  {
	// rpc client for tendermint without syncing
	client := test.NewMockClient()

	api := ethereum.NewPublicEthereumAPI(nil, client)

	syncing, err := api.Syncing()
	if err != nil {
		t.Errorf("Error calling Syncing: %+v", err)
	}

	assert.Equal(t,false, syncing)
}

func TestSyncing(t *testing.T)  {
	// rpc client for tendermint with syncing
	client := test.NewMockSyncingClient()

	api := ethereum.NewPublicEthereumAPI(nil, client)

	syncing, err := api.Syncing()
	if err != nil {
		t.Errorf("Error calling Syncing: %+v", err)
	}

	assert.Equal(t,true, syncing)
}

