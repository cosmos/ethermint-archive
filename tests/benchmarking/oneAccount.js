const config = require('config')
const Wallet = require('ethereumjs-wallet')
const Web3pool = require('./web3pool')
const utils = require('./utils')

const web3p = new Web3pool(config.get('providers'))
const web3 = web3p.web3
const wallet = Wallet.fromV3(config.get('wallet'), config.get('password'))

const walletAddress = wallet.getAddressString()
const initialNonce = web3.eth.getTransactionCount(walletAddress)
const totalTxs = config.get('n')
const blockTimeout = config.get('blockTimeout')

const transactions = []

console.log('Current block number:', web3.eth.blockNumber)
console.log(`Will send ${totalTxs} transactions and wait for ${blockTimeout} blocks`)

let privKey = wallet.getPrivateKey()
let dest = config.get('address')
let gasPrice = web3.eth.gasPrice

let cost = utils.calculateTransactionsPrice(gasPrice, totalTxs)
let balance = web3.eth.getBalance(walletAddress)

if (cost.comparedTo(balance) > 0) {
  let error = `You don't have enough money to make ${totalTxs} transactions, ` +
    `it needs ${cost} wei, but you have ${balance}`
  throw new Error(error)
}

console.log(`Generating ${totalTxs} transactions`)
for (let i = 0; i < totalTxs; i++) {
  let nonce = i + initialNonce
  let tx = utils.generateTransaction({
    from: walletAddress,
    to: dest,
    privKey: privKey,
    nonce: nonce,
    gasPrice: gasPrice
  })

  transactions.push(tx)
}
console.log('Generated.')

// Send transactions
const start = new Date()
utils.sendTransactions(web3p, transactions, (err, ms) => {
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
