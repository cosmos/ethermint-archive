const config = require('config');
const Web3 = require('web3');
const Wallet = require('ethereumjs-wallet');
const Tx = require('ethereumjs-tx');
const BigNumber = require('bignumber.js');

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

var version = _web3.version.node;
console.log(version);

console.log(_web3.eth.getBlock(100));

let nonces = {};
const _start = new Date();
const _wallet = Wallet.fromV3(config.get('wallet'), config.get('password'));
let _gasPrice = _web3.eth.gasPrice;

function getId(key) {
    return _web3.sha3(key + Date.now());
}

function getWallet() {
    return Wallet.fromV3(config.get('wallet'), config.get('password'));
}

function signAndSendRequest(wallet, gasPrice, destination) {
    const tx = new Tx();

    if (nonces[wallet.getAddressString()] === undefined) {
        nonces[wallet.getAddressString()] = _web3.eth.getTransactionCount(wallet.getAddressString());
    }

    tx.nonce = new BigNumber(nonces[wallet.getAddressString()]).toString(10);
    tx.gasPrice = gasPrice.toString(16);
    tx.to = destination;
    tx.gas = "200";
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

    const address = config.get('address');

    return signAndSendRequest(_wallet, _gasPrice, address);
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
            if (Web3.txpool.status.pending == 0) {
                console.log("processed", ((new Date() - _start) / 1000) + ' s');
                latestFilter.stopWatching();
            } else if (blocks > 10) {
                console.log("processed", "FAILED");
                latestFilter.stopWatching();
            }
        }
    });
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
