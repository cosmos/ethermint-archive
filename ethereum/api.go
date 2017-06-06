package ethereum

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// We must implement our own net service since we don't have access to `internal/ethapi`

// NetRPCService mirrors the implementation of `internal/ethapi`
type NetRPCService struct {
	networkVersion uint64
}

// NewNetRPCService creates a new net API instance.
func NewNetRPCService(networkVersion uint64) *NetRPCService {
	return &NetRPCService{networkVersion}
}

// Listening returns an indication if the node is listening for network connections.
func (n *NetRPCService) Listening() bool {
	return true // always listening
}

// PeerCount returns the number of connected peers
func (n *NetRPCService) PeerCount() hexutil.Uint {
	return hexutil.Uint(0)
}

// Version returns the current ethereum protocol version.
func (n *NetRPCService) Version() string {
	return fmt.Sprintf("%d", n.networkVersion)
}
