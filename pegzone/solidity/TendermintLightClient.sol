pragma solidity ^0.4.4;

library TendermintLightClientLib {
  struct Validator {
    address addr;
    bytes32 pubKey;
    uint votingPower;
  }
}

contract TendermintLightClient {
  using TendermintLightClientLib for TendermintLightClientLib.Validator;

  // blockHeight => ValidatorSet
  mapping (uint => TendermintLightClientLib.Validator[]) trustedValidators;
  string chainID;

  function TendermintLightClient(Validator[] validatorSet, string chainID) {


  }
}
