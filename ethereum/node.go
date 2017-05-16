package ethereum

import (
	"math/big"

	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"

	"github.com/tendermint/go-rpc/client"
)

var clientIdentifier = "geth" // Client identifier to advertise over the network

// Config for p2p and network layer
func NewNodeConfig(ctx *cli.Context) *node.Config {
	return &node.Config{
		DataDir:     utils.MakeDataDir(ctx),
		PrivateKey:  utils.MakeNodeKey(ctx),
		Name:        clientIdentifier,
		IPCPath:     utils.MakeIPCPath(ctx),
		HTTPHost:    utils.MakeHTTPRpcHost(ctx),
		HTTPPort:    ctx.GlobalInt(utils.RPCPortFlag.Name),
		HTTPCors:    ctx.GlobalString(utils.RPCCORSDomainFlag.Name),
		HTTPModules: utils.MakeRPCModules(ctx.GlobalString(utils.RPCApiFlag.Name)),
		WSHost:      utils.MakeWSRpcHost(ctx),
		WSPort:      ctx.GlobalInt(utils.WSPortFlag.Name),
		WSOrigins:   ctx.GlobalString(utils.WSAllowedOriginsFlag.Name),
		WSModules:   utils.MakeRPCModules(ctx.GlobalString(utils.WSApiFlag.Name)),
		NoDiscovery: true,
		MaxPeers:    0,
	}
}

// Config for the ethereum services
// NOTE:(go-ethereum) stack.OpenDatabase could be moved off the stack
// and then we wouldnt need it as an arg
func NewEthConfig(ctx *cli.Context, stack *node.Node) *eth.Config {

	ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	//set HomesteadBlock to 0
	chainConfig := utils.MakeChainConfig(ctx, stack)
	chainConfig.HomesteadBlock = big.NewInt(0)

	// jitEnabled := ctx.GlobalBool(utils.VMEnableJitFlag.Name)
	return &eth.Config{
		ChainConfig: chainConfig,
		// BlockChainVersion:       ctx.GlobalInt(utils.BlockchainVersionFlag.Name), TODO
		DatabaseCache:   ctx.GlobalInt(utils.CacheFlag.Name),
		DatabaseHandles: utils.MakeDatabaseHandles(),
		NetworkId:       ctx.GlobalInt(utils.NetworkIdFlag.Name),
		Etherbase:       utils.MakeEtherbase(ks, ctx),
		//EnableJit:               jitEnabled, // TODO
		//ForceJit:                ctx.GlobalBool(utils.VMForceJitFlag.Name),
		GasPrice:                common.String2Big(ctx.GlobalString(utils.GasPriceFlag.Name)),
		GpoMinGasPrice:          common.String2Big(ctx.GlobalString(utils.GpoMinGasPriceFlag.Name)),
		GpoMaxGasPrice:          common.String2Big(ctx.GlobalString(utils.GpoMaxGasPriceFlag.Name)),
		GpoFullBlockRatio:       ctx.GlobalInt(utils.GpoFullBlockRatioFlag.Name),
		GpobaseStepDown:         ctx.GlobalInt(utils.GpobaseStepDownFlag.Name),
		GpobaseStepUp:           ctx.GlobalInt(utils.GpobaseStepUpFlag.Name),
		GpobaseCorrectionFactor: ctx.GlobalInt(utils.GpobaseCorrectionFactorFlag.Name),
		SolcPath:                ctx.GlobalString(utils.SolcPathFlag.Name),
		PowFake:                 true,
	}

}

// MakeSystemNode sets up a local node and configures the services to launch
func MakeSystemNode(name, version string, ctx *cli.Context) *node.Node {

	// Sets the target gas limit
	utils.SetupNetwork(ctx)

	// Setup the node, a container for services
	// TODO: dont think we need a node.Node at all
	nodeConf := NewNodeConfig(ctx)
	stack, err := node.New(nodeConf)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}

	// Configure the eth
	ethConf := NewEthConfig(ctx, stack)

	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return NewBackend(ctx, ethConf, rpcclient.NewClientURI("tcp://localhost:46657"))
	}); err != nil {
		utils.Fatalf("Failed to register the TMSP application service: %v", err)
	}
	return stack
}
