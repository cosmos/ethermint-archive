const BigNumber = require('bignumber.js')
const Tx = require('ethereumjs-tx')
const async = require('async')

exports.extendWeb3 = (web3) => {
  web3._extend({
    property: 'txpool',
    methods: [],
    properties: [
      new web3._extend.Property({
        name: 'status',
        getter: 'txpool_status',
        outputFormatter: function (status) {
          status.pending = web3._extend.utils.toDecimal(status.pending)
          status.queued = web3._extend.utils.toDecimal(status.queued)
          return status
        }
      })
    ]
  })
}

exports.generateTransaction = (address, privKey, destination, nonce, gasPrice) => {
  const txParams = {
    nonce: '0x' + nonce.toString(16),
    gasPrice: '0x' + gasPrice.toString(16),
    gas: '0x' + new BigNumber(21024).toString(16),
    from: address,
    to: destination,
    value: '0x00',
    data: '0x'
  }

  let tx = new Tx(txParams)
  tx.sign(privKey)

  return '0x' + tx.serialize().toString('hex')
}

exports.waitProcessedInterval = function (web3, intervalMs, blockTimeout, cb) {
  let startingBlock = web3.eth.blockNumber

  console.log('Starting block:', startingBlock)
  let interval = setInterval(() => {
    let blocksGone = web3.eth.blockNumber - startingBlock
    if (blocksGone > blockTimeout) {
      clearInterval(interval)
      cb(new Error(`Pending full after ${blockTimeout} blocks`))
      return
    }

    let status = web3.txpool.status
    console.log(`Blocks Passed ${blocksGone}, ` +
      `Pending Txs: ${status.pending}, Queued Txs: ${status.queued}`)

    if (status.pending === 0 && status.queued === 0) {
      clearInterval(interval)
      cb(null, new Date())
    }
  }, intervalMs || 100)
}

exports.waitProcessedFilter = function (web3, filter, cb) {
  if (arguments.length === 2) {
    cb = filter
    filter = web3.eth.filter('latest')
  }

  let blocks = 100

  filter.watch(function (err, res) {
    if (err) {
      cb(err)
      return
    }

    blocks--

    if (web3.txpool.status.pending === 0) {
      cb(null, new Date())
      filter.stopWatching()
    }

    if (blocks < 0) {
      cb(new Error('processing failed'))
    }
  })
}

exports.sendTransactions = (web3, transactions, cb) => {
  let start = new Date()
  async.series(transactions.map((tx) => {
    return web3.eth.sendRawTransaction.bind(null, tx)
  }), (err) => {
    if (err) {
      return cb(err)
    }

    cb(null, new Date() - start)
  })
}
