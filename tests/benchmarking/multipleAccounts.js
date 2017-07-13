const config = require('config')
const Web3 = require('web3')
const Wallet = require('ethereumjs-wallet')
const utils = require('./utils')
const async = require('async')
const Web3pool = require('./web3pool')

const web3p = new Web3pool(config.get('providers'))
const web3 = web3p.web3
const wallet = Wallet.fromV3(config.get('wallet'), config.get('password'))
const totalAccounts = config.get('accounts')

const walletAddress = wallet.getAddressString()

const gasPrice = web3.eth.gasPrice
const totalTxs = config.get('n')
const balance = web3.eth.getBalance(walletAddress)
const costPerAccount = utils.calculateTransactionsPrice(gasPrice, totalTxs)
const distributingCost = utils.calculateTransactionsPrice(gasPrice, totalAccounts)
const totalCost = distributingCost.plus(costPerAccount.times(totalAccounts))

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

// wait for 200 blocks this case. Initial Fund distribution
utils.waitMultipleProcessedInterval(web3p, 100, 200, (err, endDate) => {
  if (err) {
    console.error(err)
    return
  }
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

    utils.sendTransactions(web3p, transactions, cb)
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

      utils.waitMultipleProcessedInterval(web3p, 100, blockTimeout, (err, endDate) => {
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
})
