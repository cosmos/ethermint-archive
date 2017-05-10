package ethereum

import (
	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"

	emtUtils "github.com/tendermint/ethermint/cmd/utils"
	"github.com/tendermint/go-rpc/client"
	"strings"
)

var clientIdentifier = "geth" // Client identifier to advertise over the network

// Config for p2p and network layer
func NewNodeConfig(ctx *cli.Context) *node.Config {
	return &node.Config{
		DataDir:     ethUtils.MakeDataDir(ctx),
		PrivateKey:  ethUtils.MakeNodeKey(ctx),
		Name:        clientIdentifier,
		IPCPath:     ethUtils.MakeIPCPath(ctx),
		HTTPHost:    ethUtils.MakeHTTPRpcHost(ctx),
		HTTPPort:    ctx.GlobalInt(ethUtils.RPCPortFlag.Name),
		HTTPCors:    ctx.GlobalString(ethUtils.RPCCORSDomainFlag.Name),
		HTTPModules: ethUtils.MakeRPCModules(ctx.GlobalString(ethUtils.RPCApiFlag.Name)),
		WSHost:      ethUtils.MakeWSRpcHost(ctx),
		WSPort:      ctx.GlobalInt(ethUtils.WSPortFlag.Name),
		WSOrigins:   ctx.GlobalString(ethUtils.WSAllowedOriginsFlag.Name),
		WSModules:   ethUtils.MakeRPCModules(ctx.GlobalString(ethUtils.WSApiFlag.Name)),
		NoDiscovery: true,
		MaxPeers:    0,
	}
}

// Config for the ethereum services
// NOTE:(go-ethereum) stack.OpenDatabase could be moved off the stack
// and then we wouldnt need it as an arg
func NewEthConfig(ctx *cli.Context, stack *node.Node) *eth.Config {

	ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	// jitEnabled := ctx.GlobalBool(ethUtils.VMEnableJitFlag.Name)
	return &eth.Config{
		ChainConfig: ethUtils.MakeChainConfig(ctx, stack),
		// BlockChainVersion:       ctx.GlobalInt(ethUtils.BlockchainVersionFlag.Name), TODO
		DatabaseCache:   ctx.GlobalInt(ethUtils.CacheFlag.Name),
		DatabaseHandles: ethUtils.MakeDatabaseHandles(),
		NetworkId:       ctx.GlobalInt(ethUtils.NetworkIdFlag.Name),
		Etherbase:       ethUtils.MakeEtherbase(ks, ctx),
		//EnableJit:               jitEnabled, // TODO
		//ForceJit:                ctx.GlobalBool(ethUtils.VMForceJitFlag.Name),
		GasPrice:                common.String2Big(ctx.GlobalString(ethUtils.GasPriceFlag.Name)),
		GpoMinGasPrice:          common.String2Big(ctx.GlobalString(ethUtils.GpoMinGasPriceFlag.Name)),
		GpoMaxGasPrice:          common.String2Big(ctx.GlobalString(ethUtils.GpoMaxGasPriceFlag.Name)),
		GpoFullBlockRatio:       ctx.GlobalInt(ethUtils.GpoFullBlockRatioFlag.Name),
		GpobaseStepDown:         ctx.GlobalInt(ethUtils.GpobaseStepDownFlag.Name),
		GpobaseStepUp:           ctx.GlobalInt(ethUtils.GpobaseStepUpFlag.Name),
		GpobaseCorrectionFactor: ctx.GlobalInt(ethUtils.GpobaseCorrectionFactorFlag.Name),
		SolcPath:                ctx.GlobalString(ethUtils.SolcPathFlag.Name),
		PowFake:                 true,
	}

}

// MakeSystemNode sets up a local node and configures the services to launch
func MakeSystemNode(name, version string, ctx *cli.Context) *node.Node {

	// Sets the target gas limit
	ethUtils.SetupNetwork(ctx)

	// Setup the node, a container for services
	// TODO: dont think we need a node.Node at all
	nodeConf := NewNodeConfig(ctx)
	stack, err := node.New(nodeConf)
	if err != nil {
		ethUtils.Fatalf("Failed to create the protocol stack: %v", err)
	}

	// Configure the eth
	ethConf := NewEthConfig(ctx, stack)

	//Remote tendermint RPC address
	tendermintURI := ctx.GlobalString(emtUtils.RpcLaddrFlag.Name)

	parts := strings.SplitN(tendermintURI, ":", 3)
	var port string
	if len(parts) != 3 {
		ethUtils.Fatalf("Wrong format of rpc_laddr flag")
	}
	port = parts[2]

	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return NewBackend(ctx, ethConf, rpcclient.NewClientURI("tcp://localhost:"+port))
	}); err != nil {
		ethUtils.Fatalf("Failed to register the TMSP application service: %v", err)
	}
	return stack
}
