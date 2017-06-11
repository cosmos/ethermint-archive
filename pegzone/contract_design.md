# Tendermint Light Client in Solidity
At a very basic level the light client needs to implement the same logic implemented in
tmcli. Running a light client revolves around the basic idea of tracking changes to a
validator set given a trusted genesis state. 

## contract TendermintLightClient
- chainID
- []currentValidator
- latestHeight
- mapping(blockHeight => []Validators)
  - trusted validator set at blockHeight
  - 0 index maps to the original validator set
- initialise(chainID, originalValidators)
  - assign the chainID and the originalVals
- update(Commit, Header, []Validators)
  - reject all updates where the Block is lower than latestHeight
  - validate consistency of block against chainID
  - actual verification([]currentValidator, []newValidators, chainID, commit, height)
    - need to verify that precommits and check that more than 2/3 of old vals did precommit for the block with the new validator set
