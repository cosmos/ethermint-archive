package node

import (
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/tendermint/ethermint/backend"
	"gopkg.in/urfave/cli.v1"
)

// MakeSystemNode sets up a local node and configures the services to launch
func MakeSystemNode(name, version string, ctx *cli.Context) *node.Node {
	params.TargetGasLimit = common.String2Big(ctx.GlobalString(utils.TargetGasLimitFlag.Name))

	// Configure the node's service container
	stackConf := &node.Config{
		DataDir:     utils.MustMakeDataDir(ctx),
		PrivateKey:  utils.MakeNodeKey(ctx),
		Name:        utils.MakeNodeName(name, version, ctx),
		IPCPath:     utils.MakeIPCPath(ctx),
		HTTPHost:    utils.MakeHTTPRpcHost(ctx),
		HTTPPort:    ctx.GlobalInt(utils.RPCPortFlag.Name),
		HTTPCors:    ctx.GlobalString(utils.RPCCORSDomainFlag.Name),
		HTTPModules: utils.MakeRPCModules(ctx.GlobalString(utils.RPCApiFlag.Name)),
		WSHost:      utils.MakeWSRpcHost(ctx),
		WSPort:      ctx.GlobalInt(utils.WSPortFlag.Name),
		WSOrigins:   ctx.GlobalString(utils.WSAllowedOriginsFlag.Name),
		WSModules:   utils.MakeRPCModules(ctx.GlobalString(utils.WSApiFlag.Name)),
	}
	// Configure the Ethereum service
	accman := utils.MakeAccountManager(ctx)
	jitEnabled := ctx.GlobalBool(utils.VMEnableJitFlag.Name)
	ethConf := &eth.Config{
		ChainConfig:             utils.MustMakeChainConfig(ctx),
		BlockChainVersion:       ctx.GlobalInt(utils.BlockchainVersionFlag.Name),
		DatabaseCache:           ctx.GlobalInt(utils.CacheFlag.Name),
		DatabaseHandles:         utils.MakeDatabaseHandles(),
		NetworkId:               ctx.GlobalInt(utils.NetworkIdFlag.Name),
		AccountManager:          accman,
		Etherbase:               utils.MakeEtherbase(accman, ctx),
		EnableJit:               jitEnabled,
		ForceJit:                ctx.GlobalBool(utils.VMForceJitFlag.Name),
		GasPrice:                common.String2Big(ctx.GlobalString(utils.GasPriceFlag.Name)),
		GpoMinGasPrice:          common.String2Big(ctx.GlobalString(utils.GpoMinGasPriceFlag.Name)),
		GpoMaxGasPrice:          common.String2Big(ctx.GlobalString(utils.GpoMaxGasPriceFlag.Name)),
		GpoFullBlockRatio:       ctx.GlobalInt(utils.GpoFullBlockRatioFlag.Name),
		GpobaseStepDown:         ctx.GlobalInt(utils.GpobaseStepDownFlag.Name),
		GpobaseStepUp:           ctx.GlobalInt(utils.GpobaseStepUpFlag.Name),
		GpobaseCorrectionFactor: ctx.GlobalInt(utils.GpobaseCorrectionFactorFlag.Name),
		SolcPath:                ctx.GlobalString(utils.SolcPathFlag.Name),
	}

	// Assemble and return the protocol stack
	stack, err := node.New(stackConf)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}
	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return backend.New(ctx, ethConf)
	}); err != nil {
		utils.Fatalf("Failed to register the TMSP application service: %v", err)
	}
	return stack
}
