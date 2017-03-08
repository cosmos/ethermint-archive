'use strict';

const config = require('config');
const ABI = require('./config/ABI.json');
const Tx = require('ethereumjs-tx');
const BN = require('bn.js');
const Wallet = require('ethereumjs-wallet');
const Web3 = require('web3');

const _web3 = new Web3(new Web3.providers.HttpProvider(config.get('provider')));
_web3._extend({
    property: 'txpool',
    methods: [],
    properties:
        [
            new _web3._extend.Property({
                name: 'status',
                getter: 'txpool_status',
                outputFormatter: function(status) {
                    status.pending = _web3._extend.utils.toDecimal(status.pending);
                    status.queued = _web3._extend.utils.toDecimal(status.queued);
                    return status;
                }
            })
        ]
});

const _contract = _web3.eth.contract(ABI).at(config.get('address'));

let nonces = {};
const _start = new Date();
const _wallet = getWallet();
let _gasPrice = _web3.eth.gasPrice.toString(16);

function getId(key) {
  return _web3.sha3(key + Date.now());
}

function getWallet() {
  return Wallet.fromV3(config.get('wallet'), config.get('password'));
}


function signAndSendRequest(wallet, gasPrice, request) {
  const tx = new Tx(request.params[0]);

  if (nonces[wallet.getAddressString()] === undefined) {
    nonces[wallet.getAddressString()] = _web3.eth.getTransactionCount(wallet.getAddressString());
  }

  tx.nonce = new BN(nonces[wallet.getAddressString()], 10);
  tx.gasPrice = new BN(gasPrice, 16);
  tx.sign(wallet.getPrivateKey());

  // console.log(tx.toJSON());

  return new Promise((resolve, reject) => {
    _web3.eth.sendRawTransaction('0x'+tx.serialize().toString('hex'), (err, res) => {
      if (err) {
        console.log(err);
        reject(err);
      } else {
        // console.log(res);
        nonces[wallet.getAddressString()]++;
        resolve(res);
      }
    });
  });
}

function request(i) {
  if (i % 100 === 0) {
    console.log(i, ((new Date() - _start) / 1000) + ' s');
  }

  const id = getId('product ' + i);
  // console.log(id);

  return signAndSendRequest(_wallet, _gasPrice,
    _contract.registerProduct.request(id, 'product',
      {from: _wallet.getAddressString(), gas: 150000}));
}

function waitProcessed() {
  let blocks = 0;

  const latestFilter = _web3.eth.filter('latest');
  latestFilter.watch((err, res) => {
    if (err) {
      console.log(err);
    } else {
      blocks++;
      console.log("\tblock", res);
      console.log("\ttxpool", _web3.txpool.status);
      if (_web3.txpool.status.pending == 0) {
        console.log("processed", ((new Date() - _start) / 1000) + ' s');
        latestFilter.stopWatching();
        verify();
      } else if (blocks > 10) {
        console.log("processed", "FAILED");
        latestFilter.stopWatching();
        verify();
      }
    }
  });
}

function verify() {
  const actual = _contract.getProducts.call('0x'+config.get('wallet').address);
  console.log('verify:', actual.length);

  if (actual.length !== config.get('n')) {
    console.log('verify: ERROR actual != expected');
    process.exit(1);
  }
  console.log('verify: Ok');
}

function run(index, n) {
  return request(index).
  then((res) => {
    if (index < n) {
      return run(index+1, n);
    } else {
      console.log("submitted", ((new Date() - _start) / 1000) + ' s');
      waitProcessed();
    }
  })
    .catch((e) => console.log(e));
}

run(1, config.get('n'));
