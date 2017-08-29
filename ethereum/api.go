package ethereum

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/eth"

	rpcClient "github.com/tendermint/tendermint/rpc/client"
)

// We must implement our own net service since we don't have access to `internal/ethapi`

// NetRPCService mirrors the implementation of `internal/ethapi`
// #unstable
type NetRPCService struct {
	networkVersion uint64
}

// NewNetRPCService creates a new net API instance.
// #unstable
func NewNetRPCService(networkVersion uint64) *NetRPCService {
	return &NetRPCService{networkVersion}
}

// Listening returns an indication if the node is listening for network connections.
// #unstable
func (n *NetRPCService) Listening() bool {
	return true // always listening
}

// PeerCount returns the number of connected peers
// #unstable
func (n *NetRPCService) PeerCount() hexutil.Uint {
	return hexutil.Uint(0)
}

// Version returns the current ethereum protocol version.
// #unstable
func (n *NetRPCService) Version() string {
	return fmt.Sprintf("%d", n.networkVersion)
}

// PublicEthereumAPI provides an API to access Ethereum related information.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicEthereumAPI struct {
	b      *eth.EthApiBackend
	client rpcClient.Client
}

// NewPublicEthereumAPI creates a new Ethereum protocol API.
func NewPublicEthereumAPI(b *eth.EthApiBackend, c rpcClient.Client) *PublicEthereumAPI {
	return &PublicEthereumAPI{b, c}
}

// GasPrice returns a suggestion for a gas price.
func (s *PublicEthereumAPI) GasPrice(ctx context.Context) (*big.Int, error) {
	return s.b.SuggestPrice(ctx)
}

// ProtocolVersion returns the current Ethereum protocol version this node supports
func (s *PublicEthereumAPI) ProtocolVersion() hexutil.Uint {
	return hexutil.Uint(s.b.ProtocolVersion())
}

// Syncing returns whether the underlying tendermint core instance is currently
// fast-syncing or in consensus.
func (s *PublicEthereumAPI) Syncing() (interface{}, error) {
	res, err := s.client.Status()
	if err != nil {
		return nil, err
	}
	return res.Syncing, nil
}
