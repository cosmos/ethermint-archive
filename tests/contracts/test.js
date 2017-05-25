const fs = require("fs");
const solc = require('solc')
const Web3 = require('web3');
var web3 = new Web3();

web3.setProvider(new web3.providers.HttpProvider('http://localhost:8545'));



let mySenderAddress = web3.eth.accounts[0]
console.log(mySenderAddress)
web3.personal.unlockAccount(mySenderAddress, '1234', (err) => {console.log(err)})

let source = fs.readFileSync('SimpleStorage.sol', 'utf8');
let compiledContract = solc.compile(source, 1);
let abi = compiledContract.contracts[':SimpleStorage'].interface;
let bytecode = '0x' + compiledContract.contracts[':SimpleStorage'].bytecode;
let gasEstimate = web3.eth.estimateGas({data: bytecode}) *100 ;
let MyContract = web3.eth.contract(JSON.parse(abi));

var myContractReturned = MyContract.new( {
   from:mySenderAddress,
   data:bytecode,
   gas:gasEstimate}, function(err, myContract){
	  console.log("...")
    if (err) throw err;
    // NOTE: The callback will fire twice!
    // Once the contract has the transactionHash property set and once its deployed on an address.

    if(!myContract.address) {
        console.log("hash", myContract.transactionHash)
    } else {
        console.log('addr', myContract.address)
    }
  });
