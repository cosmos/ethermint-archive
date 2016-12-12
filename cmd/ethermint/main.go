package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/tendermint/ethermint/application"
	"github.com/tendermint/ethermint/backend"
	"github.com/tendermint/ethermint/node"
	//	minerRewardStrategies "github.com/tendermint/ethermint/strategies/miner"
	//	validatorsStrategy "github.com/tendermint/ethermint/strategies/validators"
	cfg "github.com/tendermint/go-config"
	tmcfg "github.com/tendermint/tendermint/config/tendermint"
	tendermintNode "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tmsp/server"
)

var (
	config      cfg.Config
	DataDirFlag = utils.DirectoryFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases and keystore",
		Value: utils.DirectoryString{DefaultDataDir()},
	}
)

const (
	clientIdentifier = "Ethermint" // Client identifier to advertise over the network
	versionMajor     = 0           // Major version component of the current release
	versionMinor     = 1           // Minor version component of the current release
	versionPatch     = 0           // Patch version component of the current release
	versionMeta      = "unstable"  // Version metadata to append to the version string
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
	app = newCliApp(verString, "the ethermint command line interface")
	app.Action = tmspEthereumAction
	app.Commands = []cli.Command{
		{
			Action:      initCommand,
			Name:        "init",
			Usage:       "init genesis.json",
			Description: "",
		},
	}
	app.HideVersion = true // we have a command to print the version

	app.After = func(ctx *cli.Context) error {
		logger.Flush()
		return nil
	}

	logger.AddLogSystem(logger.NewStdLogSystem(os.Stdout, log.LstdFlags, logger.DebugLevel))
	glog.SetToStderr(true)
}

func main() {
	mainLogger.Infoln("Starting ethermint")
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initCommand(ctx *cli.Context) error {
	config = getTendermintConfig(ctx)
	init_files()

	genesisPath := ctx.Args().First()
	if len(genesisPath) == 0 {
		utils.Fatalf("must supply path to genesis JSON file")
	}

	chainDb, err := ethdb.NewLDBDatabase(filepath.Join(utils.MustMakeDataDir(ctx), "chaindata"), 0, 0)
	if err != nil {
		utils.Fatalf("could not open database: %v", err)
	}

	genesisFile, err := os.Open(genesisPath)
	if err != nil {
		utils.Fatalf("failed to read genesis file: %v", err)
	}

	block, err := core.WriteGenesisBlock(chainDb, genesisFile)
	if err != nil {
		utils.Fatalf("failed to write genesis block: %v", err)
	}
	glog.V(logger.Info).Infof("successfully wrote genesis block and/or chain rule set: %x", block.Hash())
	return nil
}

func getTendermintConfig(ctx *cli.Context) cfg.Config {
	datadir := ctx.GlobalString(DataDirFlag.Name)
	os.Setenv("TMROOT", datadir)
	config = tmcfg.GetConfig("")
	config.Set("node_laddr", ctx.GlobalString("node_laddr"))
	config.Set("seeds", ctx.GlobalString("seeds"))
	config.Set("fast_sync", ctx.GlobalBool("no_fast_sync"))
	config.Set("skip_upnp", ctx.GlobalBool("skip_upnp"))
	config.Set("rpc_laddr", ctx.GlobalString("rpc_laddr"))
	config.Set("proxy_app", ctx.GlobalString("addr"))
	return config
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

	config = getTendermintConfig(ctx)
	tendermintNode.RunNode(config)
	return nil
}

func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		if user, err := user.Current(); err == nil {
			p = user.HomeDir + p[1:]
		}
	}
	return path.Clean(os.ExpandEnv(p))
}

type DirectoryString struct {
	Value string
}

func (self *DirectoryString) String() string {
	return self.Value
}

func (self *DirectoryString) Set(value string) error {
	self.Value = expandPath(value)
	return nil
}

func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := HomeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "Ethermint")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "Ethermint")
		} else {
			return filepath.Join(home, ".ethermint")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
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
		DataDirFlag,
		utils.KeyStoreDirFlag,
		utils.BlockchainVersionFlag,
		utils.CacheFlag,
		utils.LightKDFFlag,
		utils.JSpathFlag,
		utils.ListenPortFlag,
		utils.MaxPeersFlag,
		utils.MaxPendingPeersFlag,
		utils.EtherbaseFlag,
		utils.TargetGasLimitFlag,
		utils.GasPriceFlag,
		utils.NATFlag,
		utils.NatspecEnabledFlag,
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
		utils.TestNetFlag,
		utils.VMForceJitFlag,
		utils.VMJitCacheFlag,
		utils.VMEnableJitFlag,
		utils.NetworkIdFlag,
		utils.RPCCORSDomainFlag,
		utils.MetricsEnabledFlag,
		utils.SolcPathFlag,
		utils.GpoMinGasPriceFlag,
		utils.GpoMaxGasPriceFlag,
		utils.GpoFullBlockRatioFlag,
		utils.GpobaseStepDownFlag,
		utils.GpobaseStepUpFlag,
		utils.GpobaseCorrectionFactorFlag,
		cli.StringFlag{
			Name:  "node_laddr",
			Value: "tcp://0.0.0.0:46656",
			Usage: "Node listen address. (0.0.0.0:0 means any interface, any port)",
		},
		cli.StringFlag{
			Name:  "seeds",
			Value: "",
			Usage: "Comma delimited host:port seed nodes",
		},
		cli.BoolFlag{
			Name:  "no_fast_sync",
			Usage: "Disable fast blockchain syncing",
		},
		cli.BoolFlag{
			Name:  "skip_upnp",
			Usage: "Skip UPNP configuration",
		},
		cli.StringFlag{
			Name:  "rpc_laddr",
			Value: "tcp://0.0.0.0:46657",
			Usage: "RPC listen address. Port required",
		},
		cli.StringFlag{
			Name:  "addr",
			Value: "tcp://0.0.0.0:46658",
			Usage: "TMSP app listen address",
		},
		cli.StringFlag{
			Name:  "tmsp",
			Value: "socket",
			Usage: "socket | grpc",
		},
	}
	return app
}
