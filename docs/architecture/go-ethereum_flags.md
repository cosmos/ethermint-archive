# Flags exposed by ethermint and why

All unexposed flags will be set internally in ethermint to an appropriate default.

## General settings
- [x] DataDirFlag
  - used to control the directory where go-ethereum stores its files
  
- [x] KeyStoreDirFlag
  - depends on DataDirFlag
  
- [x] NoUSBFlag
  - disables monitoring for and managing of USB hardware wallets
  
~~- [ ] EthashCacheDirFlag~~
  - mining and ethash should be turned off
  
~~- [ ] EthashCachesInMemoryFlag~~

~~- [ ] EthashCachesOnDiskFlag~~

~~- [ ] EthashDatasetDirFlag~~

~~- [ ] EthashDatasetsInMemoryFlag~~

~~- [ ] EthashDatasetsOnDiskFlag~~

~~- [ ] NetworkIdFlag~~
  - used to set the network (frontier, morden, ropsten, rinkeby)
  - ethermint should be agnostic as to what network it runs against, since
    it will be determined by the configuration of the tendermint core node
    
~~- [ ] TestnetFlag~~

~~- [ ] RinkebyFlag~~

~~- [ ] DevModeFlag~~

~~- [ ] IdentityFlag~~
  - allows to set a custom node name
  - ethermint should never be directly exposed or network with other nodes
    and hence the node name should not be settable
    
- [ ] DocRootFlag
  - What is this?
  
~~- [ ] FastSyncFlag~~
  - ethermint should have no notion of fast sync as it is handled by 
    tendermint core
    
~~- [ ] LightModeFlag~~
  - ethermint does not directly serve light clients
  
~~- [ ] SyncModeFlag~~
  - the syncing part is handled by tendermint core
  
~~- [ ] LightServFlag~~
  - ethermint does not serve light clients
  
~~- [ ] LightPeersFlag~~

- [ ] LightKDFFlag
  - What is this?
  
- [x] CacheFlag
  - Megabytes of memory allocated to internal cache
  - allows for optimizations of ethermint
  
- [x] TrieCacheGenFlag
  - number of trie node generations to keep in memory
  
~~- [ ] MiningEnabledFlag~~
  - ethermint does not allow go-ethereum to directly mine
  
~~- [ ] MinerThreadsFlag~~

- [x] TargetGasLimitFlag
  - this should be enabled to allow ethermint users to run economically
  - not clear yet how it will work with tendermint core
  
- [ ] EtherbaseFlag
  - can be used by reward strategies
  
- [ ] GasPriceFlag
  - this should be enabled to allow ethermint users to run economically
  - not clear yet how it will work with tendermint core
  
~~- [ ] ExtraDataFlag~~
  - allows miners to attach data to blocks
  - not used since tendermint core handles creation of blocks

## Account settings
- [x] UnlockedAccountFlag
  - allows the user to unlock accounts by default
  
- [x] PasswordFileFlag
  - password file to use for non-interactive password input
  
- [x] VMEnableDebugFlag
  - record information useful for VM and contract debugging

## Logging and debug settings
- [ ] EthStatsURLFlag
  - reporting URL of ethstats service
  - Should this be exposed or does tendermint core handle it?
  
- [ ] MetricsEnabledFlag
  - collect metrics
  - Should this be exposed or does tendermint core handle it?
  
~~- [ ] FakePoWFlag~~
  - disable proof-of-work verification
  - this is set internally
  
- [x] NoCompactionFlag
  - disables DB compaction after import

## RPC settings
- [x] RPCEnabledFlag

- [x] RPCListenAddrFlag

- [x] RPCPortFlag

- [x] RPCCORSDomainFlag

- [x] RPCApiFlag

- [x] IPCDisabledFlag

- [x] IPCPathFlag

- [x] WSEnabledFlag

- [x] WSListenAddrFlag

- [x] WSPortFlag

- [x] WSApiFlag

- [x] WSAllowedOriginsFlag

~~- [ ] ExecFlag~~
  - JS capabilities aren't exposed by ethermint
  
~~- [ ] PreloadJSFlag~~
  - preload JS files into console
  - console isn't exposed by ethermint

## Network settings
~~- [ ] MaxPeersFlag~~
  - max number of peers
  - ethermint does not connect to peers and hence this is not exposed
  
~~- [ ] MaxPendingPeersFlag~~

~~- [ ] ListenPortFlag~~

~~- [ ] BootnodesFlag~~

~~- [ ] BootnodesV4Flag~~

~~- [ ] BootnodesV5Flag~~

~~- [ ] NodeKeyFileFlag~~

~~- [ ] NodeKeyHexFlag~~

~~- [ ] NATFlag~~

~~- [ ] NoDiscoverFlag~~

~~- [ ] DiscoveryV5Flag~~

~~- [ ] NetrestrictFlag~~

- [ ] WhisperEnabledFlag
  - What is this?
  
~~- [ ] JSpathFlag~~
  - JS is not exposed by ethermint
  
- [x] GpoBlocksFlag
  - number of recent block to check for gas prices
  
- [x] GpoPercentileFlag
  - suggested gas price is the given percentile of a set of recent transactions

## Debug Flags
This package is not exposed by go-ethereum and hence we have to replicate some of their flags.

## Console Flags
It allows go-ethereum to expose an interpreted JavaScript console. In our use case
we expect users to run ```ethermint``` and then to the node using ```geth``` from 
another process.

## Custom Flags
- [x] TendermintAddrFlag
  - address used by ethermint to connect to tendermint core for transaction broadcast
  
- [x] ABCIAddrFlag
  - address used by ethermint to listen for incoming messages from tendermint core
  
- [x] ABCIProtocolFlag
  - switches between grpc and socket as a communication protocol from tendermint core
    to ethermint
    
- [x] VerbosityFlag
  - sets the level for the logger
  
- [ ] VModuleFlag
  - per module verbosity
  
- [ ] BackTraceAtFlag
  - request a stack trace at a specific logging statement
  
- [ ] DebugFlag
  - prepends log messages with call-site location
  
- [ ] PProfFlag
  - enables the pprof HTTP server
  
- [ ] PProfPortFlag
  - pprof HTTP server listening port
  
- [ ] PProfAddrFlag
  - pprof HTTP server listening interface
  
- [ ] MemProfileRateFlag
  - turns on memory profiling with the given rate
  
- [ ] BlockProfileRateFlag
  - turns on block profiling with the given rate
  
- [ ] CpuProfileFlag
  - write CPU profile to given file
  
- [ ] TraceFlag
  - write execution trace to the given file
