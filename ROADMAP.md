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


## Venus
This testnet is initially run by the Interchain Foundation and in the beginning
will have seven validators. Over time we will open this up to interested third
parties to run their own validators for the testnet. The goal is to have 100 validators
on the testnet.

## Jupiter
The live network will be secured by atom holders. It will have 100 validators initially
and those validators will initially the same as the Cosmos validators.

## Peg Zone
The peg zone is still subject to ongoing research and as such this document only
presents a very high-level and maybe inaccurate view.

The validators of Ethermint will also be running and securing the peg zone. The economic
motivation behind this is that the value of Ethermint tokens will increase of an active
peg zone is established and hence the value earned by validators through fees and rewards
will increase as well.

A validator node will run a full go-ethereum/parity node, where it listens for events fired
by the peg zone contract on the ETH network. Upon seeing a deposit (amount and destination
address) on the ethereum network, the node will post a transaction on Ethermint to the
peg zone contract that states how many pegged ERC20 tokens should be created and where they
should be send. If at least >2/3 of the addresses registered on the peg zone contract 
submit this transaction, the funds will be released. This works in a similar fashion when
converting Ethermint ERC20 tokens back into ETH. 

The smart contracts on Ethermint and ethereum track the validator set for Ethermint. Initially
it will be supplied and has to be trusted in the same fashion as the genesis block. Later
updates to it can happen through the same multisig procedure that is used for releasing funds.


## Genesis State for Jupiter
The genesis state for Jupiter will involve the allocation of native tokens to certain entities,
such as AiB, the Interchain Foundation, Tendermint and core developers.

## Token Creation for Ethermint
Validators will earn not only the transaction fee but also will receive a minting reward, as
part of creating the block. This will cause a steady and expected long-term inflation of 
around 2%.
