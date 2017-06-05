# Ethereum Peg Zones
The goal of the Ethereum peg zone is to provide a way for people to exchange ETH into CETH (COSMOS ETH) and vice versa.
Ethermint and Ethereum should be light clients to each other in order to validate the state changes. The implementation should happen within smart contracts written in solidity. Implementing the peg zone within the Ethermint codebase is dangerous because it will make it less flexible to future changes and will lead to greater divergence of the codebase from its upstream base. 

## Design
There will be two separate, but very similar sets of smart contracts, on the Ethermint and the Ethereum chain. These two smart contracts will track and lock the funds for the other chain. 
Let us say that you want to send 1 ETH to Ethermint in order to receive 1 CETH. In order to achieve this you would send 1 ETH and the destination address on Ethermint to the smart contract on Ethereum. A relay node which is constantly watching the Ethereum chain would see the Event triggered by the smart contract and would call the smart contract on Ethermint with the proof of the state change, the amount and the destination address. The second set of smart contracts would verify the proof and given its validity send 1 CETH to the destination address. 

### Security
#### Ethereum
The smart contract on Ethereum would track the validator set by virtue of receiving updates from a relay node and with that, it will be able to verify the proofs submitted to it about the state changes on the Ethermint chain.

#### Ethermint
???

#### Relay Nodes
Relay nodes would need to be constantly online to relay the newest changes from each chain as well as any transaction requests. In order to get people to run these relay nodes they will need to have a financial incentive.
A relay node could apply a percentage fee to each transaction that the smart contract would send to a separate address, which the relay node can include as extra data in the function call.

### Approaches to generating CETH
* Mine CETH when ETH is locked up on the Ethereum chain and burn CETH when it is sent from Ethermint to Ethereum
* Instantiate the Ethermint smart contract with enough CETH and lock it up and release it
**It is paramount that every CETH is backed by an ETH because otherwise, speculators will immediately attack the peg.**

Whenever one ETH is sent to Ethermint the smart contract could mine one CETH and when someone sends one CETH to Ethereum one CETH would be burned.

The genesis state of Ethermint could specify that the smart contract holds the total supply of ETH and the parameters could be set so that Ethermint follows exactly the same inflation as Ethereum. This would mean that for every transaction across the peg zones CETH and ETH are either locked up or released.


## Questions
* How can state updates on a Ethereum (PoW) be verified?
  * The problem posed is: What proof can a relay node submit to the Ethermint smart contract to trigger the release of funds?
* What will happen if speculators attack the peg zone, because they believe that 1 CETH is not worth exactly 1 ETH?
* Can there be a native currency on Ethermint that is not pegged against Ethereum?
* How will CETH be created?
* How can a smart contract mine/create the native currency of the chain that is also used to pay for gas?
* How does the smart contract choose a relay node?
* What is the market mechanism behind the pricing of relaying a transaction?
  * Maybe there is none and the fee is set as a percentage by the COSMOS validators

## Invariants
* the amount of CETH in circulation is equal to the amount of ETH held by the smart contract on the Ethereum chain
