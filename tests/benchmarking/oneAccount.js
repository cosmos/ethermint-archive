const config = require('config')
const Web3 = require('web3')
const Wallet = require('ethereumjs-wallet')
const utils = require('./utils')

const web3 = new Web3(new Web3.providers.HttpProvider(config.get('provider')))
const wallet = Wallet.fromV3(config.get('wallet'), config.get('password'))

const walletAddress = wallet.getAddressString()
const initialNonce = web3.eth.getTransactionCount(walletAddress)
const totalTxs = config.get('n')
const blockTimeout = config.get('blockTimeout')

// extend web3
utils.extendWeb3(web3)

const transactions = []

console.log('Current block number:', web3.eth.blockNumber)
console.log(`Will send ${totalTxs} transactions and wait for ${blockTimeout} blocks`)
console.log('generating transactions')

let privKey = wallet.getPrivateKey()
let dest = config.get('address')
let gasPrice = web3.eth.gasPrice

for (let i = 0; i < totalTxs; i++) {
  let nonce = i + initialNonce
  let tx = utils.generateTransaction(walletAddress, privKey, dest, nonce, gasPrice)

  console.log('generated tx: ' + i, 'nonce: ' + nonce)
  transactions.push(tx)
}

// Send transactions
const start = new Date()
utils.sendTransactions(web3, transactions, (err, ms) => {
  if (err) {
    console.error('Couldn\'t send Transactions:')
    console.error(err)
    return
  }

  utils.waitProcessedInterval(web3, 100, blockTimeout, (err, endDate) => {
    if (err) {
      console.error('Couldn\'t process transactions in blocks')
      console.error(err)
      return
    }

    let sent = transactions.length
    let processed = web3.eth.getTransactionCount(walletAddress) - initialNonce
    let timePassed = (endDate - start) / 1000
    let perSecond = processed / timePassed

    console.log(`Processed ${processed} of ${sent} transactions ` +
    `from one account in ${timePassed}s, ${perSecond} tx/s`)
  })
})
