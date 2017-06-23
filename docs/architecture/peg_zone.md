# Ethereum Peg Zones
**It is paramount that every CETH is backed by an ETH because otherwise, speculators will immediately attack the peg.**

The goal of the Ethereum peg zone is to provide a way for people to exchange ETH into CETH (COSMOS ETH) and vice versa.
The implementation should happen within smart contracts written in solidity. Implementing the peg zone within the Ethermint codebase is dangerous because it will make it less flexible to future changes and will lead to greater divergence of the codebase from its upstream base. 

## Process

### Sending Ether to Cether (Ethereum -> Ethermint)
To transfer money from ethereum to ethermint you have to send ether and a destination address to the exchange smart contract on ethereum. This smart contract will raise an event when the transfer function is invoked. Every validator
that supports ethermint will also run a full ethereum node that is connected to the mainnet. This node will listen for events on the smart contract and send data of the event to ethermint. Ethermint will then receive the amount and
destination address and will mint new CETH (through a reward strategy) for the specified destination address.

### Sending Cether to Ether (Ethermint -> Ethereum)
To transfer money from ethermint to ethereum you have to send cether and a destination address to the exchange smart contract on ethermint. This smart contract will raise an event when the transfer function is invoked. It will burn 
any CETH that it has received. Ethermint listens for these events. Once an event is received, ethermint will create, sign and broadcast a transaction that invokes the release function on the ethereum smart contract. In this function
call it will include the proof and the signatures. The smart contract on ethereum will verify the proof and signatures and conditionally release money to the destination address.

## Design
Every validator runs tendermint, ethermint and an ethereum node. Ethermint has a `--peg` flag, which will cause it to also act as a pegzone. 
The ethereum node is configured to send a notification to ethermint every time an event
is raised by the exchange contract on ethereum. This notification causes ethermint to mint new CETH out of thin air through its reward strategy.
The ethermint smart contract raises an event when it receives a deposit. After receiving a deposit it burns all deposited CETH. Ethermint listens for this event in a go-routine and after receiving the notification it will create, 
sign and broadcast a transaction which invokes the release function on the ethereum smart contract. The smart contract on ethereum then verifies the signatures and the proof and only then releases the ETH to the destination address.

## Components
### Ethereum Full Node
- sends notification to ethermint when an event is raised by a specified smart contract on ethereum

### Ethereum Smart Contract
*Tendermint light client implemented in Solidity*
- locks up ETH
- triggers event upon receiving ETH
- is a light client for tendermint and verifies the calls to its release function

### Ethermint Smart Contract
- burns CETH
- triggers event upon receiving CETH

### Ethermint
- responds to notifications from ethereum full node by minting fresh CETH
- sends transaction that invoke release function on ethereum

## Economic Incentive
Ethermint can take a percentage of the CETH is a transaction fee.
Ethermint can take a percentage of the ETH when releasing them.

## Questions
- How does a validator look that runs ethermint and basecoin at the same time?
- Are the economic incentives good enough and can there be a market mechanism around establishing the correct percentage cut?
- Is CETH used to pay for execution of the ethermint EVM?
- What will happen if speculators attack the peg zone, because they believe that 1 CETH is not worth exactly 1 ETH?
- Can there be a native currency on Ethermint that is not pegged against Ethereum?
- What is the market mechanism behind the pricing of relaying a transaction?
  - Maybe there is none and the fee is set as a percentage by the COSMOS validators
- How will IBC work in relation to ethermint?

## Invariants
- the amount of CETH in circulation is equal to the amount of ETH held by the smart contract on the Ethereum chain
