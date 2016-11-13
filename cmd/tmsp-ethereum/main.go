package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/tendermint/tmsp/server"
	"gopkg.in/urfave/cli.v1"

	"github.com/kobigurk/tmsp-ethereum/application"
	"github.com/kobigurk/tmsp-ethereum/backend"
	"github.com/kobigurk/tmsp-ethereum/node"
	minerRewardStrategies "github.com/kobigurk/tmsp-ethereum/strategies/miner"
	validatorsStrategy "github.com/kobigurk/tmsp-ethereum/strategies/validators"
)

const (
	clientIdentifier = "TMSPEthereum" // Client identifier to advertise over the network
	versionMajor     = 0              // Major version component of the current release
	versionMinor     = 1              // Minor version component of the current release
	versionPatch     = 0              // Patch version component of the current release
	versionMeta      = "unstable"     // Version metadata to append to the version string
)

var (
	verString  string // Combined textual representation of all the version components
	app        *cli.App
	mainLogger = logger.NewLogger("main")
)

func init() {
	verString = fmt.Sprintf("%d.%d.%d", versionMajor, versionMinor, versionPatch)
	if versionMeta != "" {
		verString += "-" + versionMeta
	}
	app = newCliApp(verString, "the tmsp-ethereum command line interface")
	app.Action = tmspEthereumAction
	app.HideVersion = true // we have a command to print the version

	app.After = func(ctx *cli.Context) error {
		logger.Flush()
		return nil
	}

	logger.AddLogSystem(logger.NewStdLogSystem(os.Stdout, log.LstdFlags, logger.DebugLevel))
	glog.SetToStderr(true)
}

func main() {
	mainLogger.Infoln("Starting tmsp-ethereum")
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func tmspEthereumAction(ctx *cli.Context) error {
	stack := node.MakeSystemNode(clientIdentifier, verString, ctx)
	utils.StartNode(stack)
	addr := ctx.GlobalString("addr")
	tmsp := ctx.GlobalString("tmsp")

	var backend *backend.TMSPEthereumBackend
	if err := stack.Service(&backend); err != nil {
		utils.Fatalf("backend service not running: %v", err)
	}
	client, err := stack.Attach()
	if err != nil {
		utils.Fatalf("Failed to attach to the inproc geth: %v", err)
	}
	_, err = server.NewServer(addr, tmsp, application.NewTMSPEthereumApplication(backend, client, nil, nil))
	/*
		_, err = server.NewServer(
			addr,
			tmsp,
			application.NewTMSPEthereumApplication(
				backend,
				client,
				&minerRewardStrategies.RewardConstant{},
				&validatorsStrategy.TxBasedValidatorsStrategy{},
			),
		)
	*/
	if err != nil {
		os.Exit(1)
	}

	stack.Wait()
	return nil
}

func newCliApp(version, usage string) *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = ""
	//app.Authors = nil
	app.Email = ""
	app.Version = version
	app.Usage = usage
	app.Flags = []cli.Flag{
		utils.IdentityFlag,
		utils.UnlockedAccountFlag,
		utils.PasswordFileFlag,
		utils.BootnodesFlag,
		utils.DataDirFlag,
		utils.KeyStoreDirFlag,
		utils.BlockchainVersionFlag,
		utils.OlympicFlag,
		utils.FastSyncFlag,
		utils.CacheFlag,
		utils.LightKDFFlag,
		utils.JSpathFlag,
		utils.ListenPortFlag,
		utils.MaxPeersFlag,
		utils.MaxPendingPeersFlag,
		utils.EtherbaseFlag,
		utils.GasPriceFlag,
		utils.SupportDAOFork,
		utils.OpposeDAOFork,
		utils.MinerThreadsFlag,
		utils.MiningEnabledFlag,
		utils.MiningGPUFlag,
		utils.AutoDAGFlag,
		utils.TargetGasLimitFlag,
		utils.NATFlag,
		utils.NatspecEnabledFlag,
		utils.NoDiscoverFlag,
		utils.NodeKeyFileFlag,
		utils.NodeKeyHexFlag,
		utils.RPCEnabledFlag,
		utils.RPCListenAddrFlag,
		utils.RPCPortFlag,
		utils.RPCApiFlag,
		utils.WSEnabledFlag,
		utils.WSListenAddrFlag,
		utils.WSPortFlag,
		utils.WSApiFlag,
		utils.WSAllowedOriginsFlag,
		utils.IPCDisabledFlag,
		utils.IPCApiFlag,
		utils.IPCPathFlag,
		utils.ExecFlag,
		utils.PreloadJSFlag,
		utils.WhisperEnabledFlag,
		utils.DevModeFlag,
		utils.TestNetFlag,
		utils.VMForceJitFlag,
		utils.VMJitCacheFlag,
		utils.VMEnableJitFlag,
		utils.NetworkIdFlag,
		utils.RPCCORSDomainFlag,
		utils.MetricsEnabledFlag,
		utils.FakePoWFlag,
		utils.SolcPathFlag,
		utils.GpoMinGasPriceFlag,
		utils.GpoMaxGasPriceFlag,
		utils.GpoFullBlockRatioFlag,
		utils.GpobaseStepDownFlag,
		utils.GpobaseStepUpFlag,
		utils.GpobaseCorrectionFactorFlag,
		utils.ExtraDataFlag,
		cli.StringFlag{
			Name:  "addr",
			Value: "tcp://0.0.0.0:46658",
			Usage: "Listen address",
		},
		cli.StringFlag{
			Name:  "tmsp",
			Value: "socket",
			Usage: "socket | grpc",
		},
	}
	return app
}
