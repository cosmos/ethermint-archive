package utils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"reflect"
	"strings"
	"unicode"

	cli "gopkg.in/urfave/cli.v1"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"

	"github.com/naoina/toml"
	"github.com/tendermint/ethermint/ethereum"

	rpcClient "github.com/tendermint/tendermint/rpc/lib/client"
)

const (
	// Client identifier to advertise over the network
	clientIdentifier = "go-ethereum"
	// Environment variable for home dir
	emHome = "EMHOME"
)

var (
	// GenesisTargetGasLimit is the target gas limit of the Genesis block.
	// #unstable
	GenesisTargetGasLimit = big.NewInt(100000000)
)

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

type ethstatsConfig struct {
	URL string `toml:",omitempty"`
}

type additionalConfigs struct {
	UnlockAccounts []string `toml:",omitempty"`
	Passwords      string   `toml:",omitempty"`
	TrieCacheGen   uint16   `toml:",omitempty"`
	TargetGasLimit uint64   `toml:",omitempty"`
}

// GethConfig contains configuration structure
type GethConfig struct {
	Eth        eth.Config
	Node       node.Config
	Ethstats   ethstatsConfig
	Additional additionalConfigs
}

// MakeFullNode creates a full go-ethereum node
// #unstable
func MakeFullNode(ctx *cli.Context) (*node.Node, GethConfig) {
	stack, cfg := makeConfigNode(ctx)

	tendermintLAddr := ctx.GlobalString(TendermintAddrFlag.Name)
	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return ethereum.NewBackend(ctx, &cfg.Eth, rpcClient.NewURIClient(tendermintLAddr))
	}); err != nil {
		ethUtils.Fatalf("Failed to register the ABCI application service: %v", err)
	}

	return stack, cfg
}

func makeConfigNode(ctx *cli.Context) (*node.Node, GethConfig) {
	cfg := GethConfig{
		Eth:  eth.DefaultConfig,
		Node: DefaultNodeConfig(),
	}

	ethUtils.SetNodeConfig(ctx, &cfg.Node)

	// Load config file.
	if file := ctx.GlobalString(ConfigFileFlag.Name); file != "" {
		if err := LoadConfig(file, &cfg); err != nil {
			ethUtils.Fatalf("%v", err)
		}
	}

	SetAdditionalConfig(&cfg.Additional)
	SetEthermintNodeConfig(&cfg.Node)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		ethUtils.Fatalf("Failed to create the protocol stack: %v", err)
	}

	ethUtils.SetEthConfig(ctx, stack, &cfg.Eth)
	SetEthermintEthConfig(&cfg.Eth)

	return stack, cfg
}

// DefaultNodeConfig returns the default configuration for a go-ethereum node
// #unstable
func DefaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.Version
	cfg.HTTPModules = append(cfg.HTTPModules, "eth")
	cfg.WSModules = append(cfg.WSModules, "eth")
	cfg.IPCPath = "geth.ipc"

	emHome := os.Getenv(emHome)
	if emHome != "" {
		cfg.DataDir = emHome
	}

	return cfg
}

// LoadConfig takes configuration file path and full configuration and applies file configs to it
// #unstable
func LoadConfig(file string, cfg *GethConfig) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close() // nolint: errcheck

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}

// SetEthermintNodeConfig takes a node configuration and applies ethermint specific configuration
// #unstable
func SetEthermintNodeConfig(cfg *node.Config) {
	cfg.P2P.MaxPeers = 0
	cfg.P2P.NoDiscovery = true
}

// SetEthermintEthConfig takes a ethereum configuration and applies ethermint specific configuration
// #unstable
func SetEthermintEthConfig(cfg *eth.Config) {
	cfg.MaxPeers = 0
	cfg.PowFake = true
}

// SetAdditionalConfig will setup non-geth specific configurations
// #unstable
func SetAdditionalConfig(cfg *additionalConfigs) {
	if cfg.TrieCacheGen != 0 {
		state.MaxTrieCacheGen = cfg.TrieCacheGen
	}

	if cfg.TargetGasLimit != 0 {
		params.TargetGasLimit = new(big.Int).SetUint64(cfg.TargetGasLimit)
	} else {
		params.TargetGasLimit = GenesisTargetGasLimit
	}
}

// MakeDataDir retrieves the currently requested data directory
// #unstable
func MakeDataDir(ctx *cli.Context) string {
	path := node.DefaultDataDir()

	emHome := os.Getenv(emHome)
	if emHome != "" {
		path = emHome
	}

	if ctx.GlobalIsSet(ethUtils.DataDirFlag.Name) {
		path = ctx.GlobalString(ethUtils.DataDirFlag.Name)
	}

	if path == "" {
		ethUtils.Fatalf("Cannot determine default data directory, please set manually (--datadir)")
	}

	return path
}

// MakePasswordList reads password lines from the file specified in the config file
func MakePasswordList(path string) []string {
	if path == "" {
		return nil
	}
	text, err := ioutil.ReadFile(path)
	if err != nil {
		ethUtils.Fatalf("Failed to read password file: %v", err)
	}
	lines := strings.Split(string(text), "\n")
	// Sanitise DOS line endings.
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}
	return lines
}

// DumpConfig dumps toml file from the configurations to stdout
// #unstable
func DumpConfig(cfg *GethConfig) error {
	comment := ""

	if cfg.Eth.Genesis != nil {
		cfg.Eth.Genesis = nil
		comment += "# Note: this config doesn't contain the genesis block.\n\n"
	}

	out, err := tomlSettings.Marshal(&cfg)
	if err != nil {
		return err
	}
	io.WriteString(os.Stdout, comment) // nolint: errcheck
	os.Stdout.Write(out)               // nolint: errcheck
	return nil
}
