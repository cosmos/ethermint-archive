const config = require('config')
const Web3 = require('web3')
const Wallet = require('ethereumjs-wallet')
const utils = require('./utils')

const web3 = new Web3(new Web3.providers.HttpProvider(config.get('provider')))
const wallet = Wallet.fromV3(config.get('wallet'), config.get('password'))

const initialNonce = web3.eth.getTransactionCount(wallet.getAddressString())
const totalTxs = config.get('n')

// extend web3
utils.extendWeb3(web3)

const transactions = []

console.log('generating transactions')
for (let i = 0; i < totalTxs; i++) {
  let nonce = i + initialNonce
  let tx = utils.generateTransaction(wallet, config.get('address'), nonce, web3.eth.gasPrice)

  console.log('generated tx: ' + i, 'nonce: ' + nonce)
  transactions.push(tx)
}

// Send transactions
const start = new Date()
utils.sendTransactions(web3, transactions, (err, ms) => {
  if (err) {
    console.error(err)
    return
  }

  utils.waitProcessed(web3, (err, endDate) => {
    if (err) {
      console.error('Couldn\'t process transactions in blocks')
      console.error(err)
      return
    }

    let sent = transactions.length
    let timePassed = (endDate - start) / 1000
    let perSecond = sent / timePassed

    console.log('Processed %s transaction in %s seconds from one account, %s tx/s', sent, timePassed, perSecond)
  })
})
