'use strict';

let chai = require('chai');
let assert = chai.assert;
let fs = require('fs');
let solc = require('solc');
let Web3 = require('web3');
let web3 = new Web3(new Web3.providers.HttpProvider("http://localhost:8545"));

const account = web3.eth.accounts[0];

console.log('Block number: ' + web3.eth.blockNumber)
console.log('Account: ' + account)

//unlock account
web3.personal.unlockAccount(account, "1234");
let contractSource = fs.readFileSync(__dirname + '/test.sol', 'utf-8');

console.log("Solc version is " + solc.version());

describe('gasLimit', function () {
    it('should return gas too low error', function (done) {

        const contractCompiled = solc.compile(contractSource);

        const bytecode = contractCompiled.contracts[':Test'].bytecode;
        const abi = JSON.parse(contractCompiled.contracts[':Test'].interface);

        // const estimateGas = web3.eth.estimateGas({data: '0x' + bytecode,});
        // console.log('Gas needed: ' + estimateGas);

        const testContract = web3.eth.contract(abi);
        testContract.new({
            from: account,
            data: '0x' + bytecode,
            gas: '100' //set low gas
        }, function (error, contract) {

            assert.equal(error, "Error: intrinsic gas too low");
            done();
        });
    });

    it('should deploy contract', function (done) {
        this.timeout(60 * 1000);

        const contractCompiled = solc.compile(contractSource);

        const bytecode = contractCompiled.contracts[':Test'].bytecode;
        const abi = JSON.parse(contractCompiled.contracts[':Test'].interface);

        const estimateGas = web3.eth.estimateGas({data: '0x' + bytecode,});
        // console.log('Gas needed: ' + estimateGas);

        let callbackCount = 0;

        const testContract = web3.eth.contract(abi);
        testContract.new({
            from: account,
            data: '0x' + bytecode,
            gas: '5379400' //set enough gas
        }, function (error, contract) {
            callbackCount++;

            assert.equal(error, null);

            if (error) {
                done();
            }

            assert.isNotNull(contract.transactionHash);

            if (callbackCount === 2) {
                assert.isNotNull(contract.address);
                done();
            }
        });
    });
});

describe('huge smart contract', function () {
    it('should deploy contract', function (done) {
        //set mocha timeout
        this.timeout(120 * 1000);

        //generate smart contract with many functions
        let code = codeGenerator(150);

        const contractCompiled = solc.compile(code);

        const bytecode = contractCompiled.contracts[':Test'].bytecode;
        const abi = JSON.parse(contractCompiled.contracts[':Test'].interface);

        const estimateGas = web3.eth.estimateGas({data: '0x' + bytecode,});
        // console.log('Gas needed: ' + estimateGas);

        let callbackCount = 0;

        const testContract = web3.eth.contract(abi);
        testContract.new({
            from: account,
            data: '0x' + bytecode,
            gas: estimateGas//set enough gas
        }, function (error, contract) {
            callbackCount++;

            assert.equal(error, null);

            if (error) {
                done();
            }

            assert.isNotNull(contract.transactionHash);

            if (callbackCount === 2) {
                assert.isNotNull(contract.address);
                done();
            }
        });
    });
});

function codeGenerator(functionsCount) {
    let code = `
        pragma solidity ^0.4.0;
        contract Test
        {
            event TestEvent(uint a);
            function Test() {}
        `;
    for (let i = 0; i < functionsCount; i++) {
        code += "\nfunction f" + i + `(uint i)
        {
            TestEvent(i);
        }`;
    }
    code += "\n}";

    return code;
}
