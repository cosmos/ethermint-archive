package ethereum

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethereumUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"gopkg.in/urfave/cli.v1"

	"github.com/tendermint/go-rpc/client"

	ethermintUtils "github.com/tendermint/ethermint/cmd/utils"
)

var clientIdentifier = "geth" // Client identifier to advertise over the network

// MakeSystemNode sets up a local node and configures the services to launch
func MakeSystemNode(name, version string, ctx *cli.Context) *node.Node {
	params.TargetGasLimit = common.String2Big(ctx.GlobalString(ethereumUtils.TargetGasLimitFlag.Name))
	params.MinGasLimit = common.String2Big(ctx.GlobalString(ethermintUtils.MinGasLimitFlag.Name))

	// Configure the node's service container
	stackConf := &node.Config{
		DataDir:     ethereumUtils.MakeDataDir(ctx),
		PrivateKey:  ethereumUtils.MakeNodeKey(ctx),
		Name:        clientIdentifier,
		IPCPath:     ethereumUtils.MakeIPCPath(ctx),
		HTTPHost:    ethereumUtils.MakeHTTPRpcHost(ctx),
		HTTPPort:    ctx.GlobalInt(ethereumUtils.RPCPortFlag.Name),
		HTTPCors:    ctx.GlobalString(ethereumUtils.RPCCORSDomainFlag.Name),
		HTTPModules: ethereumUtils.MakeRPCModules(ctx.GlobalString(ethereumUtils.RPCApiFlag.Name)),
		WSHost:      ethereumUtils.MakeWSRpcHost(ctx),
		WSPort:      ctx.GlobalInt(ethereumUtils.WSPortFlag.Name),
		WSOrigins:   ctx.GlobalString(ethereumUtils.WSAllowedOriginsFlag.Name),
		WSModules:   ethereumUtils.MakeRPCModules(ctx.GlobalString(ethereumUtils.WSApiFlag.Name)),
		NoDiscovery: true,
		MaxPeers:    0,
	}
	// Assemble and return the protocol stack
	stack, err := node.New(stackConf)
	if err != nil {
		ethereumUtils.Fatalf("Failed to create the protocol stack: %v", err)
	}

	// Configure the Ethereum service
	ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	// jitEnabled := ctx.GlobalBool(ethereumUtils.VMEnableJitFlag.Name)
	ethConf := &eth.Config{
		ChainConfig: ethereumUtils.MakeChainConfig(ctx, stack),
		// BlockChainVersion:       ctx.GlobalInt(ethereumUtils.BlockchainVersionFlag.Name), TODO
		DatabaseCache:   ctx.GlobalInt(ethereumUtils.CacheFlag.Name),
		DatabaseHandles: ethereumUtils.MakeDatabaseHandles(),
		NetworkId:       ctx.GlobalInt(ethereumUtils.NetworkIdFlag.Name),
		Etherbase:       ethereumUtils.MakeEtherbase(ks, ctx),
		//EnableJit:               jitEnabled, // TODO
		//ForceJit:                ctx.GlobalBool(ethereumUtils.VMForceJitFlag.Name),
		GasPrice:                common.String2Big(ctx.GlobalString(ethereumUtils.GasPriceFlag.Name)),
		GpoMinGasPrice:          common.String2Big(ctx.GlobalString(ethereumUtils.GpoMinGasPriceFlag.Name)),
		GpoMaxGasPrice:          common.String2Big(ctx.GlobalString(ethereumUtils.GpoMaxGasPriceFlag.Name)),
		GpoFullBlockRatio:       ctx.GlobalInt(ethereumUtils.GpoFullBlockRatioFlag.Name),
		GpobaseStepDown:         ctx.GlobalInt(ethereumUtils.GpobaseStepDownFlag.Name),
		GpobaseStepUp:           ctx.GlobalInt(ethereumUtils.GpobaseStepUpFlag.Name),
		GpobaseCorrectionFactor: ctx.GlobalInt(ethereumUtils.GpobaseCorrectionFactorFlag.Name),
		SolcPath:                ctx.GlobalString(ethereumUtils.SolcPathFlag.Name),
		PowFake:                 true,
	}

	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return NewBackend(ctx, ethConf, rpcclient.NewClientURI("tcp://localhost:46657"))
	}); err != nil {
		ethereumUtils.Fatalf("Failed to register the TMSP application service: %v", err)
	}
	return stack
}
