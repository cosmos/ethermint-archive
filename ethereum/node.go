package ethereum

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"gopkg.in/urfave/cli.v1"

	"github.com/tendermint/go-rpc/client"

	emtUtils "github.com/tendermint/ethermint/cmd/utils"
)

var clientIdentifier = "geth" // Client identifier to advertise over the network

// MakeSystemNode sets up a local node and configures the services to launch
func MakeSystemNode(name, version string, ctx *cli.Context) *node.Node {
	params.TargetGasLimit = common.String2Big(ctx.GlobalString(ethUtils.TargetGasLimitFlag.Name))
	params.MinGasLimit = common.String2Big(ctx.GlobalString(emtUtils.MinGasLimitFlag.Name))

	// Configure the node's service container
	stackConf := &node.Config{
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
	// Assemble and return the protocol stack
	stack, err := node.New(stackConf)
	if err != nil {
		ethUtils.Fatalf("Failed to create the protocol stack: %v", err)
	}

	// Configure the Ethereum service
	ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	// jitEnabled := ctx.GlobalBool(ethUtils.VMEnableJitFlag.Name)
	ethConf := &eth.Config{
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

	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return NewBackend(ctx, ethConf, rpcclient.NewClientURI("tcp://localhost:46657"))
	}); err != nil {
		ethUtils.Fatalf("Failed to register the TMSP application service: %v", err)
	}
	return stack
}
