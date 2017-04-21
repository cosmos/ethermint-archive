package main

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/tendermint/ethermint/app"
	"github.com/tendermint/ethermint/ethereum"
	//	minerRewardStrategies "github.com/tendermint/ethermint/strategies/miner"
	//	validatorsStrategy "github.com/tendermint/ethermint/strategies/validators"
	"github.com/ethereum/go-ethereum/params"
	"github.com/tendermint/abci/server"
	cfg "github.com/tendermint/go-config"
	tmlog "github.com/tendermint/go-logger"
	tmcfg "github.com/tendermint/tendermint/config/tendermint"
	tendermintNode "github.com/tendermint/tendermint/node"
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
	verString string // Combined textual representation of all the version components
	cliApp    *cli.App
	gitCommit = "" // Git SHA1 commit hash of the release (set via linker flags)
	// mainLogger = logger.NewLogger("main")
)

func init() {
	verString = fmt.Sprintf("%d.%d.%d", versionMajor, versionMinor, versionPatch)
	if versionMeta != "" {
		verString += "-" + versionMeta
	}
	if gitCommit != "" {
		verString += "-" + gitCommit[:8]
	}
	verString += " Ethereum/" + params.Version
	cliApp = newCliApp(verString, "the ethermint command line interface")
	cliApp.Action = abciEthereumAction
	cliApp.Commands = []cli.Command{
		{
			Action:      initCommand,
			Name:        "init",
			Usage:       "init genesis.json",
			Description: "",
		},
	}
	cliApp.HideVersion = true // we have a command to print the version

	cliApp.After = func(ctx *cli.Context) error {
		// logger.Flush()
		return nil
	}
}

func main() {
	glog.V(logger.Info).Infof("Starting ethermint")
	if err := cliApp.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initCommand(ctx *cli.Context) error {
	config = getTendermintConfig(ctx)
	init_files()

	// ethereum genesis.json
	genesisPath := ctx.Args().First()
	if len(genesisPath) == 0 {
		utils.Fatalf("must supply path to genesis JSON file")
	}

	chainDb, err := ethdb.NewLDBDatabase(filepath.Join(utils.MakeDataDir(ctx), "chaindata"), 0, 0)
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
	//set seeds only if it is not empty
	//otherwise it overwrite value from config.toml file
	if ctx.GlobalString("seeds") != "" {
		config.Set("seeds", ctx.GlobalString("seeds"))
	}
	config.Set("fast_sync", ctx.GlobalBool("no_fast_sync"))
	config.Set("skip_upnp", ctx.GlobalBool("skip_upnp"))
	config.Set("rpc_laddr", ctx.GlobalString("rpc_laddr"))
	config.Set("proxy_app", ctx.GlobalString("addr"))
	config.Set("log_level", ctx.GlobalString("log_level"))

	tmlog.SetLogLevel(config.GetString("log_level"))

	return config
}

func abciEthereumAction(ctx *cli.Context) error {
	stack := ethereum.MakeSystemNode(clientIdentifier, verString, ctx)
	utils.StartNode(stack)
	addr := ctx.GlobalString("addr")
	abci := ctx.GlobalString("abci")

	//set verbosity level for go-ethereum
	glog.SetToStderr(true)
	glog.SetV(ctx.GlobalInt(VerbosityFlag.Name))

	var backend *ethereum.Backend
	if err := stack.Service(&backend); err != nil {
		utils.Fatalf("backend service not running: %v", err)
	}
	client, err := stack.Attach()
	if err != nil {
		utils.Fatalf("Failed to attach to the inproc geth: %v", err)
	}
	ethApp, err := app.NewEthermintApplication(backend, client, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)

	}
	_, err = server.NewServer(addr, abci, ethApp)
	/*
		_, err = server.NewServer(
			addr,
			abci,
			app.NewTMSPEthereumApplication(
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
		// utils.BlockchainVersionFlag,
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
		// utils.NatspecEnabledFlag,
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

		//ethermint flags
		NodeLaddrFlag,
		LogLevelFlag,
		SeedsFlag,
		NoFastSyncFlag,
		SkipUpnpFlag,
		RpcLaddrFlag,
		AddrFlag,
		AbciFlag,
		VerbosityFlag,
	}
	return app
}
