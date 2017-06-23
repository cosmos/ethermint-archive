package utils

import (
	"gopkg.in/urfave/cli.v1"
)

var (
	// ----------------------------
	// ABCI Flags

	// TendermintAddrFlag is the address that ethermint will use to connect to the tendermint core node
	TendermintAddrFlag = cli.StringFlag{
		Name:  "tendermint_addr",
		Value: "tcp://localhost:46657",
		Usage: "This is the address that ethermint will use to connect to the tendermint core node. Please provide a port.",
	}

	// ABCIAddrFlag is the address that ethermint will use to listen to incoming ABCI connections
	ABCIAddrFlag = cli.StringFlag{
		Name:  "abci_laddr",
		Value: "tcp://0.0.0.0:46658",
		Usage: "This is the address that the ABCI server will use to listen to incoming connection from tendermint core.",
	}

	// ABCIProtocolFlag defines whether GRPC or SOCKET should be used for the ABCI connections
	ABCIProtocolFlag = cli.StringFlag{
		Name:  "abci_protocol",
		Value: "socket",
		Usage: "socket | grpc",
	}

	// VerbosityFlag defines the verbosity of the logging
	VerbosityFlag = cli.IntFlag{
		Name:  "verbosity",
		Value: 3,
		Usage: "Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=core, 5=debug, 6=detail",
	}

	// ConfigFileFlag defines the path to a TOML config for go-ethereum
	ConfigFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}

	// TargetGasLimitFlag defines gas limit of the Genesis block
	TargetGasLimitFlag = cli.Uint64Flag{
		Name:  "targetgaslimit",
		Usage: "Target gas limit sets the artificial target gas floor for the blocks to mine",
		Value: GenesisGasLimit.Uint64(),
	}
)
