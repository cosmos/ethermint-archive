# Roadmap for Ethermint

The goal of this document is to give the community the chance to understand the scope
and development progress of the Ethermint project. 

## Goals of Ethermint
*This paragraph describes the scope of Ethermint, what it is and what it is not.*

Ethermint is designed to be an implementation of the Ethereum virtual machine that runs
on top of Tendermint. It aims to provide the exact same implementation as traditional
ethereum clients such as go-ethereum or parity. The only major difference is that
it is not secured by ethereum's PoW but rather by Tendermint BFT. This of course also
entails that an efficient light-client for Ethermint is possible.

Furthermore, Ethermint will provide inter-operability with the COSMOS ecosystem. As
such it will allow to send transactions from Ethermint to Basecoin and vice versa,
which will allow users to seemlessly interoperate with any other blockchain over the
IBC (Inter Blockchain Communication) protocol. 

Lastly, Ethermint will also provide an implementation of a peg zone for ETH and ETC.
This will be achieved by modelling ETH/ETC as an ERC20 token on the Ethermint chain,
which will be fully exchangeable to and from the respective main chains. Ethermint
will have its own native token to pay for gas and validators will earn fees and
rewards in this native token. 

## Milestones
### Testnet Venus - 2017/07/01
- launch of the first public testnet of Ethermint 

### Peg zone for ETH on Venus - 2017/08/01
- launch of the first peg zone from the ETH testnet to Venus

### Live Network Jupiter - 2017/09/01
- launch of the live network of Ethermint

### Peg zone for ETH on Jupiter - 2017/10/01
- launch of the live peg zone from ETH to Jupiter

These milestones aim to provide enough time for development while also allowing the
community to grow around Ethermint. Furthermore, they aim to be synced up with the
development of Tendermint core and the Cosmos hub, as ethermint depends on them.

