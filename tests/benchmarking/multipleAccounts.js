const config = require('config')
const Web3 = require('web3')
const Wallet = require('ethereumjs-wallet')
const utils = require('./utils')
const async = require('async')

const web3 = new Web3(new Web3.providers.HttpProvider(config.get('provider')))
const wallet = Wallet.fromV3(config.get('wallet'), config.get('password'))
const totalAccounts = config.get('accounts')

const walletAddress = wallet.getAddressString()

const gasPrice = web3.eth.gasPrice
const totalTxs = config.get('n')
const balance = web3.eth.getBalance(walletAddress)
const costPerAccount = utils.calculateTransactionsPrice(gasPrice, totalTxs)
const distributingCost = utils.calculateTransactionsPrice(gasPrice, totalAccounts)
const totalCost = distributingCost.plus(costPerAccount.times(totalAccounts))

// extend web3
utils.extendWeb3(web3)

console.log(`Send ${totalTxs} transactions from each account, accounts: ${totalAccounts}`)
console.log(`Cost of each account txs: ${web3.fromWei(costPerAccount, 'ether')}`)
console.log(`Distributing cost: ${web3.fromWei(distributingCost, 'ether')}`)
console.log(`Cost of all: ${web3.fromWei(totalCost, 'ether')}`)

if (totalCost.comparedTo(balance) > 0) {
  throw new Error(`Unsufficient funds: ${web3.fromWei(balance, 'ether')}`)
}

let privKey = wallet.getPrivateKey()

console.log('Generating wallets and sending funds')
let wallets = []
let nonce = web3.eth.getTransactionCount(walletAddress)

for (let i = 0; i < totalAccounts; i++) {
  wallets.push(Wallet.generate())

  let tx = utils.generateTransaction({
    nonce: nonce + i,
    gasPrice: gasPrice,
    from: walletAddress,
    to: wallets[i].getAddressString(),
    value: costPerAccount,
    privKey: privKey
  })

  web3.eth.sendRawTransaction(tx)
}
console.log('Distributed Funds.')

// Send Transactions
let dest = config.get('address')
let initialNonces = {}
wallets.forEach((w) => {
  let addr = w.getAddressString()
  initialNonces[addr] = web3.eth.getTransactionCount(addr)
})

let runTransactionsFromAccount = (wallet, cb) => {
  let address = wallet.getAddressString()
  let privKey = wallet.getPrivateKey()
  let initialNonce = initialNonces[address]
  let transactions = []
  let totalTxs = config.get('n')

  console.log(`Generating ${totalTxs} transactions for ${address}`)
  for (let i = 0; i < totalTxs; i++) {
    let tx = utils.generateTransaction({
      nonce: initialNonce + i,
      gasPrice: gasPrice,
      from: address,
      to: dest,
      privKey: privKey
    })

    transactions.push(tx)
  }
  console.log(`Generated, starting sending transactions from ${address}`)

  utils.sendTransactions(web3, transactions, cb)
}

const start = new Date()
const blockTimeout = config.get('blockTimeout')
async.parallel(
  wallets.map((w) => runTransactionsFromAccount.bind(null, w)),
  (err, res) => {
    if (err) {
      console.log('Error on transactions:', err)
      return
    }

    utils.waitProcessedInterval(web3, 100, blockTimeout, (err, endDate) => {
      if (err) {
        console.error('Couldn\'t process transactions in blocks')
        console.error(err)
        return
      }

      let sent = totalTxs * totalAccounts
      let processed = Object.keys(initialNonces).reduce((sum, addr) => {
        return sum + (web3.eth.getTransactionCount(addr) - initialNonces[addr])
      }, 0)
      let timePassed = (endDate - start) / 1000
      let perSecond = processed / timePassed

      console.log(`Processed ${processed} of ${sent} transactions ` +
      `from one account in ${timePassed}s, ${perSecond} tx/s`)
    })
  }
)
