const BigNumber = require('bignumber.js')
const Tx = require('ethereumjs-tx')
const async = require('async')
const Web3pool = require('./web3pool')

exports.getWeb3 = (web3o) => {
  if (web3o instanceof Web3pool) {
    return web3o.web3
  }

  return web3o
}

exports.generateTransaction = (txObject) => {
  const txParams = {
    nonce: '0x' + txObject.nonce.toString(16),
    gasPrice: '0x' + txObject.gasPrice.toString(16),
    gas: '0x' + new BigNumber(21024).toString(16),
    from: txObject.from,
    to: txObject.to,
    value: txObject.value ? '0x' + txObject.value.toString(16) : '0x00',
    data: '0x'
  }

  let tx = new Tx(txParams)
  tx.sign(txObject.privKey)

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

exports.waitMultipleProcessedInterval = (web3p, intervalMs, blockTimeout, cb) => {
  let waitAll = web3p.web3s.map((web3) => {
    return exports.waitProcessedInterval.bind(null, web3, intervalMs, blockTimeout)
  })

  async.parallel(waitAll, (err, ms) => {
    if (err) {
      return cb(err)
    }

    cb(null, new Date())
  })
}

exports.sendTransactions = (web3, transactions, cb) => {
  let start = new Date()
  async.series(transactions.map((tx) => {
    let w3 = exports.getWeb3(web3)
    return w3.eth.sendRawTransaction.bind(null, tx)
  }), (err) => {
    if (err) {
      return cb(err)
    }

    cb(null, new Date() - start)
  })
}

exports.calculateTransactionsPrice = (gasPrice, txcount) => {
  let gas = 21024 // Simple transaction gas requirement

  return new BigNumber(gasPrice).times(gas).times(txcount)
}
