package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/urfave/cli.v1"

	ethereumUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"

	"github.com/tendermint/ethermint/version"

	ethermintUtils "github.com/tendermint/ethermint/cmd/utils"
	cfg "github.com/tendermint/go-config"
)

const (
	// Client identifier to advertise over the network
	clientIdentifier = "Ethermint"
)

var (
	// tendermint config
	config cfg.Config
)

func main() {
	glog.V(logger.Info).Infof("Starting ethermint")

	cliApp := newCliApp(version.Version, "the ethermint command line interface")
	cliApp.Action = ethermintCmd
	cliApp.Commands = []cli.Command{
		{
			Action:      initCmd,
			Name:        "init",
			Usage:       "init genesis.json",
			Description: "Initialize the files",
		},

		{
			Action:      versionCmd,
			Name:        "version",
			Usage:       "",
			Description: "Print the version",
		},
	}
	cliApp.HideVersion = true // we have a command to print the version

	cliApp.Before = func(ctx *cli.Context) error {
		config = ethermintUtils.GetTendermintConfig(ctx)
		return nil
	}
	cliApp.After = func(ctx *cli.Context) error {
		// logger.Flush()
		return nil
	}

	if err := cliApp.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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
		ethereumUtils.IdentityFlag,
		ethereumUtils.UnlockedAccountFlag,
		ethereumUtils.PasswordFileFlag,
		ethereumUtils.BootnodesFlag,
		ethereumUtils.KeyStoreDirFlag,
		// ethereumUtils.BlockchainVersionFlag,
		ethereumUtils.CacheFlag,
		ethereumUtils.LightKDFFlag,
		ethereumUtils.JSpathFlag,
		ethereumUtils.ListenPortFlag,
		ethereumUtils.MaxPeersFlag,
		ethereumUtils.MaxPendingPeersFlag,
		ethereumUtils.EtherbaseFlag,
		ethereumUtils.TargetGasLimitFlag,
		ethereumUtils.GasPriceFlag,
		ethereumUtils.NATFlag,
		// ethereumUtils.NatspecEnabledFlag,
		ethereumUtils.NodeKeyFileFlag,
		ethereumUtils.NodeKeyHexFlag,
		ethereumUtils.RPCEnabledFlag,
		ethereumUtils.RPCListenAddrFlag,
		ethereumUtils.RPCPortFlag,
		ethereumUtils.RPCApiFlag,
		ethereumUtils.WSEnabledFlag,
		ethereumUtils.WSListenAddrFlag,
		ethereumUtils.WSPortFlag,
		ethereumUtils.WSApiFlag,
		ethereumUtils.WSAllowedOriginsFlag,
		ethereumUtils.IPCDisabledFlag,
		ethereumUtils.IPCApiFlag,
		ethereumUtils.IPCPathFlag,
		ethereumUtils.ExecFlag,
		ethereumUtils.PreloadJSFlag,
		ethereumUtils.TestNetFlag,
		ethereumUtils.VMForceJitFlag,
		ethereumUtils.VMJitCacheFlag,
		ethereumUtils.VMEnableJitFlag,
		ethereumUtils.NetworkIdFlag,
		ethereumUtils.RPCCORSDomainFlag,
		ethereumUtils.MetricsEnabledFlag,
		ethereumUtils.SolcPathFlag,
		ethereumUtils.GpoMinGasPriceFlag,
		ethereumUtils.GpoMaxGasPriceFlag,
		ethereumUtils.GpoFullBlockRatioFlag,
		ethereumUtils.GpobaseStepDownFlag,
		ethereumUtils.GpobaseStepUpFlag,
		ethereumUtils.GpobaseCorrectionFactorFlag,
		ethermintUtils.VerbosityFlag, // not exposed by go-ethereum
		ethermintUtils.DataDirFlag,   // so we control defaults
		ethermintUtils.MinGasLimitFlag,

		//ethermint flags
		ethermintUtils.MonikerFlag,
		ethermintUtils.NodeLaddrFlag,
		ethermintUtils.LogLevelFlag,
		ethermintUtils.SeedsFlag,
		ethermintUtils.FastSyncFlag,
		ethermintUtils.SkipUpnpFlag,
		ethermintUtils.RpcLaddrFlag,
		ethermintUtils.AddrFlag,
		ethermintUtils.AbciFlag,
	}
	return app
}

func versionCmd(ctx *cli.Context) error {
	fmt.Println(version.Version)
	return nil
}
