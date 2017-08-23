pragma solidity ^0.4.4;

library TendermintLightClientLib {
  struct Validator {
    address addr;
    bytes32 pubKey;
    uint votingPower;
  }
}
