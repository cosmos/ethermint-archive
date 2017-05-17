package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/urfave/cli.v1"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/log"

	emtUtils "github.com/tendermint/ethermint/cmd/utils"
	"github.com/tendermint/ethermint/version"
)

const (
	// Client identifier to advertise over the network
	clientIdentifier = "Ethermint"
)

func main() {
	log.Info("Starting ethermint")

	cliApp := newCliApp(version.Version, "The ethermint command line interface")
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
		ethUtils.IdentityFlag,
		ethUtils.UnlockedAccountFlag,
		ethUtils.PasswordFileFlag,
		ethUtils.BootnodesFlag,
		ethUtils.KeyStoreDirFlag,
		// ethUtils.BlockchainVersionFlag,
		ethUtils.CacheFlag,
		ethUtils.LightKDFFlag,
		ethUtils.JSpathFlag,
		ethUtils.ListenPortFlag,
		ethUtils.MaxPeersFlag,
		ethUtils.MaxPendingPeersFlag,
		ethUtils.EtherbaseFlag,
		ethUtils.TargetGasLimitFlag,
		ethUtils.GasPriceFlag,
		ethUtils.NATFlag,
		// ethUtils.NatspecEnabledFlag,
		ethUtils.NodeKeyFileFlag,
		ethUtils.NodeKeyHexFlag,
		ethUtils.RPCEnabledFlag,
		ethUtils.RPCListenAddrFlag,
		ethUtils.RPCPortFlag,
		ethUtils.RPCApiFlag,
		ethUtils.WSEnabledFlag,
		ethUtils.WSListenAddrFlag,
		ethUtils.WSPortFlag,
		ethUtils.WSApiFlag,
		ethUtils.WSAllowedOriginsFlag,
		ethUtils.IPCDisabledFlag,
		ethUtils.IPCApiFlag,
		ethUtils.IPCPathFlag,
		ethUtils.ExecFlag,
		ethUtils.PreloadJSFlag,
		ethUtils.TestNetFlag,
		ethUtils.VMForceJitFlag,
		ethUtils.VMJitCacheFlag,
		ethUtils.VMEnableJitFlag,
		ethUtils.NetworkIdFlag,
		ethUtils.RPCCORSDomainFlag,
		ethUtils.MetricsEnabledFlag,
		ethUtils.SolcPathFlag,
		ethUtils.GpoMinGasPriceFlag,
		ethUtils.GpoMaxGasPriceFlag,
		ethUtils.GpoFullBlockRatioFlag,
		ethUtils.GpobaseStepDownFlag,
		ethUtils.GpobaseStepUpFlag,
		ethUtils.GpobaseCorrectionFactorFlag,

		emtUtils.VerbosityFlag, // not exposed by go-ethereum
		emtUtils.DataDirFlag,   // so we control defaults

		// ethermint flags
		emtUtils.BroadcastTxAddrFlag,
		emtUtils.AddrFlag,
		emtUtils.AbciFlag,
	}
	return app
}

func versionCmd(ctx *cli.Context) error {
	fmt.Println(version.Version)
	return nil
}
