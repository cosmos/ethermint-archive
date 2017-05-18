package main

import (
	"fmt"
	"os"

	"gopkg.in/urfave/cli.v1"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/log"

	"github.com/tendermint/ethermint/cmd/utils"
	"github.com/tendermint/ethermint/version"
)

const (
	// Client identifier to advertise over the network
	clientIdentifier = "ethermint"
)

var (
	// The app that holds all commands and flags.
	app = ethUtils.NewApp(version.Version, "the go-ethereum command line interface")
	// flags that configure the node
	nodeFlags = []cli.Flag{
		ethUtils.IdentityFlag,
		ethUtils.UnlockedAccountFlag,
		ethUtils.PasswordFileFlag,
		ethUtils.DataDirFlag,
		ethUtils.KeyStoreDirFlag,
		ethUtils.NoUSBFlag,
		ethUtils.CacheFlag,
		ethUtils.GasPriceFlag,
		ethUtils.TargetGasLimitFlag,
		ethUtils.NoDiscoverFlag,
		ethUtils.VMEnableDebugFlag,
		ethUtils.NetworkIdFlag,
		ethUtils.RPCCORSDomainFlag,
		ethUtils.EthStatsURLFlag,
		ethUtils.MetricsEnabledFlag,
		ethUtils.GpoBlocksFlag,
		ethUtils.GpoPercentileFlag,
	}

	rpcFlags = []cli.Flag{
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
		ethUtils.IPCPathFlag,
	}

	ethermintFlags = []cli.Flag{
		utils.BroadcastTxAddrFlag,
		utils.AbciFlag,
		utils.AddrFlag,
	}
)

func init() {
	log.Info("Starting ethermint")

	app.Action = ethermintCmd
	app.HideVersion = true
	app.Commands = []cli.Command{
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

	app.Flags = append(app.Flags, nodeFlags...)
	app.Flags = append(app.Flags, rpcFlags...)
	app.Flags = append(app.Flags, ethermintFlags...)

	app.Before = func(ctx *cli.Context) error {
		return nil
	}
	app.After = func(ctx *cli.Context) error {
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func versionCmd(ctx *cli.Context) error {
	fmt.Println(version.Version)
	return nil
}
